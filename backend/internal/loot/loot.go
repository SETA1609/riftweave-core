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

// Item is a stack of one item, with the material that landed on it and, for a sized weapon,
// its size, its dice after the size step, and its attack change.
type Item struct {
	Material *string `json:"material,omitempty"`
	Size     *string `json:"size,omitempty"`
	Dice     *string `json:"dice,omitempty"`
	Attack   *int64  `json:"attack,omitempty"`
	Item     string  `json:"item"`
	Quantity int64   `json:"quantity"`
}

// Result is what dropped. A table input has `table`; a monster input has `tables`.
type Result struct {
	Table        *string   `json:"table,omitempty"`
	Tables       *[]string `json:"tables,omitempty"`
	Items        []Item    `json:"items"`
	CoinsCopper  int64     `json:"coins_cp"`
	ValueCopper  int64     `json:"value_cp"`
	BudgetCopper int64     `json:"budget_cp"`
	LooterShift  int64     `json:"looter_shift"`
}

type itemStack struct {
	reference, material, size string
	quantity                  int64
}

type roller struct {
	ruleset *data.Ruleset
	err     error
	items   []itemStack
}

// add stacks an item with what lands on it (docs/core/loot.md#what-lands-on-an-item): the
// row's material if the item is found in it, else its first default; the row's size on a
// weapon only.
func (roller *roller) add(reference string, quantity int64, material, size string) {
	item, ok := roller.ruleset.Items[reference]
	if !ok {
		roller.err = &data.NotDefinedError{Kind: "item", Key: reference}
		return
	}
	if !item.TakesMaterial {
		material = ""
	}
	if material != "" && len(item.DefaultMaterials) > 0 && !slices.Contains(item.DefaultMaterials, material) {
		material = item.DefaultMaterials[0]
	}
	if !item.IsWeapon || size == "normal" {
		size = ""
	}
	for index := range roller.items {
		var stack *itemStack = &roller.items[index]
		if stack.reference == reference && stack.material == material && stack.size == size {
			stack.quantity += quantity
			return
		}
	}
	roller.items = append(roller.items, itemStack{reference: reference, material: material, size: size, quantity: quantity})
}

// sized is a weapon's dice one size step up or down the ladder, and the size's attack.
func sized(ruleset *data.Ruleset, dice string, size data.Size) (string, error) {
	for rung, dices := range ruleset.DiceLadder {
		if slices.Contains(dices, dice) {
			var target int = rung + int(size.DiceStep)
			if target < 0 || target >= len(ruleset.DiceLadder) {
				return "", fmt.Errorf("%s has no rung %d steps away", dice, size.DiceStep)
			}
			return ruleset.DiceLadder[target][0], nil
		}
	}
	return "", fmt.Errorf("%s is not on the dice ladder", dice)
}

// rollTable rolls one table and returns the copper it yields; its items are added.
func (roller *roller) rollTable(table *data.LootTable, rolls []int64, shift int64, material, size string) int64 {
	var coins int64
	for _, roll := range rolls {
		var entryIndex int8 = table.EntryIndexByRoll[min(max(roll+shift, 1), 100)-1]
		if entryIndex < 0 {
			roller.err = fmt.Errorf("%s has no entry for roll %d", table.Key, roll)
			return 0
		}
		switch entry := &table.Entries[entryIndex]; entry.Kind {
		case "coins":
			coins += entry.CoinsCopper
		case "item":
			roller.add(entry.Item, entry.Quantity, material, size)
		}
	}
	return coins
}

// result sorts and values the items at list price: the price formula at the item's cost,
// a value percent of the material's value factor times the size's percent over 100, rounded
// down, and 100 percent.
func (roller *roller) result(coins int64, level int64) (Result, error) {
	if roller.err != nil {
		return Result{}, roller.err
	}
	var ruleset *data.Ruleset = roller.ruleset
	slices.SortFunc(roller.items, func(first, second itemStack) int {
		return cmp.Or(cmp.Compare(first.reference, second.reference), cmp.Compare(first.material, second.material),
			cmp.Compare(first.size, second.size))
	})
	var result Result = Result{Items: make([]Item, 0, len(roller.items)), CoinsCopper: coins}
	var value int64 = coins
	for _, stack := range roller.items {
		var percent int64 = 100
		var item Item = Item{Item: stack.reference, Quantity: stack.quantity}
		if stack.material != "" {
			var materialKey string = stack.material[strings.LastIndex(stack.material, "/")+1:]
			materialPercent, ok := ruleset.MaterialPercent[materialKey]
			if !ok {
				return Result{}, &data.NotDefinedError{Kind: "material", Key: stack.material}
			}
			percent = materialPercent
			var material string = stack.material
			item.Material = &material
		}
		if stack.size != "" {
			size, ok := ruleset.Sizes[stack.size]
			if !ok {
				return Result{}, &data.NotDefinedError{Kind: "size", Key: stack.size}
			}
			dice, err := sized(ruleset, ruleset.Items[stack.reference].Dice, size)
			if err != nil {
				return Result{}, err
			}
			percent = percent * size.ValuePercent / 100
			var sizeKey string = stack.size
			var attack int64 = size.Attack
			item.Size, item.Dice, item.Attack = &sizeKey, &dice, &attack
		}
		price, err := economy.Price(ruleset, ruleset.Items[stack.reference].CostCopper, percent, 100)
		if err != nil {
			return Result{}, err
		}
		value += price * stack.quantity
		result.Items = append(result.Items, item)
	}
	result.ValueCopper = value
	budget, err := ruleset.Loot.BudgetCopper.Evaluate(nil, formula.Variable{Name: "level", Value: level})
	result.BudgetCopper = budget
	return result, err
}

