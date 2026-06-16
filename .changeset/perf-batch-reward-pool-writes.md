---
'@axelar-network/axelar-core': patch
---

Speed up block catch-up by persisting the reward pool once per batch instead of re-marshaling and writing it for every reward added
