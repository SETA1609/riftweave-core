// Package creation is the `creation` command (BE_S07_08, docs/backend/creation.md): the
// creation rules an app draws its form from, a check of a draft step by step, and the stored
// actor a complete draft becomes. The numbers are progression's `creation` block and
// `tag_skills` (GD_S07_07); race traits and creation perks that add to `tag_skills` or
// `creation_perks` raise those limits. It holds no number of its own.
package creation

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/SETA1609/riftweave-core/backend/internal/actor"
	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/derived"
	"github.com/SETA1609/riftweave-core/backend/internal/formula"
	"github.com/SETA1609/riftweave-core/backend/internal/modifier"
	"github.com/SETA1609/riftweave-core/backend/internal/stored"
)

// Step is one step of a draft; a check lists its problems in the order of steps.
type Step string

const (
	StepName          Step = "name"
	StepRace          Step = "race"
	StepAttributes    Step = "attributes"
	StepBackground    Step = "background"
	StepCreationPerks Step = "creation_perks"
	StepTagSkills     Step = "tag_skills"
)

var steps = []Step{StepName, StepRace, StepAttributes, StepBackground, StepCreationPerks, StepTagSkills}

// Operation is what the command does with the input.
type Operation string

const (
	OperationRules  Operation = "rules"
	OperationCheck  Operation = "check"
	OperationFinish Operation = "finish"
)

// The ref categories and lineage roles creation reads (core:<category>/<key>).
const (
	CategoryRace       = "race"
	CategoryBackground = "background"
	CategoryFeature    = "feature"
	CategorySkill      = "skill"
	roleTemplate       = "template"
	roleSubrace        = "subrace"
	typeCreation       = "creation" // a feature's type: chosen at creation
)

// Draft is a character being made: any steps, in any order. Points are spent per attribute,
// negative when one is lowered under the start.
type Draft struct {
	Name          *string          `json:"name,omitempty"`
	Race          *string          `json:"race,omitempty"`
	Points        map[string]int64 `json:"points,omitempty"`
	Background    *string          `json:"background,omitempty"`
	CreationPerks []string         `json:"creation_perks,omitempty"`
	TagSkills     []string         `json:"tag_skills,omitempty"`
}

// Input is the `creation` command's input.
type Input struct {
	Operation Operation `json:"operation"`
	Draft     *Draft    `json:"draft,omitempty"`
}

// Problem is one thing a step still needs or gets wrong.
type Problem struct {
	Step    Step   `json:"step"`
	Message string `json:"message"`
}

// Check is what a check of a draft returns: whether it is complete, its problems in step
// order, the points left, and the sheet it gives (attributes after race and perks, starting
// skills, derived stats).
type Check struct {
	Complete   bool             `json:"complete"`
	Problems   []Problem        `json:"problems"`
	PointsLeft int64            `json:"points_left"`
	Attributes map[string]int64 `json:"attributes"`
	Skills     map[string]int64 `json:"skills"`
	Derived    map[string]int64 `json:"derived"`
}

// Actor is the stored actor a finish returns, with its empty lists written out.
type Actor struct {
	Attributes   map[string]int64        `json:"attributes"`
	Level        int64                   `json:"level"`
	XP           int64                   `json:"xp"`
	Skills       map[string]int64        `json:"skills"`
	Tags         []string                `json:"tags"`
	Growth       stored.Pools            `json:"growth"`
	Resources    stored.Resources        `json:"resources"`
	Conditions   []stored.Condition      `json:"conditions"`
	Modifiers    []stored.ActiveModifier `json:"modifiers"`
	Flags        []string                `json:"flags"`
	Perks        map[string]int64        `json:"perks"`
	Spells       []string                `json:"spells"`
	KnownEffects []string                `json:"known_effects"`
	Inventory    []stored.ItemInstance   `json:"inventory"`
	Equipped     map[string]int64        `json:"equipped"`
}

