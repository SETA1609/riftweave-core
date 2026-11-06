// Package data loads ruleset/data/ into typed, flat tables, once, at start.
//
// Data-oriented means the rules never walk raw JSON: every collection a command reads is a
// dense slice of structs with a key index beside it, every cross-reference is resolved to
// an index or a bit mask when the table is built, and every formula is compiled to a flat
// program. A command then does array reads and integer math. The tables hold numbers and
// names only; the rules live in the packages that read them.
//
// The loader takes an fs.FS rooted at the data directory, so a module's data can be layered
// over core by a later issue without changing a command.
package data

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"math"
	"path"
	"slices"
	"strings"

	"github.com/SETA1609/riftweave-core/backend/internal/actor"
	"github.com/SETA1609/riftweave-core/backend/internal/formula"
	"github.com/SETA1609/riftweave-core/backend/internal/modifier"
)

// Table is a dense slice of rows with an index from key to row.
type Table[T any] struct {
	rows  []T
	index map[string]int32
}

func newTable[T any](rows []T, key func(*T) string) Table[T] {
	t := Table[T]{rows: rows, index: make(map[string]int32, len(rows))}
	for i := range rows {
		t.index[key(&rows[i])] = int32(i)
	}
	return t
}

// Get returns the row with the key.
func (t *Table[T]) Get(key string) (*T, bool) {
	i, ok := t.index[key]
	if !ok {
		return nil, false
	}
	return &t.rows[i], true
}

// Index returns the row index of a key, for callers that keep a slice aligned with the rows.
func (t *Table[T]) Index(key string) (int32, bool) {
	i, ok := t.index[key]
	return i, ok
}

// Rows are the rows in file order. Callers must not change them.
func (t *Table[T]) Rows() []T { return t.rows }

// At returns the row at an index a loader resolved.
func (t *Table[T]) At(i int32) *T { return &t.rows[i] }

// Cost is a price in a denomination.
type Cost struct {
	Quantity int64  `json:"quantity"`
	Unit     string `json:"unit"`
}

// Resolution is ruleset/data/resolution/core.json.
type Resolution struct {
	Die                  int64  `json:"die"`
	CritWindow           int64  `json:"crit_window"`
	CritWindowPerTwoLuck int64  `json:"crit_window_per_two_luck"`
	FumbleAt             int64  `json:"fumble_at"`
	TargetCap            *int64 `json:"target_cap"`
}

// DerivedStat is one row of derived/core.json.
type DerivedStat struct {
	Key             string        `json:"key"`
	Attr            string        `json:"attr"`
	Base            int64         `json:"base"`
	PerPoint        int64         `json:"per_point"`
	PerPointDivisor int64         `json:"per_point_divisor"`
	PerLevelAttr    int64         `json:"per_level_attr"`
	PerLevelFlat    int64         `json:"per_level_flat"`
	AttrID          actor.Ability `json:"-"`
}

// Action is one row of actions/core.json.
type Action struct {
	Key  string `json:"key"`
	Cost int64  `json:"cost"`
}

// ActionPool is the action point pool formula's coefficients.
type ActionPool struct {
	Base            int64         `json:"base"`
	PerPoint        int64         `json:"per_point"`
	PerPointDivisor int64         `json:"per_point_divisor"`
	Attr            string        `json:"attr"`
	AttrID          actor.Ability `json:"-"`
}

// Actions is actions/core.json.
type Actions struct {
	Pool    ActionPool
	Actions Table[Action]
}

// Edge is one directed pair of a Wuxing cycle.
type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// Gate makes a cycle apply only while the actor's magnitude is greater.
type Gate struct {
	OtherwiseCycle string `json:"otherwise_cycle"`
}

// Cycle is one Wuxing cycle. Next[from] is a bit mask of the phases it points to, so an
// edge test is one AND.
type Cycle struct {
	ID          string   `json:"id"`
	Interaction string   `json:"interaction"`
	Affects     string   `json:"affects"`
	Multiplier  float64  `json:"magnitudeMultiplier"`
	Gate        *Gate    `json:"gate"`
	Edges       []Edge   `json:"edges"`
	Next        [8]uint8 `json:"-"`
}

// Has reports whether the cycle holds the edge from -> to (phase indexes).
func (c *Cycle) Has(from, to int) bool { return c.Next[from]&(1<<to) != 0 }

// Wuxing is wuxing/core.json.
type Wuxing struct {
	Phases []string
	Cycles []Cycle
}

