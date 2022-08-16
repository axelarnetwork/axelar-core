package checks

import (
	// add packages for all modules here
	_ "github.com/axelarnetwork/axelar-core/x/ante/types"
	_ "github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	_ "github.com/axelarnetwork/axelar-core/x/evm/types"
	_ "github.com/axelarnetwork/axelar-core/x/multisig/types"
	_ "github.com/axelarnetwork/axelar-core/x/nexus/types"
	_ "github.com/axelarnetwork/axelar-core/x/permission/types"
	_ "github.com/axelarnetwork/axelar-core/x/reward/types"
	_ "github.com/axelarnetwork/axelar-core/x/snapshot/types"
	_ "github.com/axelarnetwork/axelar-core/x/tss/types"
	_ "github.com/axelarnetwork/axelar-core/x/vote/types"
)
