// Package play is the human CLI (docs/backend/cli.md#human-mode): a terminal UI that prompts
// for inputs, calls the same core as the agent CLI, and shows the result. It holds no rule.
package play

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Runner is core.Run: a command name and JSON in, JSON out.
type Runner func(name string, input []byte) ([]byte, error)

type mode int

const (
	modeMenu mode = iota
	modeForm
	modeResult
	modeCommand // the dev command line (docs/backend/dev-mode.md#the-human-cli)
)

const maximumLogLines = 6

var (
	titleStyle = lipgloss.NewStyle().Bold(true)
	hintStyle  = lipgloss.NewStyle().Faint(true)
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
)

// field is one prompt: its label, the default, and whether it is free text, not a number.
type field struct {
	label    string
	initial  string
	freeText bool
}

// form is what one menu entry asks for and how it becomes a command.
type form struct {
	title   string
	command string
	fields  []field
	build   func(values []string) map[string]any
	columns []string
	logLine func(values []string, result map[string]any) string
}

var attributeKeys = []string{"str", "per", "end", "int", "wil", "agi", "cha", "lck"}

func number(text string) int64 {
	value, _ := strconv.ParseInt(text, 10, 64)
	return value
}

func createForm() form {
	var fields []field
	for _, key := range attributeKeys {
		fields = append(fields, field{label: strings.ToUpper(key), initial: "5"})
	}
	fields = append(fields, field{label: "Level", initial: "1"})
	return form{
		title:   "Create a character",
		command: "derived",
		fields:  fields,
		build: func(values []string) map[string]any {
			attributes := map[string]int64{}
			for index, key := range attributeKeys {
				attributes[key] = number(values[index])
			}
			return map[string]any{"actor": map[string]any{"attributes": attributes, "level": number(values[len(attributeKeys)])}}
		},
		columns: []string{"hp", "mp", "sp", "ap", "carry", "initiative"},
		logLine: func(values []string, result map[string]any) string {
			return fmt.Sprintf("created: level %s, %v HP, %v MP", values[len(attributeKeys)], result["hp"], result["mp"])
		},
	}
}

func attackForm(luck string) form {
	return form{
		title:   "Resolve an attack",
		command: "combat",
		fields: []field{
			{label: "Attacker skill", initial: "40"},
			{label: "Attack roll (d100)", initial: "25"},
			{label: "Attacker LCK", initial: luck},
			{label: "Defender evasion skill", initial: "10"},
			{label: "Defender armor DR", initial: "0"},
			{label: "Defender HP", initial: "30"},
			{label: "Damage roll", initial: "8"},
			{label: "Damage type", initial: "slashing", freeText: true},
		},
		build: func(values []string) map[string]any {
			return map[string]any{
				"attacker": map[string]any{"skill": number(values[0]), "roll": number(values[1]), "attributes": map[string]int64{"lck": number(values[2])}},
				"defender": map[string]any{"evasion_skill": number(values[3]), "armor_dr": number(values[4]), "hp": number(values[5])},
				"damage":   map[string]any{"roll": number(values[6]), "type": values[7]},
			}
		},
		columns: []string{"outcome", "target", "margin", "crit", "damage", "hp"},
		logLine: func(values []string, result map[string]any) string {
			return fmt.Sprintf("attack: rolled %s against %v, %v; damage roll %s, %v lost", values[1], result["target"], result["outcome"], values[6], result["damage"])
		},
	}
}

// Model is the Bubble Tea model.
type Model struct {
	run        Runner
	mode       mode
	cursor     int
	current    form
	inputs     []textinput.Model
	focus      int
	problem    string
	result     table.Model
	log        []string
	luck       string
	lastInput  []byte
	lastOutput []byte
	// Dev mode: the panel's session, the command line, the queued d100 rolls, and the last
	// character created, which a new session starts from.
	dev       bool
	session   json.RawMessage
	line      textinput.Model
	rolls     []int64
	character map[string]any
}

var menuEntries = []string{"Create a character", "Resolve an attack", "Quit"}

// New returns the model at the main menu.
func New(run Runner) Model { return Model{run: run, luck: "5"} }

// NewDev is New with the dev panel and its short commands; the core must have dev mode on.
func NewDev(run Runner) Model { return Model{run: run, luck: "5", dev: true} }

// Session is the dev panel's session as JSON, nil before the first dev command.
func (model Model) Session() []byte { return model.session }

// Init is part of tea.Model.
func (model Model) Init() tea.Cmd { return textinput.Blink }

// LastCall is the input and output of the last command, so a test can compare it.
func (model Model) LastCall() (input, output []byte) { return model.lastInput, model.lastOutput }

