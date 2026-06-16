---
'@axelar-network/axelar-core': minor
---

Confirm transaction inclusion via CometBFT event subscriptions instead of indexer polling, so vald can broadcast against nodes with the transaction indexer disabled and no longer polls for every tx
