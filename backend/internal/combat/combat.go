// Package combat resolves one attack in the order of docs/core/combat.md: the situations
// and the called shot, one attack roll read against the defender's evasion and block bands,
// then damage through the called shot, crit, area, armor and shield DR, and resist. Dice and
// rolls are inputs. The bands and multipliers are formulas in combat/pipeline.json.
package combat

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/SETA1609/riftweave-core/backend/internal/actor"
	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/formula"
	"github.com/SETA1609/riftweave-core/backend/internal/modifier"
	"github.com/SETA1609/riftweave-core/backend/internal/resolution"
)

// Attacker is who rolls the attack.
type Attacker struct {
	Skill      int64            `json:"skill"`
	Roll       int64            `json:"roll"`
	Attributes actor.Attributes `json:"attributes"`
	Flags      []string         `json:"flags"`
	// WeaponProperties are the carried weapon's properties, which a maneuver may need.
	WeaponProperties []string `json:"weapon_properties"`
	// WeaponEffects are the weapon's on-hit effects (docs/core/combat.md#weapon-effects); nil
	// when the input names none.
	WeaponEffects []data.AppliedEffect `json:"weapon_effects"`
	// Perks are the attacker's held perks (refs); a rider perk among them puts its maneuver
	// on a plain attack (docs/core/actions.md#rider-perks).
	Perks []string `json:"perks"`
}

// Defender is who takes the attack.
type Defender struct {
	EvasionSkill         *int64           `json:"evasion_skill"`
	Attributes           actor.Attributes `json:"attributes"`
	ArmorDamageReduction int64            `json:"armor_dr"`
	// ShieldDamageReduction is a carried shield's DR; it adds to the armor DR.
	ShieldDamageReduction *int64 `json:"shield_dr"`
	ShieldsSkill          *int64 `json:"shields_skill"`
	// WeaponSkill is the held weapon's skill, which a parry band and Buckler Play read.
	WeaponSkill *int64 `json:"weapon_skill"`
	// Stance is the one the defender declared: dodge, block, or none.
	Stance    string           `json:"stance"`
	HitPoints int64            `json:"hp"`
	Resist    map[string]int64 `json:"resist"`
	Flags     []string         `json:"flags"`
	// Loadout is what the defender holds; with a weapon in each hand the defender's
	// modifiers on two_weapon_evasion_band widen the evasion band (Crossed Guard).
	Loadout *data.Loadout `json:"loadout"`
	// CreatureType is the defender's creature type, which a weapon's bane reads.
	CreatureType string `json:"creature_type"`
	// Armor is the heaviest body armor worn: none, light, medium, or heavy; "" reads none.
	// Unarmored Grace widens the band only with none.
	Armor string `json:"armor"`
	// Modifiers are the defender's perks and fighting stance: adds on evasion, evasion_band,
	// shields, block_band, shield_dr, unarmored_evasion_band, and spell_ward, and the bulwark, riposte, parry, and
	// block_with_weapon grants.
	Modifiers []modifier.Modifier `json:"modifiers"`
}

// Damage is the damage roll, its type, and the attribute that modifies it.
type Damage struct {
	Roll int64  `json:"roll"`
	Type string `json:"type"`
	// Ability names the attacker's attribute whose modifier adds to the roll, if any.
	Ability string `json:"ability"`
}

// Input is the `combat` command's input.
type Input struct {
	// Ammunition is the rounds carried and whether a loading weapon holds one; nil, rounds are not
	// counted (docs/core/combat.md#ammunition, GD_S17_02).
	Ammunition *Ammunition         `json:"ammunition"`
	Attacker   Attacker            `json:"attacker"`
	Defender   Defender            `json:"defender"`
	Damage     Damage              `json:"damage"`
	Delivery   string              `json:"delivery"`
	Modifiers  []modifier.Modifier `json:"modifiers"`
	Range      string              `json:"range"`
	Situations []string            `json:"situations"`
	CalledShot string              `json:"called_shot"`
	// Maneuver is the maneuvers/core.json key the attack is performed with, if any.
	Maneuver string `json:"maneuver"`
	// Summoning is whether the attacker has an active summon, for its while_summoning
	// modifiers; the beside_summon situation implies it (docs/core/modifier.md).
	Summoning bool `json:"summoning"`
	// Loadout is what the attacker holds: a maneuver's AP is then the loadout's attack AP.
	Loadout *data.Loadout `json:"loadout"`
}

