// Package resolution is the d100 roll-under check (docs/core/resolution.md). Combat and
// the barter check call Check; nothing else decides a success, a crit, or a fumble.
package resolution

import (
	"fmt"
	"math"

	"github.com/SETA1609/riftweave-core/backend/internal/actor"
	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/formula"
	"github.com/SETA1609/riftweave-core/backend/internal/modifier"
)

// Input is the `resolution` command's input.
type Input struct {
	Actor     actor.Actor         `json:"actor"`
	Skill     string              `json:"skill"`
	Modifiers []modifier.Modifier `json:"modifiers"`
	Roll      int64               `json:"roll"`
}

// Result is one check. Margin is nil on a failure.
type Result struct {
	Target  int64    `json:"target"`
	Success bool     `json:"success"`
	Margin  *int64   `json:"margin"`
	Tags    []string `json:"tags"`
}

// Bonus is the signed sum of every `add` modifier's value. Resolution sees bonuses and
// penalties as one list.
func Bonus(mods []modifier.Modifier) int64 {
	sum := 0.0
	for _, m := range mods {
		if m.Op == modifier.Add {
			sum += m.Value
		}
	}
	return int64(math.Floor(sum))
}

// Check rolls one check: the skill plus the bonus is the target, capped if the data caps it.
func Check(r *data.Resolution, skill, luck, bonus, roll int64) Result {
	target := skill + bonus
	if r.TargetCap != nil {
		target = min(target, *r.TargetCap)
	}
	success := roll <= target && roll < r.FumbleAt
	window := r.CritWindow + formula.FloorDiv(luck, r.CritWindowPerTwoLuck)
	res := Result{Target: target, Success: success, Tags: []string{}}
	if success {
		margin := target - roll
		res.Margin = &margin
		if roll <= window {
			res.Tags = append(res.Tags, "crit")
		}
	}
	if roll >= r.FumbleAt {
		res.Tags = append(res.Tags, "fumble")
	}
	return res
}

// Run is the `resolution` command.
func Run(rs *data.Ruleset, in Input) (Result, error) {
	skill := in.Skill
	if skill == "" {
		if len(in.Actor.Skills) != 1 {
			return Result{}, fmt.Errorf("actor has several skills; name one with `skill`")
		}
		for name := range in.Actor.Skills {
			skill = name
		}
	}
	value, ok := in.Actor.Skills[skill]
	if !ok {
		return Result{}, fmt.Errorf("actor has no skill %q", skill)
	}
	return Check(&rs.Resolution, value, in.Actor.Attributes.Get(actor.Lck), Bonus(in.Modifiers), in.Roll), nil
}
