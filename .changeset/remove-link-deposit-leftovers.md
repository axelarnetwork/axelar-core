---
'@axelar-network/axelar-core': patch
---

Remove the remaining link-deposit code from nexus and evm, and ignore the deprecated `linked_addresses` field in nexus genesis so it can no longer re-create state.
