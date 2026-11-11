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
	"cmp"
	"encoding/json"
	"errors"
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
type Table[Row any] struct {
	rows       []Row
	indexByKey map[string]int32
}

func newTable[Row any](rows []Row, keyOf func(*Row) string) Table[Row] {
	var table Table[Row] = Table[Row]{rows: rows, indexByKey: make(map[string]int32, len(rows))}
	for index := range rows {
		table.indexByKey[keyOf(&rows[index])] = int32(index)
	}
	return table
}

// Get returns the row with the key.
func (table *Table[Row]) Get(key string) (*Row, bool) {
	index, ok := table.indexByKey[key]
	if !ok {
		return nil, false
	}
	return &table.rows[index], true
}

// Index returns the row index of a key, for callers that keep a slice aligned with the rows.
func (table *Table[Row]) Index(key string) (int32, bool) {
	index, ok := table.indexByKey[key]
	return index, ok
}

// Rows are the rows in file order. Callers must not change them.
func (table *Table[Row]) Rows() []Row { return table.rows }

// At returns the row at an index a loader resolved.
func (table *Table[Row]) At(index int32) *Row { return &table.rows[index] }

// Cost is a price in a denomination.
type Cost struct {
	Quantity int64  `json:"quantity"`
	Unit     string `json:"unit"`
}

// Resolution is ruleset/data/resolution/core.json.
type Resolution struct {
	Die                      int64  `json:"die"`
	CriticalWindow           int64  `json:"crit_window"`
	CriticalWindowPerTwoLuck int64  `json:"crit_window_per_two_luck"`
	FumbleAt                 int64  `json:"fumble_at"`
	TargetCap                *int64 `json:"target_cap"`
}

