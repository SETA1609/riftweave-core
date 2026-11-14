package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"

	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/stored"
)

// SessionStep is one command in a session's log, with its input as it was given.
type SessionStep struct {
	Command string          `json:"command"`
	Input   json.RawMessage `json:"input"`
	As      string          `json:"as,omitempty"`
}

// SessionRuleset is what the run started on.
type SessionRuleset struct {
	Modules []string `json:"modules"`
}

// Session is one run (docs/core/state.md): a starting actor, a log, and the current snapshot.
type Session struct {
	Ruleset SessionRuleset `json:"ruleset"`
	Start   stored.Actor   `json:"start"`
	Log     []SessionStep  `json:"log"`
	Actor   stored.Actor   `json:"actor"`
	Dev     bool           `json:"dev,omitempty"` // a dev operation ran (docs/backend/dev-mode.md)
}

// sessionOperation is what a `session` call does. It decodes from its name, so an unknown
// operation is bad input, caught before anything runs.
type sessionOperation uint8

const (
	operationStart sessionOperation = iota
	operationStep
	operationReplay
)

var sessionOperationNames = [...]string{"start", "step", "replay"}

// UnmarshalText reads an operation name.
func (operation *sessionOperation) UnmarshalText(encoded []byte) error {
	for index, name := range sessionOperationNames {
		if name == string(encoded) {
			*operation = sessionOperation(index)
			return nil
		}
	}
	return fmt.Errorf("session operation %q is not defined", encoded)
}

// MarshalText writes the operation name.
func (operation sessionOperation) MarshalText() ([]byte, error) {
	return []byte(sessionOperationNames[operation]), nil
}

type sessionInput struct {
	Operation sessionOperation `json:"operation"`
	Start     *stored.Actor    `json:"start"`
	Session   *Session         `json:"session"`
	Step      *SessionStep     `json:"step"`
}

type sessionResult struct {
	Session Session         `json:"session"`
	Result  json.RawMessage `json:"result,omitempty"`
	Matches *bool           `json:"matches,omitempty"`
}

func init() { commands["session"] = bind(runSession) }

// runSession is the `session` command (docs/backend/sessions.md). It holds no rule.
func runSession(ruleset *data.Ruleset, input sessionInput) (sessionResult, error) {
	switch input.Operation {
	case operationStart:
		if input.Start == nil {
			return sessionResult{}, errors.New("session start needs a start actor")
		}
		var modules []string = slices.Clone(ruleset.Modules)
		if modules == nil {
			modules = []string{}
		}
		sort.Strings(modules)
		actor, err := copyActor(*input.Start)
		if err != nil {
			return sessionResult{}, err
		}
		return sessionResult{Session: Session{Ruleset: SessionRuleset{Modules: modules}, Start: *input.Start, Log: []SessionStep{}, Actor: actor}}, nil
	case operationStep:
		if input.Session == nil || input.Step == nil {
			return sessionResult{}, errors.New("session step needs a session and a step")
		}
		var session Session = *input.Session
		actor, output, err := applyStep(ruleset, session.Actor, *input.Step)
		if err != nil {
			return sessionResult{}, err
		}
		session.Actor = actor
		session.Log = append(slices.Clone(session.Log), *input.Step)
		session.Dev = session.Dev || input.Step.Command == "dev"
		return sessionResult{Session: session, Result: output}, nil
	case operationReplay:
		if input.Session == nil {
			return sessionResult{}, errors.New("session replay needs a session")
		}
		var session Session = *input.Session
		actor, err := copyActor(session.Start)
		if err != nil {
			return sessionResult{}, err
		}
		for _, step := range session.Log {
			if actor, _, err = applyStep(ruleset, actor, step); err != nil {
				return sessionResult{}, err
			}
		}
		saved, rebuilt := encodedValue(session.Actor), encodedValue(actor)
		var matches bool = reflect.DeepEqual(saved, rebuilt)
		session.Actor = actor
		return sessionResult{Session: session, Matches: &matches}, nil
	}
	return sessionResult{}, fmt.Errorf("session operation %d is not defined", input.Operation)
}

