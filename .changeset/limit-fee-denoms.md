---
'@axelar-network/axelar-core': minor
---

Limit which denominations can be used to pay transaction fees. A transaction may now pay fees in at most one denomination, and that denomination must be on a governance-controlled allowlist held by the new x/feepolicy module (default `["uaxl"]`). Enforced by an ante decorator in both CheckTx and block execution; fee-less transactions are unaffected.