func (model Model) open(current form) Model {
	model.current, model.mode, model.focus, model.problem = current, modeForm, 0, ""
	model.inputs = make([]textinput.Model, len(current.fields))
	for index, entry := range current.fields {
		input := textinput.New()
		input.Prompt = ""
		input.SetValue(entry.initial)
		model.inputs[index] = input
	}
	model.inputs[0].Focus()
	return model
}

func (model Model) moveFocus(step int) Model {
	model.inputs[model.focus].Blur()
	model.focus = (model.focus + step + len(model.inputs)) % len(model.inputs)
	model.inputs[model.focus].Focus()
	return model
}

// submit validates the fields, calls the core, and shows the result or the refusal.
func (model Model) submit() Model {
	values := make([]string, len(model.inputs))
	for index, input := range model.inputs {
		values[index] = strings.TrimSpace(input.Value())
		if _, err := strconv.ParseInt(values[index], 10, 64); err != nil && !model.current.fields[index].freeText {
			model.problem = model.current.fields[index].label + " must be a whole number"
			return model
		}
	}
	input, err := json.Marshal(model.current.build(values))
	if err != nil {
		model.problem = err.Error()
		return model
	}
	output, err := model.run(model.current.command, input)
	model.lastInput, model.lastOutput = input, output
	if err != nil {
		model.problem = err.Error()
		return model
	}
	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		model.problem = err.Error()
		return model
	}
	columns, row := make([]table.Column, 0, len(model.current.columns)), make(table.Row, 0, len(model.current.columns))
	for _, name := range model.current.columns {
		cell := "-"
		if value, present := result[name]; present && value != nil {
			cell = fmt.Sprint(value)
		}
		columns = append(columns, table.Column{Title: name, Width: max(len(name), len(cell)) + 1})
		row = append(row, cell)
	}
	model.result = table.New(table.WithColumns(columns), table.WithRows([]table.Row{row}), table.WithHeight(2))
	model.log = append(model.log, model.current.logLine(values, result))
	if len(model.log) > maximumLogLines {
		model.log = model.log[len(model.log)-maximumLogLines:]
	}
	if model.current.command == "derived" {
		model.luck = values[7]
		model.character = model.current.build(values)["actor"].(map[string]any)
	}
	model.mode, model.problem = modeResult, ""
	return model
}

// Update is part of tea.Model.
func (model Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	key, isKey := message.(tea.KeyMsg)
	if isKey && key.Type == tea.KeyCtrlC {
		return model, tea.Quit
	}
	switch model.mode {
	case modeMenu:
		if !isKey {
			return model, nil
		}
		switch key.String() {
		case "up", "k":
			model.cursor = (model.cursor + len(menuEntries) - 1) % len(menuEntries)
		case "down", "j":
			model.cursor = (model.cursor + 1) % len(menuEntries)
		case "q":
			return model, tea.Quit
		case ":":
			if model.dev {
				model.line = textinput.New()
				model.line.Prompt = ":"
				model.line.Focus()
				model.mode, model.problem = modeCommand, ""
				return model, textinput.Blink
			}
		case "enter":
			switch model.cursor {
			case 0:
				return model.open(createForm()), textinput.Blink
			case 1:
				var attack form = attackForm(model.luck)
				if len(model.rolls) > 0 {
					attack.fields[1].initial = strconv.FormatInt(model.rolls[0], 10)
					model.rolls = model.rolls[1:]
				}
				return model.open(attack), textinput.Blink
			}
			return model, tea.Quit
		}
		return model, nil
	case modeCommand:
		if isKey {
			switch key.Type {
			case tea.KeyEsc:
				model.mode = modeMenu
				return model, nil
			case tea.KeyEnter:
				return model.command(strings.TrimSpace(model.line.Value())), nil
			}
		}
		var command tea.Cmd
		model.line, command = model.line.Update(message)
		return model, command
	case modeForm:
		if isKey {
			switch key.Type {
			case tea.KeyEsc:
				model.mode = modeMenu
				return model, nil
			case tea.KeyTab, tea.KeyDown:
				return model.moveFocus(1), textinput.Blink
			case tea.KeyShiftTab, tea.KeyUp:
				return model.moveFocus(-1), textinput.Blink
			case tea.KeyEnter:
				if model.focus == len(model.inputs)-1 {
					return model.submit(), nil
				}
				return model.moveFocus(1), textinput.Blink
			}
		}
		var command tea.Cmd
		model.inputs[model.focus], command = model.inputs[model.focus].Update(message)
		return model, command
	}
	if isKey && (key.Type == tea.KeyEnter || key.Type == tea.KeyEsc) {
		model.mode = modeMenu
	}
	return model, nil
}

