---
'@axelar-network/axelar-core': patch
---

Add nexus migration (8 to 9) that eagerly reallocates existing MaintainerState bitmaps down to the reduced max size, so stored states stop carrying oversized buffers in the KV store even for maintainers on deactivated chains that never vote again.
