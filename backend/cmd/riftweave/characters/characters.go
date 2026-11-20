// Package characters is the character list (BE_S07_03, docs/backend/files.md#the-character-list):
// the saved characters, each a character file in one directory, and the sync rule that decides
// what an imported file does to the saved one. A GM loads characters into a session's party and
// exports them back, one file per player. It holds no game rule.
package characters

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/SETA1609/riftweave-core/backend/cmd/riftweave/files"
)

// Library is the saved characters of one device: a directory of character files named by id.
type Library struct {
	Directory    string // os.UserConfigDir()/riftweave/characters in the CLI
	Checker      files.Checker
	RulesVersion string
	Modules      []files.Module   // the modules enabled here, at their manifests' versions
	Now          func() time.Time // nil is time.Now
	NewID        func() string    // nil is a random UUID
}

// Entry is one row of the list: the sync fields and the level, which the list shows. A saved
// file that cannot be read is listed with its id and the problem, so one broken file neither
// hides the others nor stays out of reach of Delete.
type Entry struct {
	files.Character
	Level   int64  `json:"level"`
	Problem string `json:"problem,omitempty"`
}

// Saved is one character as the list holds it: its sync fields and its stored actor.
type Saved struct {
	Character files.Character `json:"character"`
	Data      json.RawMessage `json:"data"`
}

// The outcomes of an import (docs/backend/files.md#the-sync-rule).
const (
	Added     = "added"     // no saved character has the id
	Replaced  = "replaced"  // a higher revision, or the incoming side kept on a divergence
	Unchanged = "unchanged" // the same revision with the same content
	Kept      = "kept"      // the saved side kept on a divergence
	Refused   = "refused"   // a lower revision
	Diverged  = "diverged"  // the same revision with other content: the person picks a side
)

// Report is what an import did. A refused or diverged import changes nothing; a diverged one
// carries both sides, so the person can see them and pick one.
type Report struct {
	Outcome   string          `json:"outcome"`
	Character files.Character `json:"character"`
	Note      string          `json:"note,omitempty"`
	Saved     *Saved          `json:"saved,omitempty"`
	Incoming  *Saved          `json:"incoming,omitempty"`
	Warnings  []string        `json:"warnings"`
}

// ImportOptions are the choices an import takes: the name and the author of a file that has
// no sync fields yet, the side to keep on a divergence ("saved" or "incoming"), and whether to
// go on past a module this device does not load.
type ImportOptions struct {
	Name       string
	ChangedBy  string
	Keep       string
	AcceptLoss bool
}

// MissingModulesError stops an import that would lose the rows of a module not loaded here.
type MissingModulesError struct{ Missing []files.Missing }

func (problem *MissingModulesError) Error() string { return files.Alert(problem.Missing) }

// ErrNotFound is a character id that is not in the list.
var ErrNotFound = errors.New("no saved character has that id")

func (library Library) now() string {
	var clock = library.Now
	if clock == nil {
		clock = time.Now
	}
	return clock().UTC().Format(time.RFC3339)
}

func (library Library) newID() string {
	if library.NewID != nil {
		return library.NewID()
	}
	var bytes [16]byte
	_, _ = rand.Read(bytes[:])
	bytes[6] = bytes[6]&0x0f | 0x40 // version 4
	bytes[8] = bytes[8]&0x3f | 0x80 // RFC 4122 variant
	return fmt.Sprintf("%x-%x-%x-%x-%x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
}

var idPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func (library Library) path(id string) (string, error) {
	if !idPattern.MatchString(id) {
		return "", fmt.Errorf("%q is not a character id (a lowercase UUID)", id)
	}
	return filepath.Join(library.Directory, id+".riftweave.json"), nil
}

// open reads one saved file as it was stored, with the rules and modules it was made with.
func (library Library) open(id string) (files.Read, error) {
	path, err := library.path(id)
	if err != nil {
		return files.Read{}, err
	}
	content, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return files.Read{}, fmt.Errorf("%s: %w", id, ErrNotFound)
	}
	if err != nil {
		return files.Read{}, err
	}
	read, err := files.Open(content, library.RulesVersion, library.Modules, library.Checker)
	if err != nil {
		return files.Read{}, fmt.Errorf("%s: %v", path, err)
	}
	if read.Kind != "character" || read.Character == nil || read.Character.ID != id {
		return files.Read{}, fmt.Errorf("%s: not a saved character file", path)
	}
	return read, nil
}

