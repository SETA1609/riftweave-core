package core

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/SETA1609/riftweave-core/backend/internal/stored"
)

// docs/backend/print.md: every printed number is the number another command computes.

func printPage(t *testing.T, opened *Core, input map[string]any) string {
	t.Helper()
	body, _ := json.Marshal(input)
	out, err := opened.Run("print", body)
	if err != nil {
		t.Fatalf("print: %v: %s", err, out)
	}
	var page struct {
		Title string `json:"title"`
		HTML  string `json:"html"`
	}
	if err := json.Unmarshal(out, &page); err != nil || page.Title == "" || !strings.Contains(page.HTML, "<html") {
		t.Fatalf("print: not a page: %s", out)
	}
	return page.HTML
}

// printed reads the number a data- attribute names, as in <td data-target="blades">57</td>.
func printed(t *testing.T, html, attribute, name string) int64 {
	t.Helper()
	match := regexp.MustCompile(`data-` + attribute + `="` + regexp.QuoteMeta(name) + `"[^>]*>\s*(-?\d+)`).FindStringSubmatch(html)
	if match == nil {
		t.Fatalf("no data-%s=%q on the page", attribute, name)
	}
	value, _ := strconv.ParseInt(match[1], 10, 64)
	return value
}

func runJSON(t *testing.T, opened *Core, command string, input map[string]any) map[string]any {
	t.Helper()
	body, _ := json.Marshal(input)
	out, err := opened.Run(command, body)
	if err != nil {
		t.Fatalf("%s: %v: %s", command, err, out)
	}
	var result map[string]any
	_ = json.Unmarshal(out, &result)
	return result
}

func TestASheetPrintsEachRaisedSkillsTargetAsTheBackendComputesIt(t *testing.T) {
	opened := open(t)
	actor := startingActor()
	actor["skills"] = map[string]any{"blades": 40, "red_magic": 50}
	actor["perks"] = map[string]any{"hardened_hide": 2}
	html := printPage(t, opened, map[string]any{"kind": "sheet", "actor": actor})

	// A target is the skill's value with the actor's modifiers on it; no perk here names blades.
	if got := printed(t, html, "target", "blades"); got != 40 {
		t.Errorf("blades target %d, want 40", got)
	}
	if got := printed(t, html, "target", "red_magic"); got != 50 {
		t.Errorf("red_magic target %d, want 50", got)
	}
	if strings.Contains(html, `data-target="stealth"`) {
		t.Error("an unraised skill is not printed")
	}
	// Derived stats with the actor's modifiers: Hardened Hide at rank 2 is +30 HP.
	derived := runJSON(t, opened, "derived", map[string]any{"actor": characterOf(mustStored(t, actor)),
		"modifiers": []any{map[string]any{"target": "hp", "operation": "add", "value": 30, "source": "feature:hardened_hide", "origin": "core:origin/trained"}}})
	for _, stat := range []string{"hp", "mp", "sp", "ap"} {
		if got := printed(t, html, "stat", stat); float64(got) != derived[stat].(float64) {
			t.Errorf("%s printed %d, derived %v", stat, got, derived[stat])
		}
	}
	if got := printed(t, html, "perk", "hardened_hide"); got != 2 {
		t.Errorf("hardened_hide rank %d, want 2", got)
	}
}

