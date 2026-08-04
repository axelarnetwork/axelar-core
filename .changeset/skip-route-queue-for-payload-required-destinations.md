---
'@axelar-network/axelar-core': patch
---

Stop queueing messages that the nexus EndBlocker can never route. The EndBlocker drains the route-message queue with an empty routing context, so a destination whose route needs the original payload (cosmos chains, via the axelarnet route) failed every time.
