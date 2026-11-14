package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"sort"
	"strings"

	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/stored"
)

// The `print` command (docs/backend/print.md): a printable page from the data. It is an adapter:
// every number comes from another command's result.

type printAttack struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Weapon   string `json:"weapon"`
	Ability  string `json:"ability"`
	Delivery string `json:"delivery"`
	Phase    string `json:"phase"`
	Damage   struct {
		Bonus int64  `json:"bonus"`
		Type  string `json:"type"`
	} `json:"damage"`
}

type printMonster struct {
	Attributes map[string]int64   `json:"attributes"`
	Skills     map[string]int64   `json:"skills"`
	Level      *int64             `json:"level"`
	HitPoints  *int64             `json:"hitPoints"`
	Size       string             `json:"size"`
	Resist     map[string]float64 `json:"resist"`
	Attacks    []printAttack      `json:"attacks"`
}

type printInput struct {
	Kind      string        `json:"kind"`
	Actor     *stored.Actor `json:"actor"`
	Spell     string        `json:"spell"`
	Condition string        `json:"condition"`
	Label     string        `json:"label"`
	Monster   *printMonster `json:"monster"`
}

type printResult struct {
	Title string `json:"title"`
	HTML  string `json:"html"`
}

func init() { commands["print"] = bind(runPrint) }

// row is one line of a printed table; Attribute and Name make the number's data- attribute.
type row struct {
	Label, Attribute, Name string
	Value                  any
}

// Data is the number's data- attribute, built here because html/template refuses a dynamic
// attribute name with a hyphen in it (data-to-hit). Attribute is one of this file's fixed names,
// and Name is escaped.
func (printed row) Data() template.HTMLAttr {
	return template.HTMLAttr(fmt.Sprintf(`data-%s="%s"`, printed.Attribute, html.EscapeString(printed.Name)))
}

type pageSection struct {
	Heading string
	Rows    []row
	Notes   []string
}

func runPrint(ruleset *data.Ruleset, input printInput) (printResult, error) {
	var title string
	var sections []pageSection
	var err error
	switch input.Kind {
	case "sheet":
		title, sections, err = printSheet(ruleset, input.Actor)
	case "spell":
		title, sections, err = printSpell(ruleset, input.Spell)
	case "monster":
		title, sections, err = printMonsterBlock(ruleset, input.Label, input.Monster)
	case "condition":
		title, sections, err = printCondition(ruleset, input.Condition)
	default:
		return printResult{}, &data.NotDefinedError{Kind: "print kind", Key: input.Kind}
	}
	if err != nil {
		return printResult{}, err
	}
	var page bytes.Buffer
	if err := pageTemplate.Execute(&page, map[string]any{"Title": title, "Sections": sections}); err != nil {
		return printResult{}, err
	}
	return printResult{Title: title, HTML: page.String()}, nil
}

// run calls another command and returns its result as a map.
func run(ruleset *data.Ruleset, command string, input map[string]any) (map[string]any, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	output, err := commands[command](ruleset, body)
	if err != nil {
		return nil, err
	}
	result, _ := encodedValue(output).(map[string]any)
	return result, nil
}

