---
'@axelar-network/axelar-core': major
---

Migrate to cosmos-sdk v0.53, ibc-go v10, and wasmd v0.60 (v1.5 upgrade). Removes the x/crisis and x/capability modules and their stores, registers the 07-tendermint light client as a modular route, and runs the ibc core (6->8) and ibc transfer (5->6, DenomTrace->Denom) state migrations at the upgrade height. The ibc transfer REST/CLI query surface is renamed upstream (denom_traces -> denoms); x/tss remains for historical transaction decoding
