---
'@axelar-network/axelar-core': patch
---

Disable `axelard export`: state export to genesis is not supported, since axelar-core upgrades via in-place store migrations rather than genesis export/import. The command now returns a clear error instead of silently producing non-round-trippable genesis (in-flight IBC correlation, the nexus processing-message index, the wasm activation flag, and in-flight vote tallies are not round-trippable).
