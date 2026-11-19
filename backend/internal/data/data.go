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
	Attribute         string        `json:"attribute"`
	Base              int64         `json:"base"`
	PerPoint          int64         `json:"per_point"`
	PerPointDivisor   int64         `json:"per_point_divisor"`
	PerLevelAttribute int64         `json:"per_level_attribute"`
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
	Attribute       string        `json:"attribute"`
	Cap             int64         `json:"cap"`
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
	Category          string        `json:"category"`
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
		MaximumAttribute string `json:"max_attribute"`
	} `json:"luck_roll"`
	LuckMaximumAbility actor.Ability   `json:"-"`
	SkillStart         formula.Program `json:"skill_start"`
	PerksPerLevel      int64           `json:"perks_per_level"`
	SkillCost          struct {
		Normal int64 `json:"normal"`
		Tag    int64 `json:"tag"`
	} `json:"skill_cost"`
	SkillGroupCost struct {
		MembersPerStep int64 `json:"members_per_step"`
	} `json:"skill_group_cost"`
	// MonsterSkillGrowth is what a monster's unwritten fighting skills gain per level above 1
	// (docs/core/actor-sheet.md).
	MonsterSkillGrowth struct {
		PerLevel   int64    `json:"per_level"`
		Categories []string `json:"categories"`
	} `json:"monster_skill_growth"`
}

// SkillGroup is one row of skill_groups/core.json, the grouped sheet
// (docs/core/progression.md#the-grouped-sheet-variant).
type SkillGroup struct {
	Key        string   `json:"key"`
	Skills     []string `json:"skills"` // core:skill/<key> refs
	IsTracked  bool     `json:"is_tracked"`
	skillsKeys []string
}

// SkillKeys are the group's member skills as bare keys.
func (group *SkillGroup) SkillKeys() []string { return group.skillsKeys }

// Pipeline is combat/pipeline.json.
type Pipeline struct {
	CriticalDamage            formula.Program `json:"crit_damage"`
	AreaEvadeDamage           formula.Program `json:"area_evade_damage"`
	EvasionRate               formula.Program `json:"evasion_rate"`
	DodgeEvasionRate          formula.Program `json:"dodge_evasion_rate"`
	BlockRate                 formula.Program `json:"block_rate"`
	MagicBlockRate            formula.Program `json:"magic_block_rate"`
	BulwarkStaggerAP          int64           `json:"bulwark_stagger_ap"`
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
	// Adds is a modifier target whose attacker modifiers join the attack while the row applies.
	Adds string `json:"adds"`
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
	DamageMultiplier float64 `json:"damage_multiplier"`
}

// Maneuver is one row of maneuvers/core.json: a way to perform an action for more or fewer
// AP, with an attack add, a damage multiplier, and effects applied on a hit
// (docs/core/actions.md#maneuvers).
type Maneuver struct {
	Key              string          `json:"key"`
	Modifies         string          `json:"modifies"`
	AP               int64           `json:"ap"`
	Mana             int64           `json:"mana"`
	Attack           int64           `json:"attack"`
	DamageMultiplier float64         `json:"damage_multiplier"` // 0 when the row names none, read as 1
	Ranges           []string        `json:"ranges"`            // empty: any range
	WeaponProperties []string        `json:"weapon_properties"` // empty: any weapon
	Effects          []AppliedEffect `json:"effects"`
}

// Loadout is what an actor's two hands hold, in the mainHand and offHand vocabulary of
// ruleset/schemas/schema.json.
type Loadout struct {
	Main string `json:"main"`
	Off  string `json:"off"`
	// MainSkill and OffSkill are the weapons' skills, which a fighting stance may need.
	MainSkill string `json:"main_skill"`
	OffSkill  string `json:"off_skill"`
}

// HoldsTwoWeapons is a loadout with a weapon in each hand.
func (loadout *Loadout) HoldsTwoWeapons() bool {
	return loadout != nil && (loadout.Main == "light" || loadout.Main == "one_handed") && (loadout.Off == "light_weapon" || loadout.Off == "weapon")
}

// NoSequenceError is a loadout no attack_sequences row covers.
type NoSequenceError struct{ Loadout Loadout }

func (noSequence *NoSequenceError) Error() string {
	return fmt.Sprintf("no attack sequence fits main hand %q and off hand %q", noSequence.Loadout.Main, noSequence.Loadout.Off)
}

// Strike is one strike of an attack sequence: its hand and its attack value.
type Strike struct {
	Hand   string `json:"hand"`
	Attack int64  `json:"attack"`
}

// Jab is a loadout's off-hand strike alone: its AP and its attack value.
type Jab struct {
	AP     int64 `json:"ap"`
	Attack int64 `json:"attack"`
}

// AttackSequence is one row of attack_sequences/core.json: the loadouts it covers, the attack
// action's AP, its strikes in order, and the jab when the off hand holds a weapon
// (docs/core/combat.md#attacks-by-loadout).
type AttackSequence struct {
	Key     string   `json:"key"`
	Main    []string `json:"main"`
	Off     []string `json:"off"`
	AP      int64    `json:"ap"`
	Strikes []Strike `json:"strikes"`
	Jab     *Jab     `json:"jab"`
}

