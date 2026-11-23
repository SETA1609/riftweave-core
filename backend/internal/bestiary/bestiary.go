// Package bestiary is the `bestiary` command (docs/backend/bestiary.md, BE_S08_05): what the
// players know of each monster, and the bestiary as the GM or a player sees it. Listing and
// filtering the monsters is the catalog's; this package reads a monster's row through it.
package bestiary

import (
	"github.com/SETA1609/riftweave-core/backend/internal/actor"
	"github.com/SETA1609/riftweave-core/backend/internal/catalog"
	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/modifier"
	"github.com/SETA1609/riftweave-core/backend/internal/resolution"
)

// Operation is what the command does.
type Operation string

const (
	// OperationReveal sets the table's or a player's mark of one monster, or rolls a Lore check.
	OperationReveal Operation = "reveal"
	// OperationView shows monsters as the GM or a player sees them.
	OperationView Operation = "view"
)

// Level is how much of a monster is known.
type Level string

const (
	LevelUnknown Level = "unknown"
	LevelSeen    Level = "seen"
	LevelKnown   Level = "known"
)

// levelRank orders the levels, so the higher of two marks wins.
var levelRank = map[Level]int{LevelUnknown: 0, LevelSeen: 1, LevelKnown: 2}

// Shows is how much of an entry a viewer sees.
type Shows string

const (
	// ShowsSeen is the name, the description, and the size.
	ShowsSeen Shows = "seen"
	// ShowsFull is the whole sheet, which the caller reads with catalog and monster_sheet.
	ShowsFull Shows = "full"
)

// loreSkill is the skill a Lore check reads.
const loreSkill = "lore"

// monstersCollection is the catalog collection a monster's row is read from.
const monstersCollection = "monsters"

// Mark is one monster at a level, for the whole table (no Player) or one player.
type Mark struct {
	Monster string `json:"monster"`
	Level   Level  `json:"level"`
	Player  string `json:"player,omitempty"`
}

// Check is a Lore check by a player: the resolution command's input without its skill.
type Check struct {
	Actor     actor.Actor         `json:"actor"`
	Modifiers []modifier.Modifier `json:"modifiers"`
	Roll      int64               `json:"roll"`
}

// Input is one bestiary call.
type Input struct {
	Operation           Operation `json:"operation"`
	Knowledge           []Mark    `json:"knowledge"`
	Monster             string    `json:"monster"`
	Level               Level     `json:"level"`
	Player              string    `json:"player"`
	Check               *Check    `json:"check"`
	Monsters            []string  `json:"monsters"`
	PlayersSeeOnlyKnown bool      `json:"players_see_only_known"`
}