// besideSummon is the situation of a summon attacking the same target, so one is out.
const besideSummon = "beside_summon"

// Result is the outcome of the attack.
type Result struct {
	// Ammunition is the rounds left after the strike, with the ammunition input only.
	Ammunition  *Ammunition `json:"ammunition,omitempty"`
	Outcome     string      `json:"outcome"`
	Target      int64       `json:"target"`
	Margin      *int64      `json:"margin"`
	Critical    bool        `json:"crit"`
	Evaded      bool        `json:"evaded"`
	Blocked     bool        `json:"blocked"`
	EvasionRate int64       `json:"evasion_rate"`
	BlockRate   int64       `json:"block_rate"`
	// ParryRate is a parrying stance's band, under the block band; Parried is a hit it stopped.
	ParryRate int64 `json:"parry_rate"`
	Parried   bool  `json:"parried"`
	// Warded is Spell Ward: a spell whose roll landed in the block band, with WardPercent of
	// its damage resisted.
	Warded      bool  `json:"warded"`
	WardPercent int64 `json:"ward_percent"`
	Damage      int64 `json:"damage"`
	HitPoints   int64 `json:"hp"`
	// Counter is a riposte: a dodge's evade in the window lets the defender attack back.
	Counter bool `json:"counter"`
	// StaggerAP is Bulwark: AP the attacker loses on their next turn.
	StaggerAP int64 `json:"stagger_ap"`
	// Riders are the called shot's riders on a hit, and empty on a miss or an evade. Without
	// a called shot the field is left out.
	Riders *[]data.Rider `json:"riders,omitempty"`
	// AP and Stamina are what the attack costs when it is a maneuver: the attack's cost plus
	// the maneuver's, at least 1, and that times stamina_per_ap. Left out without a maneuver.
	AP      *int64 `json:"ap,omitempty"`
	Stamina *int64 `json:"stamina,omitempty"`
	// Effects are the maneuver's (or the rider's) effects on a hit, and empty on a miss or an
	// evade. Without a maneuver or a rider the field is left out.
	Effects *[]json.RawMessage `json:"effects,omitempty"`
	// Events are the maneuver's (or the rider's) events on a hit, such as disarm; left out
	// without a maneuver or a rider.
	Events *[]string `json:"events,omitempty"`
	// Save is the maneuver attempt on a hit, for a maneuver with a save bonus or a rider;
	// the caller rolls it and a failed save lands the effects and events.
	Save *Save `json:"save,omitempty"`
	// Rider is the maneuver key a held rider perk put on this plain attack.
	Rider string `json:"rider,omitempty"`
	// WeaponEffects are the weapon effects that land on a hit, and empty on a miss, an evade,
	// or a block. Without weapon effects the field is left out.
	WeaponEffects *[]json.RawMessage `json:"weapon_effects,omitempty"`
}

const (
	armorPierceEffect = "core:effect/armor_pierce"
	baneEffect        = "core:effect/bane"
	banishEffect      = "core:effect/banish_summoned"
	lingeringEffect   = "core:effect/lingering_damage"
)

