---
'@axelar-network/axelar-core': patch
---

Speed up SubmitPubKey by verifying the proof of ownership in the message handler (after the cheaper checks) and caching GetPermissionRole.
