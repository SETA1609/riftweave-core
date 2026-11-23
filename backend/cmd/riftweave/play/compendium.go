package play

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// The Compendium (BE_S08_10, docs/backend/app-layout.md#compendium): every table a player or a GM
// studies to plan a build, over the `catalog` command. Pick a collection, read its table, filter
// it by text, and open one row in full. It holds no rule: the rows, their columns, and the filter
// come from the core.

// catalogCommand is the command the screen runs.
const catalogCommand = "catalog"

// The catalog operations the screen sends (catalog.Operation in the core).
const (
	catalogCollections = "collections"
	catalogList        = "list"
	catalogGet         = "get"
)

// compendiumStep is where the screen is: picking a collection, reading its table, typing a
// filter, or reading one row.
type compendiumStep uint8

const (
	stepCollections compendiumStep = iota
	stepTable
	stepFilter
	stepRow
)

// compendiumHints is what each step's keys do.
var compendiumHints = [...]string{
	stepCollections: "up/down, enter to open a table, esc for the menu",
	stepTable:       "up/down to scroll, enter to open a row, / to filter, esc for the collections",
	stepFilter:      "type words in the label or description, enter to filter, esc to cancel",
	stepRow:         "esc for the table",
}

// tableHeight is how many rows the table shows at once; it scrolls past that.
const tableHeight = 15

type catalogColumn struct {
	Field string `json:"field"`
	Label string `json:"label"`
}

type catalogRow struct {
	Ref    string         `json:"ref"`
	Source string         `json:"source"`
	Row    map[string]any `json:"row"`
}

