# Keys

Author: SETA1609

A cross-reference is a namespaced string. `core:weapon/shortsword`. A numeric id is not a reference, and a row's identity is its `key`. A few vocabulary rows (abilities, Wuxing phases and cycles) have a string `id` in place of a `key`. A tier's integer `id` is its rank on the ladder, and `tier` on a material or gem holds that number. A module key is `mod:copper_fiber`.

The validator rejects a duplicate key. It does not accept `oneOf [string, id]`.

A rename or removal is a breaking change. What that costs, and how a key is deprecated first, is in [versioning.md](versioning.md).