// Phase returns the index of a phase name, or -1.
func (w *Wuxing) Phase(name string) int { return slices.Index(w.Phases, name) }

// Cycle returns the cycle with the id.
func (w *Wuxing) Cycle(id string) (*Cycle, bool) {
	for i := range w.Cycles {
		if w.Cycles[i].ID == id {
			return &w.Cycles[i], true
		}
	}
	return nil, false
}

// Recovery is one resource's rule in a rest occasion.
type Recovery struct {
	Resource    string   `json:"resource"`
	Mode        string   `json:"mode"`
	Numerator   int64    `json:"numerator"`
	Denominator int64    `json:"denominator"`
	Durations   []string `json:"durations"`
}

// Occasion is one row of rest/core.json.
type Occasion struct {
	Key      string     `json:"key"`
	Recovery []Recovery `json:"recovery"`
}

// Denomination is one coin of the currency.
type Denomination struct {
	Unit  string `json:"unit"`
	Value int64  `json:"value"`
}

// Barter is the haggling rule.
type Barter struct {
	Skill         string `json:"skill"`
	MarginDivisor int64  `json:"margin_divisor"`
	MaxShift      int64  `json:"max_shift"`
	FumblePenalty int64  `json:"fumble_penalty"`
}

// Merchant is one row of economy merchants.
type Merchant struct {
	Key         string `json:"key"`
	BuyPercent  int64  `json:"buy_percent"`
	SellPercent int64  `json:"sell_percent"`
}

// Economy is economy/core.json.
type Economy struct {
	Denominations []Denomination
	Barter        Barter
	Merchants     Table[Merchant]
}

// ToCP converts a cost to copper.
func (e *Economy) ToCP(c Cost) (int64, error) {
	for _, d := range e.Denominations {
		if d.Unit == c.Unit {
			return c.Quantity * d.Value, nil
		}
	}
	return 0, fmt.Errorf("unit %q is not a denomination", c.Unit)
}

// Progression is progression/core.json.
type Progression struct {
	StartLevel  int64           `json:"start_level"`
	MaxLevel    int64           `json:"max_level"`
	SkillCap    int64           `json:"skill_cap"`
	XPToReach   formula.Program `json:"xp_to_reach"`
	SkillPoints formula.Program `json:"skill_points"`
	LuckRoll    struct {
		Min     int64  `json:"min"`
		MaxAttr string `json:"max_attr"`
	} `json:"luck_roll"`
	LuckMaxAttr   actor.Ability `json:"-"`
	PerksPerLevel int64         `json:"perks_per_level"`
	SkillCost     struct {
		Normal int64 `json:"normal"`
		Tag    int64 `json:"tag"`
	} `json:"skill_cost"`
}

// Pipeline is combat/pipeline.json.
type Pipeline struct {
	CritDamage      formula.Program `json:"crit_damage"`
	AreaEvadeDamage formula.Program `json:"area_evade_damage"`
	PierceDR        formula.Program `json:"pierce_dr"`
	DRAppliesTag    string          `json:"dr_applies_tag"`
	PierceTag       string          `json:"pierce_tag"`
}

// DamageType is one row of combat/damage_types.json. Physical and Pierce are the pipeline's
// two tags, looked up once.
type DamageType struct {
	Key      string   `json:"key"`
	Tags     []string `json:"tags"`
	Physical bool     `json:"-"`
	Pierce   bool     `json:"-"`
}

// Condition is one row of conditions/core.json.
type Condition struct {
	Key       string              `json:"key"`
	AppliedBy []string            `json:"appliedBy"`
	RemovedBy []string            `json:"removedBy"`
	Stacking  bool                `json:"stacking"`
	MaxStacks int64               `json:"maxStacks"`
	Modifiers []modifier.Modifier `json:"modifiers"`
}

// Conditions is the table plus, for each effect ref, the conditions it applies and removes
// (as row indexes, in file order).
type Conditions struct {
	Table     Table[Condition]
	AppliedBy map[string][]int32
	RemovedBy map[string][]int32
}

// Entry is one row of a loot table. Roll[n-1] on the table points at its entry.
type Entry struct {
	Min      int64  `json:"min"`
	Max      int64  `json:"max"`
	Kind     string `json:"kind"`
	Coins    *Cost  `json:"coins"`
	Item     string `json:"item"`
	Quantity int64  `json:"quantity"`
	CoinsCP  int64  `json:"-"`
}

