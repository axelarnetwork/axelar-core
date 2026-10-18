---
'@axelar-network/axelar-core': patch
---

Warn at vald startup when the tofnd signing channel is configured with a non-loopback host: the connection is plaintext and unauthenticated, so exposing it beyond loopback lets anyone who can reach the port request signatures with the validator's keys