// AttackSequences is attack_sequences/core.json: the rows, what a full attack adds, and the
// circle cut.
type AttackSequences struct {
	FullAttack struct {
		MinExtraAP       int64   `json:"min_extra_ap"`
		ExtraMainAttacks []int64 `json:"extra_main_attacks"`
		ExtraOffAttack   int64   `json:"extra_off_attack"`
	} `json:"full_attack"`
	CircleCut struct {
		Attack int64    `json:"attack"`
		Main   []string `json:"main"`
	} `json:"circle_cut"`
	Rows []AttackSequence `json:"attack_sequences"`
}

// Covering is the row that covers a loadout, nil when none does.
func (sequences *AttackSequences) Covering(main, off string) *AttackSequence {
	for index := range sequences.Rows {
		var row *AttackSequence = &sequences.Rows[index]
		if slices.Contains(row.Main, main) && slices.Contains(row.Off, off) {
			return row
		}
	}
	return nil
}

// Stance is one row of stances/core.json: a fighting stance a perk unlocks, what the hands
// must hold, and its modifiers (docs/core/combat.md#fighting-stances). An empty list is any.
type Stance struct {
	Key        string              `json:"key"`
	Main       []string            `json:"main"`
	Off        []string            `json:"off"`
	MainSkills []string            `json:"main_skills"`
	OffSkills  []string            `json:"off_skills"`
	Modifiers  []modifier.Modifier `json:"modifiers"`
}

// Misfit is why a loadout does not hold what the stance needs, nil when it fits.
func (stance *Stance) Misfit(loadout Loadout) error {
	if (len(stance.Main) > 0 && !slices.Contains(stance.Main, loadout.Main)) || (len(stance.Off) > 0 && !slices.Contains(stance.Off, loadout.Off)) {
		return fmt.Errorf("stance %q does not fit main hand %q and off hand %q", stance.Key, loadout.Main, loadout.Off)
	}
	for _, hand := range []struct {
		name   string
		skills []string
		skill  string
	}{{"main", stance.MainSkills, loadout.MainSkill}, {"off", stance.OffSkills, loadout.OffSkill}} {
		if len(hand.skills) > 0 && !slices.Contains(hand.skills, hand.skill) {
			return fmt.Errorf("stance %q needs a %s-hand weapon of skill %s", stance.Key, hand.name, strings.Join(hand.skills, " or "))
		}
	}
	return nil
}

// Stances is stances/core.json: the AP rules and the rows.
type Stances struct {
	Rules struct {
		TurnStartAP      int64 `json:"turn_start_ap"`
		MidTurnAP        int64 `json:"mid_turn_ap"`
		ParryBandDivisor int64 `json:"parry_band_divisor"`
	} `json:"rules"`
	Rows []Stance `json:"stances"`
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
	Parameter string
	Target    string
	Magnitude float64
	Duration  float64 // seconds; 0 when the entry names none, read as 1 (instant)
	Radius    float64
}

// UnmarshalJSON keeps the entry and reads the fields dispel and the spell cost read.
func (applied *AppliedEffect) UnmarshalJSON(encoded []byte) error {
	var fields struct {
		Effect    string  `json:"effect"`
		Parameter string  `json:"parameter"`
		Target    string  `json:"target"`
		Magnitude float64 `json:"magnitude"`
		Duration  float64 `json:"duration"`
		Radius    float64 `json:"radius"`
	}
	if err := json.Unmarshal(encoded, &fields); err != nil {
		return err
	}
	applied.Encoded = append(json.RawMessage{}, encoded...)
	applied.Effect, applied.Parameter, applied.Target = fields.Effect, fields.Parameter, fields.Target
	applied.Magnitude, applied.Duration, applied.Radius = fields.Magnitude, fields.Duration, fields.Radius
	return nil
}

// Effect is what spellcraft needs of one row of effects/core.json
// (docs/modules/magic.md#uncraftable-effects).
type Effect struct {
	Craftable             *bool    `json:"craftable"` // nil reads true
	Key                   string   `json:"key"`
	Color                 string   `json:"color"`
	MagnitudeUnit         string   `json:"magnitudeUnit"`
	Channels              []string `json:"channels"`
	UncraftableParameters []string `json:"uncraftable_parameters"`
	// ParameterCategory is the kind of row a parameter must name, if any (docs/core/integrity.md).
	ParameterCategory string  `json:"parameter_category"`
	BaseCost          float64 `json:"baseCost"`
	// Phase is the effect's five-phase element, if any; a coating or an elemental enchantment
	// gives it to a bound weapon (docs/core/casting.md#bound-weapons).
	Phase string `json:"phase"`
}