// LootTable is one table of loot/core.json. ByRoll maps each d100 roll to an entry index
// (-1 where no entry covers it).
type LootTable struct {
	Key     string    `json:"key"`
	Level   int64     `json:"level"`
	Rolls   int64     `json:"rolls"`
	Entries []Entry   `json:"entries"`
	ByRoll  [100]int8 `json:"-"`
}

// DropRow is one row of a monster's loot: a table, or an item with a chance.
type DropRow struct {
	Table    string `json:"table"`
	Item     string `json:"item"`
	Quantity *int64 `json:"quantity"`
	Chance   *int64 `json:"chance"`
	Rolls    *int64 `json:"rolls"`
	Shift    int64  `json:"shift"`
	Material string `json:"material"`
	TableIdx int32  `json:"-"` // -1 for an item row
}

// Drop is a monster's loot rows.
type Drop struct {
	Monster string    `json:"monster"`
	Loot    []DropRow `json:"loot"`
}

// Loot is loot/core.json.
type Loot struct {
	Budget formula.Program `json:"budget_cp"`
	Tables Table[LootTable]
	Drops  Table[Drop]
}

// Item is what loot needs of an equipment item: its price in copper, and whether a loot row
// may set its material.
type Item struct {
	CostCP        int64
	TakesMaterial bool
}

// Ruleset is everything the commands read.
type Ruleset struct {
	Resolution  Resolution
	Derived     Table[DerivedStat]
	Actions     Actions
	Wuxing      Wuxing
	Rest        Table[Occasion]
	Economy     Economy
	Progression Progression
	Pipeline    Pipeline
	DamageTypes Table[DamageType]
	Origins     modifier.Highest
	Conditions  Conditions
	Loot        Loot

	Formulas        map[string]*formula.Program // formulas/<name>.json by name
	Price           *formula.Program
	LootLuck        *formula.Program
	Items           map[string]Item  // core:<category>/<key>
	MaterialPercent map[string]int64 // material key: valueFactor * 100
	MonsterLevel    map[string]int64 // monster key
}

type loader struct {
	fsys fs.FS
	err  error
}

func (l *loader) read(file string, v any) {
	if l.err != nil {
		return
	}
	b, err := fs.ReadFile(l.fsys, file)
	if err == nil {
		err = json.Unmarshal(b, v)
	}
	if err != nil {
		l.err = fmt.Errorf("%s: %w", file, err)
	}
}

func (l *loader) fail(format string, args ...any) {
	if l.err == nil {
		l.err = fmt.Errorf(format, args...)
	}
}

// Load reads the ruleset from the data directory held by fsys.
func Load(fsys fs.FS) (*Ruleset, error) {
	l := &loader{fsys: fsys}
	rs := &Ruleset{}

	var resolution struct{ Resolution Resolution }
	l.read("resolution/core.json", &resolution)
	if resolution.Resolution.CritWindowPerTwoLuck <= 0 {
		l.fail("resolution/core.json: crit_window_per_two_luck must be positive")
	}
	rs.Resolution = resolution.Resolution

	var derived struct{ Derived []DerivedStat }
	l.read("derived/core.json", &derived)
	for i := range derived.Derived {
		d := &derived.Derived[i]
		if d.PerPointDivisor == 0 {
			d.PerPointDivisor = 1
		}
		d.AttrID = l.ability(d.Attr)
	}
	rs.Derived = newTable(derived.Derived, func(d *DerivedStat) string { return d.Key })

	var actions struct {
		Pool    ActionPool
		Actions []Action
	}
	l.read("actions/core.json", &actions)
	actions.Pool.AttrID = l.ability(actions.Pool.Attr)
	if actions.Pool.PerPointDivisor <= 0 {
		l.fail("actions/core.json: per_point_divisor must be positive")
	}
	rs.Actions = Actions{Pool: actions.Pool, Actions: newTable(actions.Actions, func(a *Action) string { return a.Key })}

	l.loadWuxing(rs)

	var rest struct{ Rest []Occasion }
	l.read("rest/core.json", &rest)
	for _, o := range rest.Rest {
		for _, r := range o.Recovery {
			if r.Mode == "fraction" && r.Denominator <= 0 {
				l.fail("rest/core.json: %s %s has no denominator", o.Key, r.Resource)
			}
		}
	}
	rs.Rest = newTable(rest.Rest, func(o *Occasion) string { return o.Key })

	var economy struct {
		Currency  struct{ Denominations []Denomination }
		Barter    Barter
		Merchants []Merchant
	}
	l.read("economy/core.json", &economy)
	rs.Economy = Economy{
		Denominations: economy.Currency.Denominations,
		Barter:        economy.Barter,
		Merchants:     newTable(economy.Merchants, func(m *Merchant) string { return m.Key }),
	}
	if economy.Barter.MarginDivisor <= 0 {
		l.fail("economy/core.json: barter margin_divisor must be positive")
	}

	var progression struct{ Progression Progression }
	l.read("progression/core.json", &progression)
	rs.Progression = progression.Progression
	rs.Progression.LuckMaxAttr = l.ability(rs.Progression.LuckRoll.MaxAttr)
	if rs.Progression.SkillCost.Normal <= 0 || rs.Progression.SkillCost.Tag <= 0 {
		l.fail("progression/core.json: skill costs must be positive")
	}

	var pipeline struct{ Pipeline Pipeline }
	l.read("combat/pipeline.json", &pipeline)
	rs.Pipeline = pipeline.Pipeline
	var types struct {
		DamageTypes []DamageType `json:"damage_types"`
	}
	l.read("combat/damage_types.json", &types)
	for i := range types.DamageTypes {
		t := &types.DamageTypes[i]
		t.Physical = slices.Contains(t.Tags, rs.Pipeline.DRAppliesTag)
		t.Pierce = slices.Contains(t.Tags, rs.Pipeline.PierceTag)
	}
	rs.DamageTypes = newTable(types.DamageTypes, func(t *DamageType) string { return t.Key })

	var origins struct{ Origins []struct{ Key, Stack string } }
	l.read("origins/core.json", &origins)
	rs.Origins = modifier.Highest{}
	for _, o := range origins.Origins {
		rs.Origins["core:origin/"+o.Key] = o.Stack == "highest"
	}

	l.loadConditions(rs)
	l.loadFormulas(rs)
	l.loadItems(rs)
	l.loadLoot(rs)
	return rs, l.err
}