// store writes a character file into the list, keeping the rules and modules it was made with.
func (library Library) store(character files.Character, data json.RawMessage, rulesVersion string, modules []files.Module) error {
	encoded, err := files.WrapCharacter("character", &character, data, rulesVersion, modules, library.Checker)
	if err != nil {
		return err
	}
	path, err := library.path(character.ID)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(library.Directory, 0o755); err != nil {
		return err
	}
	return writeAtomically(path, encoded)
}

// writeAtomically writes a file through a temporary file in the same directory and a rename, so a
// crash mid-write leaves the old file whole instead of a broken one the list cannot read.
func writeAtomically(path string, content []byte) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), ".saving-*")
	if err != nil {
		return err
	}
	if _, err := temporary.Write(content); err != nil {
		temporary.Close()
		os.Remove(temporary.Name())
		return err
	}
	if err := temporary.Close(); err != nil {
		os.Remove(temporary.Name())
		return err
	}
	if err := os.Chmod(temporary.Name(), 0o644); err != nil {
		os.Remove(temporary.Name())
		return err
	}
	return os.Rename(temporary.Name(), path)
}

// List is every saved character, by name and then id.
func (library Library) List() ([]Entry, error) {
	var entries []Entry = []Entry{}
	names, err := filepath.Glob(filepath.Join(library.Directory, "*.riftweave.json"))
	if err != nil {
		return nil, err
	}
	for _, name := range names {
		var id string = strings.TrimSuffix(filepath.Base(name), ".riftweave.json")
		read, err := library.open(id)
		if err != nil {
			entries = append(entries, Entry{Character: files.Character{ID: id}, Problem: err.Error()})
			continue
		}
		var actor struct {
			Level int64 `json:"level"`
		}
		_ = json.Unmarshal(read.Data, &actor)
		entries = append(entries, Entry{Character: *read.Character, Level: actor.Level})
	}
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Name != entries[j].Name {
			return entries[i].Name < entries[j].Name
		}
		return entries[i].ID < entries[j].ID
	})
	return entries, nil
}

// Open is one saved character.
func (library Library) Open(id string) (Saved, error) {
	read, err := library.open(id)
	if err != nil {
		return Saved{}, err
	}
	return Saved{Character: *read.Character, Data: read.Data}, nil
}

// Export is the saved character as a file, stamped with these rules and the modules enabled
// here (APP-30). Exporting changes nothing, so the revision stays.
func (library Library) Export(id string) ([]byte, error) {
	saved, err := library.Open(id)
	if err != nil {
		return nil, err
	}
	return files.WrapCharacter("character", &saved.Character, saved.Data, library.RulesVersion, library.Modules, library.Checker)
}

// sameContent is whether two sides hold the same character: the same name and the same actor.
// Who changed it and when are not content.
func sameContent(left, right Saved) bool {
	return left.Character.Name == right.Character.Name && bytes.Equal(canonical(left.Data), canonical(right.Data))
}

func canonical(data json.RawMessage) []byte {
	var value any
	if json.Unmarshal(data, &value) != nil {
		return data
	}
	return files.Encode(value)
}