// Feature is what a dev operation needs of one row of features/core.json: its kind and ranks.
type Feature struct {
	Key          string       `json:"key"`
	Type         string       `json:"type"`
	Ranks        int64        `json:"ranks"`
	Prerequisite Prerequisite `json:"prerequisite"`
	// Modifiers are the feature's own, as written; a session emits them at the held rank
	// (docs/backend/sessions.md#what-the-actor-brings-to-a-step-app-13).
	Modifiers []map[string]any `json:"modifiers"`
	// Share is what the perk lends to allies in range (docs/core/share.md); nil for most.
	Share *Share `json:"share"`
	// Command is the orders the perk gives a summoner's summons (docs/core/commanding.md); nil for most.
	Command *Command `json:"command"`
	// Unlocks are the refs the perk unlocks; a core:material ref is a bound weapon's material.
	Unlocks []string `json:"unlocks"`
	// BoundWeaponEffects are the on-hit effects the perk adds to the holder's bound weapons
	// (docs/modules/magic.md#bound-weapons).
	BoundWeaponEffects []AppliedEffect `json:"bound_weapon_effects"`
}

// Command is a perk's command object: the AP an ordered summon adds to its next turn, whether
// one order reaches one summon or all, and for how many turns.
type Command struct {
	AP    int64  `json:"ap"`
	Reach string `json:"reach"` // "one" or "all"
	Turns int64  `json:"turns"`
}

// Share is a perk's share object: the group it lends to, how far, and for how long.
type Share struct {
	Bonus string  `json:"bonus"` // core:summon_bonus/<key>
	Range float64 `json:"range"` // units
	Turns int64   `json:"turns"`
	// RankShare is the share per rank for this perk, in place of the group's; nil uses the group's.
	RankShare []float64 `json:"rank_share"`
}

// SummonBonus is one row of summons/bonuses.json: an attribute whose modifier is lent to a
// group of attributes, with the share per rank (M3-13, M3-19, M3-14).
type SummonBonus struct {
	Key        string    `json:"key"`
	Attribute  string    `json:"attribute"`
	Attributes []string  `json:"attributes"`
	RankShare  []float64 `json:"rank_share"`
}

// EmitAt is the feature's modifiers at a rank (docs/modules/progression.md#perks-and-traits): an
// add at value × rank, a multiply of m at 1 + (m − 1) × rank, a grant or a set once. Players and
// monsters emit a perk the same way (APP-13, M3-15).
func (feature *Feature) EmitAt(rank int64) []map[string]any {
	var emitted []map[string]any = make([]map[string]any, 0, len(feature.Modifiers))
	for _, written := range feature.Modifiers {
		var modifier map[string]any = make(map[string]any, len(written))
		for field, value := range written {
			modifier[field] = value
		}
		if value, isNumber := written["value"].(float64); isNumber {
			switch written["operation"] {
			case "add":
				modifier["value"] = value * float64(rank)
			case "multiply":
				modifier["value"] = 1 + (value-1)*float64(rank)
			}
		}
		emitted = append(emitted, modifier)
	}
	return emitted
}

// Prerequisite is what a perk needs before it is taken (docs/core/progression.md).
type Prerequisite struct {
	Level     int64              `json:"level"`
	Abilities map[string]int64   `json:"abilities"`
	Skills    []SkillRequirement `json:"skills"`
	Perks     []PerkRequirement  `json:"perks"`
}

// PerkRequirement is a prerequisite perk { perk, rank }: held at that rank or above, 1 when the
// rank is left out (M3-14; one shape, M5-70).
type PerkRequirement struct {
	Perk string `json:"perk"`
	Rank int64  `json:"rank"`
}

// UnmarshalJSON reads { perk, rank }; a missing rank needs rank 1.
func (requirement *PerkRequirement) UnmarshalJSON(encoded []byte) error {
	type held PerkRequirement
	var object held
	if err := json.Unmarshal(encoded, &object); err != nil {
		return err
	}
	if object.Rank == 0 {
		object.Rank = 1
	}
	*requirement = PerkRequirement(object)
	return nil
}

// SkillRequirement { any_of, min } is met when one of its skills reads at least Minimum; a
// perk that needs one skill lists one (one shape, M5-70).
type SkillRequirement struct {
	Skills  []string `json:"any_of"`
	Minimum int64    `json:"min"`
}

// Spell is what a cast needs of one row of spells/core.json (docs/core/casting.md).
type Spell struct {
	Key         string          `json:"key"`
	Label       string          `json:"label"`
	Description string          `json:"description"`
	Color       string          `json:"color"`
	Effects     []AppliedEffect `json:"effects"`
	// Summon is a summon spell's block, in place of a core:effect/summon entry; nil otherwise.
	Summon *SummonBlock `json:"summon"`
	Cost   struct {
		Reagents []string `json:"reagents"`
		Mana     int64    `json:"mana"`
	} `json:"cost"`
	Level         int64  `json:"level"`
	Concentration bool   `json:"concentration"`
	Delivery      string `json:"delivery"`
	Save          struct {
		Attribute string `json:"attribute"`
		Skill     string `json:"skill"`
	} `json:"save"`
}

// SummonBlock is what a summon spell calls (docs/core/casting.md#summons): a group for a
// duration, and for a minion spell a count, a number or dice.
type SummonBlock struct {
	Group    string      `json:"group"`
	Duration float64     `json:"duration"`
	Minion   bool        `json:"minion"`
	Count    SummonCount `json:"count"`
}

// SummonCount is a minion count: a fixed number, or dice the caller rolls (Dice set).
type SummonCount struct {
	Fixed int64
	Dice  string
}