func (l *loader) ability(key string) actor.Ability {
	ab, ok := actor.ByKey(key)
	if !ok {
		l.fail("%q is not an attribute", key)
	}
	return ab
}

func (l *loader) loadWuxing(rs *Ruleset) {
	var raw struct {
		Phases []struct{ ID string }
		Cycles []Cycle
	}
	l.read("wuxing/core.json", &raw)
	if len(raw.Phases) > 8 {
		l.fail("wuxing/core.json: at most 8 phases")
		return
	}
	for _, p := range raw.Phases {
		rs.Wuxing.Phases = append(rs.Wuxing.Phases, p.ID)
	}
	rs.Wuxing.Cycles = raw.Cycles
	for i := range rs.Wuxing.Cycles {
		c := &rs.Wuxing.Cycles[i]
		for _, e := range c.Edges {
			from, to := rs.Wuxing.Phase(e.From), rs.Wuxing.Phase(e.To)
			if from < 0 || to < 0 {
				l.fail("wuxing/core.json: cycle %s edge %s -> %s names an unknown phase", c.ID, e.From, e.To)
				return
			}
			c.Next[from] |= 1 << to
		}
	}
}

func (l *loader) loadConditions(rs *Ruleset) {
	var raw struct{ Conditions []Condition }
	l.read("conditions/core.json", &raw)
	c := Conditions{AppliedBy: map[string][]int32{}, RemovedBy: map[string][]int32{}}
	for i := range raw.Conditions {
		row := &raw.Conditions[i]
		if row.MaxStacks == 0 {
			row.MaxStacks = 1
		}
		if !row.Stacking {
			row.MaxStacks = 1
		}
		for _, effect := range row.AppliedBy {
			c.AppliedBy[effect] = append(c.AppliedBy[effect], int32(i))
		}
		for _, effect := range row.RemovedBy {
			c.RemovedBy[effect] = append(c.RemovedBy[effect], int32(i))
		}
	}
	c.Table = newTable(raw.Conditions, func(r *Condition) string { return r.Key })
	rs.Conditions = c
}

func (l *loader) loadFormulas(rs *Ruleset) {
	rs.Formulas = map[string]*formula.Program{}
	if l.err != nil {
		return
	}
	files, err := fs.ReadDir(l.fsys, "formulas")
	if err != nil {
		l.fail("formulas: %v", err)
		return
	}
	for _, f := range files {
		var raw struct{ Expr formula.Program }
		l.read(path.Join("formulas", f.Name()), &raw)
		rs.Formulas[strings.TrimSuffix(f.Name(), ".json")] = &raw.Expr
	}
	rs.Price, rs.LootLuck = rs.Formulas["price"], rs.Formulas["loot_luck"]
	if rs.Price == nil || rs.LootLuck == nil {
		l.fail("formulas: price.json and loot_luck.json are required")
	}
}

