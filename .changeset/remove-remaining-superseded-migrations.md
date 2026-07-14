---
'@axelar-network/axelar-core': patch
---

Remove the superseded multisig (2→3), permission (1→2), reward (1→2), tss (3→4) and vote (2→3) in-place store migrations and their registrations. All networks store these modules at their current consensus versions, so the handlers can no longer be invoked. Consensus versions are unchanged.
