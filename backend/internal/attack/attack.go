// Package attack runs one attack action by loadout (docs/core/combat.md#attacks-by-loadout):
// the attack, the jab, the full attack, or the circle cut. The loadout's row of
// attack_sequences/core.json gives the AP and the strikes; perks add strikes or soften the
// penalties through modifier targets; each strike is its own roll through package combat.
package attack

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"

	"github.com/SETA1609/riftweave-core/backend/internal/actor"
	"github.com/SETA1609/riftweave-core/backend/internal/combat"
	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/modifier"
)

// Attacker is who makes the strikes.
type Attacker struct {
	Skill int64 `json:"skill"`
	// OffSkill is the off-hand weapon skill; nil reads the main-hand skill.
	OffSkill   *int64           `json:"off_skill"`
	Attributes actor.Attributes `json:"attributes"`
	Flags      []string         `json:"flags"`
	// WeaponProperties and WeaponEffects are the main-hand weapon's, as in combat.
	WeaponProperties []string             `json:"weapon_properties"`
	WeaponEffects    []data.AppliedEffect `json:"weapon_effects"`
	// OffWeaponProperties and OffWeaponEffects are the off-hand weapon's; left out, it has
	// none (an off-hand weapon never borrows the main hand's bane or binding).
	OffWeaponProperties []string             `json:"off_weapon_properties"`
	OffWeaponEffects    []data.AppliedEffect `json:"off_weapon_effects"`
	// Modifiers are the attacker's perks and fighting stance: adds on full_attack_strikes,
	// full_attack_off_strikes, mainhand_penalty, offhand_penalty, attack_roll,
	// attack_single_opponent, attack_after_parry, and unarmed_damage, and the carry, circle
	// cut, off_hand_parries, and carry_on_crit grants.
	Modifiers []modifier.Modifier `json:"modifiers"`
}

// Damage is a hand's damage type and the attribute that modifies it.
type Damage struct {
	Type    string `json:"type"`
	Ability string `json:"ability"`
}

// Roll is one strike's attack and damage roll.
type Roll struct {
	Attack int64 `json:"attack"`
	Damage int64 `json:"damage"`
}

// Input is the `attack_sequence` command's input.
type Input struct {
	Action    string              `json:"action"`
	Loadout   data.Loadout        `json:"loadout"`
	AP        *int64              `json:"ap"`
	Attacker  Attacker            `json:"attacker"`
	Foes      []json.RawMessage   `json:"foes"`
	Damage    Damage              `json:"damage"`
	OffDamage *Damage             `json:"off_damage"`
	Rolls     []Roll              `json:"rolls"`
	Targets   []int64             `json:"targets"`
	Modifiers []modifier.Modifier `json:"modifiers"`
	// Situations and Summoning go to every strike, as in combat (flanking, beside_summon, ...).
	Situations []string `json:"situations"`
	Summoning  bool     `json:"summoning"`
	// AfterParry is Espada y Daga's opening: the attacker parried a hit since their last strike.
	AfterParry bool `json:"after_parry"`
}

// StrikeResult is one strike made.
type StrikeResult struct {
	Hand    string `json:"hand"`
	Foe     int64  `json:"foe"`
	Attack  int64  `json:"attack"`
	Outcome string `json:"outcome"`
	Target  int64  `json:"target"`
	Margin  *int64 `json:"margin"`
	Crit    bool   `json:"crit"`
	Damage  int64  `json:"damage"`
	HP      int64  `json:"hp"`
	Carried bool   `json:"carried"`
	Free    bool   `json:"free"`
	Scissor bool   `json:"scissor"`
	// WeaponEffects are the effects that landed, listed only when the strike's hand has any.
	WeaponEffects *[]json.RawMessage `json:"weapon_effects,omitempty"`
}

// Result is the action's outcome.
type Result struct {
	Action   string         `json:"action"`
	Sequence string         `json:"sequence"`
	AP       int64          `json:"ap"`
	Stamina  int64          `json:"stamina"`
	Strikes  []StrikeResult `json:"strikes"`
	HP       []int64        `json:"hp"`
	Lost     int64          `json:"lost"`
}

type planned struct {
	hand    string
	attack  int64
	foe     int64
	free    bool
	carried bool
}

// soften is a negative strike value after mainhand_penalty or offhand_penalty, never past 0.
func soften(value, softening int64) int64 {
	if value < 0 {
		return min(value+softening, 0)
	}
	return value
}

