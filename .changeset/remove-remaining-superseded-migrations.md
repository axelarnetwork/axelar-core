---
'@axelar-network/axelar-core': patch
---

Remove the superseded multisig (2→3), permission (1→2), reward (1→2), tss (3→4), vote (2→3) and evm (10→11 link-deposit cleanup) in-place store migrations and their registrations. All networks store these modules at their current consensus versions, so the handlers can no longer be invoked. The evm bytecode migration stays registered wrapping a no-op. Consensus versions are unchanged.
