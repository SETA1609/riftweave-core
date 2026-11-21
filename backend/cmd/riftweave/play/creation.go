package play

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// The creation screen (BE_S07_08): the draft's steps as rows, checked by the `creation`
// command after every change, and finished into a stored actor. It holds no rule: the races,
// backgrounds, points, and limits come from `rules`, and every problem from `check`.

// Rows of the creation screen: name, race, background, the eight attributes, creation perks,
// tag skills, and finish.
const (
	rowName = iota
	rowRace
	rowBackground
	rowFirstAttribute
	rowPerks  = rowFirstAttribute + 8
	rowTags   = rowPerks + 1
	rowFinish = rowTags + 1
	rowCount  = rowFinish + 1
)

type choice struct {
	Ref   string
	Label string
}

// creator is the screen's state: the choices from `rules`, the draft, and the last check.
type creator struct {
	races, backgrounds []choice
	row                int
	name, perks, tags  string
	race, background   int // -1 until chosen
	points             map[string]int64
	check              struct {
		Problems []struct {
			Step    string `json:"step"`
			Message string `json:"message"`
		} `json:"problems"`
		PointsLeft int64            `json:"points_left"`
		Attributes map[string]int64 `json:"attributes"`
		Derived    map[string]int64 `json:"derived"`
	}
}

// refs turns typed keys, separated by spaces or commas, into refs of a category.
func refs(text, category string) []string {
	var result []string = []string{}
	for _, word := range strings.FieldsFunc(text, func(r rune) bool { return r == ' ' || r == ',' }) {
		if !strings.HasPrefix(word, "core:") {
			word = "core:" + category + "/" + word
		}
		result = append(result, word)
	}
	return result
}

func (screen *creator) draft() map[string]any {
	var draft map[string]any = map[string]any{"points": screen.points,
		"creation_perks": refs(screen.perks, "feature"), "tag_skills": refs(screen.tags, "skill")}
	if screen.name != "" {
		draft["name"] = screen.name
	}
	if screen.race >= 0 {
		draft["race"] = screen.races[screen.race].Ref
	}
	if screen.background >= 0 {
		draft["background"] = screen.backgrounds[screen.background].Ref
	}
	return draft
}

// openCreator reads the rules and opens the screen on an empty draft.
func (model Model) openCreator() Model {
	output, err := model.run("creation", []byte(`{"operation":"rules"}`))
	if err != nil {
		model.problem = err.Error()
		return model
	}
	var rules struct {
		Races []struct {
			Race  string `json:"race"`
			Label string `json:"label"`
		} `json:"races"`
		Backgrounds []struct {
			Background string `json:"background"`
			Label      string `json:"label"`
		} `json:"backgrounds"`
	}
	if err := json.Unmarshal(output, &rules); err != nil {
		model.problem = err.Error()
		return model
	}
	var screen creator = creator{race: -1, background: -1, points: map[string]int64{}}
	for _, race := range rules.Races {
		screen.races = append(screen.races, choice{race.Race, race.Label})
	}
	for _, background := range rules.Backgrounds {
		screen.backgrounds = append(screen.backgrounds, choice{background.Background, background.Label})
	}
	model.creator, model.mode, model.problem = &screen, modeCreator, ""
	return model.recheck()
}

// recheck runs `check` on the draft and keeps what it says.
func (model Model) recheck() Model {
	var screen creator = *model.creator
	encoded, _ := json.Marshal(map[string]any{"operation": "check", "draft": screen.draft()})
	output, err := model.run("creation", encoded)
	if err == nil {
		err = json.Unmarshal(output, &screen.check)
	}
	model.problem = ""
	if err != nil {
		model.problem = err.Error()
	}
	model.creator = &screen
	return model
}

// cycle moves a choice by one, from none through every row and back to none.
func cycle(index, step, count int) int {
	return (index+1+step+count+1)%(count+1) - 1
}

// rowKind is what a row of the creation screen holds, and so which keys edit it.
type rowKind uint8

const (
	kindText      rowKind = iota // name, creation perks, tag skills: typed
	kindChoice                   // race, background: left and right cycle
	kindAttribute                // the eight attributes: + and -, or right and left
	kindFinish                   // enter sends the draft to finish
)

// rowKinds is the screen as a state table: a row's kind is one array read.
var rowKinds = [rowCount]rowKind{
	rowName: kindText, rowRace: kindChoice, rowBackground: kindChoice,
	rowFirstAttribute: kindAttribute, rowFirstAttribute + 1: kindAttribute, rowFirstAttribute + 2: kindAttribute,
	rowFirstAttribute + 3: kindAttribute, rowFirstAttribute + 4: kindAttribute, rowFirstAttribute + 5: kindAttribute,
	rowFirstAttribute + 6: kindAttribute, rowFirstAttribute + 7: kindAttribute,
	rowPerks: kindText, rowTags: kindText, rowFinish: kindFinish,
}

// creatorKey moves between rows on the keys every row shares, then hands the key to the row's
// kind. Each switch is on an integer (the key type, the row kind), so it compiles to a jump table
// or a short binary search, never a chain of tests.
func (model Model) creatorKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	var screen creator = *model.creator
	switch key.Type {
	case tea.KeyEsc:
		model.mode, model.creator, model.problem = modeMenu, nil, ""
		return model, nil
	case tea.KeyUp:
		screen.row = (screen.row + rowCount - 1) % rowCount
		return model.edited(screen)
	case tea.KeyDown, tea.KeyTab:
		screen.row = (screen.row + 1) % rowCount
		return model.edited(screen)
	}
	var changed bool
	switch rowKinds[screen.row] {
	case kindText:
		changed = screen.typed(key)
	case kindChoice:
		changed = screen.chose(key)
	case kindAttribute:
		changed = screen.spent(key)
	case kindFinish:
		if key.Type == tea.KeyEnter {
			return model.finishCreation(), nil
		}
	}
	if !changed {
		return model, nil
	}
	return model.edited(screen)
}