// Import applies a character file to the list by the sync rule (docs/backend/files.md#the-sync-rule):
// a new id is added; a higher revision replaces the saved one; a lower one is refused with a
// note; the same revision with the same content changes nothing, and with other content is a
// divergence that shows both sides and changes nothing until the person keeps one. It never
// merges field by field. A file without sync fields is a new character, revision 1.
func (library Library) Import(content []byte, options ImportOptions) (Report, error) {
	read, err := files.Open(content, library.RulesVersion, library.Modules, library.Checker)
	if err != nil {
		return Report{}, err
	}
	if read.Kind != "character" {
		return Report{}, fmt.Errorf("a %s file is not a character", read.Kind)
	}
	if len(read.Missing) > 0 && !options.AcceptLoss {
		return Report{}, &MissingModulesError{Missing: read.Missing}
	}
	var report Report = Report{Warnings: read.Warnings}
	if len(read.Missing) > 0 {
		report.Warnings = append(report.Warnings, "imported without its modules: "+files.Alert(read.Missing))
	}
	if options.Keep != "" && options.Keep != "saved" && options.Keep != "incoming" {
		return Report{}, fmt.Errorf("keep %q is not saved or incoming", options.Keep)
	}
	if read.Character == nil {
		var by string = options.ChangedBy
		if by == "" {
			by = "player"
		}
		var name string = strings.TrimSpace(options.Name)
		if name == "" {
			name = "Unnamed character"
		}
		var character = files.Character{ID: library.newID(), Name: name, Revision: 1, ChangedBy: by, ChangedAt: library.now()}
		if err := library.store(character, read.Data, read.RulesVersion, read.Modules); err != nil {
			return Report{}, err
		}
		report.Outcome, report.Character = Added, character
		return report, nil
	}
	var incoming = Saved{Character: *read.Character, Data: read.Data}
	report.Character = incoming.Character
	saved, err := library.Open(incoming.Character.ID)
	switch {
	case errors.Is(err, ErrNotFound):
		report.Outcome = Added
	case err != nil:
		return Report{}, err
	case incoming.Character.Revision > saved.Character.Revision:
		report.Outcome = Replaced
	case incoming.Character.Revision < saved.Character.Revision:
		report.Outcome, report.Character = Refused, saved.Character
		report.Note = fmt.Sprintf("revision %d is older than the saved revision %d; nothing changed", incoming.Character.Revision, saved.Character.Revision)
		return report, nil
	case sameContent(saved, incoming):
		report.Outcome, report.Character = Unchanged, saved.Character
		return report, nil
	case options.Keep == "incoming":
		report.Outcome = Replaced
	case options.Keep == "saved":
		report.Outcome, report.Character = Kept, saved.Character
		return report, nil
	default:
		report.Outcome, report.Character = Diverged, saved.Character
		report.Note = fmt.Sprintf("both sides changed revision %d; keep the saved or the incoming one", saved.Character.Revision)
		report.Saved, report.Incoming = &saved, &incoming
		return report, nil
	}
	if err := library.store(incoming.Character, incoming.Data, read.RulesVersion, read.Modules); err != nil {
		return Report{}, err
	}
	return report, nil
}

// Duplicate saves a copy of a character under a new id, as a new character at revision 1.
func (library Library) Duplicate(id, changedBy string) (files.Character, error) {
	read, err := library.open(id)
	if err != nil {
		return files.Character{}, err
	}
	if changedBy == "" {
		changedBy = "player"
	}
	var copied = files.Character{ID: library.newID(), Name: read.Character.Name + " (copy)", Revision: 1, ChangedBy: changedBy, ChangedAt: library.now()}
	if err := library.store(copied, read.Data, read.RulesVersion, read.Modules); err != nil {
		return files.Character{}, err
	}
	return copied, nil
}

// Delete removes a saved character, even one whose file can no longer be read.
func (library Library) Delete(id string) error {
	path, err := library.path(id)
	if err != nil {
		return err
	}
	if err := os.Remove(path); errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("%s: %w", id, ErrNotFound)
	} else if err != nil {
		return err
	}
	return nil
}

