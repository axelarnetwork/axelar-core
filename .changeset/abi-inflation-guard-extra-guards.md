---
'@axelar-network/axelar-core': patch
---

Tighten the ABI inflation guard for CosmWasmV1 payloads: lower the inflation budget to 50kb, restrict wasm argument types to an allowlist, and charge for element types of empty dynamic arrays
