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
	Skill      int64       `json:"skill"`
	Roll       int64       `json:"roll"`
	Attributes actor.Attrs `json:"attributes"`
}

// Reaction is the defender's one reaction: an evade, or a block with a shield.
type Reaction struct {
	Kind     string `json:"kind"`
	Roll     int64  `json:"roll"`
	Skill    *int64 `json:"skill"`
	ShieldDR *int64 `json:"shield_dr"`
}

// Defender is who takes the attack.
type Defender struct {
	EvasionSkill *int64           `json:"evasion_skill"`
	Attributes   actor.Attrs      `json:"attributes"`
	ArmorDR      int64            `json:"armor_dr"`
	HP           int64            `json:"hp"`
	Resist       map[string]int64 `json:"resist"`
	Reaction     *Reaction        `json:"reaction"`
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
	Crit         bool     `json:"crit"`
	Evaded       bool     `json:"evaded"`
	Blocked      bool     `json:"blocked"`
	ReactionTags []string `json:"reaction_tags"`
	Damage       int64    `json:"damage"`
	HP           int64    `json:"hp"`
}

// Run is the `combat` command.
func Run(rs *data.Ruleset, in Input) (Result, error) {
	damageType, ok := rs.DamageTypes.Get(in.Damage.Type)
	if !ok {
		return Result{}, fmt.Errorf("damage type %q is not defined", in.Damage.Type)
	}
	delivery := in.Delivery
	if delivery == "" {
		delivery = "single"
	}
	if delivery != "single" && delivery != "area" && delivery != "instant" {
		return Result{}, fmt.Errorf("delivery %q is not defined", delivery)
	}
	def, pipe := &in.Defender, &rs.Pipeline

	attack := resolution.Check(&rs.Resolution, in.Attacker.Skill, in.Attacker.Attributes.Get(actor.Lck), resolution.Bonus(in.Modifiers), in.Attacker.Roll)
	res := Result{Target: attack.Target, Margin: attack.Margin, Crit: slices.Contains(attack.Tags, "crit"), ReactionTags: []string{}}
	if !attack.Success {
		res.Outcome, res.HP = "miss", def.HP
		if slices.Contains(attack.Tags, "fumble") {
			res.Outcome = "fumble"
		}
		return res, nil
	}

	// One reaction per attack, rolled only on a hit. Delivery decides what it can do: single
	// is evaded whole, area is halved by an evade, instant takes no reaction.
	var reaction *Reaction
	if delivery != "instant" {
		reaction = def.Reaction
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
			if def.EvasionSkill == nil {
				return Result{}, fmt.Errorf("an evade needs the defender's evasion_skill")
			}
			skill = *def.EvasionSkill
		} else {
			if reaction.Skill == nil || reaction.ShieldDR == nil {
				return Result{}, fmt.Errorf("a block needs skill and shield_dr")
			}
			skill = *reaction.Skill
		}
		check := resolution.Check(&rs.Resolution, skill, def.Attributes.Get(actor.Lck), 0, reaction.Roll)
		res.ReactionTags = check.Tags
		if check.Success && *check.Margin >= *attack.Margin {
			if reaction.Kind == "block" {
				res.Blocked = true
			} else {
				res.Evaded = true
				if delivery == "single" {
					res.Outcome, res.Damage, res.HP = "evaded", 0, def.HP
					return res, nil
				}
			}
		}
	}
	res.Outcome = "hit"

	amount := in.Damage.Roll
	var err error
	if res.Crit {
		if amount, err = pipe.CritDamage.Eval(nil, formula.Var{Name: "damage", Val: amount}); err != nil {
			return Result{}, err
		}
	}
	if res.Evaded {
		if amount, err = pipe.AreaEvadeDamage.Eval(nil, formula.Var{Name: "damage", Val: amount}); err != nil {
			return Result{}, err
		}
	}
	if res.Blocked {
		amount = max(amount-*reaction.ShieldDR, 0)
	}
	if damageType.Physical {
		dr := def.ArmorDR
		if damageType.Pierce {
			if dr, err = pipe.PierceDR.Eval(nil, formula.Var{Name: "dr", Val: dr}); err != nil {
				return Result{}, err
			}
		}
		amount = max(amount-dr, 0)
	}
	amount = max(amount-def.Resist[in.Damage.Type], 0)
	res.Damage, res.HP = amount, max(def.HP-amount, 0)
	return res, nil
}
