package play

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// The Bestiary (BE_S08_05, docs/backend/bestiary.md): the monsters of core and the enabled
// modules, listed and filtered over the `catalog` command, and what the players know of each over
// the `bestiary` command. The GM marks a monster for the table; each mark is a `bestiary` step in
// the model's session, so it replays. The players' view follows the settings box "Players see
// only what they know". It holds no rule.

// bestiaryCommand is the command the screen runs for the marks and the views.
const bestiaryCommand = "bestiary"

// The bestiary operations the screen sends (bestiary.Operation in the core).
const (
	bestiaryReveal = "reveal"
	bestiaryView   = "view"
)

// The knowledge levels and what an entry shows (bestiary.Level and bestiary.Shows in the core).
const (
	levelUnknown = "unknown"
	levelSeen    = "seen"
	levelKnown   = "known"
	showsFull    = "full"
)

// playersViewer is the player the players' view is drawn for: no player's own marks, so only
// what the whole table was shown.
const playersViewer = "table"

// markKeys is the level each marking key sets for the table.
var markKeys = map[string]string{"u": levelUnknown, "s": levelSeen, "k": levelKnown}

// The filter's field words (docs/backend/bestiary.md#the-screen): a creature type, a level or a
// range of levels, and a phase.
const (
	filterType  = "type:"
	filterLevel = "level:"
	filterPhase = "phase:"
)

// bestiaryStep is where the screen is: reading the list, typing a filter, or reading one entry.
type bestiaryStep uint8

const (
	stepBestiaryList bestiaryStep = iota
	stepBestiaryFilter
	stepBestiaryEntry
)

// bestiaryHints is what each step's keys do, by view: the GM's, then the players'.
var bestiaryHints = [...][2]string{
	stepBestiaryList: {
		"up/down, enter to open, s seen, k known, u unknown for the table, / to filter, p for the players' view, esc to go back",
		"up/down, enter to open, / to filter, p for the GM view, esc to go back",
	},
	stepBestiaryFilter: {
		"words, type:<type>, level:<n> or level:<n>-<m>, phase:<phase>; enter to filter, esc to cancel",
		"words, type:<type>, level:<n> or level:<n>-<m>, phase:<phase>; enter to filter, esc to cancel",
	},
	stepBestiaryEntry: {"esc for the list", "esc for the list"},
}

// bestiaryEntry is one monster as the viewer sees it (bestiary.Entry in the core).
type bestiaryEntry struct {
	Monster     string            `json:"monster"`
	Label       string            `json:"label"`
	Description string            `json:"description"`
	Size        string            `json:"size"`
	Knowledge   string            `json:"knowledge"`
	Shows       string            `json:"shows"`
	Players     map[string]string `json:"players"`
}

// bestiaryScreen is the screen's state.
type bestiaryScreen struct {
	step    bestiaryStep
	players bool // the players' view, not the GM's
	text    string
	rows    map[string]catalogRow // the listed monsters' catalog rows, by ref
	entries []bestiaryEntry
	table   table.Model
	filter  textinput.Model
	opened  bestiaryEntry
}

// openBestiary lists every monster in the GM's view.
func (model Model) openBestiary() Model {
	model.returnCursor, model.problem = model.cursor, ""
	return model.bestiaryListed(bestiaryScreen{})
}

// call runs one command, keeps it as the last call, and decodes its result into target.
func (model *Model) call(command string, input map[string]any, target any) error {
	encoded, _ := json.Marshal(input)
	output, err := model.run(command, encoded)
	model.lastInput, model.lastOutput = encoded, output
	if err != nil {
		return err
	}
	return json.Unmarshal(output, target)
}

// filtersOf turns the typed filter into catalog filters: the field words, then the rest as text.
func filtersOf(text string) ([]map[string]any, error) {
	var filters []map[string]any
	var words []string
	for _, word := range strings.Fields(text) {
		switch {
		case strings.HasPrefix(word, filterType):
			filters = append(filters, map[string]any{"field": "tags", "equals": strings.TrimPrefix(word, filterType)})
		case strings.HasPrefix(word, filterPhase):
			filters = append(filters, map[string]any{"field": "phase", "equals": strings.TrimPrefix(word, filterPhase)})
		case strings.HasPrefix(word, filterLevel):
			low, high, ranged := strings.Cut(strings.TrimPrefix(word, filterLevel), "-")
			if !ranged {
				high = low
			}
			minimum, lowErr := strconv.ParseFloat(low, 64)
			maximum, highErr := strconv.ParseFloat(high, 64)
			if lowErr != nil || highErr != nil {
				return nil, fmt.Errorf("%q is not a level or a range of levels", word)
			}
			filters = append(filters, map[string]any{"field": "level", "min": minimum, "max": maximum})
		default:
			words = append(words, word)
		}
	}
	if len(words) > 0 {
		filters = append(filters, map[string]any{"text": strings.Join(words, " ")})
	}
	return filters, nil
}

