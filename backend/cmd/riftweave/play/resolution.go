package play

import (
	"fmt"
	"math/rand/v2"
)

// resolutionForm is Skill resolution: one d100 check through the `resolution` command
// (docs/core/resolution.md). A blank roll is rolled here, so the input the core gets always
// names its roll and can be replayed.
func resolutionForm(luck string) form {
	return form{
		title:   "Skill resolution",
		command: "resolution",
		fields: []field{
			{label: "Skill", initial: "lockpick", freeText: true},
			{label: "Skill value", initial: "40"},
			{label: "LCK", initial: luck},
			{label: "Target modifier", initial: "0"},
			{label: "Roll (d100, blank rolls)", initial: "", optional: true},
		},
		build: func(values []string) map[string]any {
			var modifiers = []any{}
			if bonus := number(values[3]); bonus != 0 {
				modifiers = append(modifiers, map[string]any{
					"target": values[0], "operation": "add", "value": bonus, "source": "situation:gm",
				})
			}
			return map[string]any{
				"actor":     map[string]any{"skills": map[string]int64{values[0]: number(values[1])}, "attributes": map[string]int64{"lck": number(values[2])}},
				"skill":     values[0],
				"modifiers": modifiers,
				"roll":      number(values[4]),
			}
		},
		before: func(values []string) []string {
			if values[4] == "" {
				values[4] = fmt.Sprint(rand.IntN(100) + 1)
			}
			return values
		},
		columns: []string{"target", "success", "margin", "tags"},
		logLine: func(values []string, result map[string]any) string {
			return fmt.Sprintf("%s check: rolled %s against %v, %s", values[0], values[4], result["target"], successWord(result["success"]))
		},
	}
}

func successWord(value any) string {
	if value == true {
		return "success"
	}
	return "failure"
}
