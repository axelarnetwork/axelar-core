---
'@axelar-network/axelar-core': patch
---

Flatten a RefundMsgRequest's inner message in the ante handler so it passes through the message ante decorators (RestrictedTx, CheckProxy, etc.), consistent with how authz MsgExec and auxiliary BatchRequest inner messages are handled.