// UnmarshalJSON reads a count written as a number or a dice expression.
func (count *SummonCount) UnmarshalJSON(raw []byte) error {
	var dice string
	if err := json.Unmarshal(raw, &dice); err == nil {
		count.Dice = dice
		return nil
	}
	return json.Unmarshal(raw, &count.Fixed)
}

// Range is the totals the count can make: the number itself, or the dice's low and high.
func (count SummonCount) Range() (int64, int64, error) {
	if count.Dice == "" {
		return max(count.Fixed, 1), max(count.Fixed, 1), nil
	}
	var number, sides, bonus int64
	if _, err := fmt.Sscanf(strings.NewReplacer("+", " +", "-", " -").Replace(count.Dice), "%dd%d %d", &number, &sides, &bonus); err != nil {
		if _, err := fmt.Sscanf(count.Dice, "%dd%d", &number, &sides); err != nil {
			return 0, 0, fmt.Errorf("count %q is not a dice expression", count.Dice)
		}
	}
	return number + bonus, number*sides + bonus, nil
}

// AsEffect is the block priced as the summon effect it replaces: its group, its count (a minion
// count's mean), its duration, and target ranged (docs/core/casting.md#summons).
func (block *SummonBlock) AsEffect() (AppliedEffect, error) {
	var count float64 = 1
	if block.Minion {
		low, high, err := block.Count.Range()
		if err != nil {
			return AppliedEffect{}, err
		}
		count = float64(low+high) / 2
	}
	return AppliedEffect{Effect: "core:effect/summon", Parameter: block.Group, Target: "ranged", Magnitude: count, Duration: block.Duration}, nil
}

// CastingSave is how an instant spell's save picks its outcome.
type CastingSave struct {
	AttributeMultiplier int64 `json:"attribute_multiplier"`
	FullMargin          int64 `json:"full_margin"`
	PartialPercent      int64 `json:"partial_percent"`
}

// Casting is casting/core.json: the caster flags that block a cast, and the cast penalty by
// the heaviest armor category worn.
type Casting struct {
	SummonLimit   int64            `json:"summon_limit"` // active summons before perks (docs/core/casting.md#summons)
	BoundWeapon   BoundWeapon      `json:"bound_weapon"`
	Save          CastingSave      `json:"save"`
	ArmorPenalty  map[string]int64 `json:"armor_penalty"`
	PreventedBy   []string         `json:"prevented_by"`
	SkillDiscount SkillDiscount    `json:"skill_discount"`
	SpellCost     SpellCost        `json:"spell_cost"`
	Research      Research         `json:"research"`
}

// BoundWeapon is what a conjured weapon starts as (docs/modules/magic.md#bound-weapons).
type BoundWeapon struct {
	Material   string `json:"material"`    // core:material/<key>
	DamageType string `json:"damage_type"` // a damage_types.json key
}

// SpellCost holds the spell cost formula's coefficients (docs/modules/magic.md#cost-formula).
type SpellCost struct {
	TargetMultiplier map[string]float64 `json:"target_multiplier"`
	MagnitudeFactor  float64            `json:"magnitude_factor"`
	AreaFactor       float64            `json:"area_factor"`
}

// Research is what one attempt to research an effect takes (docs/modules/magic.md#spell-knowledge).
type Research struct {
	Cost Cost  `json:"cost"`
	Days int64 `json:"days"`
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
	Label         string              `json:"label"`
	Description   string              `json:"description"`
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
	// WeaponSkill is a weapon's skill key; ArmorDRBase is an armor piece's drBase, nil for an
	// item that is not armor (docs/modules/armor.md).
	WeaponSkill string
	ArmorDRBase *int64
}