func (model Model) logView() string {
	if len(model.log) == 0 {
		return ""
	}
	return "\n" + titleStyle.Render("Log") + "\n" + strings.Join(model.log, "\n") + "\n"
}

// devOperation turns a short command into a dev operation, or returns why it cannot.
func devOperation(words []string) (map[string]any, error) {
	var count = func(text string) (int64, error) {
		value, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("%q is not a whole number", text)
		}
		return value, nil
	}
	switch {
	case words[0] == "give" && (len(words) == 2 || len(words) == 3):
		operation := map[string]any{"operation": "item", "base": words[1]}
		if len(words) == 3 {
			quantity, err := count(words[2])
			if err != nil {
				return nil, err
			}
			operation["quantity"] = quantity
		}
		return operation, nil
	case words[0] == "inflict" && len(words) >= 2 && strings.HasPrefix(words[1], "core:effect/") && (len(words) == 4 || len(words) == 5):
		magnitude, err := count(words[2])
		if err != nil {
			return nil, err
		}
		operation := map[string]any{"operation": "effect", "effect": words[1], "magnitude": magnitude, "origin": words[3]}
		if len(words) == 5 {
			turns, err := count(words[4])
			if err != nil {
				return nil, err
			}
			operation["turns"] = turns
		}
		return operation, nil
	case words[0] == "inflict" && (len(words) == 2 || len(words) == 3):
		var stacks int64 = 1
		if len(words) == 3 {
			var err error
			if stacks, err = count(words[2]); err != nil {
				return nil, err
			}
		}
		return map[string]any{"operation": "condition", "key": words[1], "stacks": stacks}, nil
	case words[0] == "level" && len(words) == 2:
		level, err := count(words[1])
		if err != nil {
			return nil, err
		}
		return map[string]any{"operation": "level", "set": level}, nil
	}
	return nil, fmt.Errorf("unknown command; try :give, :inflict, :level, or :roll")
}

// startActor is the stored actor a new panel session starts from: the last character created, or
// every attribute at 5 at level 1, with full pools from the derived command.
func (model Model) startActor() (map[string]any, error) {
	var character map[string]any = model.character
	if character == nil {
		attributes := map[string]int64{}
		for _, key := range attributeKeys {
			attributes[key] = 5
		}
		character = map[string]any{"attributes": attributes, "level": int64(1)}
	}
	input, _ := json.Marshal(map[string]any{"actor": character})
	output, err := model.run("derived", input)
	if err != nil {
		return nil, err
	}
	var derived map[string]any
	if err := json.Unmarshal(output, &derived); err != nil {
		return nil, err
	}
	return map[string]any{
		"attributes": character["attributes"], "level": character["level"], "xp": 0, "skills": map[string]any{}, "tags": []any{},
		"growth":     map[string]any{"hp": 0, "mp": 0, "sp": 0},
		"resources":  map[string]any{"hp": derived["hp"], "mp": derived["mp"], "sp": derived["sp"], "ap": derived["ap"]},
		"conditions": []any{}, "modifiers": []any{}, "flags": []any{}, "inventory": []any{}, "equipped": map[string]any{},
	}, nil
}

// command runs one dev short command and returns to the menu, or stays with the problem shown.
func (model Model) command(line string) Model {
	words := strings.Fields(line)
	if len(words) == 0 {
		model.mode = modeMenu
		return model
	}
	if words[0] == "roll" {
		if len(words) == 1 {
			model.problem = "roll takes one or more d100 values"
			return model
		}
		var queued []int64
		for _, word := range words[1:] {
			value, err := strconv.ParseInt(word, 10, 64)
			if err != nil || value < 1 || value > 100 {
				model.problem = fmt.Sprintf("%q is not a d100 roll (1 to 100)", word)
				return model
			}
			queued = append(queued, value)
		}
		model.rolls = append(model.rolls, queued...)
		return model.logged("dev: queued rolls " + strings.Join(words[1:], " "))
	}
	operation, err := devOperation(words)
	if err != nil {
		model.problem = err.Error()
		return model
	}
	if model.session == nil {
		actor, err := model.startActor()
		if err != nil {
			model.problem = err.Error()
			return model
		}
		input, _ := json.Marshal(map[string]any{"operation": "start", "start": actor})
		output, err := model.run("session", input)
		if err != nil {
			model.problem = err.Error()
			return model
		}
		var started map[string]json.RawMessage
		_ = json.Unmarshal(output, &started)
		model.session = started["session"]
	}
	operation["session"] = model.session
	input, _ := json.Marshal(operation)
	output, err := model.run("dev", input)
	model.lastInput, model.lastOutput = input, output
	if err != nil {
		model.problem = err.Error()
		return model
	}
	var result map[string]json.RawMessage
	if err := json.Unmarshal(output, &result); err != nil {
		model.problem = err.Error()
		return model
	}
	model.session = result["session"]
	return model.logged("dev: " + line)
}