// Finished is a finish's result: the character's name and its stored actor.
type Finished struct {
	Name  string `json:"name"`
	Actor Actor  `json:"actor"`
}

// Run is the `creation` command.
func Run(ruleset *data.Ruleset, input Input) (any, error) {
	var draft Draft
	if input.Draft != nil {
		draft = *input.Draft
	}
	switch input.Operation {
	case OperationRules:
		return rules(ruleset), nil
	case OperationCheck:
		checked, _, err := check(ruleset, draft)
		return checked, err
	case OperationFinish:
		return finish(ruleset, draft)
	}
	return nil, fmt.Errorf("creation operation %q is not defined", input.Operation)
}

func keyOf(ref, category string) (string, bool) {
	return strings.CutPrefix(ref, "core:"+category+"/")
}

// raceParts is a race's ability modifiers summed with its parent's, and its trait refs,
// the parent's first.
func raceParts(ruleset *data.Ruleset, race *data.Race) (map[string]int64, []string) {
	var rows []*data.Race = []*data.Race{race}
	if race.Lineage.Role == roleSubrace {
		parentKey, _ := keyOf(race.Lineage.ParentRace, CategoryRace)
		if parent, ok := ruleset.Races.Get(parentKey); ok {
			rows = []*data.Race{parent, race}
		}
	}
	var modifiers map[string]int64 = map[string]int64{}
	var traits []string = []string{}
	for _, row := range rows {
		for ability, value := range row.AbilityModifiers {
			modifiers[ability] += value
		}
		for _, trait := range row.Traits {
			traits = append(traits, trait.ID)
		}
	}
	return modifiers, traits
}

type raceRule struct {
	Race             string           `json:"race"`
	Label            string           `json:"label"`
	AbilityModifiers map[string]int64 `json:"ability_modifiers"`
	Traits           []string         `json:"traits"`
}

type backgroundRule struct {
	Background   string `json:"background"`
	Label        string `json:"label"`
	SkillBonuses []struct {
		Skill string `json:"skill"`
		Bonus int64  `json:"bonus"`
	} `json:"skill_bonuses"`
	SuggestedTagSkills []string `json:"suggested_tag_skills"`
}

