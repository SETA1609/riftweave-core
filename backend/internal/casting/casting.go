// Package casting casts one spell (docs/core/casting.md): the gate, the mana and reagent
// cost, one resolution on the spell's <color>_magic skill with the armor penalty, then the
// effects, concentration, and dispel. The roll is an input.
package casting

import (
	"encoding/json"
	"math"
	"slices"
	"strings"

	"github.com/SETA1609/riftweave-core/backend/internal/actor"
	"github.com/SETA1609/riftweave-core/backend/internal/combat"
	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/modifier"
	"github.com/SETA1609/riftweave-core/backend/internal/resolution"
)

// dispelEffect is the effect whose magnitude ends standing spells of that level or below.
const dispelEffect = "core:effect/dispel"

// Caster is the one casting: skills, Luck, mana, and what it wears, carries, and holds.
type Caster struct {
	Skills        map[string]int64 `json:"skills"`
	Concentration *string          `json:"concentration"` // the held spell key, or nil
	Armor         string           `json:"armor"`         // heaviest category worn; "" is none
	Flags         []string         `json:"flags"`
	Reagents      []string         `json:"reagents"`
	Attributes    actor.Attributes `json:"attributes"`
	Mana          int64            `json:"mana"`
	Summons       []string         `json:"summons"` // active summons, oldest first: a group ref each
}

// summonEffect calls creatures: its parameter is the group, its magnitude how many.
const summonEffect = "core:effect/summon"

// Target is the defender: its save against an instant spell, and, when it has hp, what the
// cast's damage does to it through combat steps 2 to 8 (docs/core/casting.md#damage-to-a-target).
type Target struct {
	Attributes            actor.Attributes    `json:"attributes"`
	Skills                map[string]int64    `json:"skills"`
	Roll                  *int64              `json:"roll"` // the save roll
	HitPoints             *int64              `json:"hp"`
	ArmorDamageReduction  int64               `json:"armor_dr"`
	Resist                map[string]int64    `json:"resist"`
	EvasionSkill          *int64              `json:"evasion_skill"`
	ShieldDamageReduction *int64              `json:"shield_dr"`
	ShieldsSkill          *int64              `json:"shields_skill"`
	Stance                string              `json:"stance"`
	Flags                 []string            `json:"flags"`
	Modifiers             []modifier.Modifier `json:"modifiers"`
}

// Packet is one damage effect of a cast and what the combat steps did with it.
type Packet struct {
	Effect  string `json:"effect"`
	Type    string `json:"type"`
	Amount  int64  `json:"amount"` // the packet before the steps
	Damage  int64  `json:"damage"` // HP lost
	Evaded  bool   `json:"evaded"`
	Blocked bool   `json:"blocked"`
}

// damagePrefix names a damage effect: core:effect/damage_<damage type>.
const damagePrefix = "core:effect/damage_"

// Save is the defender's roll against an instant spell, and what it decided.
type Save struct {
	resolution.Result
	Outcome string `json:"result"`
}

// Input is the `casting` command's input.
type Input struct {
	Spell     string              `json:"spell"`
	Modifiers []modifier.Modifier `json:"modifiers"`
	Standing  []string            `json:"standing"` // spell keys on the target, for dispel
	Target    *Target             `json:"target"`   // the defender, for an instant spell's save
	Caster    Caster              `json:"caster"`
	Roll      int64               `json:"roll"`
}

// Result is one cast. Check is nil when nothing was rolled.
type Result struct {
	Check         *resolution.Result `json:"check"`
	Save          *Save              `json:"save"`
	Concentration *string            `json:"concentration"`
	Ended         *string            `json:"ended"`
	Outcome       string             `json:"outcome"`
	Consumed      []string           `json:"consumed"`
	Effects       []json.RawMessage  `json:"effects"`
	Dispelled     []string           `json:"dispelled"`
	Summons       []string           `json:"summons"`
	Dismissed     []string           `json:"dismissed"`
	Mana          int64              `json:"mana"`
	Damage        []Packet           `json:"damage"`
	HitPoints     *int64             `json:"hp"` // the target's after every packet; nil without a target hp
}

