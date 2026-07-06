---
'@axelar-network/axelar-core': patch
---

Add nexus migration (8 to 9) that shrinks existing MaintainerState bitmaps to the reduced max size, so stored states stop carrying oversized buffers in the KV store.