type perkRule struct {
	Perk        string `json:"perk"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type skillRule struct {
	Skill   string `json:"skill"`
	Label   string `json:"label"`
	Ability string `json:"ability"`
}

// Rules is what an app draws the creation form from.
type Rules struct {
	Creation struct {
		AttributeStart   int64 `json:"attribute_start"`
		AttributeMinimum int64 `json:"attribute_min"`
		AttributePoints  int64 `json:"attribute_points"`
		AttributeMaximum int64 `json:"attribute_max"`
		CreationPerksMax int64 `json:"creation_perks_max"`
		TagSkills        int64 `json:"tag_skills"`
	} `json:"creation"`
	Abilities     []string         `json:"abilities"`
	Races         []raceRule       `json:"races"`
	Backgrounds   []backgroundRule `json:"backgrounds"`
	CreationPerks []perkRule       `json:"creation_perks"`
	Skills        []skillRule      `json:"skills"`
}

func rules(ruleset *data.Ruleset) Rules {
	var result Rules = Rules{Abilities: []string{}, Races: []raceRule{}, Backgrounds: []backgroundRule{}, CreationPerks: []perkRule{}, Skills: []skillRule{}}
	var creation = ruleset.Progression.Creation
	result.Creation.AttributeStart, result.Creation.AttributeMinimum = creation.AttributeStart, creation.AttributeMinimum
	result.Creation.AttributePoints, result.Creation.AttributeMaximum = creation.AttributePoints, creation.AttributeMaximum
	result.Creation.CreationPerksMax, result.Creation.TagSkills = creation.CreationPerksMax, ruleset.Progression.TagSkills
	for index := actor.Ability(0); index < actor.AbilityCount; index++ {
		result.Abilities = append(result.Abilities, index.String())
	}
	for _, race := range ruleset.Races.Rows() {
		if race.Lineage.Role == roleTemplate {
			continue
		}
		modifiers, traits := raceParts(ruleset, &race)
		result.Races = append(result.Races, raceRule{Race: "core:" + CategoryRace + "/" + race.Key, Label: race.Label, AbilityModifiers: modifiers, Traits: traits})
	}
	for _, background := range ruleset.Backgrounds.Rows() {
		var row backgroundRule = backgroundRule{Background: "core:" + CategoryBackground + "/" + background.Key, Label: background.Label,
			SkillBonuses: background.SkillBonuses, SuggestedTagSkills: background.SuggestedTagSkills}
		if row.SkillBonuses == nil {
			row.SkillBonuses = []struct {
				Skill string `json:"skill"`
				Bonus int64  `json:"bonus"`
			}{}
		}
		if row.SuggestedTagSkills == nil {
			row.SuggestedTagSkills = []string{}
		}
		result.Backgrounds = append(result.Backgrounds, row)
	}
	for _, feature := range ruleset.Features.Rows() {
		if feature.Type == typeCreation {
			result.CreationPerks = append(result.CreationPerks, perkRule{Perk: "core:" + CategoryFeature + "/" + feature.Key, Label: feature.Label, Description: feature.Description})
		}
	}
	for _, skill := range ruleset.Skills.Rows() {
		result.Skills = append(result.Skills, skillRule{Skill: "core:" + CategorySkill + "/" + skill.Key, Label: skill.Label, Ability: skill.AssociatedAbility})
	}
	return result
}

// emitted is a feature's modifiers at rank 1, as the kernel's modifier type.
func emitted(feature *data.Feature) ([]modifier.Modifier, error) {
	encoded, err := json.Marshal(feature.EmitAt(1))
	if err != nil {
		return nil, err
	}
	var modifiers []modifier.Modifier
	err = json.Unmarshal(encoded, &modifiers)
	return modifiers, err
}

// chosen is the keys of a list of refs; each unknown or repeated ref is a problem.
func chosen(refs []string, category, kind string, step Step, known func(string) bool, problems *[]Problem) []string {
	var keys []string = []string{}
	for _, ref := range refs {
		key, ok := keyOf(ref, category)
		switch {
		case !ok || !known(key):
			*problems = append(*problems, Problem{step, fmt.Sprintf("%s %q is not defined", kind, ref)})
		case slices.Contains(keys, key):
			*problems = append(*problems, Problem{step, fmt.Sprintf("%s %q is chosen twice", kind, ref)})
		default:
			keys = append(keys, key)
		}
	}
	return keys
}

// what finish stores beside the check: the attributes before perks, the held features, the tags.
type storedParts struct {
	attributes map[string]int64
	held       []string
	tags       []string
}

func check(ruleset *data.Ruleset, draft Draft) (Check, storedParts, error) {
	var creation = ruleset.Progression.Creation
	var problems []Problem = []Problem{}
	var policies = ruleset.Origins
	var value = func(base int64, modifiers []modifier.Modifier, target string) (int64, error) {
		return modifier.Value(base, modifier.Target(modifiers, target), policies, 0, nil)
	}

	if draft.Name == nil || strings.TrimSpace(*draft.Name) == "" {
		problems = append(problems, Problem{StepName, "a name is needed"})
	}

	var raceModifiers map[string]int64 = map[string]int64{}
	var traits []string = []string{}
	if draft.Race == nil {
		problems = append(problems, Problem{StepRace, "a race is needed"})
	} else {
		key, _ := keyOf(*draft.Race, CategoryRace)
		race, ok := ruleset.Races.Get(key)
		if !ok {
			problems = append(problems, Problem{StepRace, fmt.Sprintf("race %q is not defined", *draft.Race)})
		} else {
			switch race.Lineage.Role {
			case roleTemplate:
				problems = append(problems, Problem{StepRace, fmt.Sprintf("race %q is a template and is never played", *draft.Race)})
			default:
				raceModifiers, traits = raceParts(ruleset, race)
			}
		}
	}

	var unknown []string
	for ability := range draft.Points {
		if _, ok := actor.ByKey(ability); !ok {
			unknown = append(unknown, ability)
		}
	}
	slices.Sort(unknown)
	for _, ability := range unknown {
		problems = append(problems, Problem{StepAttributes, fmt.Sprintf("attribute %q is not one of the eight", ability)})
	}
	var base map[string]int64 = map[string]int64{}
	var inRange bool = true
	var spent int64 = 0
	for index := actor.Ability(0); index < actor.AbilityCount; index++ {
		var key string = index.String()
		base[key] = creation.AttributeStart + draft.Points[key]
		spent += draft.Points[key]
		if base[key] > creation.AttributeMaximum {
			problems = append(problems, Problem{StepAttributes, fmt.Sprintf("%s is %d, over the maximum of %d", key, base[key], creation.AttributeMaximum)})
			inRange = false
		} else if base[key] < creation.AttributeMinimum {
			problems = append(problems, Problem{StepAttributes, fmt.Sprintf("%s is %d, under the minimum of %d", key, base[key], creation.AttributeMinimum)})
			inRange = false
		}
	}
	var left int64 = max(creation.AttributePoints-spent, 0)
	if inRange && spent > creation.AttributePoints {
		problems = append(problems, Problem{StepAttributes, fmt.Sprintf("%d points spent, more than the %d to distribute", spent, creation.AttributePoints)})
	} else if inRange && left > 0 {
		var plural string = "s"
		if left == 1 {
			plural = ""
		}
		problems = append(problems, Problem{StepAttributes, fmt.Sprintf("%d point%s left", left, plural)})
	}

	var background *data.Background
	if draft.Background == nil {
		problems = append(problems, Problem{StepBackground, "a background is needed"})
	} else {
		key, _ := keyOf(*draft.Background, CategoryBackground)
		var ok bool
		if background, ok = ruleset.Backgrounds.Get(key); !ok {
			problems = append(problems, Problem{StepBackground, fmt.Sprintf("background %q is not defined", *draft.Background)})
		}
	}

	var perks []string = chosen(draft.CreationPerks, CategoryFeature, CategoryFeature, StepCreationPerks,
		func(key string) bool { _, ok := ruleset.Features.Get(key); return ok }, &problems)
	perks = slices.DeleteFunc(perks, func(key string) bool {
		feature, _ := ruleset.Features.Get(key)
		if feature.Type != typeCreation {
			problems = append(problems, Problem{StepCreationPerks, fmt.Sprintf("feature %q is not a creation perk", "core:"+CategoryFeature+"/"+key)})
			return true
		}
		return false
	})

	// The modifiers the character starts with: its race traits and creation perks, at rank 1.
	var held []string
	var modifiers, traitModifiers []modifier.Modifier
	for index, ref := range append(slices.Clone(traits), perks...) {
		key := ref[strings.LastIndex(ref, "/")+1:]
		feature, err := ruleset.Features.Require("feature", key)
		if err != nil {
			return Check{}, storedParts{}, err
		}
		own, err := emitted(feature)
		if err != nil {
			return Check{}, storedParts{}, err
		}
		held = append(held, key)
		modifiers = append(modifiers, own...)
		if index < len(traits) {
			traitModifiers = append(traitModifiers, own...)
		}
	}
	mostPerks, err := value(creation.CreationPerksMax, traitModifiers, string(StepCreationPerks))
	if err != nil {
		return Check{}, storedParts{}, err
	}
	if int64(len(perks)) > mostPerks {
		problems = append(problems, Problem{StepCreationPerks, fmt.Sprintf("at most %d creation perks, %d chosen", mostPerks, len(perks))})
	}

	var tags []string = chosen(draft.TagSkills, CategorySkill, CategorySkill, StepTagSkills,
		func(key string) bool { _, ok := ruleset.Skills.Get(key); return ok }, &problems)
	needed, err := value(ruleset.Progression.TagSkills, modifiers, string(StepTagSkills))
	if err != nil {
		return Check{}, storedParts{}, err
	}
	if int64(len(draft.TagSkills)) != needed {
		problems = append(problems, Problem{StepTagSkills, fmt.Sprintf("%d tag skills are needed, %d chosen", needed, len(draft.TagSkills))})
	}

	var storedAttributes, attributes map[string]int64 = map[string]int64{}, map[string]int64{}
	var sheet actor.Attributes
	for index := actor.Ability(0); index < actor.AbilityCount; index++ {
		var key string = index.String()
		storedAttributes[key] = base[key] + raceModifiers[key]
		if attributes[key], err = value(storedAttributes[key], modifiers, key); err != nil {
			return Check{}, storedParts{}, err
		}
		sheet = sheet.With(index, attributes[key])
	}
	var bonuses map[string]int64 = map[string]int64{}
	if background != nil {
		for _, bonus := range background.SkillBonuses {
			bonuses[bonus.Skill[strings.LastIndex(bonus.Skill, "/")+1:]] = bonus.Bonus
		}
	}
	var skills map[string]int64 = map[string]int64{}
	for _, skill := range ruleset.Skills.Rows() {
		start, err := ruleset.Progression.SkillStart.Evaluate(&sheet, formula.Variable{Name: "attribute", Value: sheet.Get(skill.Ability)})
		if err != nil {
			return Check{}, storedParts{}, err
		}
		skills[skill.Key] = start + bonuses[skill.Key]
	}
	var level int64 = ruleset.Progression.StartLevel
	stats, err := derived.Run(ruleset, derived.Input{Actor: actor.Actor{Attributes: sheet, Level: &level}, Modifiers: modifiers})
	if err != nil {
		return Check{}, storedParts{}, err
	}
	slices.SortStableFunc(problems, func(first, second Problem) int {
		return slices.Index(steps, first.Step) - slices.Index(steps, second.Step)
	})
	if held == nil {
		held = []string{}
	}
	return Check{Complete: len(problems) == 0, Problems: problems, PointsLeft: left, Attributes: attributes, Skills: skills, Derived: stats},
		storedParts{attributes: storedAttributes, held: held, tags: tags}, nil
}

func finish(ruleset *data.Ruleset, draft Draft) (Finished, error) {
	checked, parts, err := check(ruleset, draft)
	if err != nil {
		return Finished{}, err
	}
	if !checked.Complete {
		var messages []string
		for _, problem := range checked.Problems {
			messages = append(messages, problem.Message)
		}
		return Finished{}, fmt.Errorf("the draft is not complete: %s", strings.Join(messages, "; "))
	}
	var perks map[string]int64 = map[string]int64{}
	for _, key := range parts.held {
		perks[key] = 1
	}
	var stats map[string]int64 = checked.Derived
	return Finished{Name: strings.TrimSpace(*draft.Name), Actor: Actor{
		Attributes: parts.attributes, Level: ruleset.Progression.StartLevel, XP: 0, Skills: checked.Skills, Tags: parts.tags,
		Resources:  stored.Resources{HitPoints: stats["hp"], ManaPoints: stats["mp"], Stamina: stats["sp"], ActionPoints: stats["ap"]},
		Conditions: []stored.Condition{}, Modifiers: []stored.ActiveModifier{}, Flags: []string{}, Perks: perks,
		Spells: []string{}, KnownEffects: []string{}, Inventory: []stored.ItemInstance{}, Equipped: map[string]int64{},
	}}, nil
}