// rollSave rolls the defender's save and leaves in result the effects that land.
func rollSave(ruleset *data.Ruleset, spell *data.Spell, target *Target, result *Result) error {
	var rules *data.CastingSave = &ruleset.Casting.Save
	var value int64
	if spell.Save.Skill != "" {
		value = target.Skills[spell.Save.Skill[strings.LastIndex(spell.Save.Skill, "/")+1:]]
	} else {
		ability, known := actor.ByKey(spell.Save.Attribute)
		if !known {
			return &data.NotDefinedError{Kind: "ability", Key: spell.Save.Attribute}
		}
		value = target.Attributes.Get(ability) * rules.AttributeMultiplier
	}
	var check resolution.Result = resolution.Check(&ruleset.Resolution, value, target.Attributes.Get(actor.Luck), 0, *target.Roll)
	var save Save = Save{Result: check, Outcome: "failed"}
	result.Save = &save
	if !check.Success {
		return nil
	}
	if slices.Contains(check.Tags, "crit") || *check.Margin >= rules.FullMargin {
		save.Outcome = "full"
		result.Effects = []json.RawMessage{}
		return nil
	}
	save.Outcome = "partial"
	for index, encoded := range result.Effects {
		var effect map[string]any
		if err := json.Unmarshal(encoded, &effect); err != nil {
			return err
		}
		for _, field := range []string{"magnitude", "duration"} {
			if amount, present := effect[field].(float64); present {
				effect[field] = max(1, int64(math.Floor(amount*float64(rules.PartialPercent)/100)))
			}
		}
		scaled, err := json.Marshal(effect)
		if err != nil {
			return err
		}
		result.Effects[index] = scaled
	}
	return nil
}

// Run is the `casting` command.
func Run(ruleset *data.Ruleset, input Input) (Result, error) {
	spell, err := ruleset.Spells.Require("spell", input.Spell)
	if err != nil {
		return Result{}, err
	}
	var standing []*data.Spell = make([]*data.Spell, 0, len(input.Standing))
	for _, key := range input.Standing {
		row, err := ruleset.Spells.Require("spell", key)
		if err != nil {
			return Result{}, err
		}
		standing = append(standing, row)
	}
	var caster *Caster = &input.Caster
	var skill string = spell.Color + "_magic"
	skillValue, ok := caster.Skills[skill]
	if !ok {
		return Result{}, &data.MissingInputError{Subject: "casting " + spell.Key, Needs: skill}
	}
	if spell.Delivery == "instant" && input.Target != nil && input.Target.Roll == nil {
		return Result{}, &data.MissingInputError{Subject: "an instant spell's save", Needs: "the target's roll"}
	}

	var result Result = Result{
		Mana:          caster.Mana,
		Consumed:      []string{},
		Effects:       []json.RawMessage{},
		Concentration: caster.Concentration,
		Dispelled:     []string{},
		Summons:       append([]string{}, caster.Summons...),
		Dismissed:     []string{},
		Damage:        []Packet{},
	}
	if input.Target != nil && input.Target.HitPoints != nil {
		var hp int64 = *input.Target.HitPoints
		result.HitPoints = &hp
	}
	for _, flag := range caster.Flags {
		if slices.Contains(ruleset.Casting.PreventedBy, flag) {
			result.Outcome = "prevented"
			return result, nil
		}
	}
	paid, err := paidCost(ruleset, spell.Cost.Mana, skillValue, input.Modifiers)
	if err != nil {
		return Result{}, err
	}
	if caster.Mana < paid {
		result.Outcome = "no_mana"
		return result, nil
	}
	var carried []string = slices.Clone(caster.Reagents)
	for _, reagent := range spell.Cost.Reagents {
		var index int = slices.Index(carried, reagent)
		if index < 0 {
			result.Outcome = "no_reagents"
			return result, nil
		}
		carried = slices.Delete(carried, index, index+1)
	}
	result.Mana = caster.Mana - paid
	result.Consumed = append(result.Consumed, spell.Cost.Reagents...)

	check, err := castCheck(ruleset, &input, skill, skillValue)
	if err != nil {
		return Result{}, err
	}
	result.Check = &check
	if !check.Success {
		result.Outcome = "failed"
		return result, nil
	}

	result.Outcome = "cast"
	for _, applied := range spell.Effects {
		result.Effects = append(result.Effects, applied.Encoded)
	}
	if spell.Delivery == "instant" && input.Target != nil {
		if err := rollSave(ruleset, spell, input.Target, &result); err != nil {
			return Result{}, err
		}
	}
	if input.Target != nil && input.Target.HitPoints != nil {
		if err := applyDamage(ruleset, spell, &input, &check, &result); err != nil {
			return Result{}, err
		}
	}
	if spell.Concentration {
		var held *string = caster.Concentration
		var key string = spell.Key
		result.Concentration = &key
		if held != nil && *held != spell.Key {
			result.Ended = held
		}
	}
	var summoned []string
	for _, applied := range spell.Effects {
		if applied.Effect == summonEffect {
			for range int(max(applied.Magnitude, 1)) {
				summoned = append(summoned, applied.Parameter)
			}
		}
	}
	if len(summoned) > 0 {
		// The newest join; past the limit the oldest are dismissed (docs/core/casting.md#summons).
		limit, err := modifier.Value(ruleset.Casting.SummonLimit, modifier.Target(input.Modifiers, "summon_limit"), ruleset.Origins, 0, nil)
		if err != nil {
			return Result{}, err
		}
		var summons []string = append(result.Summons, summoned...)
		var past int = max(len(summons)-int(limit), 0)
		result.Dismissed = append([]string{}, summons[:past]...)
		result.Summons = summons[past:]
	}
	for _, applied := range spell.Effects {
		if applied.Effect != dispelEffect {
			continue
		}
		for _, row := range standing {
			if float64(row.Level) <= applied.Magnitude && !slices.Contains(result.Dispelled, row.Key) {
				result.Dispelled = append(result.Dispelled, row.Key)
			}
		}
	}
	return result, nil
}