type catalogCollection struct {
	Name  string `json:"name"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

// compendium is the screen's state.
type compendium struct {
	step        compendiumStep
	collections []catalogCollection
	cursor      int // the collection picked
	columns     []catalogColumn
	rows        []catalogRow
	table       table.Model
	filter      textinput.Model
	text        string // the filter applied
	opened      catalogRow
}

// catalog runs one catalog call, keeps it as the last call, and decodes its result into target.
func (model *Model) catalog(input map[string]any, target any) error {
	encoded, _ := json.Marshal(input)
	output, err := model.run(catalogCommand, encoded)
	model.lastInput, model.lastOutput = encoded, output
	if err != nil {
		return err
	}
	return json.Unmarshal(output, target)
}

// openCompendium shows the collections.
func (model Model) openCompendium() Model {
	var result struct {
		Collections []catalogCollection `json:"collections"`
	}
	if err := model.catalog(map[string]any{"operation": catalogCollections}, &result); err != nil {
		model.problem = err.Error()
		return model
	}
	var screen compendium = compendium{collections: result.Collections}
	model.returnCursor, model.problem = model.cursor, ""
	model.compendium, model.mode = &screen, modeCompendium
	return model
}

// listed reads the picked collection's rows, with the text filter if one is set.
func (model Model) listed(screen compendium) Model {
	var input = map[string]any{"operation": catalogList, "collection": screen.collections[screen.cursor].Name}
	if screen.text != "" {
		input["filters"] = []map[string]any{{"text": screen.text}}
	}
	var result struct {
		Columns []catalogColumn `json:"columns"`
		Rows    []catalogRow    `json:"rows"`
	}
	if err := model.catalog(input, &result); err != nil {
		model.problem = err.Error()
		return model
	}
	screen.columns, screen.rows, screen.step = result.Columns, result.Rows, stepTable
	screen.table = tableOf(result.Columns, result.Rows)
	model.compendium, model.problem = &screen, ""
	return model
}

// tableOf lays the rows out as a scrollable table, each column as wide as its widest cell, up to
// a limit.
func tableOf(columns []catalogColumn, rows []catalogRow) table.Model {
	const widest = 28
	var titled []table.Column = make([]table.Column, len(columns))
	var shown []table.Row = make([]table.Row, len(rows))
	for index, column := range columns {
		titled[index] = table.Column{Title: column.Label, Width: len(column.Label)}
	}
	for rowIndex, row := range rows {
		var cells table.Row = make(table.Row, len(columns))
		for index, column := range columns {
			cells[index] = cell(lookup(row.Row, column.Field))
			titled[index].Width = min(max(titled[index].Width, len(cells[index])), widest)
		}
		shown[rowIndex] = cells
	}
	return table.New(table.WithColumns(titled), table.WithRows(shown), table.WithHeight(tableHeight), table.WithFocused(true))
}

// lookup is the value at a dotted path, nil when it is missing.
func lookup(row map[string]any, path string) any {
	var value any = row
	for _, part := range strings.Split(path, ".") {
		object, isObject := value.(map[string]any)
		if !isObject {
			return nil
		}
		value = object[part]
	}
	return value
}

// cell is a value as a table shows it: a cost as "15 gold", a list joined, an object as its
// fields, a whole number without a point.
func cell(value any) string {
	switch typed := value.(type) {
	case nil:
		return "-"
	case string:
		return typed
	case bool:
		if typed {
			return "yes"
		}
		return "no"
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case []any:
		var parts []string
		for _, item := range typed {
			parts = append(parts, cell(item))
		}
		return strings.Join(parts, ", ")
	case map[string]any:
		if quantity, priced := typed["quantity"]; priced {
			return strings.TrimSpace(cell(quantity) + " " + cell(cmpOr(typed["unit"], "")))
		}
		var fields []string
		for name := range typed {
			fields = append(fields, name)
		}
		slices.Sort(fields)
		var parts []string
		for _, name := range fields {
			parts = append(parts, name+" "+cell(typed[name]))
		}
		return strings.Join(parts, ", ")
	}
	return fmt.Sprint(value)
}

func cmpOr(value, fallback any) any {
	if value == nil {
		return fallback
	}
	return value
}

// compendiumKey runs one key on the screen.
func (model Model) compendiumKey(message tea.Msg) (tea.Model, tea.Cmd) {
	var screen compendium = *model.compendium
	key, isKey := message.(tea.KeyMsg)
	switch screen.step {
	case stepCollections:
		if !isKey {
			return model, nil
		}
		switch key.String() {
		case "up", "k":
			screen.cursor = (screen.cursor + len(screen.collections) - 1) % len(screen.collections)
		case "down", "j":
			screen.cursor = (screen.cursor + 1) % len(screen.collections)
		case "enter":
			screen.text = ""
			return model.listed(screen), nil
		case "esc":
			model.compendium, model.mode, model.cursor, model.problem = nil, modeMenu, model.returnCursor, ""
			return model, nil
		}
	case stepTable:
		if isKey {
			switch key.String() {
			case "esc":
				screen.step = stepCollections
				model.compendium, model.problem = &screen, ""
				return model, nil
			case "/":
				screen.filter = textinput.New()
				screen.filter.Prompt = "/"
				screen.filter.SetValue(screen.text)
				screen.filter.Focus()
				screen.step = stepFilter
				model.compendium = &screen
				return model, textinput.Blink
			case "enter":
				if len(screen.rows) == 0 {
					return model, nil
				}
				var picked catalogRow = screen.rows[screen.table.Cursor()]
				var result struct {
					Row catalogRow `json:"row"`
				}
				if err := model.catalog(map[string]any{"operation": catalogGet, "collection": screen.collections[screen.cursor].Name, "ref": picked.Ref}, &result); err != nil {
					model.problem = err.Error()
					return model, nil
				}
				screen.opened, screen.step = result.Row, stepRow
				model.compendium = &screen
				return model, nil
			}
		}
		var command tea.Cmd
		screen.table, command = screen.table.Update(message)
		model.compendium = &screen
		return model, command
	case stepFilter:
		if isKey {
			switch key.Type {
			case tea.KeyEsc:
				screen.step = stepTable
				model.compendium = &screen
				return model, nil
			case tea.KeyEnter:
				screen.text = strings.TrimSpace(screen.filter.Value())
				return model.listed(screen), nil
			}
		}
		var command tea.Cmd
		screen.filter, command = screen.filter.Update(message)
		model.compendium = &screen
		return model, command
	case stepRow:
		if isKey && key.String() == "esc" {
			screen.step = stepTable
		}
	}
	model.compendium = &screen
	return model, nil
}

// compendiumView draws the step the screen is at.
func (model Model) compendiumView() string {
	var screen *compendium = model.compendium
	var view strings.Builder
	switch screen.step {
	case stepCollections:
		view.WriteString(titleStyle.Render("Compendium") + "\n\n")
		for index, entry := range screen.collections {
			var marker string = "  "
			if index == screen.cursor {
				marker = "> "
			}
			fmt.Fprintf(&view, "%s%s %s\n", marker, entry.Label, dimStyle.Render(fmt.Sprintf("(%d)", entry.Count)))
		}
	case stepTable, stepFilter:
		var title string = screen.collections[screen.cursor].Label
		if screen.text != "" {
			title += " matching " + strconv.Quote(screen.text)
		}
		fmt.Fprintf(&view, "%s %s\n\n", titleStyle.Render(title), dimStyle.Render(fmt.Sprintf("(%d)", len(screen.rows))))
		if len(screen.rows) == 0 {
			view.WriteString("no rows\n")
		} else {
			view.WriteString(screen.table.View() + "\n")
		}
		if screen.step == stepFilter {
			view.WriteString("\n" + screen.filter.View() + "\n")
		}
	case stepRow:
		var label string = cell(cmpOr(screen.opened.Row["label"], screen.opened.Ref))
		fmt.Fprintf(&view, "%s\n%s\n\n", titleStyle.Render(label), dimStyle.Render(screen.opened.Ref+" from "+screen.opened.Source))
		encoded, _ := json.MarshalIndent(screen.opened.Row, "", "  ")
		view.Write(encoded)
		view.WriteString("\n")
	}
	if model.problem != "" {
		view.WriteString("\n" + errorStyle.Render(model.problem) + "\n")
	}
	view.WriteString("\n" + hintStyle.Render(compendiumHints[screen.step]) + "\n")
	return view.String()
}
