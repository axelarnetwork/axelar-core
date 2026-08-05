---
'@axelar-network/axelar-core': minor
---

Seed the x/feepolicy allowed fee denoms from the staking bond denom in the v1.5 upgrade handler, instead of relying on the module's default genesis, whose allowlist is hardcoded to `uaxl`.
