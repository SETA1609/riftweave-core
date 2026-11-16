#### APP-32 Look up items, and see what a craft is expected to give
- Status: todo
- Priority: low
- Needs: APP-27
- Modified: 2026-11-16
- Do: Owner request, 2026-11-16 (later work): an item lookup in both apps over weapons, armor, accessories, consumables, gear, materials, gems, and crafted items, resolved through their bases (composition.md), filterable by slot, material, phase, tier, and price. And an **expected crafting** view: pick a recipe, a base, and a material, and see the resulting item and its odds before crafting. The kernel has no crafting command in the contract yet [V 2026-11-16, `backend/contract/` has none], so this needs a `craft` (or `craft_preview`) command first, specced in docs/modules/crafting.md and checked by vectors like every command [A]; file that as its own issue when this starts.
- Keys: an item query command (or the `/data` reads); a crafting preview command
- Files: to split when started
- See: docs/modules/crafting.md, docs/core/composition.md, ruleset/data/crafting/
- Done when: split into issues, each with its own done line.
