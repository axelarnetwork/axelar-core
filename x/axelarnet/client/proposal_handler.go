package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/client/cli"
)

// ProposalHandler is the call contracts proposal handler.
var ProposalHandler = govclient.NewProposalHandler(cli.NewSubmitCallContractsProposalTxCmd)