// knowledge is the session's marks, none before the first.
func (model Model) knowledge() []any {
	var session struct {
		Bestiary struct {
			Knowledge []any `json:"knowledge"`
		} `json:"bestiary"`
	}
	_ = json.Unmarshal(model.session, &session)
	if session.Bestiary.Knowledge == nil {
		return []any{}
	}
	return session.Bestiary.Knowledge
}

// playersSeeOnlyKnown is the settings box (settings.bestiary.players_see_only_known).
func (model Model) playersSeeOnlyKnown() bool {
	bestiary, _ := model.settings["bestiary"].(map[string]any)
	onlyKnown, _ := bestiary["players_see_only_known"].(bool)
	return onlyKnown
}

// bestiaryListed lists the monsters that match the filter, then views them as the screen's
// viewer sees them.
func (model Model) bestiaryListed(screen bestiaryScreen) Model {
	filters, err := filtersOf(screen.text)
	if err != nil {
		model.problem = err.Error()
		return model
	}
	var input = map[string]any{"operation": catalogList, "collection": "monsters"}
	if len(filters) > 0 {
		input["filters"] = filters
	}
	var listed struct {
		Rows []catalogRow `json:"rows"`
	}
	if err := model.call(catalogCommand, input, &listed); err != nil {
		model.problem = err.Error()
		return model
	}
	var refs []string = []string{}
	screen.rows = map[string]catalogRow{}
	for _, row := range listed.Rows {
		refs = append(refs, row.Ref)
		screen.rows[row.Ref] = row
	}
	var view = map[string]any{"operation": bestiaryView, "knowledge": model.knowledge(), "monsters": refs, "players_see_only_known": model.playersSeeOnlyKnown()}
	if screen.players {
		view["player"] = playersViewer
	}
	var viewed struct {
		Entries []bestiaryEntry `json:"entries"`
	}
	if err := model.call(bestiaryCommand, view, &viewed); err != nil {
		model.problem = err.Error()
		return model
	}
	var cursor int
	if screen.entries != nil {
		cursor = screen.table.Cursor()
	}
	screen.entries, screen.step = viewed.Entries, stepBestiaryList
	screen.table = bestiaryTable(screen)
	screen.table.SetCursor(min(cursor, max(len(screen.entries)-1, 0)))
	model.bestiary, model.mode, model.problem = &screen, modeBestiary, ""
	return model
}

// bestiaryTable lays the entries out; a monster a player has only seen shows no level or phase.
func bestiaryTable(screen bestiaryScreen) table.Model {
	var mark string = "Table"
	if screen.players {
		mark = "Known"
	}
	var columns = []table.Column{{Title: "Monster", Width: 22}, {Title: "Level", Width: 5}, {Title: "Size", Width: 8}, {Title: "Phase", Width: 6}, {Title: "Type", Width: 12}, {Title: mark, Width: 8}}
	var rows []table.Row = make([]table.Row, len(screen.entries))
	for index, entry := range screen.entries {
		var row map[string]any = screen.rows[entry.Monster].Row
		var level, phase, kind string = cell(row["level"]), cell(row["phase"]), "-"
		if tags, tagged := row["tags"].([]any); tagged && len(tags) > 0 {
			kind = cell(tags[0])
		}
		if entry.Shows != showsFull {
			level, phase, kind = "?", "?", "?"
		}
		rows[index] = table.Row{entry.Label, level, entry.Size, phase, kind, entry.Knowledge}
	}
	return table.New(table.WithColumns(columns), table.WithRows(rows), table.WithHeight(tableHeight), table.WithFocused(true))
}

