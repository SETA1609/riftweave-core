package core

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// docs/backend/dev-mode.md

func devCore(t *testing.T) *Core {
	t.Helper()
	var opened *Core = open(t)
	opened.EnableDev()
	return opened
}

func startedSession(t *testing.T, opened *Core) any {
	t.Helper()
	return sessionCall(t, opened, map[string]any{"operation": "start", "start": startingActor()})["session"]
}

// dev runs one dev operation and returns the result, or fails the test.
func dev(t *testing.T, opened *Core, session any, operation map[string]any) map[string]any {
	t.Helper()
	operation["session"] = session
	body, _ := json.Marshal(operation)
	out, err := opened.Run("dev", body)
	if err != nil {
		t.Fatalf("dev %v: %v: %s", operation["operation"], err, out)
	}
	var result map[string]any
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func devRefused(t *testing.T, opened *Core, session any, operation map[string]any) string {
	t.Helper()
	operation["session"] = session
	body, _ := json.Marshal(operation)
	out, err := opened.Run("dev", body)
	if err == nil || !errors.Is(err, ErrRule) {
		t.Fatalf("dev %v: want a rule error, got %v: %s", operation["operation"], err, out)
	}
	return string(out)
}

func actorOf(result map[string]any) map[string]any {
	return result["session"].(map[string]any)["actor"].(map[string]any)
}

func TestWithoutDevEveryOperationIsRefused(t *testing.T) {
	var opened *Core = open(t)
	var session any = startedSession(t, opened)
	for _, operation := range []string{"resource", "xp", "level", "attribute", "skill", "condition", "effect", "item", "turns"} {
		out := devRefused(t, opened, session, map[string]any{"operation": operation})
		if !strings.Contains(out, "dev mode is off") {
			t.Errorf("%s: %s; want the dev mode error", operation, out)
		}
	}
}

func TestAResourceIsSetOrAdjusted(t *testing.T) {
	var opened *Core = devCore(t)
	var result map[string]any = dev(t, opened, startedSession(t, opened), map[string]any{"operation": "resource", "pool": "hp", "set": 5})
	result = dev(t, opened, result["session"], map[string]any{"operation": "resource", "pool": "mp", "add": -50})
	var resources map[string]any = actorOf(result)["resources"].(map[string]any)
	if resources["hp"] != 5.0 || resources["mp"] != 0.0 {
		t.Errorf("resources = %v; want hp 5 and mp 0 (never below 0)", resources)
	}
	if result["session"].(map[string]any)["dev"] != true {
		t.Error("the session is not marked dev")
	}
	if log := result["session"].(map[string]any)["log"].([]any); len(log) != 2 || log[0].(map[string]any)["command"] != "dev" {
		t.Errorf("log = %v; want two dev steps", log)
	}
}

func TestXPAndLevel(t *testing.T) {
	var opened *Core = devCore(t)
	var result map[string]any = dev(t, opened, startedSession(t, opened), map[string]any{"operation": "level", "set": 3})
	var actor map[string]any = actorOf(result)
	if actor["level"] != 3.0 || actor["xp"].(float64) <= 0 {
		t.Errorf("level %v xp %v; want level 3 and its XP total", actor["level"], actor["xp"])
	}
	var total float64 = actor["xp"].(float64)
	actor = actorOf(dev(t, opened, result["session"], map[string]any{"operation": "xp", "add": 7}))
	if actor["xp"] != total+7 || actor["level"] != 3.0 {
		t.Errorf("xp %v level %v; want %v and level 3", actor["xp"], actor["level"], total+7)
	}
}

func TestAnAttributeAndASkill(t *testing.T) {
	var opened *Core = devCore(t)
	var result map[string]any = dev(t, opened, startedSession(t, opened), map[string]any{"operation": "attribute", "key": "str", "add": -20})
	result = dev(t, opened, result["session"], map[string]any{"operation": "skill", "key": "blades", "set": 140})
	var actor map[string]any = actorOf(result)
	if actor["attributes"].(map[string]any)["str"] != 1.0 || actor["skills"].(map[string]any)["blades"] != 100.0 {
		t.Errorf("str %v blades %v; want 1 and 100", actor["attributes"].(map[string]any)["str"], actor["skills"].(map[string]any)["blades"])
	}
	devRefused(t, opened, result["session"], map[string]any{"operation": "skill", "key": "no_such_skill", "set": 5})
	devRefused(t, opened, result["session"], map[string]any{"operation": "attribute", "key": "luck", "set": 5})
}

func TestAConditionIsInflictedAndCleared(t *testing.T) {
	var opened *Core = devCore(t)
	var result map[string]any = dev(t, opened, startedSession(t, opened), map[string]any{"operation": "condition", "key": "poisoned", "stacks": 1})
	if conditions := actorOf(result)["conditions"].([]any); len(conditions) != 1 {
		t.Fatalf("conditions = %v; want poisoned", conditions)
	}
	result = dev(t, opened, result["session"], map[string]any{"operation": "condition", "key": "poisoned", "stacks": 0})
	if conditions := actorOf(result)["conditions"].([]any); len(conditions) != 0 {
		t.Errorf("conditions = %v; want none", conditions)
	}
	devRefused(t, opened, result["session"], map[string]any{"operation": "condition", "key": "no_such_condition", "stacks": 1})
}

// valueOf applies the actor's modifiers on target to base through the real stacking rule.
func valueOf(t *testing.T, opened *Core, actor map[string]any, target string, base int) float64 {
	t.Helper()
	var modifiers []any
	for _, held := range actor["modifiers"].([]any) {
		modifiers = append(modifiers, held.(map[string]any)["modifier"])
	}
	body, _ := json.Marshal(map[string]any{"value": base, "target": target, "modifiers": modifiers})
	out, err := opened.Run("modifiers", body)
	if err != nil {
		t.Fatalf("modifiers: %v: %s", err, out)
	}
	var result map[string]any
	_ = json.Unmarshal(out, &result)
	return result["value"].(float64)
}

func TestAnInflictedEffectStacksByItsOrigin(t *testing.T) {
	var opened *Core = devCore(t)
	var start map[string]any = startingActor()
	start["modifiers"] = []any{map[string]any{"modifier": map[string]any{"target": "str", "op": "add", "value": 2, "source": "spell:bulls_strength", "origin": "core:origin/arcane"}, "turns_elapsed": 0}}
	var session any = sessionCall(t, opened, map[string]any{"operation": "start", "start": start})["session"]
	var arcane map[string]any = dev(t, opened, session, map[string]any{"operation": "effect", "effect": "core:effect/fortify_attribute", "parameter": "str", "magnitude": 3, "origin": "core:origin/arcane"})
	if value := valueOf(t, opened, actorOf(arcane), "str", 5); value != 8 {
		t.Errorf("str with two arcane bonuses = %v; want 8, the strongest only", value)
	}
	var cleared map[string]any = dev(t, opened, arcane["session"], map[string]any{"operation": "effect", "effect": "core:effect/fortify_attribute", "clear": true})
	if value := valueOf(t, opened, actorOf(cleared), "str", 5); value != 7 {
		t.Errorf("str after clearing = %v; want 7", value)
	}
	devRefused(t, opened, session, map[string]any{"operation": "effect", "effect": "core:effect/no_such_effect", "magnitude": 1, "origin": "core:origin/arcane"})
	devRefused(t, opened, session, map[string]any{"operation": "effect", "effect": "core:effect/fortify_attribute", "magnitude": 1, "origin": "core:origin/no_such_origin"})
}

func TestAnItemIsAddedAndRemoved(t *testing.T) {
	var opened *Core = devCore(t)
	var result map[string]any = dev(t, opened, startedSession(t, opened), map[string]any{"operation": "item", "base": "core:weapon/dagger"})
	result = dev(t, opened, result["session"], map[string]any{"operation": "item", "base": "core:consumable/potion_healing", "quantity": 3})
	var inventory []any = actorOf(result)["inventory"].([]any)
	if len(inventory) != 2 || inventory[1].(map[string]any)["id"] != 2.0 || inventory[1].(map[string]any)["quantity"] != 3.0 {
		t.Fatalf("inventory = %v; want a dagger (id 1) and three potions (id 2)", inventory)
	}
	result = dev(t, opened, result["session"], map[string]any{"operation": "item", "remove": 1})
	if inventory = actorOf(result)["inventory"].([]any); len(inventory) != 1 {
		t.Errorf("inventory = %v; want the potions only", inventory)
	}
	devRefused(t, opened, result["session"], map[string]any{"operation": "item", "base": "core:weapon/no_such_weapon"})
	devRefused(t, opened, result["session"], map[string]any{"operation": "item", "remove": 9})
}

func TestTurnsEndATimedEffect(t *testing.T) {
	var opened *Core = devCore(t)
	var result map[string]any = dev(t, opened, startedSession(t, opened), map[string]any{"operation": "effect", "effect": "core:effect/fortify_attribute", "parameter": "agi", "magnitude": 2, "origin": "core:origin/arcane", "turns": 3})
	result = dev(t, opened, result["session"], map[string]any{"operation": "turns", "turns": 2})
	if modifiers := actorOf(result)["modifiers"].([]any); len(modifiers) != 1 || modifiers[0].(map[string]any)["turns_elapsed"] != 2.0 {
		t.Fatalf("modifiers = %v; want one at 2 turns elapsed", modifiers)
	}
	result = dev(t, opened, result["session"], map[string]any{"operation": "turns", "turns": 1})
	if modifiers := actorOf(result)["modifiers"].([]any); len(modifiers) != 0 {
		t.Errorf("modifiers = %v; want none after 3 turns", modifiers)
	}
}

func TestABadOperationChangesNothing(t *testing.T) {
	var opened *Core = devCore(t)
	var session any = startedSession(t, opened)
	devRefused(t, opened, session, map[string]any{"operation": "resource", "pool": "hp", "set": 5, "add": 1})
	devRefused(t, opened, session, map[string]any{"operation": "resource", "pool": "hp"})
	devRefused(t, opened, session, map[string]any{"operation": "no_such_operation"})
}

func TestADevSessionReplaysAndStaysMarked(t *testing.T) {
	var opened *Core = devCore(t)
	var result map[string]any = dev(t, opened, startedSession(t, opened), map[string]any{"operation": "resource", "pool": "hp", "set": 9})
	result = dev(t, opened, result["session"], map[string]any{"operation": "item", "base": "core:weapon/dagger"})
	var replayed map[string]any = sessionCall(t, opened, map[string]any{"operation": "replay", "session": result["session"]})
	if replayed["matches"] != true || replayed["session"].(map[string]any)["dev"] != true {
		t.Errorf("replay matches %v, dev %v; want both true", replayed["matches"], replayed["session"].(map[string]any)["dev"])
	}
	body, _ := json.Marshal(map[string]any{"operation": "replay", "session": result["session"]})
	if out, err := open(t).Run("session", body); err == nil || !strings.Contains(string(out), "dev mode is off") {
		t.Errorf("a dev session replayed without --dev: %v %s", err, out)
	}
}
