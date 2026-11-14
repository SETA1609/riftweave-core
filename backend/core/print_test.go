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
	actor["perks"] = map[string]any{"toughness": 2}
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
	// Derived stats with the actor's modifiers: Toughness at rank 2 is +30 HP.
	derived := runJSON(t, opened, "derived", map[string]any{"actor": characterOf(mustStored(t, actor)),
		"modifiers": []any{map[string]any{"target": "hp", "operation": "add", "value": 30, "source": "feature:toughness", "origin": "core:origin/trained"}}})
	for _, stat := range []string{"hp", "mp", "sp", "ap"} {
		if got := printed(t, html, "stat", stat); float64(got) != derived[stat].(float64) {
			t.Errorf("%s printed %d, derived %v", stat, got, derived[stat])
		}
	}
	if got := printed(t, html, "perk", "toughness"); got != 2 {
		t.Errorf("toughness rank %d, want 2", got)
	}
}

func TestAMonsterStatBlockPrintsItsToHitFromTheMonsterSheet(t *testing.T) {
	opened := open(t)
	wolf := map[string]any{
		"attributes": map[string]any{"str": 6, "agi": 5, "per": 6, "end": 4, "wil": 2}, "level": 5, "hitPoints": 59,
		"attacks": []any{map[string]any{"key": "bite", "label": "Bite", "weapon": "core:natural_weapon/bite"}},
		"resist":  map[string]any{"cold": 5},
	}
	html := printPage(t, opened, map[string]any{"kind": "monster", "label": "Wolf", "monster": wolf})
	sheet := runJSON(t, opened, "monster_sheet", map[string]any{"monster": map[string]any{
		"attributes": wolf["attributes"], "level": 5, "hitPoints": 59}, "read_skills": []any{"unarmed", "evasion"}})
	skills := sheet["skills"].(map[string]any)
	if got := printed(t, html, "to-hit", "bite"); float64(got) != skills["unarmed"].(float64) {
		t.Errorf("bite to-hit %d, monster_sheet unarmed %v", got, skills["unarmed"])
	}
	if got := printed(t, html, "stat", "evasion"); float64(got) != skills["evasion"].(float64) {
		t.Errorf("evasion %d, monster_sheet %v", got, skills["evasion"])
	}
	if got := printed(t, html, "stat", "hp"); got != 59 {
		t.Errorf("hp %d, want the authored 59", got)
	}
	if got := printed(t, html, "resist", "cold"); got != 5 {
		t.Errorf("cold resist %d, want 5", got)
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