// landedWeaponEffects are the weapon effects a hit lists for the caller: lingering damage of
// the attack's type, a banish only on a summoned defender, and any other effect as written.
// Armor pierce and bane already acted on the damage.
func landedWeaponEffects(effects []data.AppliedEffect, damageType string, defender *Defender) ([]json.RawMessage, error) {
	var landed []json.RawMessage = []json.RawMessage{}
	for _, applied := range effects {
		switch applied.Effect {
		case armorPierceEffect, baneEffect:
			continue
		case banishEffect:
			if !slices.Contains(defender.Flags, "summoned") {
				continue
			}
		case lingeringEffect:
			var effect map[string]any
			if err := json.Unmarshal(applied.Encoded, &effect); err != nil {
				return nil, err
			}
			effect["parameter"] = damageType
			encoded, err := json.Marshal(effect)
			if err != nil {
				return nil, err
			}
			landed = append(landed, encoded)
			continue
		}
		landed = append(landed, applied.Encoded)
	}
	return landed, nil
}

// maneuverOf is the maneuver the attack is performed with, checked against the attack's
// range and the weapon's properties; nil without one.
func maneuverOf(ruleset *data.Ruleset, input *Input) (*data.Maneuver, error) {
	if input.Maneuver == "" {
		return nil, nil
	}
	maneuver, err := ruleset.Maneuvers.Require("maneuver", input.Maneuver)
	if err != nil {
		return nil, err
	}
	var attackRange string = input.Range
	if attackRange == "" {
		attackRange = "melee"
	}
	if len(maneuver.Ranges) > 0 && !slices.Contains(maneuver.Ranges, attackRange) {
		return nil, &data.MissingInputError{Subject: maneuver.Key, Needs: either(maneuver.Ranges, "attack")}
	}
	if len(maneuver.WeaponProperties) > 0 && !slices.ContainsFunc(maneuver.WeaponProperties, func(property string) bool {
		return slices.Contains(input.Attacker.WeaponProperties, property)
	}) {
		return nil, &data.MissingInputError{Subject: maneuver.Key, Needs: either(maneuver.WeaponProperties, "weapon")}
	}
	return maneuver, nil
}

// riderOf is the maneuver a held rider perk puts on a plain attack: nil when none is held,
// the attack is a maneuver, or no rider's maneuver fits the attack. Two riders that fit the
// same attack are an error.
func riderOf(ruleset *data.Ruleset, input *Input, maneuver *data.Maneuver) (*data.Maneuver, error) {
	var attackRange string = input.Range
	if attackRange == "" {
		attackRange = "melee"
	}
	var fitting []string
	var rider *data.Maneuver
	for _, ref := range input.Attacker.Perks {
		feature, err := ruleset.Features.Require("perk", ref[strings.LastIndex(ref, "/")+1:])
		if err != nil {
			return nil, err
		}
		if feature.Rider == "" {
			continue
		}
		row, err := ruleset.Maneuvers.Require("maneuver", feature.Rider[strings.LastIndex(feature.Rider, "/")+1:])
		if err != nil {
			return nil, err
		}
		if row.Fits(attackRange, input.Attacker.WeaponProperties) {
			fitting, rider = append(fitting, feature.Key), row
		}
	}
	if len(fitting) > 1 {
		return nil, fmt.Errorf("two rider perks fit this attack: %s", strings.Join(fitting, ", "))
	}
	if maneuver != nil {
		return nil, nil
	}
	return rider, nil
}

// Save is a maneuver attempt's save (docs/core/actions.md#the-maneuver-attempt).
type Save struct {
	Attribute string `json:"attribute"`
	DC        int64  `json:"dc"`
	Target    int64  `json:"target"`
}

// saveOf is the attempt's save: the DC is the weapon attribute times the pipeline's
// multiplier plus the bonus, and the defender's target is 100 minus it, at least 0.
func saveOf(input *Input, multiplier, bonus int64) *Save {
	var dc int64 = bonus
	if ability, known := actor.ByKey(input.Damage.Ability); known {
		dc += input.Attacker.Attributes.Get(ability) * multiplier
	}
	return &Save{Attribute: input.Damage.Ability, DC: dc, Target: max(100-dc, 0)}
}

// either names the choices as "a melee attack or a ranged attack".
func either(choices []string, noun string) string {
	var names []string = make([]string, 0, len(choices))
	for _, choice := range choices {
		names = append(names, "a "+choice+" "+noun)
	}
	return strings.Join(names, " or ")
}

