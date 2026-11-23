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
	"github.com/SETA1609/riftweave-core/backend/internal/formula"
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
	// ShieldDR is the carried shield's DR, which a Boss Strike bash adds to its damage roll.
	ShieldDR *int64 `json:"shield_dr"`
	// Perks are the attacker's perk refs, which every strike reads for a rider, as in combat.
	Perks []string `json:"perks"`
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

// The actions out of turn (docs/core/combat.md#attacks-out-of-turn), and why one did not fire.
const (
	ActionOpportunity = "opportunity"
	ActionReadied     = "readied"
	// actionAttack is the action an attack out of turn runs: the loadout's attack.
	actionAttack = "attack"

	ReasonDisengaged     = "disengaged"
	ReasonNotTriggered   = "not_triggered"
	ReasonNotEnoughHeld  = "not_enough_held"
	ReasonNoReactionLeft = "no_reaction_left"

	// reactionsFromAGI is the flag that makes the AGI modifier the most reactions between two turns,
	// and reactionsPerRound the modifier target that adds to it (Phalanx).
	reactionsFromAGI  = "reactions_from_agi"
	reactionsPerRound = "reactions_per_round"
)

// outOfTurnSubject names each action out of turn in a refusal.
var outOfTurnSubject = map[string]string{ActionOpportunity: "an opportunity attack", ActionReadied: "a readied attack"}

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
	// AfterBlock is Boss Strike's last tier: the attacker just blocked a hit in their band.
	AfterBlock bool `json:"after_block"`
	// Trigger is what a readied attack waits for, and Event what happened (enters_reach,
	// acts_in_reach, or none); Disengaged is a foe that left reach with Disengage (GD_S19_04).
	Trigger    string `json:"trigger"`
	Event      string `json:"event"`
	Disengaged bool   `json:"disengaged"`
	// ReactionsUsed is the attacks out of turn the attacker already made since its last turn.
	ReactionsUsed int64 `json:"reactions_used"`
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
	// Bash marks an off-hand strike made with the shield (Boss Strike), and StaggerAP is the AP
	// it costs the target. Both are left out on any other strike.
	Bash      bool   `json:"bash,omitempty"`
	StaggerAP *int64 `json:"stagger_ap,omitempty"`
	// Rider, Effects, Events, and Save are a rider perk's attempt on the strike, as in combat;
	// left out on a strike no rider rides.
	Rider   string             `json:"rider,omitempty"`
	Effects *[]json.RawMessage `json:"effects,omitempty"`
	Events  *[]string          `json:"events,omitempty"`
	Save    *combat.Save       `json:"save,omitempty"`
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
	// Fired, Reason, and GuardLost are an attack out of turn's: whether it fired, why not, and
	// the guard the AP it spent took. Left out on an attack on the actor's own turn.
	Fired     *bool  `json:"fired,omitempty"`
	Reason    string `json:"reason,omitempty"`
	GuardLost *int64 `json:"guard_lost,omitempty"`
}

// reactionsMost is the reactions an actor has between two of its turns: reactions_per_round, or the
// AGI modifier (at least 1) with reactions_from_agi, then its reactions_per_round modifiers (Phalanx).
func reactionsMost(ruleset *data.Ruleset, attacker Attacker) (int64, error) {
	var most int64 = ruleset.Actions.Pool.ReactionsPerRound
	if slices.Contains(modifier.Flags(attacker.Modifiers, 0, nil), reactionsFromAGI) {
		agility, err := ruleset.AttributeModifier.Evaluate(nil, formula.Variable{Name: "attribute", Value: attacker.Attributes.Get(actor.Agility)})
		if err != nil {
			return 0, err
		}
		most = max(1, agility)
	}
	return modifier.Value(most, modifier.Target(attacker.Modifiers, reactionsPerRound), ruleset.Origins, 0, nil)
}

