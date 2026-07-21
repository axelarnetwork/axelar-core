---
'@axelar-network/axelar-core': patch
---

Rebuild the nexus processing-message index from message status on genesis import (the index is not part of serialized genesis state), and reject genesis files with duplicate message IDs.
