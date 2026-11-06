// Package combat resolves one attack in the order of docs/core/combat.md: the attack roll,
// one reaction, then damage through crit, area, block, armor, and resist. Dice and rolls
// are inputs. The multipliers are formulas in combat/pipeline.json, not constants here.
package combat

import (
	"fmt"
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
}

// Damage is the damage roll and its type.
type Damage struct {
	Roll int64  `json:"roll"`
	Type string `json:"type"`
}

// Input is the `combat` command's input.
type Input struct {
	Attacker  Attacker            `json:"attacker"`
	Defender  Defender            `json:"defender"`
	Damage    Damage              `json:"damage"`
	Delivery  string              `json:"delivery"`
	Modifiers []modifier.Modifier `json:"modifiers"`
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
}

// Run is the `combat` command.
func Run(ruleset *data.Ruleset, input Input) (Result, error) {
	damageType, ok := ruleset.DamageTypes.Get(input.Damage.Type)
	if !ok {
		return Result{}, fmt.Errorf("damage type %q is not defined", input.Damage.Type)
	}
	var delivery string = input.Delivery
	if delivery == "" {
		delivery = "single"
	}
	if delivery != "single" && delivery != "area" && delivery != "instant" {
		return Result{}, fmt.Errorf("delivery %q is not defined", delivery)
	}
	var defender *Defender = &input.Defender
	var pipeline *data.Pipeline = &ruleset.Pipeline

	var attackerLuck int64 = input.Attacker.Attributes.Get(actor.Luck)
	var attack resolution.Result = resolution.Check(&ruleset.Resolution, input.Attacker.Skill, attackerLuck, resolution.Bonus(input.Modifiers), input.Attacker.Roll)
	var result Result = Result{
		Target:       attack.Target,
		Margin:       attack.Margin,
		Critical:     slices.Contains(attack.Tags, "crit"),
		ReactionTags: []string{},
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
	if reaction != nil {
		if reaction.Kind != "evade" && reaction.Kind != "block" {
			return Result{}, fmt.Errorf("reaction kind %q is not defined", reaction.Kind)
		}
		if reaction.Kind == "block" && !damageType.Physical {
			reaction = nil
		}
	}
	if reaction != nil {
		// The reaction is a d100 check like the attack: same fumble and crit window.
		var skill int64
		if reaction.Kind == "evade" {
			if defender.EvasionSkill == nil {
				return Result{}, fmt.Errorf("an evade needs the defender's evasion_skill")
			}
			skill = *defender.EvasionSkill
		} else {
			if reaction.Skill == nil || reaction.ShieldDamageReduction == nil {
				return Result{}, fmt.Errorf("a block needs skill and shield_dr")
			}
			skill = *reaction.Skill
		}
		var check resolution.Result = resolution.Check(&ruleset.Resolution, skill, defender.Attributes.Get(actor.Luck), 0, reaction.Roll)
		result.ReactionTags = check.Tags
		if check.Success && *check.Margin >= *attack.Margin {
			if reaction.Kind == "block" {
				result.Blocked = true
			} else {
				result.Evaded = true
				if delivery == "single" {
					result.Outcome, result.Damage, result.HitPoints = "evaded", 0, defender.HitPoints
					return result, nil
				}
			}
		}
	}
	result.Outcome = "hit"

	var amount int64 = input.Damage.Roll
	var err error
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
		amount = max(amount-*reaction.ShieldDamageReduction, 0)
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
	return result, nil
}
