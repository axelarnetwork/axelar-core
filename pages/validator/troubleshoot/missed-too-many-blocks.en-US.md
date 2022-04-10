# "Missed too many blocks" status

import Callout from 'nextra-theme-docs/callout'

If your validator misses 6 or more of the last 100 blocks then your Axelar status becomes `missed_too_many_blocks` and your [health check](../setup/health-check) prints something like:

```yaml
tofnd check: passed
broadcaster check: passed
operator check: failed (health check to operator MY_VALIDATOR_ADDRESS failed due to the following issues: {"missed_too_many_blocks":true})
```

You can restore your validator to healthy status simply by waiting --- `missed_too_many_blocks` is dropped automatically as soon as 100 blocks have passed in which you've missed 5 or fewer blocks.

<Callout emoji="💡">
  Tip: If you missed 50 or more of the last 100 blocks then your validator status becomes `jailed`. In this case, see [Unjail](../troubleshoot/unjail) for instructions on how to restore your validator to healthy status.
</Callout>