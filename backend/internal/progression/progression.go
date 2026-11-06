// Package progression is XP, level, the skill points of a level-up, and spending them on a
// skill (docs/core/progression.md). The curve, the caps, and the costs are in the data.
package progression

import (
	"errors"
	"fmt"

	"github.com/SETA1609/riftweave-core/backend/internal/actor"
	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/formula"
)

// Input is the `progression` command's input; `op` says which fields it reads.
type Input struct {
	Op       string       `json:"op"`
	Level    *int64       `json:"level"`
	XP       *int64       `json:"xp"`
	Actor    *actor.Actor `json:"actor"`
	LuckRoll *int64       `json:"luck_roll"`
	Skill    *int64       `json:"skill"`
	Tagged   *bool        `json:"tagged"`
	Points   *int64       `json:"points"`
}

type xpResult struct {
	Total int64  `json:"xp_total"`
	Step  *int64 `json:"xp_step"`
}

type levelResult struct {
	Level int64 `json:"level"`
}

type levelUpResult struct {
	SkillPoints int64 `json:"skill_points"`
	Perks       int64 `json:"perks"`
}

type spendResult struct {
	Ranks int64 `json:"ranks"`
	Spent int64 `json:"spent"`
	Left  int64 `json:"left"`
}

func need(op, field string, given bool) error {
	if given {
		return nil
	}
	return fmt.Errorf("progression %s needs %s", op, field)
}

// Run is the `progression` command. Its result type depends on `op`.
func Run(rs *data.Ruleset, in Input) (any, error) {
	p := &rs.Progression
	switch in.Op {
	case "xp":
		if err := need(in.Op, "level", in.Level != nil); err != nil {
			return nil, err
		}
		total, err := p.XPToReach.Eval(nil, formula.Var{Name: "level", Val: *in.Level})
		if err != nil {
			return nil, err
		}
		res := xpResult{Total: total}
		if *in.Level > p.StartLevel {
			prev, err := p.XPToReach.Eval(nil, formula.Var{Name: "level", Val: *in.Level - 1})
			if err != nil {
				return nil, err
			}
			step := total - prev
			res.Step = &step
		}
		return res, nil
	case "level":
		if err := need(in.Op, "xp", in.XP != nil); err != nil {
			return nil, err
		}
		level := p.StartLevel
		for level < p.MaxLevel {
			needed, err := p.XPToReach.Eval(nil, formula.Var{Name: "level", Val: level + 1})
			if err != nil {
				return nil, err
			}
			if needed > *in.XP {
				break
			}
			level++
		}
		return levelResult{level}, nil
	case "level_up":
		if err := need(in.Op, "actor and luck_roll", in.Actor != nil && in.LuckRoll != nil); err != nil {
			return nil, err
		}
		attrs := in.Actor.Attributes
		if !attrs.Has(p.LuckMaxAttr) {
			return nil, fmt.Errorf("progression level_up needs the %s attribute", p.LuckMaxAttr)
		}
		if *in.LuckRoll < p.LuckRoll.Min || *in.LuckRoll > attrs.Get(p.LuckMaxAttr) {
			return nil, errors.New("luck_roll out of range")
		}
		points, err := p.SkillPoints.Eval(&attrs, formula.Var{Name: "luck_roll", Val: *in.LuckRoll})
		if err != nil {
			return nil, err
		}
		return levelUpResult{points, p.PerksPerLevel}, nil
	case "spend":
		if err := need(in.Op, "skill, tagged and points", in.Skill != nil && in.Tagged != nil && in.Points != nil); err != nil {
			return nil, err
		}
		cost := p.SkillCost.Normal
		if *in.Tagged {
			cost = p.SkillCost.Tag
		}
		ranks := min(*in.Points/cost, max(p.SkillCap-*in.Skill, 0))
		return spendResult{ranks, ranks * cost, *in.Points - ranks*cost}, nil
	}
	return nil, fmt.Errorf("progression op %q is not defined", in.Op)
}