// marked records the table's mark of the monster under the cursor as a session step, then lists
// again.
func (model Model) marked(screen bestiaryScreen, level string) Model {
	if len(screen.entries) == 0 {
		return model
	}
	if err := model.ensureSession(); err != nil {
		model.problem = err.Error()
		return model
	}
	var step = map[string]any{"command": bestiaryCommand, "input": map[string]any{"operation": bestiaryReveal, "monster": screen.entries[screen.table.Cursor()].Monster, "level": level}}
	var result struct {
		Session json.RawMessage `json:"session"`
	}
	if err := model.call("session", map[string]any{"operation": "step", "session": model.session, "step": step}, &result); err != nil {
		model.problem = err.Error()
		return model
	}
	model.session = result.Session
	return model.bestiaryListed(screen)
}

// bestiaryKey runs one key on the screen.
func (model Model) bestiaryKey(message tea.Msg) (tea.Model, tea.Cmd) {
	var screen bestiaryScreen = *model.bestiary
	key, isKey := message.(tea.KeyMsg)
	switch screen.step {
	case stepBestiaryList:
		if isKey {
			if level, marks := markKeys[key.String()]; marks && !screen.players {
				return model.marked(screen, level), nil
			}
			switch key.String() {
			case "esc":
				if screen.text != "" {
					screen.text = ""
					return model.bestiaryListed(screen), nil
				}
				model.bestiary, model.mode, model.cursor, model.problem = nil, modeMenu, model.returnCursor, ""
				return model, nil
			case "p":
				screen.players = !screen.players
				return model.bestiaryListed(screen), nil
			case "/":
				screen.filter = textinput.New()
				screen.filter.Prompt = "/"
				screen.filter.SetValue(screen.text)
				screen.filter.Focus()
				screen.step = stepBestiaryFilter
				model.bestiary = &screen
				return model, textinput.Blink
			case "enter":
				if len(screen.entries) == 0 {
					return model, nil
				}
				screen.opened, screen.step = screen.entries[screen.table.Cursor()], stepBestiaryEntry
				model.bestiary = &screen
				return model, nil
			}
		}
		var command tea.Cmd
		screen.table, command = screen.table.Update(message)
		model.bestiary = &screen
		return model, command
	case stepBestiaryFilter:
		if isKey {
			switch key.Type {
			case tea.KeyEsc:
				screen.step = stepBestiaryList
				model.bestiary = &screen
				return model, nil
			case tea.KeyEnter:
				screen.text = strings.TrimSpace(screen.filter.Value())
				return model.bestiaryListed(screen), nil
			}
		}
		var command tea.Cmd
		screen.filter, command = screen.filter.Update(message)
		model.bestiary = &screen
		return model, command
	case stepBestiaryEntry:
		if isKey && key.String() == "esc" {
			screen.step = stepBestiaryList
		}
	}
	model.bestiary = &screen
	return model, nil
}

// bestiaryView draws the step the screen is at.
func (model Model) bestiaryView() string {
	var screen *bestiaryScreen = model.bestiary
	var view strings.Builder
	var viewer string = "GM view"
	var hint int = 0
	if screen.players {
		viewer, hint = "Players' view", 1
	}
	switch screen.step {
	case stepBestiaryList, stepBestiaryFilter:
		var title string = "Bestiary"
		if screen.text != "" {
			title += " matching " + strconv.Quote(screen.text)
		}
		fmt.Fprintf(&view, "%s %s %s\n\n", titleStyle.Render(title), dimStyle.Render(fmt.Sprintf("(%d)", len(screen.entries))), dimStyle.Render(viewer))
		if len(screen.entries) == 0 {
			view.WriteString("no monsters\n")
		} else {
			view.WriteString(screen.table.View() + "\n")
		}
		if screen.step == stepBestiaryFilter {
			view.WriteString("\n" + screen.filter.View() + "\n")
		}
	case stepBestiaryEntry:
		var opened bestiaryEntry = screen.opened
		fmt.Fprintf(&view, "%s\n%s\n\n%s\n", titleStyle.Render(opened.Label), dimStyle.Render(opened.Monster+", "+opened.Size+", "+opened.Knowledge), opened.Description)
		if opened.Shows == showsFull {
			encoded, _ := json.MarshalIndent(screen.rows[opened.Monster].Row, "", "  ")
			view.WriteString("\n")
			view.Write(encoded)
			view.WriteString("\n")
		}
		for player, level := range opened.Players {
			fmt.Fprintf(&view, "%s: %s\n", player, level)
		}
	}
	if model.problem != "" {
		view.WriteString("\n" + errorStyle.Render(model.problem) + "\n")
	}
	view.WriteString("\n" + hintStyle.Render(bestiaryHints[screen.step][hint]) + "\n")
	return view.String()
}
