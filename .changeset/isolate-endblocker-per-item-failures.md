---
'@axelar-network/axelar-core': patch
---

Isolate per-item work in the vote, multisig, axelarnet, nexus, and reward EndBlockers so a single failing item is dead-lettered and logged instead of aborting FinalizeBlock and halting the chain.
