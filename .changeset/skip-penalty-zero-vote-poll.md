---
'@axelar-network/axelar-core': patch
---

Skip the missing-vote penalty in the EVM `HandleExpiredPoll` when a poll expires with zero votes, so a poll that no validator was able to vote on can no longer mark every maintainer missing and clear their rewards, removing the maintainer-deregistration pressure from that path.