func (model Model) logged(line string) Model {
	model.log = append(model.log, line)
	if len(model.log) > maximumLogLines {
		model.log = model.log[len(model.log)-maximumLogLines:]
	}
	model.mode, model.problem = modeMenu, ""
	return model
}

// panelView is the dev panel: the session's actor, and the queued rolls.
func (model Model) panelView() string {
	var view strings.Builder
	view.WriteString("\n" + titleStyle.Render("Dev") + "\n")
	if model.session == nil {
		view.WriteString("no session yet; : runs a dev command\n")
	} else {
		var session struct {
			Actor struct {
				Level      int64            `json:"level"`
				XP         int64            `json:"xp"`
				Resources  map[string]int64 `json:"resources"`
				Conditions []struct {
					Key    string `json:"key"`
					Stacks int64  `json:"stacks"`
				} `json:"conditions"`
				Modifiers []any `json:"modifiers"`
				Inventory []any `json:"inventory"`
			} `json:"actor"`
		}
		_ = json.Unmarshal(model.session, &session)
		var actor = session.Actor
		fmt.Fprintf(&view, "level %d, %d XP; HP %d, MP %d, SP %d, AP %d\n", actor.Level, actor.XP,
			actor.Resources["hp"], actor.Resources["mp"], actor.Resources["sp"], actor.Resources["ap"])
		var conditions []string
		for _, condition := range actor.Conditions {
			conditions = append(conditions, fmt.Sprintf("%s %d", condition.Key, condition.Stacks))
		}
		fmt.Fprintf(&view, "conditions: %s; effects held: %d; items: %d\n", cmpOrNone(strings.Join(conditions, ", ")),
			len(actor.Modifiers), len(actor.Inventory))
	}
	if len(model.rolls) > 0 {
		var queued []string
		for _, roll := range model.rolls {
			queued = append(queued, strconv.FormatInt(roll, 10))
		}
		view.WriteString("queued rolls: " + strings.Join(queued, " ") + "\n")
	}
	return view.String()
}

func cmpOrNone(text string) string {
	if text == "" {
		return "none"
	}
	return text
}

// View is part of tea.Model.
func (model Model) View() string {
	var view strings.Builder
	switch model.mode {
	case modeMenu:
		view.WriteString(titleStyle.Render("Riftweave") + "\n\n")
		for index, entry := range menuEntries {
			marker := "  "
			if index == model.cursor {
				marker = "> "
			}
			view.WriteString(marker + entry + "\n")
		}
		view.WriteString(model.logView())
		if model.dev {
			view.WriteString(model.panelView())
			view.WriteString("\n" + hintStyle.Render("up/down, enter, : for a dev command, q to quit") + "\n")
		} else {
			view.WriteString("\n" + hintStyle.Render("up/down, enter, q to quit") + "\n")
		}
	case modeCommand:
		view.WriteString(titleStyle.Render("Dev command") + "\n\n" + model.line.View() + "\n")
		if model.problem != "" {
			view.WriteString("\n" + errorStyle.Render(model.problem) + "\n")
		}
		view.WriteString("\n" + hintStyle.Render(":give <item ref> [quantity], :inflict <condition> [stacks] or <effect ref> <magnitude> <origin ref> [turns], :level <n>, :roll <d100>...; enter to run, esc for the menu") + "\n")
	case modeForm:
		view.WriteString(titleStyle.Render(model.current.title) + "\n\n")
		for index, entry := range model.current.fields {
			fmt.Fprintf(&view, "%-24s %s\n", entry.label, model.inputs[index].View())
		}
		if model.problem != "" {
			view.WriteString("\n" + errorStyle.Render(model.problem) + "\n")
		}
		view.WriteString("\n" + hintStyle.Render("tab or enter for the next field, enter on the last to run, esc for the menu") + "\n")
	default:
		view.WriteString(titleStyle.Render(model.current.title) + "\n\n" + model.result.View() + "\n")
		view.WriteString(model.logView())
		view.WriteString("\n" + hintStyle.Render("enter for the menu") + "\n")
	}
	return view.String()
}