// runOutOfTurn is an opportunity or a readied attack: one attack of the loadout. An opportunity
// attack is a reaction and costs no AP; a readied attack is paid from the AP held (`ap`) and spends
// no reaction. When it does not fire, nothing is spent.
func runOutOfTurn(ruleset *data.Ruleset, input Input) (Result, error) {
	var subject string = outOfTurnSubject[input.Action]
	if input.Action == ActionReadied && input.AP == nil {
		return Result{}, &data.MissingInputError{Subject: subject, Needs: "ap"}
	}
	if input.Action == ActionReadied && input.Trigger == "" {
		return Result{}, &data.MissingInputError{Subject: subject, Needs: "trigger"}
	}
	var row *data.AttackSequence = ruleset.AttackSequences.Covering(input.Loadout.Main, input.Loadout.Off)
	if row == nil {
		return Result{}, &data.NoSequenceError{Loadout: input.Loadout}
	}
	var cost int64
	var reason string
	switch input.Action {
	case ActionOpportunity:
		most, err := reactionsMost(ruleset, input.Attacker)
		if err != nil {
			return Result{}, err
		}
		if input.Disengaged {
			reason = ReasonDisengaged
		} else if input.ReactionsUsed >= most {
			reason = ReasonNoReactionLeft
		}
	case ActionReadied:
		cost = row.AP
		if input.Event != input.Trigger {
			reason = ReasonNotTriggered
		} else if *input.AP < cost {
			reason = ReasonNotEnoughHeld
		}
	}
	if reason != "" {
		var hp []int64 = make([]int64, len(input.Foes))
		for index, encoded := range input.Foes {
			var foe combat.Defender
			if err := json.Unmarshal(encoded, &foe); err != nil {
				return Result{}, fmt.Errorf("foe %d: %w", index, err)
			}
			hp[index] = foe.HitPoints
		}
		var fired bool
		var none int64
		return Result{Action: input.Action, Sequence: row.Key, Strikes: []StrikeResult{}, HP: hp, Fired: &fired, Reason: reason, GuardLost: &none}, nil
	}
	var inner Input = input
	inner.Action, inner.AP, inner.Trigger, inner.Event, inner.Disengaged, inner.ReactionsUsed = actionAttack, nil, "", "", false, 0
	result, err := Run(ruleset, inner)
	if err != nil {
		return Result{}, err
	}
	var fired bool = true
	var guardLost int64 = cost * ruleset.Actions.Pool.GuardPerAP
	result.Action, result.AP, result.Stamina, result.Fired, result.GuardLost = input.Action, cost, cost*ruleset.StaminaPerAP, &fired, &guardLost
	return result, nil
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
	case ActionOpportunity, ActionReadied:
		return runOutOfTurn(ruleset, input)
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
	// Boss Strike: from its jab tier a jab with a shield is a bash, and from its attack tier
	// the attack strikes with the shield too, as the row of a light off-hand weapon.
	var bash = &ruleset.AttackSequences.ShieldBash
	tier, err := on("boss_strike")
	if err != nil {
		return Result{}, err
	}
	var needed int64 = bash.Tiers.Attack
	if input.Action == "jab" {
		needed = bash.Tiers.Jab
	}
	var bashing bool = slices.Contains(bash.Off, input.Loadout.Off) && tier >= needed && input.Action != "circle_cut"
	if bashing {
		if input.Attacker.ShieldDR == nil {
			return Result{}, &data.MissingInputError{Subject: "a shield bash", Needs: "shield_dr"}
		}
		// The bash rolls on the Shields skill, never silently on the main-hand weapon's.
		if input.Attacker.OffSkill == nil {
			return Result{}, &data.MissingInputError{Subject: "a shield bash", Needs: "off_skill"}
		}
		var asLight data.Loadout = input.Loadout
		asLight.Off = bash.AsOff
		if row = ruleset.AttackSequences.Covering(asLight.Main, asLight.Off); row == nil {
			return Result{}, &data.NoSequenceError{Loadout: asLight}
		}
		if input.Action == "jab" && input.AfterBlock && tier >= bash.Tiers.AfterBlock && row.Jab != nil {
			var cheap data.AttackSequence = *row
			var jab data.Jab = *row.Jab
			jab.AP = bash.AfterBlockAP
			cheap.Jab = &jab
			row = &cheap
		}
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
		// A shield has no weapon dice, so a bash is never the Scissor Cut strike.
		var scissor bool = scissorLeft && strike.hand == "off" && struckByMain[foe] && !bashing
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
		var isBash bool = bashing && strike.hand == "off"
		var perks []string = input.Attacker.Perks
		if isBash {
			// A shield has no weapon attribute, so no rider rides a bash (GD_S05_04).
			damage, damageRoll, perks = Damage{Type: bash.DamageType}, *input.Attacker.ShieldDR+roll.Damage, nil
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
				WeaponProperties: properties, WeaponEffects: effects, Perks: perks,
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
			Rider: outcome.Rider, Effects: outcome.Effects, Events: outcome.Events, Save: outcome.Save,
		})
		if len(effects) > 0 {
			var landed []json.RawMessage = []json.RawMessage{}
			if outcome.WeaponEffects != nil {
				landed = *outcome.WeaponEffects
			}
			result.Strikes[len(result.Strikes)-1].WeaponEffects = &landed
		}
		if isBash {
			var stagger int64
			if hit && tier >= bash.Tiers.Stagger {
				stagger = bash.StaggerAP
			}
			result.Strikes[len(result.Strikes)-1].Bash, result.Strikes[len(result.Strikes)-1].StaggerAP = true, &stagger
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
