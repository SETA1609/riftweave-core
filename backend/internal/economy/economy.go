// Package economy converts a cost to copper and prices one trade (docs/core/economy.md).
// The price is the formula in formulas/price.json; the barter check is a resolution.
package economy

import (
	"fmt"
	"slices"

	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/formula"
	"github.com/SETA1609/riftweave-core/backend/internal/resolution"
)

// Barter is the haggling roll of a trade.
type Barter struct {
	SkillValue int64 `json:"skill_value"`
	Roll       int64 `json:"roll"`
}

// Input is the `economy` command's input; `op` says which fields it reads.
type Input struct {
	Operation    string     `json:"op"`
	Cost         *data.Cost `json:"cost"`
	Side         string     `json:"side"`
	Merchant     string     `json:"merchant"`
	ItemCost     *data.Cost `json:"item_cost"`
	ValuePercent *int64     `json:"value_percent"`
	Barter       *Barter    `json:"barter"`
}

// Result has `cp` for a conversion, the other fields for a trade.
type Result struct {
	Copper      *int64 `json:"cp,omitempty"`
	BaseCopper  *int64 `json:"base_cp,omitempty"`
	Shift       *int64 `json:"shift,omitempty"`
	Percent     *int64 `json:"percent,omitempty"`
	PriceCopper *int64 `json:"price_cp,omitempty"`
}

// Price is the base price in copper scaled by a value percent and a merchant percent, with
// the data's rounding.
func Price(ruleset *data.Ruleset, baseCopper, valuePercent, percent int64) (int64, error) {
	return ruleset.Price.Evaluate(nil,
		formula.Variable{Name: "base_cp", Value: baseCopper},
		formula.Variable{Name: "value_percent", Value: valuePercent},
		formula.Variable{Name: "percent", Value: percent})
}

// Run is the `economy` command.
func Run(ruleset *data.Ruleset, input Input) (Result, error) {
	switch input.Operation {
	case "convert":
		if input.Cost == nil {
			return Result{}, fmt.Errorf("economy convert needs cost")
		}
		copper, err := ruleset.Economy.ToCopper(*input.Cost)
		return Result{Copper: &copper}, err
	case "trade":
		return trade(ruleset, input)
	}
	return Result{}, fmt.Errorf("economy op %q is not defined", input.Operation)
}

func trade(ruleset *data.Ruleset, input Input) (Result, error) {
	if input.ItemCost == nil || (input.Side != "buy" && input.Side != "sell") {
		return Result{}, fmt.Errorf("economy trade needs side (buy or sell), merchant and item_cost")
	}
	merchant, ok := ruleset.Economy.Merchants.Get(input.Merchant)
	if !ok {
		return Result{}, fmt.Errorf("merchant %q is not defined", input.Merchant)
	}
	var rules *data.Barter = &ruleset.Economy.Barter
	var shift int64
	if input.Barter != nil {
		var check resolution.Result = resolution.Check(&ruleset.Resolution, input.Barter.SkillValue, 0, 0, input.Barter.Roll)
		switch {
		case slices.Contains(check.Tags, "fumble"):
			shift = -rules.FumblePenalty
		case check.Success:
			shift = min(rules.MaximumShift, formula.FloorDivide(*check.Margin, rules.MarginDivisor))
		}
	}
	var percent int64 = merchant.SellPercent + shift
	if input.Side == "buy" {
		percent = merchant.BuyPercent - shift
	}
	baseCopper, err := ruleset.Economy.ToCopper(*input.ItemCost)
	if err != nil {
		return Result{}, err
	}
	var valuePercent int64 = 100
	if input.ValuePercent != nil {
		valuePercent = *input.ValuePercent
	}
	price, err := Price(ruleset, baseCopper, valuePercent, percent)
	if err != nil {
		return Result{}, err
	}
	return Result{BaseCopper: &baseCopper, Shift: &shift, Percent: &percent, PriceCopper: &price}, nil
}
