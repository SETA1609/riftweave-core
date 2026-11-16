// Package combat resolves one attack in the order of docs/core/combat.md: the situations
// and the called shot, one attack roll read against the defender's evasion and block bands,
// then damage through the called shot, crit, area, armor and shield DR, and resist. Dice and
// rolls are inputs. The bands and multipliers are formulas in combat/pipeline.json.
package combat

import (
	"fmt"
	"math"
	"slices"

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
}

// Defender is who takes the attack.
type Defender struct {
	EvasionSkill         *int64           `json:"evasion_skill"`
	Attributes           actor.Attributes `json:"attributes"`
	ArmorDamageReduction int64            `json:"armor_dr"`
	// ShieldDamageReduction is a carried shield's DR; it adds to the armor DR.
	ShieldDamageReduction *int64 `json:"shield_dr"`
	ShieldsSkill          *int64 `json:"shields_skill"`
	// Stance is the one the defender declared: dodge, block, or none.
	Stance    string           `json:"stance"`
	HitPoints int64            `json:"hp"`
	Resist    map[string]int64 `json:"resist"`
	Flags     []string         `json:"flags"`
	// Modifiers are the defender's perks: adds on evasion, shields, and shield_dr, and the
	// block_magic, bulwark, and riposte grants.
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
	Attacker   Attacker            `json:"attacker"`
	Defender   Defender            `json:"defender"`
	Damage     Damage              `json:"damage"`
	Delivery   string              `json:"delivery"`
	Modifiers  []modifier.Modifier `json:"modifiers"`
	Range      string              `json:"range"`
	Situations []string            `json:"situations"`
	CalledShot string              `json:"called_shot"`
}

// Result is the outcome of the attack.
type Result struct {
	Outcome     string `json:"outcome"`
	Target      int64  `json:"target"`
	Margin      *int64 `json:"margin"`
	Critical    bool   `json:"crit"`
	Evaded      bool   `json:"evaded"`
	Blocked     bool   `json:"blocked"`
	EvasionRate int64  `json:"evasion_rate"`
	BlockRate   int64  `json:"block_rate"`
	Damage      int64  `json:"damage"`
	HitPoints   int64  `json:"hp"`
	// Counter is a riposte: a dodge's evade in the window lets the defender attack back.
	Counter bool `json:"counter"`
	// StaggerAP is Bulwark: AP the attacker loses on their next turn.
	StaggerAP int64 `json:"stagger_ap"`
	// Riders are the called shot's riders on a hit, and empty on a miss or an evade. Without
	// a called shot the field is left out.
	Riders *[]data.Rider `json:"riders,omitempty"`
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
func Run(ruleset *data.Ruleset, input Input) (Result, error) {
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
	if defender.Stance != "" && defender.Stance != "dodge" && defender.Stance != "block" {
		return Result{}, &data.NotDefinedError{Kind: "stance", Key: defender.Stance}
	}
	if defender.Stance == "block" && (defender.ShieldDamageReduction == nil || defender.ShieldsSkill == nil) {
		return Result{}, &data.MissingInputError{Subject: "a block", Needs: "shield_dr and shields_skill"}
	}
	evasionRate, blockRate, err := bands(ruleset, defender, damageType.Physical, delivery)
	if err != nil {
		return Result{}, err
	}
	bonus, err := resolution.Bonus(ruleset.Origins, slices.Concat(input.Modifiers, attackModifiers))
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
	if calledShot != nil {
		var noRiders []data.Rider = []data.Rider{}
		result.Riders = &noRiders
	}
	if !attack.Success {
		result.Outcome, result.HitPoints = "miss", defender.HitPoints
		if slices.Contains(attack.Tags, "fumble") {
			result.Outcome = "fumble"
		}
		return result, nil
	}

	// The same roll read in bands: the evasion band, then a block stance's band under it.
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
	case margin < evasionRate+blockRate:
		result.Outcome, result.Blocked, result.Damage, result.HitPoints = "blocked", true, 0, defender.HitPoints
		if slices.Contains(perkFlags, "bulwark") && margin >= evasionRate+blockRate-window {
			result.StaggerAP = pipeline.BulwarkStaggerAP
		}
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
		amount = max(amount-damageReduction, 0)
	}
	amount = max(amount-defender.Resist[input.Damage.Type], 0)
	result.Damage, result.HitPoints = amount, max(defender.HitPoints-amount, 0)
	if calledShot != nil && !result.Evaded {
		var riders []data.Rider = append([]data.Rider{}, calledShot.Riders...)
		result.Riders = &riders
	}
	return result, nil
}

// bands is step 2's rates: the evasion band (doubled by the dodge stance) and the block
// stance's band (halved against magic with block_magic, none without it). Instant meets
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
	var blockMagic bool = slices.Contains(modifier.Flags(defender.Modifiers, 0, nil), "block_magic")
	if defender.Stance != "block" || (!physical && !blockMagic) {
		return evasionRate, 0, nil
	}
	shields, err := modifier.Value(*defender.ShieldsSkill, modifier.Target(defender.Modifiers, "shields"), ruleset.Origins, 0, nil)
	if err != nil {
		return 0, 0, err
	}
	blockRate, err := pipeline.BlockRate.Evaluate(nil, formula.Variable{Name: "skill", Value: shields})
	if err != nil {
		return 0, 0, err
	}
	if !physical {
		if blockRate, err = pipeline.MagicBlockRate.Evaluate(nil, formula.Variable{Name: "rate", Value: blockRate}); err != nil {
			return 0, 0, err
		}
	}
	return evasionRate, blockRate, nil
}
