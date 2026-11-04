# Keys

Author: SETA1609

A cross-reference is a namespaced string. `core:weapon/shortsword`. A numeric id is not a reference. A module key is `mod:copper_fiber`.

The validator rejects a duplicate key. It does not accept `oneOf [string, id]`.

A rename or removal is a breaking change. What that costs, and how a key is deprecated first, is in [versioning.md](versioning.md).