// DerivedStat is one row of derived/core.json.
type DerivedStat struct {
	Key               string        `json:"key"`
	Attribute         string        `json:"attr"`
	Base              int64         `json:"base"`
	PerPoint          int64         `json:"per_point"`
	PerPointDivisor   int64         `json:"per_point_divisor"`
	PerLevelAttribute int64         `json:"per_level_attr"`
	PerLevelFlat      int64         `json:"per_level_flat"`
	Ability           actor.Ability `json:"-"`
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
	Attribute       string        `json:"attr"`
	Ability         actor.Ability `json:"-"`
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

// Cycle is one Wuxing cycle. NextPhasesMask[from] is a bit mask of the phases it points to,
// so an edge test is one AND.
type Cycle struct {
	Identifier     string   `json:"id"`
	Interaction    string   `json:"interaction"`
	Affects        string   `json:"affects"`
	Multiplier     float64  `json:"magnitudeMultiplier"`
	Gate           *Gate    `json:"gate"`
	Edges          []Edge   `json:"edges"`
	NextPhasesMask [8]uint8 `json:"-"`
}

// Has reports whether the cycle holds the edge from -> to (phase indexes).
func (cycle *Cycle) Has(from, to int) bool { return cycle.NextPhasesMask[from]&(1<<to) != 0 }

// Wuxing is the wuxing module's data (ruleset/modules/wuxing/).
type Wuxing struct {
	Phases []string
	Cycles []Cycle
}

// Phase returns the index of a phase name, or -1.
func (wuxing *Wuxing) Phase(name string) int { return slices.Index(wuxing.Phases, name) }

// Cycle returns the cycle with the identifier.
func (wuxing *Wuxing) Cycle(identifier string) (*Cycle, bool) {
	for index := range wuxing.Cycles {
		if wuxing.Cycles[index].Identifier == identifier {
			return &wuxing.Cycles[index], true
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
	MaximumShift  int64  `json:"max_shift"`
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

// ToCopper converts a cost to copper.
func (economy *Economy) ToCopper(cost Cost) (int64, error) {
	for _, denomination := range economy.Denominations {
		if denomination.Unit == cost.Unit {
			return cost.Quantity * denomination.Value, nil
		}
	}
	return 0, fmt.Errorf("unit %q is not a denomination", cost.Unit)
}

// Skill is what the rules need of one row of skills/core.json.
type Skill struct {
	Key               string        `json:"key"`
	AssociatedAbility string        `json:"associatedAbility"`
	Ability           actor.Ability `json:"-"`
}

// Progression is progression/core.json.
type Progression struct {
	StartLevel        int64           `json:"start_level"`
	MaximumLevel      int64           `json:"max_level"`
	SkillCap          int64           `json:"skill_cap"`
	ExperienceToReach formula.Program `json:"xp_to_reach"`
	SkillPoints       formula.Program `json:"skill_points"`
	PoolRollMaximum   formula.Program `json:"pool_roll_max"`
	LuckRoll          struct {
		Minimum          int64  `json:"min"`
		MaximumAttribute string `json:"max_attr"`
	} `json:"luck_roll"`
	LuckMaximumAbility actor.Ability   `json:"-"`
	SkillStart         formula.Program `json:"skill_start"`
	PerksPerLevel      int64           `json:"perks_per_level"`
	SkillCost          struct {
		Normal int64 `json:"normal"`
		Tag    int64 `json:"tag"`
	} `json:"skill_cost"`
}

// Pipeline is combat/pipeline.json.
type Pipeline struct {
	CriticalDamage            formula.Program `json:"crit_damage"`
	AreaEvadeDamage           formula.Program `json:"area_evade_damage"`
	PierceDamageReduction     formula.Program `json:"pierce_dr"`
	DamageReductionAppliesTag string          `json:"dr_applies_tag"`
	PierceTag                 string          `json:"pierce_tag"`
}

// DamageType is one row of combat/damage_types.json. Physical and Pierce are the pipeline's
// two tags, looked up once.
type DamageType struct {
	Key      string   `json:"key"`
	Tags     []string `json:"tags"`
	Physical bool     `json:"-"`
	Pierce   bool     `json:"-"`
}

// Situation is one row of combat/situations.json: attack modifiers per range, named by
// the caller or applied by a flag (docs/core/combat.md, step 0).
type Situation struct {
	When struct {
		Attacker string `json:"attacker"` // applies when the attacker has this flag
		Defender string `json:"defender"` // applies when the defender has this flag
	} `json:"when"`
	Key        string              `json:"key"`
	Melee      []modifier.Modifier `json:"melee"`
	Ranged     []modifier.Modifier `json:"ranged"`
	Forbids    []string            `json:"forbids"`
	Deliveries []string            `json:"deliveries"` // nil: every delivery
}

// RangeModifiers returns the row's modifiers for a range, "melee" or "ranged".
func (situation *Situation) RangeModifiers(attackRange string) []modifier.Modifier {
	if attackRange == "ranged" {
		return situation.Ranged
	}
	return situation.Melee
}

// Rider is what a called shot does on a hit, for the caller to apply.
type Rider struct {
	Condition string `json:"condition,omitempty"`
	Event     string `json:"event,omitempty"`
	Check     string `json:"check,omitempty"`
	Turns     int64  `json:"turns,omitempty"`
}

// CalledShot is one called_shots row of combat/situations.json.
type CalledShot struct {
	Key              string  `json:"key"`
	Riders           []Rider `json:"riders"`
	Attack           int64   `json:"attack"`
	DamageMultiplier float64 `json:"damage_mul"`
}

// HazardCheck is a check a hazard asks for: a row of checks/core.json, and the hazard's
// difficulty as an add on its target. Skill is the check's skill key, resolved at load.
type HazardCheck struct {
	Check      string `json:"check"`
	OnAvoid    string `json:"on_avoid"` // avoid only: "none" or "half"
	Skill      string `json:"-"`
	Difficulty int64  `json:"difficulty"`
}

// Hazard is one row of hazards/core.json (docs/core/hazards.md).
type Hazard struct {
	Detect HazardCheck  `json:"detect"`
	Disarm *HazardCheck `json:"disarm"` // nil: cannot be disarmed
	Avoid  *HazardCheck `json:"avoid"`  // nil: no chance to avoid
	Damage *struct {
		Type string `json:"type"`
	} `json:"damage"`
	Key     string   `json:"key"`
	Nature  string   `json:"nature"` // mechanical, natural, or arcane
	Color   string   `json:"color"`  // arcane only: detect and disarm roll <color>_magic
	Effects []string `json:"effects"`
	Level   int64    `json:"level"`
}

// AppliedEffect is one entry of a spell's effects. Encoded keeps the entry as written, so
// a cast returns it unchanged; Effect and Magnitude are what dispel reads.
type AppliedEffect struct {
	Encoded   json.RawMessage
	Effect    string
	Magnitude float64
}

// UnmarshalJSON keeps the entry and reads its effect ref and magnitude.
func (applied *AppliedEffect) UnmarshalJSON(encoded []byte) error {
	var fields struct {
		Effect    string  `json:"effect"`
		Magnitude float64 `json:"magnitude"`
	}
	if err := json.Unmarshal(encoded, &fields); err != nil {
		return err
	}
	applied.Encoded = append(json.RawMessage{}, encoded...)
	applied.Effect, applied.Magnitude = fields.Effect, fields.Magnitude
	return nil
}

// Spell is what a cast needs of one row of spells/core.json (docs/core/casting.md).
type Spell struct {
	Key     string          `json:"key"`
	Color   string          `json:"color"`
	Effects []AppliedEffect `json:"effects"`
	Cost    struct {
		Reagents []string `json:"reagents"`
		Mana     int64    `json:"mana"`
	} `json:"cost"`
	Level         int64 `json:"level"`
	Concentration bool  `json:"concentration"`
}

// Casting is casting/core.json: the caster flags that block a cast, and the cast penalty by
// the heaviest armor category worn.
type Casting struct {
	ArmorPenalty  map[string]int64 `json:"armor_penalty"`
	PreventedBy   []string         `json:"prevented_by"`
	SkillDiscount SkillDiscount    `json:"skill_discount"`
}

// SkillDiscount is what skill takes off a spell's mana cost: min(skill / Divisor, Cap).
type SkillDiscount struct {
	Divisor     float64 `json:"divisor"`
	Cap         float64 `json:"cap"`
	MinimumPaid int64   `json:"minimum_paid"`
}

// Condition is one row of conditions/core.json.
type Condition struct {
	Key           string              `json:"key"`
	AppliedBy     []string            `json:"appliedBy"`
	RemovedBy     []string            `json:"removedBy"`
	Stacking      bool                `json:"stacking"`
	MaximumStacks int64               `json:"maxStacks"`
	Modifiers     []modifier.Modifier `json:"modifiers"`
}

// Conditions is the table plus, for each effect ref, the conditions it applies and removes
// (as row indexes, in file order).
type Conditions struct {
	Table     Table[Condition]
	AppliedBy map[string][]int32
	RemovedBy map[string][]int32
}

// LootEntry is one row of a loot table. EntryIndexByRoll[n-1] on the table points at its
// entry.
type LootEntry struct {
	Minimum     int64  `json:"min"`
	Maximum     int64  `json:"max"`
	Kind        string `json:"kind"`
	Coins       *Cost  `json:"coins"`
	Item        string `json:"item"`
	Quantity    int64  `json:"quantity"`
	CoinsCopper int64  `json:"-"`
}

// LootTable is one table of loot/core.json. EntryIndexByRoll maps each d100 roll to an
// entry index (-1 where no entry covers it).
type LootTable struct {
	Key              string      `json:"key"`
	Level            int64       `json:"level"`
	Rolls            int64       `json:"rolls"`
	Entries          []LootEntry `json:"entries"`
	EntryIndexByRoll [100]int8   `json:"-"`
}

// DropRow is one row of a monster's loot: a table, or an item with a chance.
type DropRow struct {
	Table      string `json:"table"`
	Item       string `json:"item"`
	Quantity   *int64 `json:"quantity"`
	Chance     *int64 `json:"chance"`
	Rolls      *int64 `json:"rolls"`
	Shift      int64  `json:"shift"`
	Material   string `json:"material"`
	Size       string `json:"size"`
	TableIndex int32  `json:"-"` // -1 for an item row
}

// Drop is a monster's loot rows.
type Drop struct {
	Monster string    `json:"monster"`
	Loot    []DropRow `json:"loot"`
}

// Loot is loot/core.json.
type Loot struct {
	BudgetCopper formula.Program `json:"budget_cp"`
	Tables       Table[LootTable]
	Drops        Table[Drop]
}

// Item is what loot needs of an equipment item: its price in copper, whether a loot row may
// set its material, the materials it is found in (the first is its fallback), and for a
// weapon its dice, which a loot row's size moves (docs/core/loot.md#what-lands-on-an-item).
type Item struct {
	Dice             string
	DefaultMaterials []string
	CostCopper       int64
	TakesMaterial    bool
	IsWeapon         bool
	Type             string
	Consumable       *Consumable
}

// Consumable is how an item is used and the effects it applies.
type Consumable struct {
	Delivery string
	Effects  []ConsumableEffect
}

// ConsumableEffect is one applied effect on a consumable; Magnitude is nil when it has none.
type ConsumableEffect struct {
	Effect    string
	Parameter string
	Magnitude *int64
	Duration  int64
	Origin    string
}

// Size is one row of combat/dice_ladder.json's sizes: the rungs the dice move, the attack
// change, and the value percent.
type Size struct {
	DiceStep     int64 `json:"dice_step"`
	Attack       int64 `json:"attack"`
	ValuePercent int64 `json:"value_percent"`
}

// Ruleset is everything the commands read.
type Ruleset struct {
	Resolution     Resolution
	Derived        Table[DerivedStat]
	Skills         Table[Skill]
	Actions        Actions
	Wuxing         Wuxing
	Modules        []string // ids of the loaded modules
	Rest           Table[Occasion]
	Economy        Economy
	Progression    Progression
	Pipeline       Pipeline
	DamageTypes    Table[DamageType]
	Situations     Table[Situation]
	CalledShots    Table[CalledShot]
	Hazards        Table[Hazard]
	Spells         Table[Spell]
	Casting        Casting
	Origins        modifier.OriginPolicies
	OriginReplaces map[string]bool   // origin ref: its policy is `replace`
	OriginFor      map[string]string // emitter: the origin ref it defaults to
	SecondsPerTurn int64             // actions/core.json
	Conditions     Conditions
	Loot           Loot

	Formulas          map[string]*formula.Program // formulas/<name>.json by name
	AttributeModifier *formula.Program            // formulas/attribute_modifier.json
	Price             *formula.Program
	LootLuck          *formula.Program
	Items             map[string]Item  // core:<category>/<key>
	MaterialPercent   map[string]int64 // material key: valueFactor * 100
	DiceLadder        [][]string       // combat/dice_ladder.json: rungs of equal dice, smallest first
	Sizes             map[string]Size  // combat/dice_ladder.json: undersized, normal, oversized
	MonsterLevel      map[string]int64 // monster key
}

// rulesetLoader reads every file even after one fails, and joins every error, so one run
// reports every broken file and row. A check runs only when its file was read, so a file
// that failed is reported once, not again for each zero value it left behind.
type rulesetLoader struct {
	fileSystem fs.FS
	errs       []error
}

// read decodes one file into target and reports whether it could.
func (loader *rulesetLoader) read(file string, target any) bool {
	content, err := fs.ReadFile(loader.fileSystem, file)
	if err == nil {
		err = json.Unmarshal(content, target)
	}
	if err != nil {
		loader.errs = append(loader.errs, fmt.Errorf("%s: %w", file, err))
		return false
	}
	return true
}

func (loader *rulesetLoader) fail(format string, arguments ...any) {
	loader.errs = append(loader.errs, fmt.Errorf(format, arguments...))
}

// Load reads the ruleset from the data directory held by fileSystem. The error, if any,
// joins one error per broken file or row.
func Load(fileSystem fs.FS, modules ...fs.FS) (*Ruleset, error) {
	var loader *rulesetLoader = &rulesetLoader{fileSystem: fileSystem}
	var ruleset *Ruleset = &Ruleset{}

	var resolution struct{ Resolution Resolution }
	if loader.read("resolution/core.json", &resolution) && resolution.Resolution.CriticalWindowPerTwoLuck <= 0 {
		loader.fail("resolution/core.json: crit_window_per_two_luck must be positive")
	}
	ruleset.Resolution = resolution.Resolution

	var derived struct{ Derived []DerivedStat }
	if loader.read("derived/core.json", &derived) {
		for index := range derived.Derived {
			var stat *DerivedStat = &derived.Derived[index]
			if stat.PerPointDivisor == 0 {
				stat.PerPointDivisor = 1
			}
			stat.Ability = loader.ability("derived/core.json", stat.Attribute)
		}
	}
	ruleset.Derived = newTable(derived.Derived, func(stat *DerivedStat) string { return stat.Key })

	var actions struct {
		Pool           ActionPool
		SecondsPerTurn int64 `json:"seconds_per_turn"`
		Actions        []Action
	}
	if loader.read("actions/core.json", &actions) {
		actions.Pool.Ability = loader.ability("actions/core.json", actions.Pool.Attribute)
		if actions.Pool.PerPointDivisor <= 0 {
			loader.fail("actions/core.json: per_point_divisor must be positive")
		}
	}
	ruleset.SecondsPerTurn = actions.SecondsPerTurn
	ruleset.Actions = Actions{Pool: actions.Pool, Actions: newTable(actions.Actions, func(action *Action) string { return action.Key })}

	var rest struct{ Rest []Occasion }
	loader.read("rest/core.json", &rest)
	for _, occasion := range rest.Rest {
		for _, recovery := range occasion.Recovery {
			if recovery.Mode == "fraction" && recovery.Denominator <= 0 {
				loader.fail("rest/core.json: %s %s has no denominator", occasion.Key, recovery.Resource)
			}
		}
	}
	ruleset.Rest = newTable(rest.Rest, func(occasion *Occasion) string { return occasion.Key })

	var economy struct {
		Currency  struct{ Denominations []Denomination }
		Barter    Barter
		Merchants []Merchant
	}
	var economyLoaded bool = loader.read("economy/core.json", &economy)
	ruleset.Economy = Economy{
		Denominations: economy.Currency.Denominations,
		Barter:        economy.Barter,
		Merchants:     newTable(economy.Merchants, func(merchant *Merchant) string { return merchant.Key }),
	}
	if economyLoaded && economy.Barter.MarginDivisor <= 0 {
		loader.fail("economy/core.json: barter margin_divisor must be positive")
	}

	var skills struct{ Skills []Skill }
	if loader.read("skills/core.json", &skills) {
		for index := range skills.Skills {
			var skill *Skill = &skills.Skills[index]
			skill.Ability = loader.ability("skills/core.json", skill.AssociatedAbility)
		}
	}
	ruleset.Skills = newTable(skills.Skills, func(skill *Skill) string { return skill.Key })

	var progression struct{ Progression Progression }
	if loader.read("progression/core.json", &progression) {
		var rules *Progression = &progression.Progression
		rules.LuckMaximumAbility = loader.ability("progression/core.json", rules.LuckRoll.MaximumAttribute)
		if rules.SkillCost.Normal <= 0 || rules.SkillCost.Tag <= 0 {
			loader.fail("progression/core.json: skill costs must be positive")
		}
	}
	ruleset.Progression = progression.Progression

	var pipeline struct{ Pipeline Pipeline }
	loader.read("combat/pipeline.json", &pipeline)
	ruleset.Pipeline = pipeline.Pipeline
	var damageTypes struct {
		DamageTypes []DamageType `json:"damage_types"`
	}
	loader.read("combat/damage_types.json", &damageTypes)
	for index := range damageTypes.DamageTypes {
		var damageType *DamageType = &damageTypes.DamageTypes[index]
		damageType.Physical = slices.Contains(damageType.Tags, ruleset.Pipeline.DamageReductionAppliesTag)
		damageType.Pierce = slices.Contains(damageType.Tags, ruleset.Pipeline.PierceTag)
	}
	ruleset.DamageTypes = newTable(damageTypes.DamageTypes, func(damageType *DamageType) string { return damageType.Key })

	var situations struct {
		Situations  []Situation  `json:"situations"`
		CalledShots []CalledShot `json:"called_shots"`
	}
	loader.read("combat/situations.json", &situations)
	ruleset.Situations = newTable(situations.Situations, func(situation *Situation) string { return situation.Key })
	ruleset.CalledShots = newTable(situations.CalledShots, func(calledShot *CalledShot) string { return calledShot.Key })

	var origins struct {
		Origins []struct {
			Key, Stack string
			DefaultFor []string `json:"defaultFor"`
		}
	}
	loader.read("origins/core.json", &origins)
	ruleset.Origins = modifier.OriginPolicies{}
	ruleset.OriginReplaces, ruleset.OriginFor = map[string]bool{}, map[string]string{}
	for _, origin := range origins.Origins {
		var ref string = "core:origin/" + origin.Key
		ruleset.Origins[ref] = origin.Stack == "highest"
		ruleset.OriginReplaces[ref] = origin.Stack == "replace"
		for _, emitter := range origin.DefaultFor {
			ruleset.OriginFor[emitter] = ref
		}
	}

	loader.loadConditions(ruleset)
	loader.loadHazards(ruleset)

	var spells struct{ Spells []Spell }
	loader.read("spells/core.json", &spells)
	ruleset.Spells = newTable(spells.Spells, func(spell *Spell) string { return spell.Key })
	var casting struct{ Casting Casting }
	loader.read("casting/core.json", &casting)
	ruleset.Casting = casting.Casting

	loader.loadFormulas(ruleset)
	// Items and loot are priced in copper, so they need the economy's denominations.
	if economyLoaded {
		loader.loadItems(ruleset)
		loader.loadLoot(ruleset)
	} else {
		loader.fail("equipment and loot: not loaded, because economy/core.json did not load")
	}
	for _, module := range modules {
		loader.loadModule(ruleset, module)
	}
	return ruleset, errors.Join(loader.errs...)
}

func (loader *rulesetLoader) ability(file, key string) actor.Ability {
	ability, ok := actor.ByKey(key)
	if !ok {
		loader.fail("%s: %q is not an attribute", file, key)
	}
	return ability
}

// loadModule reads one module's manifest and the data of the modules the core reads.
func (loader *rulesetLoader) loadModule(ruleset *Ruleset, module fs.FS) {
	var core fs.FS = loader.fileSystem
	loader.fileSystem = module
	defer func() { loader.fileSystem = core }()
	var manifest struct {
		Identifier string `json:"id"`
		Data       []string
	}
	if !loader.read("manifest.json", &manifest) {
		return
	}
	ruleset.Modules = append(ruleset.Modules, manifest.Identifier)
	if manifest.Identifier == "wuxing" {
		for _, file := range manifest.Data {
			loader.loadWuxing(ruleset, file)
		}
	}
}

func (loader *rulesetLoader) loadWuxing(ruleset *Ruleset, file string) {
	var raw struct {
		Phases []struct {
			Identifier string `json:"id"`
		}
		Cycles []Cycle
	}
	if !loader.read(file, &raw) {
		return
	}
	if len(raw.Phases) > 8 {
		loader.fail("%s: at most 8 phases", file)
		return
	}
	for _, phase := range raw.Phases {
		ruleset.Wuxing.Phases = append(ruleset.Wuxing.Phases, phase.Identifier)
	}
	ruleset.Wuxing.Cycles = raw.Cycles
	for index := range ruleset.Wuxing.Cycles {
		var cycle *Cycle = &ruleset.Wuxing.Cycles[index]
		for _, edge := range cycle.Edges {
			var from, to int = ruleset.Wuxing.Phase(edge.From), ruleset.Wuxing.Phase(edge.To)
			if from < 0 || to < 0 {
				loader.fail("%s: cycle %s edge %s -> %s names an unknown phase", file, cycle.Identifier, edge.From, edge.To)
				continue
			}
			cycle.NextPhasesMask[from] |= 1 << to
		}
	}
}

func (loader *rulesetLoader) loadConditions(ruleset *Ruleset) {
	var raw struct{ Conditions []Condition }
	loader.read("conditions/core.json", &raw)
	var conditions Conditions = Conditions{AppliedBy: map[string][]int32{}, RemovedBy: map[string][]int32{}}
	for index := range raw.Conditions {
		var row *Condition = &raw.Conditions[index]
		if row.MaximumStacks == 0 {
			row.MaximumStacks = 1
		}
		if !row.Stacking {
			row.MaximumStacks = 1
		}
		for _, effect := range row.AppliedBy {
			conditions.AppliedBy[effect] = append(conditions.AppliedBy[effect], int32(index))
		}
		for _, effect := range row.RemovedBy {
			conditions.RemovedBy[effect] = append(conditions.RemovedBy[effect], int32(index))
		}
	}
	conditions.Table = newTable(raw.Conditions, func(row *Condition) string { return row.Key })
	ruleset.Conditions = conditions
}

// loadHazards reads the hazards and resolves each check ref to the skill it rolls. An arcane
// hazard's detect and disarm name no check: they roll its color's magic skill.
func (loader *rulesetLoader) loadHazards(ruleset *Ruleset) {
	var checks struct {
		Checks []struct{ Key, Skill string }
	}
	var hazards struct{ Hazards []Hazard }
	var checksLoaded bool = loader.read("checks/core.json", &checks)
	var hazardsLoaded bool = loader.read("hazards/core.json", &hazards)
	if !checksLoaded || !hazardsLoaded {
		return
	}
	var skillByCheck map[string]string = make(map[string]string, len(checks.Checks))
	for _, check := range checks.Checks {
		skillByCheck["core:check/"+check.Key] = check.Skill[strings.LastIndex(check.Skill, "/")+1:]
	}
	for index := range hazards.Hazards {
		var hazard *Hazard = &hazards.Hazards[index]
		for _, check := range []*HazardCheck{&hazard.Detect, hazard.Disarm, hazard.Avoid} {
			if check == nil {
				continue
			}
			if check.Check == "" && hazard.Nature == "arcane" {
				check.Skill = hazard.Color + "_magic"
				continue
			}
			skill, ok := skillByCheck[check.Check]
			if !ok {
				loader.fail("hazards/core.json: %s names the unknown check %q", hazard.Key, check.Check)
				continue
			}
			check.Skill = skill
		}
	}
	ruleset.Hazards = newTable(hazards.Hazards, func(hazard *Hazard) string { return hazard.Key })
}

func (loader *rulesetLoader) loadFormulas(ruleset *Ruleset) {
	ruleset.Formulas = map[string]*formula.Program{}
	files, err := fs.ReadDir(loader.fileSystem, "formulas")
	if err != nil {
		loader.fail("formulas: %v", err)
		return
	}
	for _, file := range files {
		var raw struct {
			Expression formula.Program `json:"expr"`
		}
		loader.read(path.Join("formulas", file.Name()), &raw)
		ruleset.Formulas[strings.TrimSuffix(file.Name(), ".json")] = &raw.Expression
	}
	ruleset.Price, ruleset.LootLuck = ruleset.Formulas["price"], ruleset.Formulas["loot_luck"]
	ruleset.AttributeModifier = ruleset.Formulas["attribute_modifier"]
	if ruleset.Price == nil || ruleset.LootLuck == nil || ruleset.AttributeModifier == nil {
		loader.fail("formulas: price.json, loot_luck.json, and attribute_modifier.json are required")
	}
}

// loadItems prices every equipment item and material, and reads every monster's level.
func (loader *rulesetLoader) loadItems(ruleset *Ruleset) {
	ruleset.Items = map[string]Item{}
	var categories []struct{ category, file, baseDirectory string } = []struct{ category, file, baseDirectory string }{
		{"weapon", "weapons", "weapons"},
		{"armor", "armor", "armor"},
		{"consumable", "consumables", ""},
		{"accessory", "accessories", "accessories"},
		{"gear", "gear", ""},
	}
	for _, category := range categories {
		type itemFields struct {
			Cost             *Cost
			Weapon           *struct{ Damage struct{ Dice string } }
			DefaultMaterials []string `json:"default_materials"`
			Type             string
			Consumable       *Consumable
		}
		var raw struct {
			Equipment []struct {
				itemFields
				Key       string
				Base      string
				Overrides itemFields
			}
		}
		loader.read("equipment/"+category.file+".json", &raw)
		for _, row := range raw.Equipment {
			var base itemFields
			if row.Base != "" {
				loader.read(fmt.Sprintf("equipment/bases/%s/%s.json", category.baseDirectory, row.Base), &base)
			}
			// A field is the row's own, its overrides', or its base's.
			var cost *Cost = cmp.Or(row.Cost, row.Overrides.Cost, base.Cost)
			var weapon *struct{ Damage struct{ Dice string } } = cmp.Or(row.Weapon, row.Overrides.Weapon, base.Weapon)
			var defaults []string = row.DefaultMaterials
			if defaults == nil {
				defaults = row.Overrides.DefaultMaterials
			}
			if defaults == nil {
				defaults = base.DefaultMaterials
			}
			if cost == nil {
				loader.fail("equipment/%s.json: %s has no cost", category.file, row.Key)
				continue
			}
			copper, err := ruleset.Economy.ToCopper(*cost)
			if err != nil {
				loader.fail("equipment/%s.json: %s: %v", category.file, row.Key, err)
				continue
			}
			var item Item = Item{
				CostCopper:       copper,
				TakesMaterial:    category.category == "weapon" || category.category == "armor",
				IsWeapon:         category.category == "weapon",
				DefaultMaterials: defaults,
				Type:             row.Type,
				Consumable:       row.Consumable,
			}
			if weapon != nil {
				item.Dice = weapon.Damage.Dice
			}
			ruleset.Items["core:"+category.category+"/"+row.Key] = item
		}
	}

	var ladder struct {
		DiceLadder [][]string      `json:"dice_ladder"`
		Sizes      map[string]Size `json:"sizes"`
	}
	loader.read("combat/dice_ladder.json", &ladder)
	ruleset.DiceLadder, ruleset.Sizes = ladder.DiceLadder, ladder.Sizes

	var materials struct {
		Materials []struct {
			Key         string
			ValueFactor float64
		}
	}
	loader.read("materials/core.json", &materials)
	ruleset.MaterialPercent = map[string]int64{}
	for _, material := range materials.Materials {
		ruleset.MaterialPercent[material.Key] = int64(math.RoundToEven(material.ValueFactor * 100))
	}

	var monsters struct {
		Monsters []struct {
			Key       string
			Base      string
			Level     *int64
			Overrides struct{ Level *int64 }
		}
	}
	loader.read("monsters/core.json", &monsters)
	ruleset.MonsterLevel = map[string]int64{}
	for _, monster := range monsters.Monsters {
		var level *int64 = monster.Overrides.Level
		if level == nil {
			level = monster.Level
		}
		if level == nil {
			var base struct{ Level int64 }
			if !loader.read("monsters/bases/"+monster.Base+".json", &base) {
				continue
			}
			level = &base.Level
		}
		ruleset.MonsterLevel[monster.Key] = *level
	}
}

func (loader *rulesetLoader) loadLoot(ruleset *Ruleset) {
	var raw struct {
		BudgetCopper formula.Program `json:"budget_cp"`
		Tables       []LootTable
		Drops        []Drop
	}
	if !loader.read("loot/core.json", &raw) {
		return
	}
	for tableIndex := range raw.Tables {
		var table *LootTable = &raw.Tables[tableIndex]
		for roll := range table.EntryIndexByRoll {
			table.EntryIndexByRoll[roll] = -1
		}
		for entryIndex := range table.Entries {
			var entry *LootEntry = &table.Entries[entryIndex]
			if entry.Coins != nil {
				copper, err := ruleset.Economy.ToCopper(*entry.Coins)
				if err != nil {
					loader.fail("loot/core.json: %s: %v", table.Key, err)
					continue
				}
				entry.CoinsCopper = copper
			}
			for roll := max(entry.Minimum, 1); roll <= min(entry.Maximum, 100); roll++ {
				table.EntryIndexByRoll[roll-1] = int8(entryIndex)
			}
		}
	}
	ruleset.Loot = Loot{
		BudgetCopper: raw.BudgetCopper,
		Tables:       newTable(raw.Tables, func(table *LootTable) string { return table.Key }),
		Drops:        newTable(raw.Drops, func(drop *Drop) string { return drop.Monster }),
	}
	for dropIndex := range raw.Drops {
		var drop *Drop = &raw.Drops[dropIndex]
		for rowIndex := range drop.Loot {
			var row *DropRow = &drop.Loot[rowIndex]
			row.TableIndex = -1
			if row.Table == "" && row.Item == "" {
				loader.fail("loot/core.json: %s has a loot row with neither table nor item", drop.Monster)
				continue
			}
			if row.Table != "" {
				tableIndex, ok := ruleset.Loot.Tables.indexByKey[row.Table]
				if !ok {
					loader.fail("loot/core.json: %s drops unknown table %q", drop.Monster, row.Table)
					continue
				}
				row.TableIndex = tableIndex
			}
		}
	}
}
