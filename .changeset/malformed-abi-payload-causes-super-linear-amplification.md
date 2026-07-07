---
'@axelar-network/axelar-core': patch
---

Guard against super-linear ABI amplification when routing a MsgRouteMessage with a malformed payload by bounding the decoded size before unpacking