// situationModifiers is step 0: the attack modifiers of every situation that applies (the
// ones the caller names, then the ones a flag applies, each once) and of the called shot.
// A situation that forbids the range is an error.
func situationModifiers(ruleset *data.Ruleset, input *Input, delivery string) ([]modifier.Modifier, *data.CalledShot, error) {
	var attackRange string = input.Range
	if attackRange == "" {
		attackRange = "melee"
	}
	if attackRange != "melee" && attackRange != "ranged" {
		return nil, nil, &data.NotDefinedError{Kind: "range", Key: attackRange}
	}
	var applied []*data.Situation
	for _, key := range input.Situations {
		situation, err := ruleset.Situations.Require("situation", key)
		if err != nil {
			return nil, nil, err
		}
		applied = append(applied, situation)
	}
	var rows []data.Situation = ruleset.Situations.Rows()
	for index := range rows {
		var situation *data.Situation = &rows[index]
		var attackerFlagged bool = situation.When.Attacker != "" && slices.Contains(input.Attacker.Flags, situation.When.Attacker)
		var defenderFlagged bool = situation.When.Defender != "" && slices.Contains(input.Defender.Flags, situation.When.Defender)
		if (attackerFlagged || defenderFlagged) && !slices.Contains(applied, situation) {
			applied = append(applied, situation)
		}
	}
	var modifiers []modifier.Modifier
	for _, situation := range applied {
		if slices.Contains(situation.Forbids, attackRange) {
			return nil, nil, fmt.Errorf("%s forbids a %s attack", situation.Key, attackRange)
		}
		if situation.Deliveries == nil || slices.Contains(situation.Deliveries, delivery) {
			modifiers = append(modifiers, situation.RangeModifiers(attackRange)...)
		}
	}
	// A row's adds: the attacker's modifiers on that target join the attack while the row applies.
	for _, situation := range applied {
		if situation.Adds == "" {
			continue
		}
		for _, attackModifier := range modifier.Target(input.Modifiers, situation.Adds) {
			attackModifier.Target = "attack_roll"
			modifiers = append(modifiers, attackModifier)
		}
	}
	if input.CalledShot == "" {
		return modifiers, nil, nil
	}
	calledShot, err := ruleset.CalledShots.Require("called shot", input.CalledShot)
	if err != nil {
		return nil, nil, err
	}
	if calledShot.Attack != 0 {
		modifiers = append(modifiers, modifier.Modifier{
			Target:    "attack_roll",
			Operation: modifier.Add,
			Value:     float64(calledShot.Attack),
			Source:    "called_shot:" + calledShot.Key,
		})
	}
	return modifiers, calledShot, nil
}

// Run is the `combat` command.
// Ammunition is a weapon's rounds: how many the attacker carries, and for a loading weapon
// whether it holds one.
type Ammunition struct {
	Rounds int64 `json:"rounds"`
	Loaded *bool `json:"loaded,omitempty"`
}

// The weapon properties that spend a round, and the one that must hold it first.
const (
	propertyAmmunition = "ammunition"
	propertyThrown     = "thrown"
	propertyLoading    = "loading"
)

// Run is one attack. With the ammunition input, a weapon that fires or throws spends one round,
// and a loading weapon must hold one (docs/core/combat.md#ammunition).
func Run(ruleset *data.Ruleset, input Input) (Result, error) {
	var properties []string = input.Attacker.WeaponProperties
	if input.Ammunition == nil || !(slices.Contains(properties, propertyAmmunition) || slices.Contains(properties, propertyThrown)) {
		return strike(ruleset, input)
	}
	if input.Ammunition.Rounds <= 0 {
		return Result{}, errors.New("the weapon has no rounds left")
	}
	var loading bool = slices.Contains(properties, propertyLoading)
	if loading && (input.Ammunition.Loaded == nil || !*input.Ammunition.Loaded) {
		return Result{}, errors.New("the weapon must reload first")
	}
	var rounds Ammunition = Ammunition{Rounds: input.Ammunition.Rounds - 1}
	if loading {
		var empty bool
		rounds.Loaded = &empty
	}
	input.Ammunition = nil
	result, err := strike(ruleset, input)
	if err != nil {
		return Result{}, err
	}
	result.Ammunition = &rounds
	return result, nil
}