// Consumable is how an item is used and the effects it applies.
type Consumable struct {
	Delivery string
	Effects  []ConsumableEffect
	Toxicity int64 // what drinking it adds (docs/core/modifier.md#toxicity)
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
	Resolution  Resolution
	Derived     Table[DerivedStat]
	Skills      Table[Skill]
	SkillGroups Table[SkillGroup]
	// GroupOfSkill is the skill group each skill belongs to, by skill key.
	GroupOfSkill map[string]*SkillGroup
	Actions      Actions
	Wuxing       Wuxing
	Modules      []string // ids of the loaded modules
	Rest         Table[Occasion]
	Economy      Economy
	Progression  Progression
	Pipeline     Pipeline
	DamageTypes  Table[DamageType]
	Situations   Table[Situation]
	CalledShots  Table[CalledShot]
	Maneuvers    Table[Maneuver] // maneuvers/core.json, by key
	// AttackSequences is attack_sequences/core.json: how strikes stack by loadout.
	AttackSequences AttackSequences
	// Stances is stances/core.json: the fighting stances a perk unlocks.
	Stances        Stances
	Hazards        Table[Hazard]
	Spells         Table[Spell]
	Features       Table[Feature]     // features/core.json, for dev perks (docs/backend/dev-mode.md)
	SummonBonuses  Table[SummonBonus] // summons/bonuses.json, by ref core:summon_bonus/<key>
	Effects        Table[Effect]      // by ref, core:effect/<key>
	Casting        Casting
	Origins        modifier.OriginPolicies
	OriginReplaces map[string]bool   // origin ref: its policy is `replace`
	OriginFor      map[string]string // emitter: the origin ref it defaults to
	SecondsPerTurn int64             // actions/core.json
	StaminaPerAP   int64             // actions/core.json: stamina one AP is worth (docs/core/actions.md)
	// ToxicityDecayPerSecond is how fast stored toxicity falls (rest/core.json).
	ToxicityDecayPerSecond int64
	Conditions             Conditions
	Loot                   Loot

	Formulas map[string]*formula.Program // formulas/<name>.json by name
	// Files and ModuleFiles are what the ruleset was loaded from, for a command that reads rows
	// no table holds (author's ref index); Schemas is ruleset/schemas, nil unless given.
	Files             fs.FS
	ModuleFiles       []fs.FS
	Schemas           fs.FS
	AttributeModifier *formula.Program // formulas/attribute_modifier.json
	Price             *formula.Program
	LootLuck          *formula.Program
	Items             map[string]Item      // core:<category>/<key>
	MaterialPercent   map[string]int64     // material key: valueFactor * 100
	MaterialDefense   map[string]int64     // a material ref: its armor.defense, added to an armor piece's drBase
	MaterialAttack    map[string]int64     // a material ref: its weapon.attack (a bound weapon's attack)
	TemplateDamage    map[string]string    // a monster template ref: its damage_type, "" for none (element templates)
	DiceLadder        [][]string           // combat/dice_ladder.json: rungs of equal dice, smallest first
	Sizes             map[string]Size      // combat/dice_ladder.json: undersized, normal, oversized
	MonsterLevel      map[string]int64     // monster key
	MonsterLoot       map[string][]DropRow // monster key: its resolved loot rows

	// EffectRefs is every ref in effects/core.json, for the refs a dev operation names.
	EffectRefs map[string]bool
	// DevMode is set by the binary's --dev flag, never by data or input
	// (docs/backend/dev-mode.md).
	DevMode bool
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

// spellIdentity is what a spell built on a spell keeps of its own row, beside its key and source
// (docs/core/composition.md#a-spell-built-on-a-spell).
var spellIdentity = [...]string{"key", "label", "description", "source"}

// readSpells reads a spell file and resolves its spell bases before decoding, so every reader sees
// the flat row: a row with `base` is a copy of that spell with its `overrides` on top, an object
// field merged per field. A base may not itself name a base.
func (loader *rulesetLoader) readSpells(file string, target *[]Spell) {
	var raw struct {
		Spells []map[string]any `json:"spells"`
	}
	if !loader.read(file, &raw) {
		return
	}
	var byKey map[string]map[string]any = make(map[string]map[string]any, len(raw.Spells))
	for _, row := range raw.Spells {
		if key, ok := row["key"].(string); ok {
			byKey[key] = row
		}
	}
	var flat []map[string]any = make([]map[string]any, 0, len(raw.Spells))
	for _, row := range raw.Spells {
		baseRef, composed := row["base"].(string)
		if !composed {
			flat = append(flat, row)
			continue
		}
		base, exists := byKey[strings.TrimPrefix(baseRef, "core:spell/")]
		if !exists || !strings.HasPrefix(baseRef, "core:spell/") {
			loader.fail("%s: spell '%v': base spell '%s' does not exist", file, row["key"], baseRef)
			continue
		}
		if _, chained := base["base"]; chained {
			loader.fail("%s: spell '%v': base spell '%s' names a base itself; a spell composes one level", file, row["key"], baseRef)
			continue
		}
		var merged map[string]any = make(map[string]any, len(base))
		for field, value := range base {
			merged[field] = value
		}
		overrides, _ := row["overrides"].(map[string]any)
		for field, value := range overrides {
			block, isBlock := value.(map[string]any)
			inherited, hadBlock := merged[field].(map[string]any)
			if isBlock && hadBlock {
				var joined map[string]any = make(map[string]any, len(inherited)+len(block))
				for inner, innerValue := range inherited {
					joined[inner] = innerValue
				}
				for inner, innerValue := range block {
					joined[inner] = innerValue
				}
				value = joined
			}
			merged[field] = value
		}
		for _, field := range spellIdentity {
			delete(merged, field)
			if value, written := row[field]; written {
				merged[field] = value
			}
		}
		flat = append(flat, merged)
	}
	encoded, err := json.Marshal(flat)
	if err == nil {
		err = json.Unmarshal(encoded, target)
	}
	if err != nil {
		loader.fail("%s: %v", file, err)
	}
}

func (loader *rulesetLoader) fail(format string, arguments ...any) {
	loader.errs = append(loader.errs, fmt.Errorf(format, arguments...))
}

// Load reads the ruleset from the data directory held by fileSystem. The error, if any,
// joins one error per broken file or row.
func Load(fileSystem fs.FS, modules ...fs.FS) (*Ruleset, error) {
	var loader *rulesetLoader = &rulesetLoader{fileSystem: fileSystem}
	var ruleset *Ruleset = &Ruleset{Files: fileSystem, ModuleFiles: modules}

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
		StaminaPerAP   int64 `json:"stamina_per_ap"`
		Actions        []Action
	}
	if loader.read("actions/core.json", &actions) {
		actions.Pool.Ability = loader.ability("actions/core.json", actions.Pool.Attribute)
		if actions.Pool.PerPointDivisor <= 0 {
			loader.fail("actions/core.json: per_point_divisor must be positive")
		}
	}
	ruleset.SecondsPerTurn = actions.SecondsPerTurn
	ruleset.StaminaPerAP = actions.StaminaPerAP
	ruleset.Actions = Actions{Pool: actions.Pool, Actions: newTable(actions.Actions, func(action *Action) string { return action.Key })}

	var rest struct {
		Rest          []Occasion
		ToxicityDecay struct {
			PerSecond int64 `json:"per_second"`
		} `json:"toxicity_decay"`
	}
	loader.read("rest/core.json", &rest)
	ruleset.ToxicityDecayPerSecond = rest.ToxicityDecay.PerSecond
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

	var skillGroups struct {
		SkillGroups []SkillGroup `json:"skill_groups"`
	}
	loader.read("skill_groups/core.json", &skillGroups)
	for index := range skillGroups.SkillGroups {
		var group *SkillGroup = &skillGroups.SkillGroups[index]
		for _, ref := range group.Skills {
			group.skillsKeys = append(group.skillsKeys, ref[strings.LastIndex(ref, "/")+1:])
		}
	}
	ruleset.SkillGroups = newTable(skillGroups.SkillGroups, func(group *SkillGroup) string { return group.Key })
	ruleset.GroupOfSkill = make(map[string]*SkillGroup)
	for index := range ruleset.SkillGroups.rows {
		var group *SkillGroup = &ruleset.SkillGroups.rows[index]
		for _, key := range group.skillsKeys {
			ruleset.GroupOfSkill[key] = group
		}
	}

	var progression struct{ Progression Progression }
	if loader.read("progression/core.json", &progression) {
		var rules *Progression = &progression.Progression
		rules.LuckMaximumAbility = loader.ability("progression/core.json", rules.LuckRoll.MaximumAttribute)
		if rules.SkillCost.Normal <= 0 || rules.SkillCost.Tag <= 0 {
			loader.fail("progression/core.json: skill costs must be positive")
		}
		if rules.SkillGroupCost.MembersPerStep <= 0 {
			loader.fail("progression/core.json: skill_group_cost members_per_step must be positive")
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

	var maneuvers struct {
		Maneuvers []Maneuver `json:"maneuvers"`
	}
	loader.read("maneuvers/core.json", &maneuvers)
	loader.read("attack_sequences/core.json", &ruleset.AttackSequences)
	loader.read("stances/core.json", &ruleset.Stances)
	ruleset.Maneuvers = newTable(maneuvers.Maneuvers, func(maneuver *Maneuver) string { return maneuver.Key })

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

	loader.loadEffectRefs(ruleset)
	loader.loadConditions(ruleset)
	loader.loadHazards(ruleset)

	var spells struct{ Spells []Spell }
	loader.readSpells("spells/core.json", &spells.Spells)
	ruleset.Spells = newTable(spells.Spells, func(spell *Spell) string { return spell.Key })
	var features struct{ Features []Feature }
	loader.read("features/core.json", &features)
	ruleset.Features = newTable(features.Features, func(feature *Feature) string { return feature.Key })
	var summonBonuses struct {
		SummonBonuses []SummonBonus `json:"summon_bonuses"`
	}
	loader.read("summons/bonuses.json", &summonBonuses)
	ruleset.SummonBonuses = newTable(summonBonuses.SummonBonuses, func(bonus *SummonBonus) string { return "core:summon_bonus/" + bonus.Key })
	var effects struct{ Effects []Effect }
	loader.read("effects/core.json", &effects)
	ruleset.Effects = newTable(effects.Effects, func(effect *Effect) string { return "core:effect/" + effect.Key })
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
			Expression formula.Program `json:"expression"`
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

// baseName is the file a typed base reference names (docs/core/keys.md).
func baseName(ref string) string { return ref[strings.LastIndex(ref, "/")+1:] }

// itemBases lists, in file-name order, the keys of the bases in equipment/bases/<directory>
// that state is_item: true. Each is the item of its own key with no overrides
// (docs/core/composition.md#a-base-that-is-an-item).
func (loader *rulesetLoader) itemBases(directory string) []string {
	if directory == "" {
		return nil
	}
	var path string = "equipment/bases/" + directory
	entries, err := fs.ReadDir(loader.fileSystem, path)
	if err != nil {
		loader.fail("%s: %v", path, err)
		return nil
	}
	var keys []string
	for _, entry := range entries {
		var name string = entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".json") {
			continue
		}
		var base struct {
			IsItem *bool `json:"is_item"`
		}
		if !loader.read(path+"/"+name, &base) {
			continue
		}
		if base.IsItem == nil {
			loader.fail("%s/%s: is_item must be true or false", path, name)
			continue
		}
		if *base.IsItem {
			keys = append(keys, strings.TrimSuffix(name, ".json"))
		}
	}
	return keys
}

// MergeOverride merges an override value into its base value, as validate.py's _merge_field
// (docs/core/overrides.md#how-an-override-merges): an object merges field by field, each field by
// this same rule; `{"$extend": [...]}` appends to the base's list; `{"$replace": v}` is v whole;
// any other value replaces the base's.
func MergeOverride(base, override any) any {
	object, isObject := override.(map[string]any)
	if !isObject {
		return override
	}
	if items, extends := object["$extend"]; extends {
		list, isList := items.([]any)
		if !isList {
			return base
		}
		if baseList, hadList := base.([]any); hadList {
			return append(slices.Clone(baseList), list...)
		}
		return slices.Clone(list)
	}
	if value, replaces := object["$replace"]; replaces {
		return value
	}
	baseObject, hadObject := base.(map[string]any)
	if !hadObject {
		return override
	}
	var merged map[string]any = make(map[string]any, len(baseObject)+len(object))
	for field, value := range baseObject {
		merged[field] = value
	}
	for field, value := range object {
		merged[field] = MergeOverride(baseObject[field], value)
	}
	return merged
}

// weaponFields is what an item row says of its weapon: its dice and its skill ref.
type weaponFields struct {
	Damage struct{ Dice string }
	Skill  string
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
			Cost   *Cost
			Weapon *weaponFields
			Armor  *struct {
				DRBase int64 `json:"drBase"`
			}
			DefaultMaterials []string `json:"default_materials"`
			Type             string
			Consumable       *Consumable
		}
		var raw struct{ Equipment []map[string]any }
		loader.read("equipment/"+category.file+".json", &raw)
		var rows []map[string]any
		for _, key := range loader.itemBases(category.baseDirectory) {
			rows = append(rows, map[string]any{"key": key, "base": "core:" + category.category + "_base/" + key})
		}
		rows = append(rows, raw.Equipment...)
		for _, row := range rows {
			// A row is its base with its overrides merged in (as validate.py's compose_row),
			// then its own fields on top.
			var composed map[string]any = map[string]any{}
			if baseRef, _ := row["base"].(string); baseRef != "" {
				var base map[string]any
				loader.read(fmt.Sprintf("equipment/bases/%s/%s.json", category.baseDirectory, baseName(baseRef)), &base)
				composed = base
				if overrides, ok := row["overrides"].(map[string]any); ok {
					composed, _ = MergeOverride(base, overrides).(map[string]any)
				}
			}
			for field, value := range row {
				if field != "base" && field != "overrides" {
					composed[field] = value
				}
			}
			var fields struct {
				itemFields
				Key string
			}
			if encoded, err := json.Marshal(composed); err != nil || json.Unmarshal(encoded, &fields) != nil {
				loader.fail("equipment/%s.json: %v does not decode", category.file, row["key"])
				continue
			}
			var cost *Cost = fields.Cost
			var weapon *weaponFields = fields.Weapon
			var armor = fields.Armor
			var defaults []string = fields.DefaultMaterials
			if cost == nil {
				loader.fail("equipment/%s.json: %s has no cost", category.file, fields.Key)
				continue
			}
			copper, err := ruleset.Economy.ToCopper(*cost)
			if err != nil {
				loader.fail("equipment/%s.json: %s: %v", category.file, fields.Key, err)
				continue
			}
			var item Item = Item{
				CostCopper:       copper,
				TakesMaterial:    category.category == "weapon" || category.category == "armor",
				IsWeapon:         category.category == "weapon",
				DefaultMaterials: defaults,
				Type:             fields.Type,
				Consumable:       fields.Consumable,
			}
			if weapon != nil {
				item.Dice = weapon.Damage.Dice
				item.WeaponSkill = baseName(weapon.Skill)
			}
			if armor != nil && category.category == "armor" {
				var dr int64 = armor.DRBase
				item.ArmorDRBase = &dr
			}
			ruleset.Items["core:"+category.category+"/"+fields.Key] = item
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
			Armor       struct{ Defense int64 }
			Weapon      struct{ Attack int64 }
		}
	}
	loader.read("materials/core.json", &materials)
	ruleset.MaterialPercent = map[string]int64{}
	ruleset.MaterialDefense = map[string]int64{}
	ruleset.MaterialAttack = map[string]int64{}
	for _, material := range materials.Materials {
		ruleset.MaterialAttack["core:material/"+material.Key] = material.Weapon.Attack
		ruleset.MaterialPercent[material.Key] = int64(math.RoundToEven(material.ValueFactor * 100))
		ruleset.MaterialDefense["core:material/"+material.Key] = material.Armor.Defense
	}

	var monsters struct {
		Monsters []struct {
			Key       string
			Base      string
			Level     *int64
			Templates []string
			Overrides struct {
				Level *int64
				Loot  []DropRow
			}
		}
	}
	loader.read("monsters/core.json", &monsters)
	var templates struct {
		Templates []struct {
			Key        string
			Loot       []DropRow
			DamageType string `json:"damage_type"`
		} `json:"monster_templates"`
	}
	loader.read("monsters/templates.json", &templates)
	ruleset.TemplateDamage = map[string]string{}
	for _, template := range templates.Templates {
		ruleset.TemplateDamage["core:monster_template/"+template.Key] = template.DamageType
	}
	var templateLoot map[string][]DropRow = map[string][]DropRow{}
	for _, template := range templates.Templates {
		templateLoot[template.Key] = template.Loot
	}
	ruleset.MonsterLevel = map[string]int64{}
	var byKey = make(map[string]int, len(monsters.Monsters))
	for index, monster := range monsters.Monsters {
		byKey[monster.Key] = index
	}
	// A monster's level: its overrides', its row's, or its base monster's or creature type's,
	// since an override lists only what differs from its base (M5-73).
	var resolveLevel func(key string, seen map[string]bool) (int64, bool)
	resolveLevel = func(key string, seen map[string]bool) (int64, bool) {
		if level, done := ruleset.MonsterLevel[key]; done {
			return level, true
		}
		index, known := byKey[key]
		if !known || seen[key] {
			return 0, false
		}
		seen[key] = true
		var monster = monsters.Monsters[index]
		var level *int64 = cmp.Or(monster.Overrides.Level, monster.Level)
		if level == nil {
			if parent, derived := strings.CutPrefix(monster.Base, "core:monster/"); derived {
				inherited, found := resolveLevel(parent, seen)
				if !found {
					return 0, false
				}
				level = &inherited
			} else {
				var base struct{ Level int64 }
				if !loader.read("monsters/bases/"+baseName(monster.Base)+".json", &base) {
					return 0, false
				}
				level = &base.Level
			}
		}
		ruleset.MonsterLevel[key] = *level
		return *level, true
	}
	for _, monster := range monsters.Monsters {
		resolveLevel(monster.Key, map[string]bool{})
	}

	// A monster's loot: its type's or base monster's, then its templates', then its own.
	ruleset.MonsterLoot = map[string][]DropRow{}
	var resolve func(key string, seen map[string]bool) []DropRow
	resolve = func(key string, seen map[string]bool) []DropRow {
		if loot, done := ruleset.MonsterLoot[key]; done {
			return loot
		}
		index, known := byKey[key]
		if !known || seen[key] {
			return nil
		}
		seen[key] = true
		var monster = monsters.Monsters[index]
		var loot []DropRow
		if parent, derived := strings.CutPrefix(monster.Base, "core:monster/"); derived {
			loot = slices.Clone(resolve(parent, seen))
		} else {
			var base struct{ Loot []DropRow }
			loader.read("monsters/bases/"+baseName(monster.Base)+".json", &base)
			loot = base.Loot
		}
		for _, template := range monster.Templates {
			loot = append(slices.Clone(loot), templateLoot[template]...)
		}
		if monster.Overrides.Loot != nil {
			loot = monster.Overrides.Loot
		}
		ruleset.MonsterLoot[key] = loot
		return loot
	}
	for _, monster := range monsters.Monsters {
		resolve(monster.Key, map[string]bool{})
	}
}