func TestAMonsterStatBlockPrintsItsToHitFromTheMonsterSheet(t *testing.T) {
	opened := open(t)
	wolf := map[string]any{
		"attributes": map[string]any{"str": 6, "agi": 5, "per": 6, "end": 4, "wil": 2}, "level": 5,
		"attacks": []any{map[string]any{"key": "bite", "label": "Bite", "weapon": "core:natural_weapon/bite"}},
		"resist":  map[string]any{"cold": 5},
	}
	html := printPage(t, opened, map[string]any{"kind": "monster", "label": "Wolf", "monster": wolf})
	sheet := runJSON(t, opened, "monster_sheet", map[string]any{"monster": map[string]any{
		"attributes": wolf["attributes"], "level": 5}, "read_skills": []any{"unarmed", "evasion"}})
	skills := sheet["skills"].(map[string]any)
	if got := printed(t, html, "to-hit", "bite"); float64(got) != skills["unarmed"].(float64) {
		t.Errorf("bite to-hit %d, monster_sheet unarmed %v", got, skills["unarmed"])
	}
	if got := printed(t, html, "stat", "evasion"); float64(got) != skills["evasion"].(float64) {
		t.Errorf("evasion %d, monster_sheet %v", got, skills["evasion"])
	}
	// HP is computed (DE_S02_03): floor((15 + 8 x END 4 + 4 x level 5) x 0.4) for a standard monster.
	if got := printed(t, html, "stat", "hp"); got != 26 || float64(got) != sheet["hp"].(float64) {
		t.Errorf("hp %d, want the computed 26 the monster_sheet gives (%v)", got, sheet["hp"])
	}
	if got := printed(t, html, "resist", "cold"); got != 5 {
		t.Errorf("cold resist %d, want 5", got)
	}
}

// DE_S02_02: an attack names its weapon, and its label is the weapon row's unless it writes its own.
func TestAMonsterAttackWithoutALabelPrintsItsWeaponsLabel(t *testing.T) {
	opened := open(t)
	monster := map[string]any{
		"attributes": map[string]any{"str": 5, "agi": 5}, "level": 1,
		"attacks": []any{
			map[string]any{"key": "bite", "weapon": "core:natural_weapon/bite"},
			map[string]any{"key": "cut", "weapon": "core:weapon/shortsword"},
			map[string]any{"key": "gnaw", "label": "Gnaw", "weapon": "core:natural_weapon/bite"},
		},
	}
	html := printPage(t, opened, map[string]any{"kind": "monster", "label": "Bandit", "monster": monster})
	for _, want := range []string{"Bite (core:natural_weapon/bite", "Shortsword (core:weapon/shortsword", "Gnaw (core:natural_weapon/bite"} {
		if !strings.Contains(html, want) {
			t.Errorf("the page does not print %q", want)
		}
	}
}

func TestASpellCardAndAConditionCardPrintFromTheData(t *testing.T) {
	opened := open(t)
	spell := printPage(t, opened, map[string]any{"kind": "spell", "spell": "fire_bolt"})
	if !strings.Contains(spell, "Fire Bolt") || !strings.Contains(spell, "red_magic") {
		t.Errorf("the spell card names the spell and its skill: %.300s", spell)
	}
	if got := printed(t, spell, "stat", "mana"); got != 6 {
		t.Errorf("fire_bolt mana %d, want 6", got)
	}
	condition := printPage(t, opened, map[string]any{"kind": "condition", "condition": "blinded"})
	if !strings.Contains(condition, "Blinded") || !strings.Contains(condition, "core:effect/blind") {
		t.Errorf("the condition card names it and what applies it: %.300s", condition)
	}
}

func TestPrintRefusesAnUnknownKey(t *testing.T) {
	body, _ := json.Marshal(map[string]any{"kind": "spell", "spell": "nope"})
	if _, err := open(t).Run("print", body); err == nil {
		t.Error("an unknown spell is refused")
	}
}

func mustStored(t *testing.T, raw map[string]any) *stored.Actor {
	t.Helper()
	body, _ := json.Marshal(raw)
	var actor stored.Actor
	if err := json.Unmarshal(body, &actor); err != nil {
		t.Fatal(err)
	}
	return &actor
}

func TestThePageInlinesItsStylesheetUnescaped(t *testing.T) {
	page := printPage(t, open(t), map[string]any{"kind": "condition", "condition": "blinded"})
	if !strings.Contains(page, "<style>\n@page { size: A4; margin: 15mm; }") || strings.Contains(page, "ZgotmplZ") {
		t.Errorf("templates/print.css is not in the page's <style> as written: %.400s", page)
	}
}