// strike is one attack, in the order of docs/core/combat.md.
func strike(ruleset *data.Ruleset, input Input) (Result, error) {
	// The attacker's while_summoning modifiers, resolved once (docs/core/modifier.md).
	input.Modifiers = modifier.Summoning(input.Modifiers, input.Summoning || slices.Contains(input.Situations, besideSummon))
	damageType, err := ruleset.DamageTypes.Require("damage type", input.Damage.Type)
	if err != nil {
		return Result{}, err
	}
	var delivery string = input.Delivery
	if delivery == "" {
		delivery = "single"
	}
	if delivery != "single" && delivery != "area" && delivery != "instant" {
		return Result{}, &data.NotDefinedError{Kind: "delivery", Key: delivery}
	}
	var defender *Defender = &input.Defender
	var pipeline *data.Pipeline = &ruleset.Pipeline
	attackModifiers, calledShot, err := situationModifiers(ruleset, &input, delivery)
	if err != nil {
		return Result{}, err
	}
	maneuver, err := maneuverOf(ruleset, &input)
	if err != nil {
		return Result{}, err
	}
	rider, err := riderOf(ruleset, &input, maneuver)
	if err != nil {
		return Result{}, err
	}
	// A maneuver attempt (a save bonus row, or a rider) saves against the weapon's attribute.
	if (rider != nil || (maneuver != nil && maneuver.SaveBonus != nil)) && input.Damage.Ability == "" {
		return Result{}, fmt.Errorf("a maneuver attempt needs the weapon's attribute (damage.ability)")
	}
	var sequence *data.AttackSequence
	if input.Loadout != nil {
		if sequence = ruleset.AttackSequences.Covering(input.Loadout.Main, input.Loadout.Off); sequence == nil {
			return Result{}, &data.NoSequenceError{Loadout: *input.Loadout}
		}
	}
	if maneuver != nil && maneuver.Attack != 0 {
		attackModifiers = append(attackModifiers, modifier.Modifier{
			Target:    "attack_roll",
			Operation: modifier.Add,
			Value:     float64(maneuver.Attack),
			Source:    "maneuver:" + maneuver.Key,
		})
	}
	if defender.Stance != "" && defender.Stance != "dodge" && defender.Stance != "block" {
		return Result{}, &data.NotDefinedError{Kind: "stance", Key: defender.Stance}
	}
	if defender.Stance == "block" && (defender.ShieldDamageReduction == nil || defender.ShieldsSkill == nil) {
		return Result{}, &data.MissingInputError{Subject: "a block", Needs: "shield_dr and shields_skill"}
	}
	if slices.Contains(modifier.Flags(defender.Modifiers, 0, nil), "parry") && defender.WeaponSkill == nil {
		return Result{}, &data.MissingInputError{Subject: "a parry", Needs: "weapon_skill"}
	}
	evasionRate, blockRate, err := bands(ruleset, defender, damageType.Physical, delivery)
	if err != nil {
		return Result{}, err
	}
	// A situation's adds target joins the attack only through its row (situationModifiers).
	var adds map[string]bool = map[string]bool{}
	for _, row := range ruleset.Situations.Rows() {
		if row.Adds != "" {
			adds[row.Adds] = true
		}
	}
	var own []modifier.Modifier = slices.DeleteFunc(slices.Clone(input.Modifiers), func(m modifier.Modifier) bool { return adds[m.Target] })
	bonus, err := resolution.Bonus(ruleset.Origins, slices.Concat(own, attackModifiers))
	if err != nil {
		return Result{}, err
	}
	var attackerLuck int64 = input.Attacker.Attributes.Get(actor.Luck)
	var attack resolution.Result = resolution.Check(&ruleset.Resolution, input.Attacker.Skill, attackerLuck, bonus, input.Attacker.Roll)
	var result Result = Result{
		Target:      attack.Target,
		Margin:      attack.Margin,
		Critical:    slices.Contains(attack.Tags, "crit"),
		EvasionRate: evasionRate,
		BlockRate:   blockRate,
	}
	if delivery != "instant" && damageType.Physical && slices.Contains(modifier.Flags(defender.Modifiers, 0, nil), "parry") {
		result.ParryRate = formula.FloorDivide(*defender.WeaponSkill, ruleset.Stances.Rules.ParryBandDivisor)
	}
	if calledShot != nil {
		var noRiders []data.Rider = []data.Rider{}
		result.Riders = &noRiders
	}
	if input.Attacker.WeaponEffects != nil {
		var none []json.RawMessage = []json.RawMessage{}
		result.WeaponEffects = &none
	}
	if maneuver != nil {
		action, err := ruleset.Actions.Actions.Require("action", maneuver.Modifies)
		if err != nil {
			return Result{}, err
		}
		var cost int64 = action.Cost
		if sequence != nil && maneuver.Modifies == "attack" {
			cost = sequence.AP
		}
		var actionPoints int64 = max(cost+maneuver.AP, 1)
		var stamina int64 = actionPoints * ruleset.StaminaPerAP
		var noEffects []json.RawMessage = []json.RawMessage{}
		result.AP, result.Stamina, result.Effects = &actionPoints, &stamina, &noEffects
	}
	var carried *data.Maneuver = maneuver
	if rider != nil {
		var noEffects []json.RawMessage = []json.RawMessage{}
		carried, result.Rider, result.Effects = rider, rider.Key, &noEffects
	}
	if carried != nil {
		var noEvents []string = []string{}
		result.Events = &noEvents
	}
	if !attack.Success {
		result.Outcome, result.HitPoints = "miss", defender.HitPoints
		if slices.Contains(attack.Tags, "fumble") {
			result.Outcome = "fumble"
		}
		return result, nil
	}

	// The same roll read in bands: the evasion band, then a block stance's band under it, then
	// a parrying stance's band under that.
	// A crit always hits. Riposte and Bulwark read the band's lowest rolls, as wide as the
	// defender's crit window.
	var perkFlags []string = modifier.Flags(defender.Modifiers, 0, nil)
	var margin int64 = *attack.Margin
	var window int64 = ruleset.Resolution.CriticalWindow + formula.FloorDivide(defender.Attributes.Get(actor.Luck), ruleset.Resolution.CriticalWindowPerTwoLuck)
	switch {
	case result.Critical:
	case margin < evasionRate:
		result.Evaded = true
		result.Counter = defender.Stance == "dodge" && delivery == "single" && slices.Contains(perkFlags, "riposte") && margin >= evasionRate-window
		if delivery == "single" {
			result.Outcome, result.Damage, result.HitPoints = "evaded", 0, defender.HitPoints
			return result, nil
		}
	case margin < evasionRate+blockRate && !damageType.Physical:
		tier, err := modifier.Value(0, modifier.Target(defender.Modifiers, "spell_ward"), ruleset.Origins, 0, nil)
		if err != nil {
			return Result{}, err
		}
		ward, err := pipeline.SpellWardPercent.Evaluate(nil, formula.Variable{Name: "end", Value: defender.Attributes.Get(actor.Endurance)}, formula.Variable{Name: "tier", Value: tier})
		if err != nil {
			return Result{}, err
		}
		result.Warded, result.WardPercent = true, min(ward, pipeline.SpellWardMaxPercent)
	case margin < evasionRate+blockRate:
		result.Outcome, result.Blocked, result.Damage, result.HitPoints = "blocked", true, 0, defender.HitPoints
		if slices.Contains(perkFlags, "bulwark") && margin >= evasionRate+blockRate-window {
			result.StaggerAP = pipeline.BulwarkStaggerAP
		}
		return result, nil
	case margin < evasionRate+blockRate+result.ParryRate:
		result.Outcome, result.Blocked, result.Parried, result.Damage, result.HitPoints = "blocked", true, true, 0, defender.HitPoints
		return result, nil
	}
	result.Outcome = "hit"

	var amount int64 = input.Damage.Roll
	if input.Damage.Ability != "" {
		ability, known := actor.ByKey(input.Damage.Ability)
		if !known {
			return Result{}, &data.NotDefinedError{Kind: "ability", Key: input.Damage.Ability}
		}
		var attribute int64 = input.Attacker.Attributes.Get(ability)
		modifier, err := ruleset.AttributeModifier.Evaluate(nil, formula.Variable{Name: "attribute", Value: attribute})
		if err != nil {
			return Result{}, err
		}
		amount = max(amount+modifier, 0)
	}
	if calledShot != nil {
		amount = int64(math.Floor(float64(amount) * calledShot.DamageMultiplier))
	}
	if maneuver != nil && maneuver.DamageMultiplier != 0 {
		amount = int64(math.Floor(float64(amount) * maneuver.DamageMultiplier))
	}
	// The weapon's banes (docs/core/combat.md#weapon-effects), and its armor pierce for step 6.
	var pierce int64
	for _, applied := range input.Attacker.WeaponEffects {
		switch {
		case applied.Effect == baneEffect && applied.Parameter != "" && applied.Parameter == defender.CreatureType:
			amount += int64(applied.Magnitude)
		case applied.Effect == armorPierceEffect:
			pierce += int64(applied.Magnitude)
		}
	}
	pierce = min(pierce, 100)
	if result.Critical {
		if amount, err = pipeline.CriticalDamage.Evaluate(nil, formula.Variable{Name: "damage", Value: amount}); err != nil {
			return Result{}, err
		}
	}
	if result.Evaded {
		if amount, err = pipeline.AreaEvadeDamage.Evaluate(nil, formula.Variable{Name: "damage", Value: amount}); err != nil {
			return Result{}, err
		}
	}
	if damageType.Physical {
		var damageReduction int64 = defender.ArmorDamageReduction
		if defender.ShieldDamageReduction != nil {
			shield, err := modifier.Value(*defender.ShieldDamageReduction, modifier.Target(defender.Modifiers, "shield_dr"), ruleset.Origins, 0, nil)
			if err != nil {
				return Result{}, err
			}
			damageReduction += shield
		}
		if damageType.Pierce {
			if damageReduction, err = pipeline.PierceDamageReduction.Evaluate(nil, formula.Variable{Name: "dr", Value: damageReduction}); err != nil {
				return Result{}, err
			}
		}
		damageReduction -= damageReduction * pierce / 100
		amount = max(amount-damageReduction, 0)
	}
	amount = max(amount-defender.Resist[input.Damage.Type], 0)
	amount -= amount * result.WardPercent / 100
	result.Damage, result.HitPoints = amount, max(defender.HitPoints-amount, 0)
	if calledShot != nil && !result.Evaded {
		var riders []data.Rider = append([]data.Rider{}, calledShot.Riders...)
		result.Riders = &riders
	}
	if maneuver != nil && !result.Evaded {
		var effects []json.RawMessage = make([]json.RawMessage, 0, len(maneuver.Effects))
		for _, applied := range maneuver.Effects {
			effects = append(effects, applied.Encoded)
		}
		result.Effects = &effects
		var events []string = append([]string{}, maneuver.Events...)
		result.Events = &events
		if maneuver.SaveBonus != nil {
			result.Save = saveOf(&input, pipeline.ManeuverSaveMultiplier, *maneuver.SaveBonus)
		}
	}
	// A rider: the attack resolves as a plain one, then on a hit the maneuver's attempt
	// follows, with no save bonus.
	if rider != nil && !result.Evaded {
		var effects []json.RawMessage = make([]json.RawMessage, 0, len(rider.Effects))
		for _, applied := range rider.Effects {
			effects = append(effects, applied.Encoded)
		}
		var events []string = append([]string{}, rider.Events...)
		result.Effects, result.Events, result.Save = &effects, &events, saveOf(&input, pipeline.ManeuverSaveMultiplier, 0)
	}
	if input.Attacker.WeaponEffects != nil && !result.Evaded {
		landed, err := landedWeaponEffects(input.Attacker.WeaponEffects, input.Damage.Type, defender)
		if err != nil {
			return Result{}, err
		}
		result.WeaponEffects = &landed
	}
	return result, nil
}

