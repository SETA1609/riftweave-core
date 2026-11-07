// Package casting casts one spell (docs/core/casting.md): the gate, the mana and reagent
// cost, one resolution on the spell's <color>_magic skill with the armor penalty, then the
// effects, concentration, and dispel. The roll is an input.
package casting

import (
	"encoding/json"
	"slices"

	"github.com/SETA1609/riftweave-core/backend/internal/actor"
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
}

// Input is the `casting` command's input.
type Input struct {
	Spell     string              `json:"spell"`
	Modifiers []modifier.Modifier `json:"modifiers"`
	Standing  []string            `json:"standing"` // spell keys on the target, for dispel
	Caster    Caster              `json:"caster"`
	Roll      int64               `json:"roll"`
}

// Result is one cast. Check is nil when nothing was rolled.
type Result struct {
	Check         *resolution.Result `json:"check"`
	Concentration *string            `json:"concentration"`
	Ended         *string            `json:"ended"`
	Outcome       string             `json:"outcome"`
	Consumed      []string           `json:"consumed"`
	Effects       []json.RawMessage  `json:"effects"`
	Dispelled     []string           `json:"dispelled"`
	Mana          int64              `json:"mana"`
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

	var result Result = Result{
		Mana:          caster.Mana,
		Consumed:      []string{},
		Effects:       []json.RawMessage{},
		Concentration: caster.Concentration,
		Dispelled:     []string{},
	}
	for _, flag := range caster.Flags {
		if slices.Contains(ruleset.Casting.PreventedBy, flag) {
			result.Outcome = "prevented"
			return result, nil
		}
	}
	if caster.Mana < spell.Cost.Mana {
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
	result.Mana = caster.Mana - spell.Cost.Mana
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
	if spell.Concentration {
		var held *string = caster.Concentration
		var key string = spell.Key
		result.Concentration = &key
		if held != nil && *held != spell.Key {
			result.Ended = held
		}
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
