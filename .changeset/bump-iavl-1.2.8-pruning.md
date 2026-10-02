---
'@axelar-network/axelar-core': patch
---

Bump IAVL from v1.2.4 to v1.2.8 so the async pruner keeps going when a version is missing instead of stopping for good (iavl#1065), and to pick up the fixes for the latest-version lookup after a failed legacy prune (iavl#1067) and the fast-node cache/commit race (iavl#1142). Node-local; does not affect consensus.