func encodedValue(value any) any {
	body, _ := json.Marshal(value)
	var decoded any
	_ = json.Unmarshal(body, &decoded)
	return decoded
}

func copyActor(actor stored.Actor) (stored.Actor, error) {
	body, err := json.Marshal(actor)
	if err != nil {
		return stored.Actor{}, err
	}
	var copied stored.Actor
	err = json.Unmarshal(body, &copied)
	return copied, err
}

// section returns the object at key, creating it when missing.
func section(object map[string]any, key string) map[string]any {
	if inner, ok := object[key].(map[string]any); ok {
		return inner
	}
	inner := map[string]any{}
	object[key] = inner
	return inner
}

func setDefault(object map[string]any, key string, value any) {
	if _, present := object[key]; !present {
		object[key] = value
	}
}

func derivedStats(ruleset *data.Ruleset, actor *stored.Actor) (map[string]any, error) {
	body, err := json.Marshal(map[string]any{"actor": characterOf(actor)})
	if err != nil {
		return nil, err
	}
	output, err := commands["derived"](ruleset, body)
	if err != nil {
		return nil, err
	}
	stats, _ := encodedValue(output).(map[string]any)
	return stats, nil
}

// characterOf is the part of a stored actor the `actor` input of a command reads.
func characterOf(actor *stored.Actor) map[string]any {
	return map[string]any{"attributes": actor.Attributes, "level": actor.Level, "growth": actor.Growth, "skills": actor.Skills}
}

// sessionAdapter is how a session reads one command: fill adds what the actor knows to the
// fields a step did not name, and write changes the actor by the command's result. A command
// with no adapter runs and changes nothing.
type sessionAdapter struct {
	fill  func(ruleset *data.Ruleset, actor *stored.Actor, step *SessionStep, input map[string]any) error
	write func(actor *stored.Actor, step *SessionStep, input, result map[string]any)
}

// sessionAdapters is the one table of what each command reads from and writes to the actor
// (docs/backend/sessions.md). Adding a command to sessions is one entry here.
var sessionAdapters = map[string]sessionAdapter{
	"derived":     {fill: fillDerived},
	"combat":      {fill: fillCombat, write: writeCombat},
	"rest":        {fill: fillRest, write: writeRest},
	"casting":     {fill: fillCasting, write: writeCasting},
	"progression": {fill: fillProgression, write: writeProgression},
	"consume":     {fill: fillConsume, write: writeConsume},
	"conditions":  {fill: fillConditions, write: writeConditions},
	"spellcraft":  {fill: fillSpellcraft, write: writeSpellcraft},
}

func fillDerived(ruleset *data.Ruleset, actor *stored.Actor, _ *SessionStep, input map[string]any) error {
	setDefault(input, "actor", characterOf(actor))
	setDefault(input, "modifiers", actorModifiers(ruleset, actor))
	return nil
}