// loadItems prices every equipment item and material, and reads every monster's level.
func (l *loader) loadItems(rs *Ruleset) {
	rs.Items = map[string]Item{}
	categories := []struct{ category, file, baseDir string }{
		{"weapon", "weapons", "weapons"},
		{"armor", "armor", "armor"},
		{"consumable", "consumables", ""},
		{"accessory", "accessories", "accessories"},
	}
	for _, c := range categories {
		var raw struct {
			Equipment []struct {
				Key       string
				Base      string
				Cost      *Cost
				Overrides struct{ Cost *Cost }
			}
		}
		l.read("equipment/"+c.file+".json", &raw)
		for _, row := range raw.Equipment {
			cost := row.Cost
			if cost == nil {
				cost = row.Overrides.Cost
			}
			if cost == nil {
				var base struct{ Cost *Cost }
				l.read(fmt.Sprintf("equipment/bases/%s/%s.json", c.baseDir, row.Base), &base)
				cost = base.Cost
			}
			if cost == nil {
				l.fail("equipment/%s.json: %s has no cost", c.file, row.Key)
				return
			}
			cp, err := rs.Economy.ToCP(*cost)
			if err != nil {
				l.fail("equipment/%s.json: %s: %v", c.file, row.Key, err)
				return
			}
			rs.Items["core:"+c.category+"/"+row.Key] = Item{CostCP: cp, TakesMaterial: c.category == "weapon" || c.category == "armor"}
		}
	}

	var materials struct {
		Materials []struct {
			Key         string
			ValueFactor float64
		}
	}
	l.read("materials/core.json", &materials)
	rs.MaterialPercent = map[string]int64{}
	for _, m := range materials.Materials {
		rs.MaterialPercent[m.Key] = int64(math.RoundToEven(m.ValueFactor * 100))
	}

	var monsters struct {
		Monsters []struct {
			Key       string
			Base      string
			Level     *int64
			Overrides struct{ Level *int64 }
		}
	}
	l.read("monsters/core.json", &monsters)
	rs.MonsterLevel = map[string]int64{}
	for _, m := range monsters.Monsters {
		level := m.Overrides.Level
		if level == nil {
			level = m.Level
		}
		if level == nil {
			var base struct{ Level int64 }
			l.read("monsters/bases/"+m.Base+".json", &base)
			level = &base.Level
		}
		rs.MonsterLevel[m.Key] = *level
	}
}

func (l *loader) loadLoot(rs *Ruleset) {
	var raw struct {
		Budget formula.Program `json:"budget_cp"`
		Tables []LootTable
		Drops  []Drop
	}
	l.read("loot/core.json", &raw)
	if l.err != nil {
		return
	}
	for i := range raw.Tables {
		t := &raw.Tables[i]
		for n := range t.ByRoll {
			t.ByRoll[n] = -1
		}
		for n := range t.Entries {
			e := &t.Entries[n]
			if e.Coins != nil {
				cp, err := rs.Economy.ToCP(*e.Coins)
				if err != nil {
					l.fail("loot/core.json: %s: %v", t.Key, err)
					return
				}
				e.CoinsCP = cp
			}
			for roll := max(e.Min, 1); roll <= min(e.Max, 100); roll++ {
				t.ByRoll[roll-1] = int8(n)
			}
		}
	}
	rs.Loot = Loot{
		Budget: raw.Budget,
		Tables: newTable(raw.Tables, func(t *LootTable) string { return t.Key }),
		Drops:  newTable(raw.Drops, func(d *Drop) string { return d.Monster }),
	}
	for i := range raw.Drops {
		for n := range raw.Drops[i].Loot {
			row := &raw.Drops[i].Loot[n]
			row.TableIdx = -1
			if row.Table == "" && row.Item == "" {
				l.fail("loot/core.json: %s has a loot row with neither table nor item", raw.Drops[i].Monster)
				return
			}
			if row.Table != "" {
				idx, ok := rs.Loot.Tables.index[row.Table]
				if !ok {
					l.fail("loot/core.json: %s drops unknown table %q", raw.Drops[i].Monster, row.Table)
					return
				}
				row.TableIdx = idx
			}
		}
	}
}