// plan is the action's strikes in order, as hand and attack value, and the AP it spends.
func plan(ruleset *data.Ruleset, input *Input, row *data.AttackSequence, flags []string, on func(string) (int64, error)) ([]planned, int64, error) {
	var sequences *data.AttackSequences = &ruleset.AttackSequences
	mainSoft, err := on("mainhand_penalty")
	if err != nil {
		return nil, 0, err
	}
	offSoft, err := on("offhand_penalty")
	if err != nil {
		return nil, 0, err
	}
	var softening map[string]int64 = map[string]int64{"main": mainSoft, "off": offSoft}
	var table []planned = make([]planned, 0, len(row.Strikes))
	for _, strike := range row.Strikes {
		table = append(table, planned{hand: strike.Hand, attack: soften(strike.Attack, softening[strike.Hand])})
	}
	switch input.Action {
	case "attack", "jab":
		var cost int64 = row.AP
		if input.Action == "jab" {
			if row.Jab == nil {
				return nil, 0, fmt.Errorf("a jab needs a weapon in the off hand (%s has none)", row.Key)
			}
			table, cost = []planned{{hand: "off", attack: soften(row.Jab.Attack, offSoft)}}, row.Jab.AP
		}
		if input.AP != nil && *input.AP < cost {
			return nil, 0, fmt.Errorf("%s needs %d AP", input.Action, cost)
		}
		return table, cost, nil
	}
	var need int64 = row.AP + sequences.FullAttack.MinExtraAP
	if input.AP == nil || *input.AP < need {
		return nil, 0, fmt.Errorf("%s needs %d AP", input.Action, need)
	}
	if input.Action == "circle_cut" {
		if !slices.Contains(flags, "can_circle_cut") {
			return nil, 0, &data.MissingInputError{Subject: "circle_cut", Needs: "a perk that grants can_circle_cut"}
		}
		if !slices.Contains(sequences.CircleCut.Main, input.Loadout.Main) {
			return nil, 0, &data.MissingInputError{Subject: "circle_cut", Needs: either(sequences.CircleCut.Main)}
		}
		var strikes []planned = make([]planned, len(input.Foes))
		for index := range strikes {
			strikes[index] = planned{hand: "main", attack: sequences.CircleCut.Attack}
		}
		return strikes, *input.AP, nil
	}
	extra, err := on("full_attack_strikes")
	if err != nil {
		return nil, 0, err
	}
	var count int64 = max(min(extra, int64(len(sequences.FullAttack.ExtraMainAttacks))), 0)
	for _, value := range sequences.FullAttack.ExtraMainAttacks[:count] {
		table = append(table, planned{hand: "main", attack: value})
	}
	offExtra, err := on("full_attack_off_strikes")
	if err != nil {
		return nil, 0, err
	}
	for _, strike := range table {
		if strike.hand == "off" {
			if offExtra >= 1 {
				table = append(table, planned{hand: "off", attack: strike.attack + sequences.FullAttack.ExtraOffAttack})
			}
			break
		}
	}
	if slices.Contains(flags, "off_hand_parries") {
		// Espada y Daga: the off-hand weapon parries instead of striking in a full attack.
		table = slices.DeleteFunc(table, func(strike planned) bool { return strike.hand == "off" })
	}
	return table, *input.AP, nil
}

// either names the main hands as "a two_handed main hand or ...".
func either(mains []string) string {
	var text string
	for index, main := range mains {
		if index > 0 {
			text += " or "
		}
		text += "a " + main + " main hand"
	}
	return text
}