// bands is step 2's rates: the evasion band (doubled by the dodge stance) and the block
// stance's band (physical hits, and with Spell Ward the spells it wards). Instant meets
// neither.
func bands(ruleset *data.Ruleset, defender *Defender, physical bool, delivery string) (int64, int64, error) {
	if delivery == "instant" {
		return 0, 0, nil
	}
	var pipeline *data.Pipeline = &ruleset.Pipeline
	var evasionSkill int64
	if defender.EvasionSkill != nil {
		evasionSkill = *defender.EvasionSkill
	}
	skill, err := modifier.Value(evasionSkill, modifier.Target(defender.Modifiers, "evasion"), ruleset.Origins, 0, nil)
	if err != nil {
		return 0, 0, err
	}
	evasionRate, err := pipeline.EvasionRate.Evaluate(nil, formula.Variable{Name: "skill", Value: skill})
	if err != nil {
		return 0, 0, err
	}
	if defender.Stance == "dodge" {
		if evasionRate, err = pipeline.DodgeEvasionRate.Evaluate(nil, formula.Variable{Name: "rate", Value: evasionRate}); err != nil {
			return 0, 0, err
		}
	}
	if defender.Loadout.HoldsTwoWeapons() {
		widened, err := modifier.Value(evasionRate, modifier.Target(defender.Modifiers, "two_weapon_evasion_band"), ruleset.Origins, 0, nil)
		if err != nil {
			return 0, 0, err
		}
		evasionRate = max(widened, 0)
	}
	// A fighting stance's evasion_band (forward, low, and long guard).
	banded, err := modifier.Value(evasionRate, modifier.Target(defender.Modifiers, "evasion_band"), ruleset.Origins, 0, nil)
	if err != nil {
		return 0, 0, err
	}
	evasionRate = max(banded, 0)
	// Unarmored Grace: no body armor widens the band.
	if defender.Armor == "" || defender.Armor == "none" {
		unarmored, err := modifier.Value(evasionRate, modifier.Target(defender.Modifiers, "unarmored_evasion_band"), ruleset.Origins, 0, nil)
		if err != nil {
			return 0, 0, err
		}
		evasionRate = max(unarmored, 0)
	}
	wardTier, err := modifier.Value(0, modifier.Target(defender.Modifiers, "spell_ward"), ruleset.Origins, 0, nil)
	if err != nil {
		return 0, 0, err
	}
	if defender.Stance != "block" || (!physical && wardTier <= 0) {
		return evasionRate, 0, nil
	}
	shields, err := modifier.Value(*defender.ShieldsSkill, modifier.Target(defender.Modifiers, "shields"), ruleset.Origins, 0, nil)
	if err != nil {
		return 0, 0, err
	}
	if defender.WeaponSkill != nil && slices.Contains(modifier.Flags(defender.Modifiers, 0, nil), "block_with_weapon") {
		shields = max(shields, *defender.WeaponSkill)
	}
	blockRate, err := pipeline.BlockRate.Evaluate(nil, formula.Variable{Name: "skill", Value: shields})
	if err != nil {
		return 0, 0, err
	}
	widened, err := modifier.Value(blockRate, modifier.Target(defender.Modifiers, "block_band"), ruleset.Origins, 0, nil)
	if err != nil {
		return 0, 0, err
	}
	return evasionRate, max(widened, 0), nil
}
