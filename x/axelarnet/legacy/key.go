package legacy

import "github.com/axelarnetwork/axelar-core/utils"

var (
	// ChainByAssetPrefix is legacy chain_by_asset prefix
	ChainByAssetPrefix = utils.KeyFromStr("chain_by_asset")
	// AssetByChainPrefix is legacy asset_by_chain prefix
	AssetByChainPrefix = utils.KeyFromStr("asset_by_chain")
	// PathPrefix is legacy path prefix
	PathPrefix = utils.KeyFromStr("path")
)