// woundedActor is mid-campaign (BE_S09_03): HP and mana spent, frightened, a held -10 on blades,
// and a shortsword worn.
func woundedActor() map[string]any {
	actor := equippedActor()
	actor["skills"] = map[string]any{"blades": 40, "red_magic": 50}
	actor["perks"] = map[string]any{"hardened_hide": 2}
	actor["resources"] = map[string]any{"hp": 9, "mp": 4, "sp": 3, "ap": 0}
	actor["conditions"] = []any{map[string]any{"key": "frightened", "stacks": 1}}
	actor["modifiers"] = []any{map[string]any{"modifier": map[string]any{"target": "blades", "operation": "add", "value": -10, "source": "situation:gm", "origin": "core:origin/morale"}, "turns_elapsed": 0}}
	return actor
}

func TestACurrentSheetPrintsTheSessionsNumbers(t *testing.T) {
	opened := open(t)
	actor := woundedActor()
	html := printPage(t, opened, map[string]any{"kind": "sheet", "actor": actor, "state": "current"})
	for pool, want := range map[string]int64{"hp": 9, "mp": 4, "sp": 3} {
		if got := printed(t, html, "current", pool); got != want {
			t.Errorf("current %s %d, want %d, the session's", pool, got, want)
		}
	}
	// The target is the session's: the skill read as progression reads it, with the perks and
	// the held modifiers on it, through the modifiers command.
	stored := mustStored(t, actor)
	value, err := readSkillValue(opened.rules, stored, "blades")
	if err != nil {
		t.Fatal(err)
	}
	target := runJSON(t, opened, "modifiers", map[string]any{"value": value, "target": "blades", "modifiers": actorModifiers(opened.rules, stored)})
	if got := printed(t, html, "target", "blades"); float64(got) != target["value"].(float64) || got != 30 {
		t.Errorf("blades target %d, the session's %v, want 30", got, target["value"])
	}
	if got := printed(t, html, "condition", "frightened"); got != 1 {
		t.Errorf("frightened stacks %d, want 1", got)
	}
	if got := printed(t, html, "item", "core:weapon/shortsword"); got != 1 || !strings.Contains(html, "main_hand") {
		t.Errorf("the worn shortsword: %d, %s", got, html)
	}
	if !strings.Contains(html, "situation:gm") {
		t.Error("a held modifier is listed with its source")
	}
}

func TestAFreshSheetPrintsFullResourcesAndNoConditions(t *testing.T) {
	opened := open(t)
	html := printPage(t, opened, map[string]any{"kind": "sheet", "actor": woundedActor(), "state": "fresh"})
	for _, pool := range []string{"hp", "mp", "sp"} {
		if current, maximum := printed(t, html, "current", pool), printed(t, html, "stat", pool); current != maximum {
			t.Errorf("fresh %s %d, want the maximum %d", pool, current, maximum)
		}
	}
	if strings.Contains(html, `data-condition=`) || strings.Contains(html, "situation:gm") {
		t.Error("a fresh sheet has no conditions and no held modifiers")
	}
	if got := printed(t, html, "target", "blades"); got != 40 {
		t.Errorf("fresh blades target %d, want 40 without the held -10", got)
	}
	if got := printed(t, html, "item", "core:weapon/shortsword"); got != 1 {
		t.Errorf("a fresh sheet keeps the gear: %d", got)
	}
}

func TestASheetStateIsCurrentOrFresh(t *testing.T) {
	body, _ := json.Marshal(map[string]any{"kind": "sheet", "actor": woundedActor(), "state": "tomorrow"})
	if _, err := open(t).Run("print", body); err == nil || !strings.Contains(err.Error(), "tomorrow") {
		t.Errorf("an unknown state: %v", err)
	}
	if html := printPage(t, open(t), map[string]any{"kind": "sheet", "actor": woundedActor()}); printed(t, html, "current", "hp") != 9 {
		t.Error("without a state the sheet is current")
	}
}