func (loader *rulesetLoader) loadLoot(ruleset *Ruleset) {
	var raw struct {
		BudgetCopper formula.Program `json:"budget_cp"`
		RollsByLevel formula.Program `json:"rolls_by_level"`
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
	// A `drops` row replaces the bestiary's loot for its monster.
	var replaced map[string]bool = map[string]bool{}
	for _, drop := range raw.Drops {
		replaced[drop.Monster] = true
	}
	for key, rows := range ruleset.MonsterLoot {
		var monster string = "core:monster/" + key
		if !replaced[monster] {
			raw.Drops = append(raw.Drops, Drop{Monster: monster, Loot: slices.Clone(rows)})
		}
	}
	slices.SortFunc(raw.Drops, func(first, second Drop) int { return strings.Compare(first.Monster, second.Monster) })
	ruleset.Loot = Loot{
		BudgetCopper: raw.BudgetCopper,
		Tables:       newTable(raw.Tables, func(table *LootTable) string { return table.Key }),
		Drops:        newTable(raw.Drops, func(drop *Drop) string { return drop.Monster }),
	}
	for dropIndex := range raw.Drops {
		var drop *Drop = &raw.Drops[dropIndex]
		level, levelKnown := ruleset.MonsterLevel[drop.Monster[strings.LastIndex(drop.Monster, "/")+1:]]
		for rowIndex := range drop.Loot {
			var row *DropRow = &drop.Loot[rowIndex]
			if row.Table != "" && row.Rolls == nil && levelKnown {
				count, err := raw.RollsByLevel.Evaluate(nil, formula.Variable{Name: "level", Value: level})
				if err != nil {
					loader.fail("loot/core.json: rolls_by_level: %v", err)
					continue
				}
				row.Rolls = &count
			}
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

// loadEffectRefs indexes the effect registry by ref.
func (loader *rulesetLoader) loadEffectRefs(ruleset *Ruleset) {
	var effects struct{ Effects []struct{ Key string } }
	loader.read("effects/core.json", &effects)
	ruleset.EffectRefs = make(map[string]bool, len(effects.Effects))
	for _, effect := range effects.Effects {
		ruleset.EffectRefs["core:effect/"+effect.Key] = true
	}
}
