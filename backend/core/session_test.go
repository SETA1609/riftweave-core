package core

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// The chain is ruleset/tests/vectors/playthrough.json run as session steps: each step drops
// the fields the session fills from the actor and must still match its vector's expect.

func sessionCall(t *testing.T, opened *Core, input map[string]any) map[string]any {
	t.Helper()
	body, _ := json.Marshal(input)
	out, err := opened.Run("session", body)
	if err != nil {
		t.Fatalf("session: %v: %s", err, out)
	}
	var result map[string]any
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func startingActor() map[string]any {
	return map[string]any{
		"attributes": map[string]any{"str": 6, "per": 5, "end": 6, "int": 5, "wil": 4, "agi": 5, "cha": 4, "lck": 5},
		"level":      1, "xp": 0, "skills": map[string]any{"red_magic": 50}, "tags": []any{},
		"growth":     map[string]any{"hp": 0, "mp": 0, "sp": 0},
		"resources":  map[string]any{"hp": 67, "mp": 42, "sp": 10, "ap": 0},
		"conditions": []any{}, "modifiers": []any{}, "flags": []any{}, "inventory": []any{}, "equipped": map[string]any{},
	}
}

func without(section map[string]any, fields ...string) {
	for _, field := range fields {
		delete(section, field)
	}
}

// runChain plays the playthrough through a session and returns the final session.
func runChain(t *testing.T, opened *Core) map[string]any {
	t.Helper()
	vectors := readJSON(t, filepath.Join(repo, "ruleset", "tests", "vectors", "playthrough.json")).([]any)
	var session any = sessionCall(t, opened, map[string]any{"operation": "start", "start": startingActor()})["session"]
	for _, raw := range vectors {
		var vector map[string]any = raw.(map[string]any)
		name := vector["name"].(string)
		command := vector["function"].(string)
		expect := vector["expect"].(map[string]any)
		input := map[string]any{}
		for key, value := range vector {
			if key != "name" && key != "function" && key != "expect" {
				input[key] = value
			}
		}
		step := map[string]any{"command": command, "input": input}
		switch {
		case strings.HasPrefix(name, "create"):
			without(input, "actor")
		case strings.HasPrefix(name, "fight: the zombie"):
			step["as"] = "defender"
			without(input["defender"].(map[string]any), "hp")
		case strings.HasPrefix(name, "fight"):
			without(input["attacker"].(map[string]any), "attributes")
		case strings.HasPrefix(name, "rest"), strings.HasPrefix(name, "recover"):
			without(input, "actor", "current")
		case strings.HasPrefix(name, "level: the"):
			without(input, "actor")
		case strings.HasPrefix(name, "cast"):
			without(input["caster"].(map[string]any), "mana", "attributes", "skills")
		}
		answer := sessionCall(t, opened, map[string]any{"operation": "step", "session": session, "step": step})
		session = answer["session"]
		result := answer["result"].(map[string]any)
		for field, want := range expect {
			if !reflect.DeepEqual(result[field], want) {
				t.Errorf("%s: field %q: got %v, want %v", name, field, result[field], want)
			}
		}
	}
	return session.(map[string]any)
}

func TestTheChainMatchesItsVectors(t *testing.T) {
	var session map[string]any = runChain(t, open(t))
	var actor map[string]any = session["actor"].(map[string]any)
	if actor["level"] != float64(2) || actor["xp"] != float64(900) {
		t.Errorf("level %v, xp %v", actor["level"], actor["xp"])
	}
	var resources map[string]any = actor["resources"].(map[string]any)
	if resources["hp"] != float64(73) || resources["mp"] != float64(46) {
		t.Errorf("resources %v", resources)
	}
	if len(actor["conditions"].([]any)) != 0 {
		t.Errorf("the long rest should clear blinded: %v", actor["conditions"])
	}
	if len(session["log"].([]any)) != 12 {
		t.Errorf("log has %d entries", len(session["log"].([]any)))
	}
}

func TestASavedSessionReplaysToTheSameResult(t *testing.T) {
	opened := open(t)
	session := runChain(t, opened)
	saved, _ := json.Marshal(session)
	var loaded map[string]any
	if err := json.Unmarshal(saved, &loaded); err != nil {
		t.Fatal(err)
	}
	answer := sessionCall(t, opened, map[string]any{"operation": "replay", "session": loaded})
	if answer["matches"] != true || !reflect.DeepEqual(answer["session"], session) {
		t.Errorf("replay differs: matches %v", answer["matches"])
	}
}

func TestAReplayNoticesAnActorThatDisagreesWithTheLog(t *testing.T) {
	opened := open(t)
	session := runChain(t, opened)
	session["actor"].(map[string]any)["xp"] = 5
	answer := sessionCall(t, opened, map[string]any{"operation": "replay", "session": session})
	if answer["matches"] != false {
		t.Error("a changed actor should not match")
	}
	if answer["session"].(map[string]any)["actor"].(map[string]any)["xp"] != float64(900) {
		t.Error("the log should win")
	}
}

func TestAFailedStepChangesNothing(t *testing.T) {
	opened := open(t)
	session := sessionCall(t, opened, map[string]any{"operation": "start", "start": startingActor()})["session"]
	body, _ := json.Marshal(map[string]any{"operation": "step", "session": session, "step": map[string]any{"command": "casting", "input": map[string]any{"spell": "no_such_spell", "roll": 1}}})
	if _, err := opened.Run("session", body); !errors.Is(err, ErrRule) {
		t.Errorf("want a rule error, got %v", err)
	}
}

func TestAnUnknownOperationIsBadInput(t *testing.T) {
	if _, err := open(t).Run("session", []byte(`{"operation":"rewind"}`)); !errors.Is(err, ErrInput) {
		t.Errorf("want bad input, got %v", err)
	}
}

func TestEveryAdaptedCommandExists(t *testing.T) {
	for name := range sessionAdapters {
		if _, known := commands[name]; !known {
			t.Errorf("session adapter for unknown command %q", name)
		}
	}
}

// step runs one session step and returns the new session and the step's result.
func step(t *testing.T, opened *Core, session any, command string, input map[string]any) (map[string]any, map[string]any) {
	t.Helper()
	answer := sessionCall(t, opened, map[string]any{"operation": "step", "session": session, "step": map[string]any{"command": command, "input": input}})
	result, _ := answer["result"].(map[string]any)
	return answer["session"].(map[string]any), result
}

// APP-12: a level-up's choices are session steps, so a played run needs no dev step.
func TestALevelUpsChoicesAreWrittenWithNoDevStep(t *testing.T) {
	opened := open(t)
	start := startingActor()
	start["tags"] = []any{"blades"}
	var session map[string]any = sessionCall(t, opened, map[string]any{"operation": "start", "start": start})["session"].(map[string]any)

	session, spent := step(t, opened, session, "progression", map[string]any{"operation": "spend", "key": "blades", "points": 10})
	if spent["ranks"] != float64(10) {
		t.Fatalf("a tag skill buys 10 ranks for 10 points: %v", spent)
	}
	session, _ = step(t, opened, session, "progression", map[string]any{"operation": "spend", "key": "evasion", "points": 7})
	skills := session["actor"].(map[string]any)["skills"].(map[string]any)
	if skills["blades"] != float64(27) || skills["evasion"] != float64(18) {
		t.Errorf("blades 17 + 10 and evasion 15 + 3 written: %v", skills)
	}

	session, _ = step(t, opened, session, "progression", map[string]any{"operation": "take_perk", "perk": "toughness"})
	session, _ = step(t, opened, session, "progression", map[string]any{"operation": "take_perk", "perk": "toughness"})
	session, refused := step(t, opened, session, "progression", map[string]any{"operation": "take_perk", "perk": "power_attack"})
	if refused["taken"] != false {
		t.Errorf("power_attack needs blades 40: %v", refused)
	}
	perks := session["actor"].(map[string]any)["perks"].(map[string]any)
	if !reflect.DeepEqual(perks, map[string]any{"toughness": float64(2)}) {
		t.Errorf("toughness at rank 2 and nothing else: %v", perks)
	}

	session, _ = step(t, opened, session, "spellcraft", map[string]any{"operation": "learn", "spell": "fire_bolt"})
	actor := session["actor"].(map[string]any)
	if !reflect.DeepEqual(actor["spells"], []any{"fire_bolt"}) || !reflect.DeepEqual(actor["known_effects"], []any{"damage_fire"}) {
		t.Errorf("fire_bolt and its fire learned: %v %v", actor["spells"], actor["known_effects"])
	}

	if session["dev"] == true {
		t.Error("no dev step ran")
	}
	if len(session["log"].([]any)) != 6 {
		t.Errorf("six steps logged, the refused perk too: %d", len(session["log"].([]any)))
	}
	answer := sessionCall(t, opened, map[string]any{"operation": "replay", "session": session})
	if answer["matches"] != true {
		t.Error("the session replays to the same actor")
	}
}