// PartyAdd loads a saved character into a session's party, as it is now: the session keeps the
// revision it was loaded at (docs/backend/sessions.md#the-party). A character already in the
// party is refused.
func (library Library) PartyAdd(session []byte, id string) ([]byte, error) {
	saved, err := library.Open(id)
	if err != nil {
		return nil, err
	}
	var value map[string]any
	if err := json.Unmarshal(session, &value); err != nil {
		return nil, fmt.Errorf("the session is not JSON: %v", err)
	}
	party, _ := value["party"].([]any)
	for _, raw := range party {
		member, _ := raw.(map[string]any)
		character, _ := member["character"].(map[string]any)
		if character["id"] == id {
			return nil, fmt.Errorf("%s (%s) is already in the party", saved.Character.Name, id)
		}
	}
	var actor any
	_ = json.Unmarshal(saved.Data, &actor)
	party = append(party, map[string]any{"character": saved.Character, "start": actor, "actor": actor})
	value["party"] = party
	var encoded []byte = files.Encode(value)
	if err := files.Check("session", encoded, library.Checker); err != nil {
		return nil, err
	}
	return encoded, nil
}

// PartyExport is the session's party as character files, one per player, named by the file
// name it should be saved as (docs/backend/sessions.md#the-party), and the session with the
// export recorded. A character whose actor changed since it was loaded or last exported goes out
// one revision higher, changed by the GM, now; one that did not change goes out as it is. The
// returned session holds each member's exported sync fields and its actor as exported (`start`),
// so a second export from the same session counts on from the first instead of sending the same
// revision with other content, which a player would see as a divergence.
func (library Library) PartyExport(session []byte) (map[string][]byte, []byte, error) {
	if err := files.Check("session", session, library.Checker); err != nil {
		return nil, nil, err
	}
	var value struct {
		Party []struct {
			Character files.Character `json:"character"`
			Start     json.RawMessage `json:"start"`
			Actor     json.RawMessage `json:"actor"`
		} `json:"party"`
	}
	if err := json.Unmarshal(session, &value); err != nil {
		return nil, nil, err
	}
	if len(value.Party) == 0 {
		return nil, nil, errors.New("the session has no party")
	}
	var whole map[string]any
	if err := json.Unmarshal(session, &whole); err != nil {
		return nil, nil, err
	}
	party, _ := whole["party"].([]any)
	var exported = map[string][]byte{}
	for index, member := range value.Party {
		var character files.Character = member.Character
		if !bytes.Equal(canonical(member.Start), canonical(member.Actor)) {
			character.Revision++
			character.ChangedBy, character.ChangedAt = "gm", library.now()
		}
		encoded, err := files.WrapCharacter("character", &character, member.Actor, library.RulesVersion, library.Modules, library.Checker)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %v", character.Name, err)
		}
		var name string = FileName(character)
		if _, taken := exported[name]; taken {
			return nil, nil, fmt.Errorf("two party members would be saved as %s", name)
		}
		exported[name] = encoded
		if index >= len(party) {
			continue
		}
		if recorded, ok := party[index].(map[string]any); ok {
			var actor any
			_ = json.Unmarshal(member.Actor, &actor)
			recorded["character"], recorded["start"] = character, actor
		}
	}
	var updated []byte = files.Encode(whole)
	if err := files.Check("session", updated, library.Checker); err != nil {
		return nil, nil, err
	}
	return exported, updated, nil
}

var notInFileName = regexp.MustCompile(`[^a-z0-9]+`)

// FileName is the name a character's file is saved as: its name in lowercase words, and the
// start of its id, so two characters of one name do not overwrite each other.
func FileName(character files.Character) string {
	var slug string = strings.Trim(notInFileName.ReplaceAllString(strings.ToLower(character.Name), "-"), "-")
	if slug == "" {
		slug = "character"
	}
	return slug + "-" + character.ID[:8] + ".riftweave.json"
}
