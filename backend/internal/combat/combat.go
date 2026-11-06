// Package combat resolves one attack in the order of docs/core/combat.md: the situations
// and the called shot, the attack roll, one reaction, then damage through the called shot,
// crit, area, block, armor, and resist. Dice and rolls
// are inputs. The multipliers are formulas in combat/pipeline.json, not constants here.
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

// Reaction is the defender's one reaction: an evade, or a block with a shield.
type Reaction struct {
	Kind                  string `json:"kind"`
	Roll                  int64  `json:"roll"`
	Skill                 *int64 `json:"skill"`
	ShieldDamageReduction *int64 `json:"shield_dr"`
}

// Defender is who takes the attack.
type Defender struct {
	EvasionSkill         *int64           `json:"evasion_skill"`
	Attributes           actor.Attributes `json:"attributes"`
	ArmorDamageReduction int64            `json:"armor_dr"`
	HitPoints            int64            `json:"hp"`
	Resist               map[string]int64 `json:"resist"`
	Reaction             *Reaction        `json:"reaction"`
	Flags                []string         `json:"flags"`
	// Modifiers are the defender's perks: adds on evasion, block, and shield_dr, and the
	// block_magic, bulwark, and riposte grants.
	Modifiers []modifier.Modifier `json:"modifiers"`
}

// Damage is the damage roll and its type.
type Damage struct {
	Roll int64  `json:"roll"`
	Type string `json:"type"`
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
	Outcome      string   `json:"outcome"`
	Target       int64    `json:"target"`
	Margin       *int64   `json:"margin"`
	Critical     bool     `json:"crit"`
	Evaded       bool     `json:"evaded"`
	Blocked      bool     `json:"blocked"`
	ReactionTags []string `json:"reaction_tags"`
	Damage       int64    `json:"damage"`
	HitPoints    int64    `json:"hp"`
	// Counter is a riposte: a crit evade of a single attack lets the defender attack back.
	Counter bool `json:"counter"`
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
			Target:    "attack",
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
	bonus, err := resolution.Bonus(ruleset.Origins, slices.Concat(input.Modifiers, attackModifiers))
	if err != nil {
		return Result{}, err
	}
	var attackerLuck int64 = input.Attacker.Attributes.Get(actor.Luck)
	var attack resolution.Result = resolution.Check(&ruleset.Resolution, input.Attacker.Skill, attackerLuck, bonus, input.Attacker.Roll)
	var result Result = Result{
		Target:       attack.Target,
		Margin:       attack.Margin,
		Critical:     slices.Contains(attack.Tags, "crit"),
		ReactionTags: []string{},
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

	// One reaction per attack, rolled only on a hit. Delivery decides what it can do: single
	// is evaded whole, area is halved by an evade, instant takes no reaction.
	var reaction *Reaction
	if delivery != "instant" {
		reaction = defender.Reaction
	}
	// The defender's perks: adds on the reaction targets and shield_dr, and grant flags.
	var perkFlags []string = modifier.Flags(defender.Modifiers, 0, nil)
	if reaction != nil {
		if reaction.Kind != "evade" && reaction.Kind != "block" {
			return Result{}, &data.NotDefinedError{Kind: "reaction kind", Key: reaction.Kind}
		}
		if reaction.Kind == "block" && !damageType.Physical && !slices.Contains(perkFlags, "block_magic") {
			reaction = nil
		}
	}
	if reaction != nil {
		// The reaction is a d100 check like the attack: same fumble and crit window.
		var skill int64
		var reactionTarget string = "block"
		if reaction.Kind == "evade" {
			reactionTarget = "evasion"
			if defender.EvasionSkill == nil {
				return Result{}, &data.MissingInputError{Subject: "an evade", Needs: "the defender's evasion_skill"}
			}
			skill = *defender.EvasionSkill
		} else {
			if reaction.Skill == nil || reaction.ShieldDamageReduction == nil {
				return Result{}, &data.MissingInputError{Subject: "a block", Needs: "skill and shield_dr"}
			}
			skill = *reaction.Skill
		}
		reactionBonus, err := resolution.Bonus(ruleset.Origins, modifier.Target(defender.Modifiers, reactionTarget))
		if err != nil {
			return Result{}, err
		}
		var check resolution.Result = resolution.Check(&ruleset.Resolution, skill, defender.Attributes.Get(actor.Luck), reactionBonus, reaction.Roll)
		result.ReactionTags = check.Tags
		if check.Success && *check.Margin >= *attack.Margin {
			if reaction.Kind == "block" {
				result.Blocked = true
			} else {
				result.Evaded = true
				result.Counter = delivery == "single" && slices.Contains(check.Tags, "crit") && slices.Contains(perkFlags, "riposte")
				if delivery == "single" {
					result.Outcome, result.Damage, result.HitPoints = "evaded", 0, defender.HitPoints
					return result, nil
				}
			}
		}
	}
	result.Outcome = "hit"

	var amount int64 = input.Damage.Roll
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
	if result.Blocked {
		shieldDamageReduction, err := modifier.Value(*reaction.ShieldDamageReduction, modifier.Target(defender.Modifiers, "shield_dr"), ruleset.Origins, 0, nil)
		if err != nil {
			return Result{}, err
		}
		if !damageType.Physical {
			shieldDamageReduction = formula.FloorDivide(shieldDamageReduction, 2)
		}
		amount = max(amount-shieldDamageReduction, 0)
		if slices.Contains(perkFlags, "bulwark") && slices.Contains(result.ReactionTags, "crit") {
			amount = 0
		}
	}
	if damageType.Physical {
		var damageReduction int64 = defender.ArmorDamageReduction
		if damageType.Pierce {
			if damageReduction, err = pipeline.PierceDamageReduction.Evaluate(nil, formula.Variable{Name: "dr", Value: damageReduction}); err != nil {
				return Result{}, err
			}
		}
		amount = max(amount-damageReduction, 0)
	}
	amount = max(amount-defender.Resist[input.Damage.Type], 0)
	result.Damage, result.HitPoints = amount, max(defender.HitPoints-amount, 0)
	if calledShot != nil {
		var riders []data.Rider = append([]data.Rider{}, calledShot.Riders...)
		result.Riders = &riders
	}
	return result, nil
}
