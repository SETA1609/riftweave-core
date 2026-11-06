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
	Op           string     `json:"op"`
	Cost         *data.Cost `json:"cost"`
	Side         string     `json:"side"`
	Merchant     string     `json:"merchant"`
	ItemCost     *data.Cost `json:"item_cost"`
	ValuePercent *int64     `json:"value_percent"`
	Barter       *Barter    `json:"barter"`
}

// Result has `cp` for a conversion, the other fields for a trade.
type Result struct {
	CP      *int64 `json:"cp,omitempty"`
	BaseCP  *int64 `json:"base_cp,omitempty"`
	Shift   *int64 `json:"shift,omitempty"`
	Percent *int64 `json:"percent,omitempty"`
	PriceCP *int64 `json:"price_cp,omitempty"`
}

// Price is base_cp scaled by a value percent and a merchant percent, with the data's rounding.
func Price(rs *data.Ruleset, baseCP, valuePercent, percent int64) (int64, error) {
	return rs.Price.Eval(nil,
		formula.Var{Name: "base_cp", Val: baseCP},
		formula.Var{Name: "value_percent", Val: valuePercent},
		formula.Var{Name: "percent", Val: percent})
}

// Run is the `economy` command.
func Run(rs *data.Ruleset, in Input) (Result, error) {
	switch in.Op {
	case "convert":
		if in.Cost == nil {
			return Result{}, fmt.Errorf("economy convert needs cost")
		}
		cp, err := rs.Economy.ToCP(*in.Cost)
		return Result{CP: &cp}, err
	case "trade":
		return trade(rs, in)
	}
	return Result{}, fmt.Errorf("economy op %q is not defined", in.Op)
}

func trade(rs *data.Ruleset, in Input) (Result, error) {
	if in.ItemCost == nil || (in.Side != "buy" && in.Side != "sell") {
		return Result{}, fmt.Errorf("economy trade needs side (buy or sell), merchant and item_cost")
	}
	merchant, ok := rs.Economy.Merchants.Get(in.Merchant)
	if !ok {
		return Result{}, fmt.Errorf("merchant %q is not defined", in.Merchant)
	}
	rules := &rs.Economy.Barter
	var shift int64
	if in.Barter != nil {
		check := resolution.Check(&rs.Resolution, in.Barter.SkillValue, 0, 0, in.Barter.Roll)
		switch {
		case slices.Contains(check.Tags, "fumble"):
			shift = -rules.FumblePenalty
		case check.Success:
			shift = min(rules.MaxShift, formula.FloorDiv(*check.Margin, rules.MarginDivisor))
		}
	}
	percent := merchant.SellPercent + shift
	if in.Side == "buy" {
		percent = merchant.BuyPercent - shift
	}
	baseCP, err := rs.Economy.ToCP(*in.ItemCost)
	if err != nil {
		return Result{}, err
	}
	valuePercent := int64(100)
	if in.ValuePercent != nil {
		valuePercent = *in.ValuePercent
	}
	price, err := Price(rs, baseCP, valuePercent, percent)
	if err != nil {
		return Result{}, err
	}
	return Result{BaseCP: &baseCP, Shift: &shift, Percent: &percent, PriceCP: &price}, nil
}