func (model Model) edited(screen creator) (tea.Model, tea.Cmd) {
	model.creator = &screen
	return model.recheck(), nil
}

// text is the typed value of a text row.
func (screen *creator) text() *string {
	switch screen.row {
	case rowName:
		return &screen.name
	case rowPerks:
		return &screen.perks
	}
	return &screen.tags
}

func (screen *creator) typed(key tea.KeyMsg) bool {
	var text *string = screen.text()
	switch key.Type {
	case tea.KeyEnter:
		screen.row++
	case tea.KeyBackspace:
		if *text == "" {
			return false
		}
		*text = (*text)[:len(*text)-1]
	case tea.KeySpace:
		*text += " "
	case tea.KeyRunes:
		*text += string(key.Runes)
	default:
		return false
	}
	return true
}

func (screen *creator) chose(key tea.KeyMsg) bool {
	var step int
	switch key.Type {
	case tea.KeyEnter:
		screen.row++
		return true
	case tea.KeyLeft:
		step = -1
	case tea.KeyRight:
		step = 1
	default:
		return false
	}
	if screen.row == rowRace {
		screen.race = cycle(screen.race, step, len(screen.races))
	} else {
		screen.background = cycle(screen.background, step, len(screen.backgrounds))
	}
	return true
}

func (screen *creator) spent(key tea.KeyMsg) bool {
	var step int64
	switch key.Type {
	case tea.KeyEnter:
		screen.row++
		return true
	case tea.KeyRight:
		step = 1
	case tea.KeyLeft:
		step = -1
	case tea.KeyRunes:
		switch key.String() {
		case "+":
			step = 1
		case "-":
			step = -1
		default:
			return false
		}
	default:
		return false
	}
	screen.points = clonePoints(screen.points)
	screen.points[attributeKeys[screen.row-rowFirstAttribute]] += step
	return true
}

func clonePoints(points map[string]int64) map[string]int64 {
	var copied map[string]int64 = make(map[string]int64, len(points)+1)
	for key, value := range points {
		copied[key] = value
	}
	return copied
}

// finishCreation sends the draft to `finish`; a stored actor becomes the character a dev
// session and the attack form start from, and a refusal stays on the screen.
func (model Model) finishCreation() Model {
	encoded, _ := json.Marshal(map[string]any{"operation": "finish", "draft": model.creator.draft()})
	output, err := model.run("creation", encoded)
	model.lastInput, model.lastOutput = encoded, output
	var result struct {
		Error string         `json:"error"`
		Name  string         `json:"name"`
		Actor map[string]any `json:"actor"`
	}
	if decodeErr := json.Unmarshal(output, &result); err != nil || decodeErr != nil {
		model.problem = result.Error
		if model.problem == "" && err != nil {
			model.problem = err.Error()
		}
		return model
	}
	var attributes, _ = result.Actor["attributes"].(map[string]any)
	var resources, _ = result.Actor["resources"].(map[string]any)
	model.character = result.Actor
	model.luck = fmt.Sprint(attributes["lck"])
	model.creator = nil
	return model.logged(fmt.Sprintf("created: %s, level 1, %v HP", result.Name, resources["hp"]))
}

func (model Model) creatorView() string {
	var screen *creator = model.creator
	var view strings.Builder
	view.WriteString(titleStyle.Render("Create a character") + "\n\n")
	var row = func(index int, label, value string) {
		marker := "  "
		if index == screen.row {
			marker = "> "
		}
		fmt.Fprintf(&view, "%s%-16s %s\n", marker, label, value)
	}
	var pick = func(choices []choice, index int) string {
		if index < 0 {
			return dimStyle.Render("none, left/right to choose")
		}
		return "< " + choices[index].Label + " >"
	}
	row(rowName, "Name", screen.name)
	row(rowRace, "Race", pick(screen.races, screen.race))
	row(rowBackground, "Background", pick(screen.backgrounds, screen.background))
	for index, key := range attributeKeys {
		row(rowFirstAttribute+index, strings.ToUpper(key), fmt.Sprintf("%d (%+d)", screen.check.Attributes[key], screen.points[key]))
	}
	row(rowPerks, "Creation perks", screen.perks)
	row(rowTags, "Tag skills", screen.tags)
	row(rowFinish, "Finish", "")
	fmt.Fprintf(&view, "\n%d points left; HP %d, MP %d, SP %d, AP %d, initiative %d\n", screen.check.PointsLeft,
		screen.check.Derived["hp"], screen.check.Derived["mp"], screen.check.Derived["sp"], screen.check.Derived["ap"], screen.check.Derived["initiative"])
	for _, problem := range screen.check.Problems {
		view.WriteString(dimStyle.Render(problem.Step+": ") + problem.Message + "\n")
	}
	if model.problem != "" {
		view.WriteString("\n" + errorStyle.Render(model.problem) + "\n")
	}
	view.WriteString("\n" + hintStyle.Render("up/down for a step; left/right picks; + and - change an attribute; type keys for perks and tag skills; enter on Finish creates; esc for the menu") + "\n")
	return view.String()
}