// Run is the `attack_sequence` command.
func Run(ruleset *data.Ruleset, input Input) (Result, error) {
	switch input.Action {
	case "attack", "jab", "full_attack", "circle_cut":
	default:
		return Result{}, &data.NotDefinedError{Kind: "action", Key: input.Action}
	}
	var row *data.AttackSequence = ruleset.AttackSequences.Covering(input.Loadout.Main, input.Loadout.Off)
	if row == nil {
		return Result{}, &data.NoSequenceError{Loadout: input.Loadout}
	}
	var flags []string = modifier.Flags(input.Attacker.Modifiers, 0, nil)
	var on func(string) (int64, error) = func(target string) (int64, error) {
		return modifier.Value(0, modifier.Target(input.Attacker.Modifiers, target), ruleset.Origins, 0, nil)
	}
	strikes, actionPoints, err := plan(ruleset, &input, row, flags, on)
	if err != nil {
		return Result{}, err
	}
	var foes []combat.Defender = make([]combat.Defender, len(input.Foes))
	for index, encoded := range input.Foes {
		if err := json.Unmarshal(encoded, &foes[index]); err != nil {
			return Result{}, fmt.Errorf("foe %d: %w", index, err)
		}
	}
	var aims []int64
	switch input.Action {
	case "full_attack":
		aims = make([]int64, len(strikes))
	case "circle_cut":
		for index := range foes {
			aims = append(aims, int64(index))
		}
	default:
		aims = input.Targets
		if aims == nil {
			aims = make([]int64, len(strikes))
		}
		if len(aims) != len(strikes) {
			return Result{}, fmt.Errorf("targets names %d foes for %d strikes", len(aims), len(strikes))
		}
	}
	for _, aim := range aims {
		if aim < 0 || aim >= int64(len(foes)) {
			return Result{}, fmt.Errorf("target %d is not a foe in reach", aim)
		}
		if foes[aim].HitPoints == 0 {
			return Result{}, fmt.Errorf("foe %d is already down", aim)
		}
	}
	var hitPoints []int64 = make([]int64, len(foes))
	for index := range foes {
		hitPoints[index] = foes[index].HitPoints
	}
	// A circle cut strikes each foe once, so nothing carries.
	var carries int64
	switch {
	case input.Action == "circle_cut":
	case slices.Contains(flags, "reaping_arc"):
		carries = math.MaxInt64
	case slices.Contains(flags, "carry_through"):
		carries = 1
	}
	for index := range strikes {
		strikes[index].foe = aims[index]
	}
	var queue []planned = strikes
	// A fighting stance's attack terms: attack_roll on every strike, attack_single_opponent
	// with one foe in reach, attack_after_parry on the first main-hand strike after a parry,
	// and unarmed_damage on an empty main hand.
	var stanceAttack []modifier.Modifier = modifier.Target(input.Attacker.Modifiers, "attack_roll")
	if len(foes) == 1 {
		stanceAttack = append(stanceAttack, modifier.Target(input.Attacker.Modifiers, "attack_single_opponent")...)
	}
	var afterParry []modifier.Modifier
	if input.AfterParry {
		afterParry = modifier.Target(input.Attacker.Modifiers, "attack_after_parry")
	}
	var unarmed int64
	if input.Loadout.Main == "empty" {
		if unarmed, err = on("unarmed_damage"); err != nil {
			return Result{}, err
		}
	}
	var asAttackRoll func([]modifier.Modifier) []modifier.Modifier = func(mods []modifier.Modifier) []modifier.Modifier {
		var renamed []modifier.Modifier = make([]modifier.Modifier, len(mods))
		for index, mod := range mods {
			renamed[index] = mod
			renamed[index].Target = "attack_roll"
		}
		return renamed
	}
	var rolls []Roll = input.Rolls
	var struckByMain map[int64]bool = map[int64]bool{}
	var scissorLeft bool = input.Action == "full_attack" && slices.Contains(flags, "scissor_cut")
	var result Result = Result{Action: input.Action, Sequence: row.Key, AP: actionPoints, Strikes: []StrikeResult{}, HP: hitPoints}
	for index := 0; index < len(queue); index++ {
		var strike *planned = &queue[index]
		var carried bool = strike.carried
		if hitPoints[strike.foe] == 0 {
			var standing int64 = -1
			for foe, left := range hitPoints {
				if left > 0 {
					standing = int64(foe)
					break
				}
			}
			if carries == 0 || standing < 0 {
				result.Lost++
				continue
			}
			carries--
			carried = true
			if input.Action == "full_attack" {
				for later := index; later < len(queue); later++ {
					queue[later].foe = standing
				}
			}
			strike.foe = standing
		}
		if len(rolls) == 0 {
			return Result{}, fmt.Errorf("the %s needs a roll for strike %d", input.Action, len(result.Strikes)+1)
		}
		var roll Roll = rolls[0]
		rolls = rolls[1:]
		var foe int64 = strike.foe
		var scissor bool = scissorLeft && strike.hand == "off" && struckByMain[foe]
		var damage Damage = input.Damage
		var skill int64 = input.Attacker.Skill
		var properties []string = input.Attacker.WeaponProperties
		var effects []data.AppliedEffect = input.Attacker.WeaponEffects
		if strike.hand == "off" {
			if input.OffDamage != nil {
				damage = *input.OffDamage
			}
			if input.Attacker.OffSkill != nil {
				skill = *input.Attacker.OffSkill
			}
			properties, effects = input.Attacker.OffWeaponProperties, input.Attacker.OffWeaponEffects
		}
		var damageRoll int64 = roll.Damage + unarmed
		if scissor {
			damageRoll *= 2
		}
		var strikeModifiers []modifier.Modifier = slices.Clone(input.Modifiers)
		if strike.attack != 0 {
			strikeModifiers = append(strikeModifiers, modifier.Modifier{Target: "attack_roll", Operation: modifier.Add, Value: float64(strike.attack), Source: "strike:" + strike.hand})
		}
		strikeModifiers = append(strikeModifiers, asAttackRoll(stanceAttack)...)
		if strike.hand == "main" && len(afterParry) > 0 {
			strikeModifiers = append(strikeModifiers, asAttackRoll(afterParry)...)
			afterParry = nil
		}
		var defender combat.Defender = foes[foe]
		defender.HitPoints = hitPoints[foe]
		outcome, err := combat.Run(ruleset, combat.Input{
			Attacker: combat.Attacker{
				Skill: skill, Roll: roll.Attack, Attributes: input.Attacker.Attributes, Flags: input.Attacker.Flags,
				WeaponProperties: properties, WeaponEffects: effects,
			},
			Defender:   defender,
			Damage:     combat.Damage{Roll: damageRoll, Type: damage.Type, Ability: damage.Ability},
			Modifiers:  strikeModifiers,
			Situations: input.Situations,
			Summoning:  input.Summoning,
		})
		if err != nil {
			return Result{}, err
		}
		var hit bool = outcome.Outcome == "hit"
		if scissor && hit {
			scissorLeft = false
		} else if scissor {
			scissor = false
		}
		if hit && strike.hand == "main" {
			struckByMain[foe] = true
		}
		var dropped bool = hitPoints[foe] > 0 && outcome.HitPoints == 0
		hitPoints[foe] = outcome.HitPoints
		result.Strikes = append(result.Strikes, StrikeResult{
			Hand: strike.hand, Foe: foe, Attack: strike.attack, Outcome: outcome.Outcome, Target: outcome.Target,
			Margin: outcome.Margin, Crit: outcome.Critical, Damage: outcome.Damage, HP: outcome.HitPoints,
			Carried: carried, Free: strike.free, Scissor: scissor && hit,
		})
		if len(effects) > 0 {
			var landed []json.RawMessage = []json.RawMessage{}
			if outcome.WeaponEffects != nil {
				landed = *outcome.WeaponEffects
			}
			result.Strikes[len(result.Strikes)-1].WeaponEffects = &landed
		}
		if dropped && index == len(queue)-1 && carries > 0 && slices.ContainsFunc(hitPoints, func(left int64) bool { return left > 0 }) {
			queue = append(queue, planned{hand: strike.hand, attack: strike.attack, foe: foe, free: true})
		}
		// Long Guard: a crit that leaves its foe standing carries as a drop would, to the first
		// other standing foe.
		var other int64 = -1
		for index, left := range hitPoints {
			if left > 0 && int64(index) != foe {
				other = int64(index)
				break
			}
		}
		// In an attack, only a next strike aimed at the critted foe is carried; one the caller
		// aimed elsewhere keeps its foe, and the carry is not spent.
		var aimedElsewhere bool = input.Action != "full_attack" && index < len(queue)-1 && queue[index+1].foe != foe
		if outcome.Critical && !dropped && slices.Contains(flags, "carry_on_crit") && carries > 0 && other >= 0 && !aimedElsewhere {
			carries--
			switch {
			case index < len(queue)-1 && input.Action == "full_attack":
				for later := index + 1; later < len(queue); later++ {
					queue[later].foe = other
				}
				queue[index+1].carried = true
			case index < len(queue)-1:
				queue[index+1].foe, queue[index+1].carried = other, true
			default:
				queue = append(queue, planned{hand: strike.hand, attack: strike.attack, foe: other, free: true, carried: true})
			}
		}
	}
	result.Stamina = actionPoints * ruleset.StaminaPerAP
	return result, nil
}
