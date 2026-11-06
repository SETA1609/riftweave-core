// Package loot rolls a loot table, or a monster's loot rows, and values what drops on the
// price scale (docs/core/loot.md). Rolls are inputs. Each table keeps a 100-slot lookup from
// roll to entry, built when the data loads, so a roll is one array read.
package loot

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/SETA1609/riftweave-core/backend/internal/actor"
	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/economy"
	"github.com/SETA1609/riftweave-core/backend/internal/formula"
	"github.com/SETA1609/riftweave-core/backend/internal/modifier"
)

// Input is the `loot` command's input: a table or a monster, and one d100 per roll.
type Input struct {
	Table     string              `json:"table"`
	Monster   string              `json:"monster"`
	Rolls     []int64             `json:"rolls"`
	Actor     *actor.Actor        `json:"actor"`
	Modifiers []modifier.Modifier `json:"modifiers"`
}

// Item is a stack of one item, with the material a loot row set.
type Item struct {
	Item     string  `json:"item"`
	Quantity int64   `json:"quantity"`
	Material *string `json:"material,omitempty"`
}

// Result is what dropped. A table input has `table`; a monster input has `tables`.
type Result struct {
	Table       *string   `json:"table,omitempty"`
	Tables      *[]string `json:"tables,omitempty"`
	Items       []Item    `json:"items"`
	CoinsCP     int64     `json:"coins_cp"`
	ValueCP     int64     `json:"value_cp"`
	BudgetCP    int64     `json:"budget_cp"`
	LooterShift int64     `json:"looter_shift"`
}

type stack struct {
	ref, material string
	quantity      int64
}

type roller struct {
	rs    *data.Ruleset
	items []stack
	err   error
}

func (r *roller) add(ref string, quantity int64, material string) {
	item, ok := r.rs.Items[ref]
	if !ok {
		r.err = fmt.Errorf("item %q is not defined", ref)
		return
	}
	if !item.TakesMaterial {
		material = ""
	}
	for i := range r.items {
		if r.items[i].ref == ref && r.items[i].material == material {
			r.items[i].quantity += quantity
			return
		}
	}
	r.items = append(r.items, stack{ref, material, quantity})
}

// table rolls one table and returns the copper it yields; its items are added.
func (r *roller) table(t *data.LootTable, rolls []int64, shift int64, material string) int64 {
	var coins int64
	for _, roll := range rolls {
		slot := t.ByRoll[min(max(roll+shift, 1), 100)-1]
		if slot < 0 {
			r.err = fmt.Errorf("%s has no entry for roll %d", t.Key, roll)
			return 0
		}
		switch entry := &t.Entries[slot]; entry.Kind {
		case "coins":
			coins += entry.CoinsCP
		case "item":
			r.add(entry.Item, entry.Quantity, material)
		}
	}
	return coins
}

// result sorts and values the items at list price: the price formula at the item's cost,
// the material's value factor, and 100 percent.
func (r *roller) result(coins int64, level int64) (Result, error) {
	if r.err != nil {
		return Result{}, r.err
	}
	rs := r.rs
	slices.SortFunc(r.items, func(a, b stack) int {
		return cmp.Or(cmp.Compare(a.ref, b.ref), cmp.Compare(a.material, b.material))
	})
	res := Result{Items: make([]Item, 0, len(r.items)), CoinsCP: coins}
	value := coins
	for _, s := range r.items {
		percent := int64(100)
		item := Item{Item: s.ref, Quantity: s.quantity}
		if s.material != "" {
			key := s.material[strings.LastIndex(s.material, "/")+1:]
			p, ok := rs.MaterialPercent[key]
			if !ok {
				return Result{}, fmt.Errorf("material %q is not defined", s.material)
			}
			percent = p
			material := s.material
			item.Material = &material
		}
		price, err := economy.Price(rs, rs.Items[s.ref].CostCP, percent, 100)
		if err != nil {
			return Result{}, err
		}
		value += price * s.quantity
		res.Items = append(res.Items, item)
	}
	res.ValueCP = value
	budget, err := rs.Loot.Budget.Eval(nil, formula.Var{Name: "level", Val: level})
	res.BudgetCP = budget
	return res, err
}