// paidCost is what the caster pays (docs/core/casting.md#paid-cost): the skill discount,
// rounded up, then the modifiers on spell_cost by the kernel rules, at least 1.
func paidCost(ruleset *data.Ruleset, authored int64, skill int64, modifiers []modifier.Modifier) (int64, error) {
	var rule *data.SkillDiscount = &ruleset.Casting.SkillDiscount
	var discount float64 = min(float64(skill)/rule.Divisor, rule.Cap)
	var discounted float64 = float64(authored) * (1 - discount)
	var base int64 = int64(math.Ceil(math.Round(discounted*1e6) / 1e6))
	value, err := modifier.Value(base, modifier.Target(modifiers, "spell_cost"), ruleset.Origins, 0, nil)
	if err != nil {
		return 0, err
	}
	return max(rule.MinimumPaid, value), nil
}

// castCheck is step 5: the modifiers on the spell's skill, plus the armor penalty as one more
// add with source armor:<category>, through the resolution function.
func castCheck(ruleset *data.Ruleset, input *Input, skill string, skillValue int64) (resolution.Result, error) {
	var armor string = input.Caster.Armor
	if armor == "" {
		armor = "none"
	}
	penalty, ok := ruleset.Casting.ArmorPenalty[armor]
	if !ok {
		return resolution.Result{}, &data.NotDefinedError{Kind: "armor category", Key: armor}
	}
	var modifiers []modifier.Modifier = modifier.Target(input.Modifiers, skill)
	if penalty != 0 {
		modifiers = append(modifiers, modifier.Modifier{
			Target:    skill,
			Operation: modifier.Add,
			Value:     float64(penalty),
			Source:    "armor:" + armor,
		})
	}
	bonus, err := resolution.Bonus(ruleset.Origins, modifiers)
	if err != nil {
		return resolution.Result{}, err
	}
	var luck int64 = input.Caster.Attributes.Get(actor.Luck)
	return resolution.Check(&ruleset.Resolution, skillValue, luck, bonus, input.Roll), nil
}

// applyDamage runs each damage packet through combat steps 2 to 8, the cast check standing in
// for the attack roll: the same target, roll, and LCK give the same margin and crit.
func applyDamage(ruleset *data.Ruleset, spell *data.Spell, input *Input, check *resolution.Result, result *Result) error {
	var target *Target = input.Target
	var delivery string = spell.Delivery
	if delivery == "" {
		delivery = "single"
	}
	for _, encoded := range result.Effects {
		var effect struct {
			Effect    string   `json:"effect"`
			Magnitude float64  `json:"magnitude"`
			Duration  *float64 `json:"duration"`
		}
		if err := json.Unmarshal(encoded, &effect); err != nil {
			return err
		}
		if !strings.HasPrefix(effect.Effect, damagePrefix) || (effect.Duration != nil && *effect.Duration > 1) {
			continue
		}
		var kind string = strings.TrimPrefix(effect.Effect, damagePrefix)
		if _, isType := ruleset.DamageTypes.Get(kind); !isType {
			continue
		}
		var amount int64 = int64(effect.Magnitude)
		hit, err := combat.Run(ruleset, combat.Input{
			Attacker: combat.Attacker{Skill: check.Target, Roll: input.Roll, Attributes: input.Caster.Attributes},
			Defender: combat.Defender{
				EvasionSkill:          target.EvasionSkill,
				Attributes:            target.Attributes,
				ArmorDamageReduction:  target.ArmorDamageReduction,
				HitPoints:             *result.HitPoints,
				Resist:                target.Resist,
				ShieldDamageReduction: target.ShieldDamageReduction,
				ShieldsSkill:          target.ShieldsSkill,
				Stance:                target.Stance,
				Flags:                 target.Flags,
				Modifiers:             target.Modifiers,
			},
			Damage:   combat.Damage{Roll: amount, Type: kind},
			Delivery: delivery,
		})
		if err != nil {
			return err
		}
		*result.HitPoints = hit.HitPoints
		result.Damage = append(result.Damage, Packet{Effect: effect.Effect, Type: kind, Amount: amount, Damage: hit.Damage, Evaded: hit.Evaded, Blocked: hit.Blocked})
	}
	return nil
}