// Entry is one monster as the viewer sees it.
type Entry struct {
	Monster     string `json:"monster"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Size        string `json:"size"`
	Knowledge   Level  `json:"knowledge"`
	Shows       Shows  `json:"shows"`
	// Players is each player's own mark, on the GM's view only.
	Players *map[string]Level `json:"players,omitempty"`
}

// Output is the result: the marks after a reveal, or the entries of a view.
type Output struct {
	Knowledge []Mark             `json:"knowledge,omitempty"`
	Level     Level              `json:"level,omitempty"`
	Check     *resolution.Result `json:"check,omitempty"`
	Entries   *[]Entry           `json:"entries,omitempty"`
}

// Run runs one bestiary call.
func Run(ruleset *data.Ruleset, input Input) (Output, error) {
	switch input.Operation {
	case OperationReveal:
		return reveal(ruleset, input)
	case OperationView:
		return view(ruleset, input)
	}
	return Output{}, &data.NotDefinedError{Kind: "bestiary operation", Key: string(input.Operation)}
}

// KnownLevel is what a player knows of a monster: the higher of the table's mark and their own.
// For the table (player ""), only the table's marks count.
func KnownLevel(knowledge []Mark, monster, player string) Level {
	var best Level = LevelUnknown
	for _, mark := range knowledge {
		if mark.Monster == monster && (mark.Player == "" || mark.Player == player) && levelRank[mark.Level] > levelRank[best] {
			best = mark.Level
		}
	}
	return best
}

// marked is the marks with this holder's mark of the monster set to level, in place if it had one.
func marked(knowledge []Mark, monster string, level Level, player string) []Mark {
	var marks []Mark = append([]Mark{}, knowledge...)
	var mark Mark = Mark{Monster: monster, Level: level, Player: player}
	for index, existing := range marks {
		if existing.Monster == monster && existing.Player == player {
			marks[index] = mark
			return marks
		}
	}
	return append(marks, mark)
}

// row is the monster's resolved row, through the catalog.
func row(ruleset *data.Ruleset, ref string) (map[string]any, error) {
	output, err := catalog.Run(ruleset, catalog.Input{Operation: catalog.OperationGet, Collection: monstersCollection, Ref: ref})
	if err != nil {
		return nil, err
	}
	return output.Row.Row, nil
}

func reveal(ruleset *data.Ruleset, input Input) (Output, error) {
	if _, err := row(ruleset, input.Monster); err != nil {
		return Output{}, err
	}
	var knowledge []Mark = append([]Mark{}, input.Knowledge...)
	var output Output
	switch {
	case input.Check != nil:
		if input.Player == "" {
			return Output{}, &data.MissingInputError{Subject: "a lore check", Needs: "player"}
		}
		rolled, err := resolution.Run(ruleset, resolution.Input{Actor: input.Check.Actor, Skill: loreSkill, Modifiers: input.Check.Modifiers, Roll: input.Check.Roll})
		if err != nil {
			return Output{}, err
		}
		output.Check = &rolled
		if rolled.Success {
			knowledge = marked(knowledge, input.Monster, LevelKnown, input.Player)
		}
	case input.Level == "":
		return Output{}, &data.MissingInputError{Subject: "a reveal", Needs: "level"}
	default:
		if _, defined := levelRank[input.Level]; !defined {
			return Output{}, &data.NotDefinedError{Kind: "knowledge level", Key: string(input.Level)}
		}
		knowledge = marked(knowledge, input.Monster, input.Level, input.Player)
	}
	output.Knowledge, output.Level = knowledge, KnownLevel(knowledge, input.Monster, input.Player)
	return output, nil
}

func entryOf(ref string, monster map[string]any, level Level, shows Shows) Entry {
	label, _ := monster["label"].(string)
	description, _ := monster["description"].(string)
	size, _ := monster["size"].(string)
	return Entry{Monster: ref, Label: label, Description: description, Size: size, Knowledge: level, Shows: shows}
}

func view(ruleset *data.Ruleset, input Input) (Output, error) {
	if input.Monsters == nil {
		return Output{}, &data.MissingInputError{Subject: "a view", Needs: "monsters"}
	}
	var entries []Entry = []Entry{}
	for _, ref := range input.Monsters {
		monster, err := row(ruleset, ref)
		if err != nil {
			return Output{}, err
		}
		if input.Player == "" {
			var players map[string]Level = map[string]Level{}
			for _, mark := range input.Knowledge {
				if mark.Monster == ref && mark.Player != "" {
					players[mark.Player] = mark.Level
				}
			}
			var entry Entry = entryOf(ref, monster, KnownLevel(input.Knowledge, ref, ""), ShowsFull)
			entry.Players = &players
			entries = append(entries, entry)
			continue
		}
		var level Level = KnownLevel(input.Knowledge, ref, input.Player)
		switch {
		case !input.PlayersSeeOnlyKnown:
			entries = append(entries, entryOf(ref, monster, level, ShowsFull))
		case level == LevelSeen:
			entries = append(entries, entryOf(ref, monster, level, ShowsSeen))
		case level == LevelKnown:
			entries = append(entries, entryOf(ref, monster, level, ShowsFull))
		}
	}
	return Output{Entries: &entries}, nil
}
