# GD_S19_09_initiative-ties One rule for initiative ties

Date: 2026-11-21

- **Decision:** an initiative tie breaks by the higher AGI, then the higher LCK, then a d100 roll-off, lowest first. The rule is written once, in [derived.md](../core/derived.md#initiative-ties); [combat.md §7](../modules/combat.md#7-initiative) and `BE_S08_07_encounter-tracker-cli` link to it, and `test_initiative_ties.py` fails a rule doc that states its own. [V 2026-11-21]
- **Why:** combat.md said AGI then a d100 roll-off, and the encounter tracker's issue said AGI, then LCK, then the GM's order.
- **Offline:** the roll-off is made once, when the encounter starts, by the device that holds it, and kept in the encounter's session step, so the order replays on every device without a shared roll. [A] The GM's order was dropped as the last step: it is not a rule two apps can replay alike.
- **Open:** initiative itself stays fixed (`PER`, no roll). No command returns a tie list yet; the encounter tracker (`BE_S08_07`) computes the order and stores the roll-off. [A]