// looterShift is Luck plus the looter's loot_roll modifiers, after kernel stacking. Luck
// emits an innate add on loot_roll from formulas/loot_luck.json. No actor and no modifiers: 0.
func looterShift(ruleset *data.Ruleset, input Input) (int64, error) {
	var modifiers []modifier.Modifier = modifier.Target(input.Modifiers, "loot_roll")
	if input.Actor != nil {
		var attributes actor.Attributes = actor.Attributes{}.With(actor.Luck, input.Actor.Attributes.Get(actor.Luck))
		luck, err := ruleset.LootLuck.Evaluate(&attributes)
		if err != nil {
			return 0, err
		}
		var luckModifier modifier.Modifier = modifier.Modifier{
			Target:    "loot_roll",
			Operation: modifier.Add,
			Value:     float64(luck),
			Source:    "attr:lck",
			Origin:    "core:origin/innate",
		}
		modifiers = append([]modifier.Modifier{luckModifier}, modifiers...)
	}
	return modifier.Value(0, modifiers, ruleset.Origins, 0, nil)
}

// Run is the `loot` command. A rule error, such as the wrong number of rolls, is an error.
func Run(ruleset *data.Ruleset, input Input) (Result, error) {
	for _, roll := range input.Rolls {
		if roll < 1 || roll > 100 {
			return Result{}, fmt.Errorf("roll %d is not 1 to 100", roll)
		}
	}
	looter, err := looterShift(ruleset, input)
	if err != nil {
		return Result{}, err
	}
	var lootRoller *roller = &roller{ruleset: ruleset}
	if input.Table != "" {
		table, err := ruleset.Loot.Tables.Require("table", input.Table)
		if err != nil {
			return Result{}, err
		}
		if int64(len(input.Rolls)) != table.Rolls {
			return Result{}, fmt.Errorf("%s takes %d rolls, got %d", table.Key, table.Rolls, len(input.Rolls))
		}
		var coins int64 = lootRoller.rollTable(table, input.Rolls, looter, "", "")
		result, err := lootRoller.result(coins, table.Level)
		result.Table, result.LooterShift = &table.Key, looter
		return result, err
	}

	drop, ok := ruleset.Loot.Drops.Get(input.Monster)
	if !ok {
		return Result{}, fmt.Errorf("%q has no loot rows", input.Monster)
	}
	level, ok := ruleset.MonsterLevel[input.Monster[strings.LastIndex(input.Monster, "/")+1:]]
	if !ok {
		return Result{}, &data.NotDefinedError{Kind: "monster", Key: input.Monster}
	}
	var coins int64
	var used int = 0
	var take func(count int64) ([]int64, bool) = func(count int64) ([]int64, bool) {
		if int64(len(input.Rolls)-used) < count {
			return nil, false
		}
		var taken []int64 = input.Rolls[used : used+int(count)]
		used += int(count)
		return taken, true
	}
	var tables []string = []string{}
	for _, row := range drop.Loot {
		if row.Chance != nil {
			chance, ok := take(1)
			if !ok {
				return Result{}, fmt.Errorf("%s: ran out of rolls after %d", input.Monster, used)
			}
			// A row's chance after the looter shift, kept between 1 and 99.
			if chance[0] > min(max(*row.Chance+looter, 1), 99) {
				continue
			}
		}
		if row.Item != "" {
			var quantity int64 = 1
			if row.Quantity != nil {
				quantity = *row.Quantity
			}
			lootRoller.add(row.Item, quantity, row.Material, row.Size)
			continue
		}
		var table *data.LootTable = ruleset.Loot.Tables.At(row.TableIndex)
		tables = append(tables, table.Key)
		var rollCount int64 = table.Rolls
		if row.Rolls != nil {
			rollCount = *row.Rolls
		}
		rolls, ok := take(rollCount)
		if !ok {
			return Result{}, fmt.Errorf("%s: ran out of rolls after %d", input.Monster, used)
		}
		coins += lootRoller.rollTable(table, rolls, row.Shift+looter, row.Material, row.Size)
	}
	if used != len(input.Rolls) {
		return Result{}, fmt.Errorf("%s: %d rolls given, %d used", input.Monster, len(input.Rolls), used)
	}
	result, err := lootRoller.result(coins, level)
	result.Tables, result.LooterShift = &tables, looter
	return result, err
}