func sortedKeys[V any](values map[string]V) []string {
	var keys []string = make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func attributeRows(attributes map[string]int64) []row {
	var rows []row
	for _, key := range []string{"str", "per", "end", "int", "wil", "agi", "cha", "lck"} {
		if value, has := attributes[key]; has {
			rows = append(rows, row{Label: strings.ToUpper(key), Attribute: "attribute", Name: key, Value: value})
		}
	}
	return rows
}

func printSheet(ruleset *data.Ruleset, actor *stored.Actor) (string, []pageSection, error) {
	if actor == nil {
		return "", nil, &data.MissingInputError{Subject: "print sheet", Needs: "actor"}
	}
	modifiers := actorModifiers(ruleset, actor)
	stats, err := run(ruleset, "derived", map[string]any{"actor": characterOf(actor), "modifiers": modifiers})
	if err != nil {
		return "", nil, err
	}
	var derived []row
	for _, stat := range ruleset.Derived.Rows() {
		derived = append(derived, row{Label: stat.Key, Attribute: "stat", Name: stat.Key, Value: stats[stat.Key]})
	}
	var skills []row
	for _, key := range sortedKeys(actor.Skills) {
		value, err := readSkillValue(ruleset, actor, key)
		if err != nil {
			return "", nil, err
		}
		target, err := run(ruleset, "modifiers", map[string]any{"value": value, "target": key, "modifiers": modifiers})
		if err != nil {
			return "", nil, err
		}
		skills = append(skills, row{Label: key, Attribute: "target", Name: key, Value: target["value"]})
	}
	var perks []row
	for _, key := range sortedKeys(actor.Perks) {
		perks = append(perks, row{Label: key, Attribute: "perk", Name: key, Value: actor.Perks[key]})
	}
	var spells []string = append([]string{}, actor.Spells...)
	sort.Strings(spells)
	return fmt.Sprintf("Character, level %d", actor.Level), []pageSection{
		{Heading: "Attributes", Rows: attributeRows(actor.Attributes)},
		{Heading: "Derived", Rows: derived},
		{Heading: "Skills: roll d100 at or under the target", Rows: skills},
		{Heading: "Perks (rank)", Rows: perks},
		{Heading: "Spells known", Notes: spells},
	}, nil
}

func printSpell(ruleset *data.Ruleset, key string) (string, []pageSection, error) {
	spell, err := ruleset.Spells.Require("spell", key)
	if err != nil {
		return "", nil, err
	}
	rows := []row{
		{Label: "skill", Value: spell.Color + "_magic"},
		{Label: "level", Attribute: "stat", Name: "level", Value: spell.Level},
		{Label: "mana", Attribute: "stat", Name: "mana", Value: spell.Cost.Mana},
	}
	if len(spell.Cost.Reagents) > 0 {
		rows = append(rows, row{Label: "reagents", Value: strings.Join(spell.Cost.Reagents, ", ")})
	}
	if spell.Delivery != "" {
		rows = append(rows, row{Label: "delivery", Value: spell.Delivery})
	}
	if spell.Concentration {
		rows = append(rows, row{Label: "concentration", Value: "yes"})
	}
	if spell.Save.Attribute != "" || spell.Save.Skill != "" {
		rows = append(rows, row{Label: "save", Value: strings.Trim(spell.Save.Attribute+" "+spell.Save.Skill, " ")})
	}
	var effects []string
	for _, applied := range spell.Effects {
		encoded, _ := json.Marshal(applied)
		effects = append(effects, string(encoded))
	}
	return spell.Label, []pageSection{{Heading: spell.Description, Rows: rows}, {Heading: "Effects", Notes: effects}}, nil
}

func printMonsterBlock(ruleset *data.Ruleset, label string, monster *printMonster) (string, []pageSection, error) {
	if monster == nil {
		return "", nil, &data.MissingInputError{Subject: "print monster", Needs: "monster"}
	}
	// The skill each attack rolls with: unarmed for a natural weapon, else the weapon's own.
	var attackSkills []string
	read := []any{"evasion"}
	for _, attack := range monster.Attacks {
		var skill string = "unarmed"
		if !strings.HasPrefix(attack.Weapon, "core:natural_weapon/") {
			item, known := ruleset.Items[attack.Weapon]
			if !known || item.WeaponSkill == "" {
				return "", nil, &data.NotDefinedError{Kind: "weapon", Key: attack.Weapon}
			}
			skill = item.WeaponSkill
		}
		attackSkills = append(attackSkills, skill)
		read = append(read, skill)
	}
	sheet, err := run(ruleset, "monster_sheet", map[string]any{"monster": map[string]any{
		"attributes": monster.Attributes, "skills": monster.Skills, "level": monster.Level, "hitPoints": monster.HitPoints,
	}, "read_skills": read})
	if err != nil {
		return "", nil, err
	}
	skills, _ := sheet["skills"].(map[string]any)
	stats := []row{
		{Label: "HP", Attribute: "stat", Name: "hp", Value: sheet["hp"]},
		{Label: "MP", Attribute: "stat", Name: "mp", Value: sheet["mp"]},
		{Label: "AP", Attribute: "stat", Name: "ap", Value: sheet["ap"]},
		{Label: "evasion", Attribute: "stat", Name: "evasion", Value: skills["evasion"]},
	}
	var attacks []row
	for index, attack := range monster.Attacks {
		attacks = append(attacks, row{
			Label:     fmt.Sprintf("%s (%s, %s; to-hit with %s)", attack.Label, attack.Weapon, cmpOr(attack.Damage.Type, "its weapon's type"), attackSkills[index]),
			Attribute: "to-hit", Name: attack.Key, Value: skills[attackSkills[index]],
		})
	}
	var resist []row
	for _, kind := range sortedKeys(monster.Resist) {
		resist = append(resist, row{Label: kind, Attribute: "resist", Name: kind, Value: monster.Resist[kind]})
	}
	var level int64 = 1
	if monster.Level != nil {
		level = *monster.Level
	}
	return fmt.Sprintf("%s, level %d", cmpOr(label, "Monster"), level), []pageSection{
		{Heading: "Attributes", Rows: attributeRows(monster.Attributes)},
		{Heading: "Defense", Rows: stats},
		{Heading: "Attacks: to-hit is the target on d100", Rows: attacks},
		{Heading: "Resistances", Rows: resist},
	}, nil
}

func printCondition(ruleset *data.Ruleset, key string) (string, []pageSection, error) {
	condition, err := ruleset.Conditions.Table.Require("condition", key)
	if err != nil {
		return "", nil, err
	}
	rows := []row{
		{Label: "applied by", Value: strings.Join(condition.AppliedBy, ", ")},
		{Label: "removed by", Value: strings.Join(condition.RemovedBy, ", ")},
	}
	if condition.Stacking {
		rows = append(rows, row{Label: "stacks", Attribute: "stat", Name: "max_stacks", Value: condition.MaximumStacks})
	}
	var modifiers []string
	for _, held := range condition.Modifiers {
		encoded, _ := json.Marshal(held)
		modifiers = append(modifiers, string(encoded))
	}
	return condition.Label, []pageSection{{Heading: condition.Description, Rows: rows}, {Heading: "Modifiers", Notes: modifiers}}, nil
}

var pageTemplate = template.Must(template.New("page").Parse(`<!doctype html>
<html lang="en"><head><meta charset="utf-8"><title>{{.Title}}</title>
<style>
@page { size: A4; margin: 15mm; }
body { font: 11pt/1.4 system-ui, sans-serif; color: #212529; margin: 0 auto; max-width: 180mm; }
h1 { font-size: 16pt; margin: 0 0 4mm; border-bottom: 2px solid #212529; }
h2 { font-size: 11pt; margin: 5mm 0 1mm; }
table { border-collapse: collapse; width: 100%; }
td, th { border-bottom: 1px solid #dee2e6; padding: 1mm 2mm; text-align: left; }
td.number { text-align: right; font-weight: 600; width: 25mm; }
ul { margin: 0; padding-left: 5mm; }
section { break-inside: avoid; }
</style></head>
<body><h1>{{.Title}}</h1>
{{range .Sections}}<section><h2>{{.Heading}}</h2>
{{if .Rows}}<table>{{range .Rows}}<tr><th>{{.Label}}</th>{{if .Attribute}}<td class="number" {{.Data}}>{{.Value}}</td>{{else}}<td>{{.Value}}</td>{{end}}</tr>{{end}}</table>{{end}}
{{if .Notes}}<ul>{{range .Notes}}<li>{{.}}</li>{{end}}</ul>{{end}}
</section>
{{end}}</body></html>
`))
