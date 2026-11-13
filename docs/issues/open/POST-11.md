#### POST-11 Table module: result bands and crit on doubles
- Status: todo
- Priority: medium
- Needs: POST-10
- Do: Comparing two margins means subtracting at the table. Add two table variants: results read from the tens digit (a tens digit under the target's tens digit is a strong success, an equal one a plain success [A]), and a crit on doubles (11, 22, 33...), with Luck adding one more doubled number [A]. Show how each maps onto the core margin and crit window, so a referee can switch mid-campaign.
- Keys: result-band and crit rows in `ruleset/modules/table/`
- Files: ~5 (ruleset/modules/table/, docs/modules/table.md, ruleset/tests/vectors/)
- See: docs/core/resolution.md, docs/modules/table.md
- Done when: both variants have vectors, including the edges (00, 100, a tens digit equal to the target's); with the module off, every core vector is unchanged.
