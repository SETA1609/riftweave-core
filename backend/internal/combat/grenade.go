package combat

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/SETA1609/riftweave-core/backend/internal/data"
)

// grenadeType is the item type of a grenade, and areaDelivery the only delivery it has.
const (
	grenadeType  = "grenade"
	areaDelivery = "area"
	outside      = "outside"
)

// throwGrenade is a grenade thrown as an area attack (docs/core/combat.md#grenades): its own dice
// and damage type with no ability added, nothing past its radius, and its riders on a hit that is
// not evaded. The checks run in the order of kernel_reference.throw_grenade.
func throwGrenade(ruleset *data.Ruleset, input Input) (Result, error) {
	item, found := ruleset.Items[input.Grenade]
	if !found || item.Type != grenadeType || item.Consumable == nil {
		return Result{}, fmt.Errorf("%s is not a grenade", input.Grenade)
	}
	if input.Maneuver != "" || input.CalledShot != "" || input.Ammunition != nil {
		return Result{}, errors.New("a grenade is thrown without a maneuver, a called shot, or ammunition")
	}
	if input.Delivery != "" && input.Delivery != areaDelivery {
		return Result{}, errors.New("a grenade is an area attack")
	}
	if input.Damage.Ability != "" {
		return Result{}, errors.New("a grenade's damage adds no ability")
	}
	var dealt *data.GrenadeDamage = item.Consumable.Damage
	if dealt == nil {
		return Result{}, fmt.Errorf("the %s deals no damage: it is no attack", item.Label)
	}
	low, high, err := diceSpan(dealt.Dice)
	if err != nil {
		return Result{}, fmt.Errorf("%s: %w", input.Grenade, err)
	}
	if input.Damage.Roll < low || input.Damage.Roll > high {
		return Result{}, fmt.Errorf("the %s rolls %s: %d is outside %d to %d", item.Label, dealt.Dice, input.Damage.Roll, low, high)
	}
	if input.Damage.Type != "" && input.Damage.Type != dealt.Type {
		return Result{}, fmt.Errorf("the %s deals %s damage, not %s", item.Label, dealt.Type, input.Damage.Type)
	}
	var none []json.RawMessage = []json.RawMessage{}
	if input.BurstDistance > item.Consumable.Radius {
		return Result{Outcome: outside, HitPoints: input.Defender.HitPoints, Effects: &none}, nil
	}
	input.Grenade, input.BurstDistance = "", 0
	input.Delivery, input.Damage.Type = areaDelivery, dealt.Type
	result, err := strike(ruleset, input)
	if err != nil {
		return Result{}, err
	}
	result.Effects = &none
	if result.Outcome == "hit" && !result.Evaded {
		var riders []json.RawMessage = make([]json.RawMessage, 0, len(item.Consumable.Effects))
		for _, applied := range item.Consumable.Effects {
			encoded, err := json.Marshal(appliedEffectJSON(applied))
			if err != nil {
				return Result{}, err
			}
			riders = append(riders, encoded)
		}
		result.Effects = &riders
	}
	return result, nil
}

// appliedEffectJSON writes a consumable's applied effect as the data file has it.
func appliedEffectJSON(applied data.ConsumableEffect) map[string]any {
	var row map[string]any = map[string]any{"effect": applied.Effect}
	if applied.Parameter != "" {
		row["parameter"] = applied.Parameter
	}
	if applied.Magnitude != nil {
		row["magnitude"] = *applied.Magnitude
	}
	if applied.Duration != 0 {
		row["duration"] = applied.Duration
	}
	if applied.Origin != "" {
		row["origin"] = applied.Origin
	}
	return row
}

// diceSpan is the lowest and highest total of dice such as 2d6.
func diceSpan(dice string) (int64, int64, error) {
	count, sides, found := strings.Cut(dice, "d")
	n, errCount := strconv.ParseInt(count, 10, 64)
	s, errSides := strconv.ParseInt(sides, 10, 64)
	if !found || errCount != nil || errSides != nil || n < 1 || s < 1 {
		return 0, 0, fmt.Errorf("dice %q are not NdS", dice)
	}
	return n, n * s, nil
}