// looterShift is Luck plus the looter's loot_roll modifiers, after kernel stacking. Luck
// emits an innate add on loot_roll from formulas/loot_luck.json. No actor and no modifiers: 0.
func looterShift(rs *data.Ruleset, in Input) (int64, error) {
	mods := modifier.Target(in.Modifiers, "loot_roll")
	if in.Actor != nil {
		attrs := actor.Attrs{}.With(actor.Lck, in.Actor.Attributes.Get(actor.Lck))
		luck, err := rs.LootLuck.Eval(&attrs)
		if err != nil {
			return 0, err
		}
		luckMod := modifier.Modifier{Target: "loot_roll", Op: modifier.Add, Value: float64(luck), Source: "attr:lck", Origin: "core:origin/innate"}
		mods = append([]modifier.Modifier{luckMod}, mods...)
	}
	return modifier.Value(0, mods, rs.Origins, 0, nil)
}

// Run is the `loot` command. A rule error, such as the wrong number of rolls, is an error.
func Run(rs *data.Ruleset, in Input) (Result, error) {
	for _, roll := range in.Rolls {
		if roll < 1 || roll > 100 {
			return Result{}, fmt.Errorf("roll %d is not 1 to 100", roll)
		}
	}
	looter, err := looterShift(rs, in)
	if err != nil {
		return Result{}, err
	}
	r := &roller{rs: rs}
	if in.Table != "" {
		t, ok := rs.Loot.Tables.Get(in.Table)
		if !ok {
			return Result{}, fmt.Errorf("table %q is not defined", in.Table)
		}
		if int64(len(in.Rolls)) != t.Rolls {
			return Result{}, fmt.Errorf("%s takes %d rolls, got %d", t.Key, t.Rolls, len(in.Rolls))
		}
		coins := r.table(t, in.Rolls, looter, "")
		res, err := r.result(coins, t.Level)
		res.Table, res.LooterShift = &t.Key, looter
		return res, err
	}

	drop, ok := rs.Loot.Drops.Get(in.Monster)
	if !ok {
		return Result{}, fmt.Errorf("%q has no loot rows", in.Monster)
	}
	level, ok := rs.MonsterLevel[in.Monster[strings.LastIndex(in.Monster, "/")+1:]]
	if !ok {
		return Result{}, fmt.Errorf("monster %q is not defined", in.Monster)
	}
	var coins int64
	used := 0
	take := func(n int64) ([]int64, bool) {
		if int64(len(in.Rolls)-used) < n {
			return nil, false
		}
		taken := in.Rolls[used : used+int(n)]
		used += int(n)
		return taken, true
	}
	tables := []string{}
	for _, row := range drop.Loot {
		if row.Chance != nil {
			chance, ok := take(1)
			if !ok {
				return Result{}, fmt.Errorf("%s: ran out of rolls after %d", in.Monster, used)
			}
			// A row's chance after the looter shift, kept between 1 and 99.
			if chance[0] > min(max(*row.Chance+looter, 1), 99) {
				continue
			}
		}
		if row.Item != "" {
			quantity := int64(1)
			if row.Quantity != nil {
				quantity = *row.Quantity
			}
			r.add(row.Item, quantity, row.Material)
			continue
		}
		t := rs.Loot.Tables.At(row.TableIdx)
		tables = append(tables, t.Key)
		n := t.Rolls
		if row.Rolls != nil {
			n = *row.Rolls
		}
		rolls, ok := take(n)
		if !ok {
			return Result{}, fmt.Errorf("%s: ran out of rolls after %d", in.Monster, used)
		}
		coins += r.table(t, rolls, row.Shift+looter, row.Material)
	}
	if used != len(in.Rolls) {
		return Result{}, fmt.Errorf("%s: %d rolls given, %d used", in.Monster, len(in.Rolls), used)
	}
	res, err := r.result(coins, level)
	res.Tables, res.LooterShift = &tables, looter
	return res, err
}
