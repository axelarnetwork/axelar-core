---
'@axelar-network/axelar-core': minor
---

Add the x/authz module in the v1.5 upgrade, enabling scoped, revocable, expiring authorization grants (e.g. a validator delegating governance voting to an operational key). authz MsgExec is restricted to flat messages: it cannot wrap another MsgExec or a batch request, and a batch request cannot wrap a MsgExec.