// actorModifiers are the actor's perks at their rank, then its held modifiers
// (docs/backend/sessions.md#what-the-actor-brings-to-a-step-app-13). A perk at rank r emits an
// add at value × r and a multiply of m at 1 + (m − 1) × r; a grant or a set once.
func actorModifiers(ruleset *data.Ruleset, actor *stored.Actor) []any {
	var keys []string = make([]string, 0, len(actor.Perks))
	for key := range actor.Perks {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var modifiers []any = []any{}
	for _, key := range keys {
		feature, exists := ruleset.Features.Get(key)
		if !exists {
			continue
		}
		var rank float64 = float64(actor.Perks[key])
		for _, written := range feature.Modifiers {
			var emitted map[string]any = make(map[string]any, len(written))
			for field, value := range written {
				emitted[field] = value
			}
			if value, isNumber := written["value"].(float64); isNumber {
				switch written["operation"] {
				case "add":
					emitted["value"] = value * rank
				case "multiply":
					emitted["value"] = 1 + (value-1)*rank
				}
			}
			modifiers = append(modifiers, emitted)
		}
	}
	for _, held := range actor.Modifiers {
		modifiers = append(modifiers, encodedValue(held.Modifier))
	}
	return modifiers
}

// readSkillValue reads one of the actor's skills as `progression read_skill` does.
func readSkillValue(ruleset *data.Ruleset, actor *stored.Actor, key string) (any, error) {
	body, err := json.Marshal(map[string]any{"operation": "read_skill", "actor": characterOf(actor), "read": key})
	if err != nil {
		return nil, err
	}
	output, err := commands["progression"](ruleset, body)
	if err != nil {
		return nil, err
	}
	read, _ := encodedValue(output).(map[string]any)
	return read["value"], nil
}

// equippedItems are the actor's equipped inventory items, by slot name.
func equippedItems(actor *stored.Actor) []stored.ItemInstance {
	var slots []string = make([]string, 0, len(actor.Equipped))
	for slot := range actor.Equipped {
		slots = append(slots, slot)
	}
	sort.Strings(slots)
	var items []stored.ItemInstance
	for _, slot := range slots {
		for _, item := range actor.Inventory {
			if item.ID == actor.Equipped[slot] {
				items = append(items, item)
			}
		}
	}
	return items
}

// attackSkill is the skill of the first equipped weapon, by slot name, else unarmed.
func attackSkill(ruleset *data.Ruleset, actor *stored.Actor) string {
	for _, item := range equippedItems(actor) {
		if row, known := ruleset.Items[item.Base]; known && row.IsWeapon && row.WeaponSkill != "" {
			return row.WeaponSkill
		}
	}
	return "unarmed"
}

// armorDR is the sum over equipped armor of drBase plus the material's armor defense; a piece
// with no material takes the first of its default materials (docs/modules/armor.md).
func armorDR(ruleset *data.Ruleset, actor *stored.Actor) int64 {
	var total int64
	for _, item := range equippedItems(actor) {
		row, known := ruleset.Items[item.Base]
		if !known || row.ArmorDRBase == nil {
			continue
		}
		var material string = item.Material
		if material == "" && len(row.DefaultMaterials) > 0 {
			material = row.DefaultMaterials[0]
		}
		total += *row.ArmorDRBase + ruleset.MaterialDefense[material]
	}
	return total
}

func fillCombat(ruleset *data.Ruleset, actor *stored.Actor, step *SessionStep, input map[string]any) error {
	if step.As == "defender" {
		defender := section(input, "defender")
		setDefault(defender, "hp", actor.Resources.HitPoints)
		setDefault(defender, "attributes", actor.Attributes)
		setDefault(defender, "armor_dr", armorDR(ruleset, actor))
		setDefault(defender, "modifiers", actorModifiers(ruleset, actor))
		return nil
	}
	attacker := section(input, "attacker")
	setDefault(attacker, "attributes", actor.Attributes)
	var skill string = attackSkill(ruleset, actor)
	if _, named := attacker["skill"]; !named {
		value, err := readSkillValue(ruleset, actor, skill)
		if err != nil {
			return err
		}
		attacker["skill"] = value
	}
	// The attacker's modifiers all count on the attack, so only those on its skill go in.
	if _, named := input["modifiers"]; !named {
		var onSkill []any = []any{}
		for _, raw := range actorModifiers(ruleset, actor) {
			if modifier, _ := raw.(map[string]any); modifier["target"] == skill {
				onSkill = append(onSkill, modifier)
			}
		}
		input["modifiers"] = onSkill
	}
	return nil
}

func writeCombat(actor *stored.Actor, step *SessionStep, _, result map[string]any) {
	if step.As == "defender" {
		actor.Resources.HitPoints = number(result["hp"])
	}
}

func fillRest(ruleset *data.Ruleset, actor *stored.Actor, _ *SessionStep, input map[string]any) error {
	setDefault(input, "actor", characterOf(actor))
	setDefault(input, "current", map[string]any{
		"hp": actor.Resources.HitPoints, "mp": actor.Resources.ManaPoints,
		"stamina": actor.Resources.Stamina, "ap": actor.Resources.ActionPoints,
	})
	if _, present := section(input, "max")["stamina"]; !present {
		stats, err := derivedStats(ruleset, actor)
		if err != nil {
			return err
		}
		section(input, "max")["stamina"] = stats["sp"]
	}
	return nil
}

func writeRest(actor *stored.Actor, _ *SessionStep, input, result map[string]any) {
	actor.Resources.HitPoints, actor.Resources.ManaPoints = number(result["hp"]), number(result["mp"])
	actor.Resources.Stamina, actor.Resources.ActionPoints = number(result["stamina"]), number(result["ap"])
	kept, _ := result["conditions"].([]any)
	named, _ := input["conditions"].([]any)
	var cleared []string
	for _, entry := range named {
		key, _ := entry.(map[string]any)["key"].(string)
		if !slices.Contains(kept, any(key)) {
			cleared = append(cleared, key)
		}
	}
	actor.Conditions = slices.DeleteFunc(actor.Conditions, func(condition stored.Condition) bool {
		return slices.Contains(cleared, condition.Key)
	})
	for _, key := range cleared {
		actor.Flags = slices.DeleteFunc(actor.Flags, func(flag string) bool { return flag == key })
	}
}

func fillCasting(ruleset *data.Ruleset, actor *stored.Actor, _ *SessionStep, input map[string]any) error {
	setDefault(input, "modifiers", actorModifiers(ruleset, actor))
	caster := section(input, "caster")
	setDefault(caster, "skills", actor.Skills)
	setDefault(caster, "attributes", actor.Attributes)
	setDefault(caster, "mana", actor.Resources.ManaPoints)
	return nil
}

func writeCasting(actor *stored.Actor, _ *SessionStep, _, result map[string]any) {
	actor.Resources.ManaPoints = number(result["mana"])
}

func fillProgression(ruleset *data.Ruleset, actor *stored.Actor, _ *SessionStep, input map[string]any) error {
	switch input["operation"] {
	case "level_up":
		setDefault(input, "actor", map[string]any{"attributes": actor.Attributes})
	case "spend":
		key, named := input["key"].(string)
		if !named {
			return nil
		}
		if _, present := input["skill"]; !present {
			value, err := readSkillValue(ruleset, actor, key)
			if err != nil {
				return err
			}
			input["skill"] = value
		}
		setDefault(input, "tagged", slices.Contains(actor.Tags, key))
	case "take_perk":
		setDefault(input, "actor", characterOf(actor))
		perks := map[string]any{}
		for key, rank := range actor.Perks {
			perks[key] = rank
		}
		setDefault(input, "perks", perks)
	}
	return nil
}

func writeProgression(actor *stored.Actor, _ *SessionStep, input, result map[string]any) {
	switch input["operation"] {
	case "level":
		actor.XP, actor.Level = number(input["xp"]), number(result["level"])
	case "level_up":
		actor.Growth.HitPoints += number(input["hp_roll"])
		actor.Growth.ManaPoints += number(input["mp_roll"])
		actor.Growth.Stamina += number(input["sp_roll"])
	case "spend":
		key, named := input["key"].(string)
		if !named || number(result["ranks"]) == 0 {
			return
		}
		if actor.Skills == nil {
			actor.Skills = map[string]int64{}
		}
		actor.Skills[key] = number(input["skill"]) + number(result["ranks"])
	case "take_perk":
		if result["taken"] != true {
			return
		}
		if actor.Perks == nil {
			actor.Perks = map[string]int64{}
		}
		actor.Perks[result["perk"].(string)] = number(result["rank"])
	}
}

// effectRefPrefix is how a stored actor's known effect key reads as a ref.
const effectRefPrefix = "core:effect/"

func fillSpellcraft(_ *data.Ruleset, actor *stored.Actor, _ *SessionStep, input map[string]any) error {
	if input["operation"] != "learn" {
		return nil
	}
	known := make([]any, 0, len(actor.KnownEffects))
	for _, key := range actor.KnownEffects {
		known = append(known, effectRefPrefix+key)
	}
	setDefault(section(input, "caster"), "known", known)
	return nil
}

func writeSpellcraft(actor *stored.Actor, _ *SessionStep, input, result map[string]any) {
	if input["operation"] != "learn" {
		return
	}
	spell, _ := input["spell"].(string)
	spell = spell[strings.LastIndex(spell, "/")+1:]
	if !slices.Contains(actor.Spells, spell) {
		actor.Spells = append(actor.Spells, spell)
	}
	known, _ := result["known"].([]any)
	actor.KnownEffects = make([]string, 0, len(known))
	for _, ref := range known {
		text, _ := ref.(string)
		actor.KnownEffects = append(actor.KnownEffects, text[strings.LastIndex(text, "/")+1:])
	}
}

func fillConsume(ruleset *data.Ruleset, actor *stored.Actor, _ *SessionStep, input map[string]any) error {
	stats, err := derivedStats(ruleset, actor)
	if err != nil {
		return err
	}
	setDefault(input, "current", map[string]any{"hp": actor.Resources.HitPoints, "mp": actor.Resources.ManaPoints, "stamina": actor.Resources.Stamina})
	setDefault(input, "max", map[string]any{"hp": stats["hp"], "mp": stats["mp"], "stamina": stats["sp"]})
	return nil
}

func writeConsume(actor *stored.Actor, _ *SessionStep, _, result map[string]any) {
	actor.Resources.HitPoints, actor.Resources.ManaPoints = number(result["hp"]), number(result["mp"])
	actor.Resources.Stamina = number(result["stamina"])
}

func fillConditions(_ *data.Ruleset, actor *stored.Actor, _ *SessionStep, input map[string]any) error {
	if _, present := input["conditions"]; !present {
		held := map[string]int64{}
		for _, condition := range actor.Conditions {
			held[condition.Key] = condition.Stacks
		}
		input["conditions"] = held
	}
	return nil
}

func writeConditions(actor *stored.Actor, _ *SessionStep, _, result map[string]any) {
	held, _ := result["conditions"].(map[string]any)
	keys := make([]string, 0, len(held))
	for key := range held {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	actor.Conditions = []stored.Condition{}
	for _, key := range keys {
		actor.Conditions = append(actor.Conditions, stored.Condition{Key: key, Stacks: number(held[key])})
	}
	flags, _ := result["flags"].([]any)
	actor.Flags = []string{}
	for _, flag := range flags {
		actor.Flags = append(actor.Flags, flag.(string))
	}
}

func number(value any) int64 {
	amount, _ := value.(float64)
	return int64(amount)
}

// applyStep runs one step on a copy of the actor, so a step that fails changes nothing.
func applyStep(ruleset *data.Ruleset, current stored.Actor, step SessionStep) (stored.Actor, json.RawMessage, error) {
	if step.Command == "dev" {
		actor, err := copyActor(current)
		if err != nil {
			return stored.Actor{}, nil, err
		}
		if err := applyDev(ruleset, &actor, step.Input); err != nil {
			return stored.Actor{}, nil, err
		}
		return actor, json.RawMessage(`{}`), nil
	}
	handler, known := commands[step.Command]
	if !known || step.Command == "session" {
		return stored.Actor{}, nil, fmt.Errorf("session step: unknown command %q", step.Command)
	}
	actor, err := copyActor(current)
	if err != nil {
		return stored.Actor{}, nil, err
	}
	var input map[string]any
	if err := json.Unmarshal(step.Input, &input); err != nil {
		return stored.Actor{}, nil, err
	}
	var adapter sessionAdapter = sessionAdapters[step.Command]
	if adapter.fill != nil {
		if err := adapter.fill(ruleset, &actor, &step, input); err != nil {
			return stored.Actor{}, nil, err
		}
	}
	body, err := json.Marshal(input)
	if err != nil {
		return stored.Actor{}, nil, err
	}
	output, err := handler(ruleset, body)
	if err != nil {
		return stored.Actor{}, nil, err
	}
	encoded, err := json.Marshal(output)
	if err != nil {
		return stored.Actor{}, nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(encoded, &result); err != nil {
		return stored.Actor{}, nil, err
	}
	if adapter.write != nil {
		adapter.write(&actor, &step, input, result)
	}
	return actor, encoded, nil
}
