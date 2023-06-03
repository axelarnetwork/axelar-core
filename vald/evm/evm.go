package evm

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	geth "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	tmLog "github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/sdk-utils/broadcast"
	"github.com/axelarnetwork/axelar-core/utils/errors"
	"github.com/axelarnetwork/axelar-core/vald/evm/rpc"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	voteTypes "github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

// Smart contract event signatures
var (
	ERC20TransferSig                = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
	ERC20TokenDeploymentSig         = crypto.Keccak256Hash([]byte("TokenDeployed(string,address)"))
	MultisigTransferOperatorshipSig = crypto.Keccak256Hash([]byte("OperatorshipTransferred(bytes)"))
	ContractCallSig                 = crypto.Keccak256Hash([]byte("ContractCall(address,string,string,bytes32,bytes)"))
	ContractCallWithTokenSig        = crypto.Keccak256Hash([]byte("ContractCallWithToken(address,string,string,bytes32,bytes,string,uint256)"))
	TokenSentSig                    = crypto.Keccak256Hash([]byte("TokenSent(address,string,string,string,uint256)"))
)

// Mgr manages all communication with Ethereum
type Mgr struct {
	logger                    tmLog.Logger
	rpcs                      map[string]rpc.Client
	broadcaster               broadcast.Broadcaster
	validator                 sdk.ValAddress
	proxy                     sdk.AccAddress
	latestFinalizedBlockCache LatestFinalizedBlockCache
	axelarChainID             string
}

// NewMgr returns a new Mgr instance
func NewMgr(rpcs map[string]rpc.Client, broadcaster broadcast.Broadcaster, logger tmLog.Logger, valAddr sdk.ValAddress, proxy sdk.AccAddress, latestFinalizedBlockCache LatestFinalizedBlockCache, axelarChainID string) *Mgr {
	return &Mgr{
		rpcs:                      rpcs,
		proxy:                     proxy,
		broadcaster:               broadcaster,
		logger:                    logger.With("listener", "evm"),
		validator:                 valAddr,
		latestFinalizedBlockCache: latestFinalizedBlockCache,
		axelarChainID:             axelarChainID,
	}
}

// ProcessNewChain notifies the operator that vald needs to be restarted/udpated for a new chain
func (mgr Mgr) ProcessNewChain(event *types.ChainAdded) (err error) {
	mgr.logger.Info(fmt.Sprintf("VALD needs to be updated and restarted for new chain %s", event.Chain.String()))
	return nil
}

// ProcessDepositConfirmation votes on the correctness of an EVM chain token deposit
func (mgr Mgr) ProcessDepositConfirmation(event *types.ConfirmDepositStarted) error {
	if !slices.Any(event.Participants, func(v sdk.ValAddress) bool { return v.Equals(mgr.validator) }) {
		mgr.logger.Debug("ignoring deposit confirmation poll: not a participant", "pollID", event.PollID)
		return nil
	}

	txReceipt, err := mgr.GetTxReceiptIfFinalized(event.Chain, common.Hash(event.TxID), event.ConfirmationHeight)
	if err != nil {
		return err
	}
	if txReceipt == nil {
		mgr.logger.Info(fmt.Sprintf("broadcasting empty vote for poll %s", event.PollID.String()))
		_, err := mgr.broadcaster.Broadcast(context.TODO(), voteTypes.NewVoteRequest(mgr.proxy, event.PollID, types.NewVoteEvents(event.Chain)))

		return err
	}

	var events []types.Event
	for i, log := range txReceipt.Logs {
		if log.Topics[0] != ERC20TransferSig {
			continue
		}

		if !bytes.Equal(event.TokenAddress.Bytes(), log.Address.Bytes()) {
			continue
		}

		erc20Event, err := DecodeERC20TransferEvent(log)
		if err != nil {
			mgr.logger.Debug(sdkerrors.Wrap(err, "decode event Transfer failed").Error())
			continue
		}

		if erc20Event.To != event.DepositAddress {
			continue
		}

		if err := erc20Event.ValidateBasic(); err != nil {
			mgr.logger.Debug(sdkerrors.Wrap(err, "invalid event Transfer").Error())
			continue
		}

		events = append(events, types.Event{
			Chain: event.Chain,
			TxID:  event.TxID,
			Index: uint64(i),
			Event: &types.Event_Transfer{
				Transfer: &erc20Event,
			},
		})
	}

	mgr.logger.Info(fmt.Sprintf("broadcasting vote %v for poll %s", events, event.PollID.String()))
	_, err = mgr.broadcaster.Broadcast(context.TODO(), voteTypes.NewVoteRequest(mgr.proxy, event.PollID, types.NewVoteEvents(event.Chain, events...)))

	return err
}

// ProcessTokenConfirmation votes on the correctness of an EVM chain token deployment
func (mgr Mgr) ProcessTokenConfirmation(event *types.ConfirmTokenStarted) error {
	if !slices.Any(event.Participants, func(v sdk.ValAddress) bool { return v.Equals(mgr.validator) }) {
		mgr.logger.Debug("ignoring token confirmation poll: not a participant", "pollID", event.PollID)
		return nil
	}

	txReceipt, err := mgr.GetTxReceiptIfFinalized(event.Chain, common.Hash(event.TxID), event.ConfirmationHeight)
	if err != nil {
		return err
	}
	if txReceipt == nil {
		mgr.logger.Info(fmt.Sprintf("broadcasting empty vote for poll %s", event.PollID.String()))
		_, err := mgr.broadcaster.Broadcast(context.TODO(), voteTypes.NewVoteRequest(mgr.proxy, event.PollID, types.NewVoteEvents(event.Chain)))

		return err
	}

	var events []types.Event
	for i, log := range txReceipt.Logs {
		if log.Topics[0] != ERC20TokenDeploymentSig {
			continue
		}

		if !bytes.Equal(event.GatewayAddress.Bytes(), log.Address.Bytes()) {
			continue
		}

		erc20Event, err := DecodeERC20TokenDeploymentEvent(log)
		if err != nil {
			mgr.logger.Debug(sdkerrors.Wrap(err, "decode event TokenDeployed failed").Error())
			continue
		}

		if erc20Event.TokenAddress != event.TokenAddress || erc20Event.Symbol != event.TokenDetails.Symbol {
			continue
		}

		if err := erc20Event.ValidateBasic(); err != nil {
			mgr.logger.Debug(sdkerrors.Wrap(err, "invalid event ERC20TokenDeployment").Error())
			continue
		}

		events = append(events, types.Event{
			Chain: event.Chain,
			TxID:  event.TxID,
			Index: uint64(i),
			Event: &types.Event_TokenDeployed{
				TokenDeployed: &erc20Event,
			},
		})
		break
	}

	mgr.logger.Info(fmt.Sprintf("broadcasting vote %v for poll %s", events, event.PollID.String()))
	_, err = mgr.broadcaster.Broadcast(context.TODO(), voteTypes.NewVoteRequest(mgr.proxy, event.PollID, types.NewVoteEvents(event.Chain, events...)))

	return err
}

// This is a temporary workaround to allow a Filecoin key transfer tx to be processed.
// This workaround allows validators to avoid having to sync an archival node for Filecoin to process this tx
// The tx receipt is hardcoded below and can be verified against your own archival node or the explorer
// https://filfox.info/en/tx/0x44ca58be45e862026202bc7a3c4dc897804264d29f087b9a1249f9f4bf1e31d7
var (
	// FilecoinTransferKeyTxID is the tx hash of the Filecoin key transfer
	FilecoinTransferKeyTxID = types.Hash(common.BytesToHash(funcs.Must(
		hexutil.Decode("0x44ca58be45e862026202bc7a3c4dc897804264d29f087b9a1249f9f4bf1e31d7"),
	)))
)

// GetFilecoinTransferKeyTxReceipt returns the hardcoded tx receipt for the Filecoin key transfer
func GetFilecoinTransferKeyTxReceipt() *geth.Receipt {
	return &geth.Receipt{
		Logs: []*geth.Log{
			{
				Address: common.BytesToAddress(funcs.Must(hexutil.Decode("0x1a920B29eBD437074225cAeE44f78FC700B27a5d"))),
				Topics: []common.Hash{
					crypto.Keccak256Hash([]byte("OperatorshipTransferred(address[],uint256[],uint256)")),
				},
				Data: funcs.Must(hexutil.Decode("0x00000000000000000000000000000000000000000000000000000000000000600000000000000000000000000000000000000000000000000000000000000940000000000000000000000000000000000000000000000000000000000001f88c0000000000000000000000000000000000000000000000000000000000000046000000000000000000000000005b9b7dbb5e1a804e71e8e5c2c1dc9fcede357500000000000000000000000002932f1646cc648a85924224515a79ff0aa6faef000000000000000000000000048aa902b03034e33d7c2f937d7b86759df418220000000000000000000000000999d3bac6083beaa5d0827605ce45ccf7e907870000000000000000000000000a9f3166966f717ec2bffe4cdc6c6951a993bd0e00000000000000000000000010abe6b12dae0776abcbc22d0aa8388b9116d8b400000000000000000000000012cd106fbe74746e056aa904f78884822b70140e00000000000000000000000018b191e2fab94aaa230a15eb284b205da684432c0000000000000000000000001f5e43c87f14a3f426d1854cfa5804057c8c22cd0000000000000000000000001ffd39e4625d0891a802a728e43ca3fe63abecd00000000000000000000000002129e0b16f89e72bf2243f758f55814854e7cebf00000000000000000000000022ca820e861b88075a3d8b8fd63953d983d4965d00000000000000000000000026d7a0fb41cb4b62702ce5e4db5b664c7249abda00000000000000000000000026ed865c17b688bd1a49edec7c7ee8e11656087a00000000000000000000000029bbdeff9bf1b480013e47582397b621de59f36c0000000000000000000000002c18de214d948e61d8dcf072e98c7ef0d540c14000000000000000000000000030d1f743437d682f08bda172b1580e9dd0190a1100000000000000000000000032486eff582fcdfd7357fb000c4ea9502c889f3100000000000000000000000034b4ad263716685103080123f9b429a9394ed6550000000000000000000000003ab066f3bb35a9db96ba2c12cb8b788dc38b0d4e0000000000000000000000004671e7cedd74ab77cd217f5c2a4d8c7e14be619e00000000000000000000000046d843f398d680b8087fc5b5435877924cdeceff0000000000000000000000004b4bcd0251a6a0a1bd8a4809f6f40037b8f492a500000000000000000000000053ed56a92e4caa5485c87dab1372f27242e1ea5f000000000000000000000000542b5328c8207bef168367123eb5b30fd1d62bcc00000000000000000000000054c2144fe51d5e9d18da0ed27f4de9d3fea8a39c00000000000000000000000057deaa35a513e30888772e2a39c03177c3bbb8ed000000000000000000000000591530866608961c587d7e6d0e1120032fe54bc00000000000000000000000005d82d074a0f79d2b7950e27962d33854c6d2db4a0000000000000000000000005eb054b5dd4e4bfd24eca253cf248ae60d31eb720000000000000000000000006075ca701e28b452a70ab4822ffb7bc78949e7b6000000000000000000000000697fea4aabdf019a94dcd283bd6b24930b5e893c0000000000000000000000006a8ab627fffd53979ab43845e70d89530ae413360000000000000000000000006aa18708efe9d692ae2260c1ef09c07e188bf2be000000000000000000000000707ffd30f1002b34698c0e97f883f773bcdf3e4b0000000000000000000000007db6268e8036977408252b95cb816ae72312fced0000000000000000000000007e4df9aae47bc229fa8379ecc2d3955fd50403f10000000000000000000000007e6be8cf7d3da441b25ee6dc55c242c5114f1fdb0000000000000000000000007fc92bcb8f6d62b74ca0a679da8b6a08aa5a35a70000000000000000000000008052130c949208b4944a5ec86aa3eef874ff340d00000000000000000000000082735375563eb6d606ec54c747b2b6d4ee0cb77f00000000000000000000000085f447d76710df97e40163f5d555463562dbd52d0000000000000000000000008627ba680fc3f877080036aa4b16026e9f008df90000000000000000000000008635e3b81013ed6e75a60373f71f107c748434bb0000000000000000000000008b9cb38e0f8bb288a7f3035eb3c43bd8a82fb33400000000000000000000000092446060c330564cd0d4b54bf41db6fec992c91100000000000000000000000093efdb3bb91d0f56d81763cd27e0b4feda3d5265000000000000000000000000961df4753b167e6d2d9402ce2656e5fc2318206c0000000000000000000000009cedcc28f2b44b32f07ebad02993f528f2a22eda000000000000000000000000a404c72151139e9ec7b1e5fd7a32d2a0a47ff594000000000000000000000000a6728d808425845c265358cd0a1acfd400cf295e000000000000000000000000aa5d27cbeee6302b4f2e121d58b96977969891ed000000000000000000000000ad67d60e92c090725244fae9338cde9b685b3890000000000000000000000000af9d39013cf82194596ef6ba0e1359fe7a2a83d5000000000000000000000000b213c42985595dc5c71bb133858b1e769c4d0ac3000000000000000000000000b3e0f556209e6413e9946799812b8e112ebc7312000000000000000000000000bd57952b2357a3bb0566b7f6e98018d9c8a674e7000000000000000000000000c14ab693b122bcee9c7a51c147fd6c02a6defc8a000000000000000000000000c64e38381e6c78ea5a80f01399a908fe248a1539000000000000000000000000c6ee966ccbb94751525dc4ac592a754723e3c32b000000000000000000000000d0d219137e23c13b6dffdf128bc7cdfbe364511f000000000000000000000000d58732034c4a1c114aaa83fe3cb444cea695e126000000000000000000000000d685c77717ea6291fe88abb5d9073a8e11c37a6b000000000000000000000000d9284d8656da02bc87db515568366aa813d922f7000000000000000000000000dc10d87afe6a097bc2ddb365593c167ea57cf28f000000000000000000000000dd48e6e9feff9275289f48259896841ccba11bc3000000000000000000000000e294053589d99ca5fd956836f4e4d8294dce4b5c000000000000000000000000ee53ab644757afd82ef9418ffcb203635b59ae29000000000000000000000000f1cabb77a5bd1629d0e88a1d0bd4354a61edfe9b000000000000000000000000fb62dd6470d083fc3ca2df5e4f28b43915179b4f00000000000000000000000000000000000000000000000000000000000000460000000000000000000000000000000000000000000000000000000000000bbe0000000000000000000000000000000000000000000000000000000000000756000000000000000000000000000000000000000000000000000000000000094c00000000000000000000000000000000000000000000000000000000000006eb000000000000000000000000000000000000000000000000000000000000101f000000000000000000000000000000000000000000000000000000000000118d0000000000000000000000000000000000000000000000000000000000000d620000000000000000000000000000000000000000000000000000000000000ac70000000000000000000000000000000000000000000000000000000000000fa6000000000000000000000000000000000000000000000000000000000000069b0000000000000000000000000000000000000000000000000000000000000a8500000000000000000000000000000000000000000000000000000000000007690000000000000000000000000000000000000000000000000000000000000e370000000000000000000000000000000000000000000000000000000000000dd300000000000000000000000000000000000000000000000000000000000007540000000000000000000000000000000000000000000000000000000000000f380000000000000000000000000000000000000000000000000000000000000c5a0000000000000000000000000000000000000000000000000000000000001a660000000000000000000000000000000000000000000000000000000000000b54000000000000000000000000000000000000000000000000000000000000091f0000000000000000000000000000000000000000000000000000000000000e24000000000000000000000000000000000000000000000000000000000000060800000000000000000000000000000000000000000000000000000000000007190000000000000000000000000000000000000000000000000000000000000e620000000000000000000000000000000000000000000000000000000000000e6a000000000000000000000000000000000000000000000000000000000000081200000000000000000000000000000000000000000000000000000000000008f900000000000000000000000000000000000000000000000000000000000010d80000000000000000000000000000000000000000000000000000000000000c860000000000000000000000000000000000000000000000000000000000000cf30000000000000000000000000000000000000000000000000000000000000e600000000000000000000000000000000000000000000000000000000000000d5900000000000000000000000000000000000000000000000000000000000016c40000000000000000000000000000000000000000000000000000000000000cd90000000000000000000000000000000000000000000000000000000000000eae000000000000000000000000000000000000000000000000000000000000084900000000000000000000000000000000000000000000000000000000000006420000000000000000000000000000000000000000000000000000000000000d3a00000000000000000000000000000000000000000000000000000000000007e600000000000000000000000000000000000000000000000000000000000011960000000000000000000000000000000000000000000000000000000000000a3b0000000000000000000000000000000000000000000000000000000000000c9800000000000000000000000000000000000000000000000000000000000005ae0000000000000000000000000000000000000000000000000000000000000ea70000000000000000000000000000000000000000000000000000000000000bd600000000000000000000000000000000000000000000000000000000000007d200000000000000000000000000000000000000000000000000000000000006cb0000000000000000000000000000000000000000000000000000000000000ff60000000000000000000000000000000000000000000000000000000000000f070000000000000000000000000000000000000000000000000000000000000bea000000000000000000000000000000000000000000000000000000000000075a0000000000000000000000000000000000000000000000000000000000000ccc000000000000000000000000000000000000000000000000000000000000086200000000000000000000000000000000000000000000000000000000000007540000000000000000000000000000000000000000000000000000000000000f530000000000000000000000000000000000000000000000000000000000000ed6000000000000000000000000000000000000000000000000000000000000149700000000000000000000000000000000000000000000000000000000000013c500000000000000000000000000000000000000000000000000000000000011e20000000000000000000000000000000000000000000000000000000000000cf60000000000000000000000000000000000000000000000000000000000000b3e00000000000000000000000000000000000000000000000000000000000006100000000000000000000000000000000000000000000000000000000000000b61000000000000000000000000000000000000000000000000000000000000074f000000000000000000000000000000000000000000000000000000000000079c00000000000000000000000000000000000000000000000000000000000006150000000000000000000000000000000000000000000000000000000000000a0400000000000000000000000000000000000000000000000000000000000010980000000000000000000000000000000000000000000000000000000000000dd40000000000000000000000000000000000000000000000000000000000000b45")),
			},
			{
				Address: common.BytesToAddress(funcs.Must(hexutil.Decode("0xe432150cce91c13a887f7D836923d5597adD8E31"))),
				Topics: []common.Hash{
					crypto.Keccak256Hash([]byte("OperatorshipTransferred(bytes)")),
				},
				Data: funcs.Must(hexutil.Decode("0x0000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000122000000000000000000000000000000000000000000000000000000000000000600000000000000000000000000000000000000000000000000000000000000940000000000000000000000000000000000000000000000000000000000001f88c0000000000000000000000000000000000000000000000000000000000000046000000000000000000000000005b9b7dbb5e1a804e71e8e5c2c1dc9fcede357500000000000000000000000002932f1646cc648a85924224515a79ff0aa6faef000000000000000000000000048aa902b03034e33d7c2f937d7b86759df418220000000000000000000000000999d3bac6083beaa5d0827605ce45ccf7e907870000000000000000000000000a9f3166966f717ec2bffe4cdc6c6951a993bd0e00000000000000000000000010abe6b12dae0776abcbc22d0aa8388b9116d8b400000000000000000000000012cd106fbe74746e056aa904f78884822b70140e00000000000000000000000018b191e2fab94aaa230a15eb284b205da684432c0000000000000000000000001f5e43c87f14a3f426d1854cfa5804057c8c22cd0000000000000000000000001ffd39e4625d0891a802a728e43ca3fe63abecd00000000000000000000000002129e0b16f89e72bf2243f758f55814854e7cebf00000000000000000000000022ca820e861b88075a3d8b8fd63953d983d4965d00000000000000000000000026d7a0fb41cb4b62702ce5e4db5b664c7249abda00000000000000000000000026ed865c17b688bd1a49edec7c7ee8e11656087a00000000000000000000000029bbdeff9bf1b480013e47582397b621de59f36c0000000000000000000000002c18de214d948e61d8dcf072e98c7ef0d540c14000000000000000000000000030d1f743437d682f08bda172b1580e9dd0190a1100000000000000000000000032486eff582fcdfd7357fb000c4ea9502c889f3100000000000000000000000034b4ad263716685103080123f9b429a9394ed6550000000000000000000000003ab066f3bb35a9db96ba2c12cb8b788dc38b0d4e0000000000000000000000004671e7cedd74ab77cd217f5c2a4d8c7e14be619e00000000000000000000000046d843f398d680b8087fc5b5435877924cdeceff0000000000000000000000004b4bcd0251a6a0a1bd8a4809f6f40037b8f492a500000000000000000000000053ed56a92e4caa5485c87dab1372f27242e1ea5f000000000000000000000000542b5328c8207bef168367123eb5b30fd1d62bcc00000000000000000000000054c2144fe51d5e9d18da0ed27f4de9d3fea8a39c00000000000000000000000057deaa35a513e30888772e2a39c03177c3bbb8ed000000000000000000000000591530866608961c587d7e6d0e1120032fe54bc00000000000000000000000005d82d074a0f79d2b7950e27962d33854c6d2db4a0000000000000000000000005eb054b5dd4e4bfd24eca253cf248ae60d31eb720000000000000000000000006075ca701e28b452a70ab4822ffb7bc78949e7b6000000000000000000000000697fea4aabdf019a94dcd283bd6b24930b5e893c0000000000000000000000006a8ab627fffd53979ab43845e70d89530ae413360000000000000000000000006aa18708efe9d692ae2260c1ef09c07e188bf2be000000000000000000000000707ffd30f1002b34698c0e97f883f773bcdf3e4b0000000000000000000000007db6268e8036977408252b95cb816ae72312fced0000000000000000000000007e4df9aae47bc229fa8379ecc2d3955fd50403f10000000000000000000000007e6be8cf7d3da441b25ee6dc55c242c5114f1fdb0000000000000000000000007fc92bcb8f6d62b74ca0a679da8b6a08aa5a35a70000000000000000000000008052130c949208b4944a5ec86aa3eef874ff340d00000000000000000000000082735375563eb6d606ec54c747b2b6d4ee0cb77f00000000000000000000000085f447d76710df97e40163f5d555463562dbd52d0000000000000000000000008627ba680fc3f877080036aa4b16026e9f008df90000000000000000000000008635e3b81013ed6e75a60373f71f107c748434bb0000000000000000000000008b9cb38e0f8bb288a7f3035eb3c43bd8a82fb33400000000000000000000000092446060c330564cd0d4b54bf41db6fec992c91100000000000000000000000093efdb3bb91d0f56d81763cd27e0b4feda3d5265000000000000000000000000961df4753b167e6d2d9402ce2656e5fc2318206c0000000000000000000000009cedcc28f2b44b32f07ebad02993f528f2a22eda000000000000000000000000a404c72151139e9ec7b1e5fd7a32d2a0a47ff594000000000000000000000000a6728d808425845c265358cd0a1acfd400cf295e000000000000000000000000aa5d27cbeee6302b4f2e121d58b96977969891ed000000000000000000000000ad67d60e92c090725244fae9338cde9b685b3890000000000000000000000000af9d39013cf82194596ef6ba0e1359fe7a2a83d5000000000000000000000000b213c42985595dc5c71bb133858b1e769c4d0ac3000000000000000000000000b3e0f556209e6413e9946799812b8e112ebc7312000000000000000000000000bd57952b2357a3bb0566b7f6e98018d9c8a674e7000000000000000000000000c14ab693b122bcee9c7a51c147fd6c02a6defc8a000000000000000000000000c64e38381e6c78ea5a80f01399a908fe248a1539000000000000000000000000c6ee966ccbb94751525dc4ac592a754723e3c32b000000000000000000000000d0d219137e23c13b6dffdf128bc7cdfbe364511f000000000000000000000000d58732034c4a1c114aaa83fe3cb444cea695e126000000000000000000000000d685c77717ea6291fe88abb5d9073a8e11c37a6b000000000000000000000000d9284d8656da02bc87db515568366aa813d922f7000000000000000000000000dc10d87afe6a097bc2ddb365593c167ea57cf28f000000000000000000000000dd48e6e9feff9275289f48259896841ccba11bc3000000000000000000000000e294053589d99ca5fd956836f4e4d8294dce4b5c000000000000000000000000ee53ab644757afd82ef9418ffcb203635b59ae29000000000000000000000000f1cabb77a5bd1629d0e88a1d0bd4354a61edfe9b000000000000000000000000fb62dd6470d083fc3ca2df5e4f28b43915179b4f00000000000000000000000000000000000000000000000000000000000000460000000000000000000000000000000000000000000000000000000000000bbe0000000000000000000000000000000000000000000000000000000000000756000000000000000000000000000000000000000000000000000000000000094c00000000000000000000000000000000000000000000000000000000000006eb000000000000000000000000000000000000000000000000000000000000101f000000000000000000000000000000000000000000000000000000000000118d0000000000000000000000000000000000000000000000000000000000000d620000000000000000000000000000000000000000000000000000000000000ac70000000000000000000000000000000000000000000000000000000000000fa6000000000000000000000000000000000000000000000000000000000000069b0000000000000000000000000000000000000000000000000000000000000a8500000000000000000000000000000000000000000000000000000000000007690000000000000000000000000000000000000000000000000000000000000e370000000000000000000000000000000000000000000000000000000000000dd300000000000000000000000000000000000000000000000000000000000007540000000000000000000000000000000000000000000000000000000000000f380000000000000000000000000000000000000000000000000000000000000c5a0000000000000000000000000000000000000000000000000000000000001a660000000000000000000000000000000000000000000000000000000000000b54000000000000000000000000000000000000000000000000000000000000091f0000000000000000000000000000000000000000000000000000000000000e24000000000000000000000000000000000000000000000000000000000000060800000000000000000000000000000000000000000000000000000000000007190000000000000000000000000000000000000000000000000000000000000e620000000000000000000000000000000000000000000000000000000000000e6a000000000000000000000000000000000000000000000000000000000000081200000000000000000000000000000000000000000000000000000000000008f900000000000000000000000000000000000000000000000000000000000010d80000000000000000000000000000000000000000000000000000000000000c860000000000000000000000000000000000000000000000000000000000000cf30000000000000000000000000000000000000000000000000000000000000e600000000000000000000000000000000000000000000000000000000000000d5900000000000000000000000000000000000000000000000000000000000016c40000000000000000000000000000000000000000000000000000000000000cd90000000000000000000000000000000000000000000000000000000000000eae000000000000000000000000000000000000000000000000000000000000084900000000000000000000000000000000000000000000000000000000000006420000000000000000000000000000000000000000000000000000000000000d3a00000000000000000000000000000000000000000000000000000000000007e600000000000000000000000000000000000000000000000000000000000011960000000000000000000000000000000000000000000000000000000000000a3b0000000000000000000000000000000000000000000000000000000000000c9800000000000000000000000000000000000000000000000000000000000005ae0000000000000000000000000000000000000000000000000000000000000ea70000000000000000000000000000000000000000000000000000000000000bd600000000000000000000000000000000000000000000000000000000000007d200000000000000000000000000000000000000000000000000000000000006cb0000000000000000000000000000000000000000000000000000000000000ff60000000000000000000000000000000000000000000000000000000000000f070000000000000000000000000000000000000000000000000000000000000bea000000000000000000000000000000000000000000000000000000000000075a0000000000000000000000000000000000000000000000000000000000000ccc000000000000000000000000000000000000000000000000000000000000086200000000000000000000000000000000000000000000000000000000000007540000000000000000000000000000000000000000000000000000000000000f530000000000000000000000000000000000000000000000000000000000000ed6000000000000000000000000000000000000000000000000000000000000149700000000000000000000000000000000000000000000000000000000000013c500000000000000000000000000000000000000000000000000000000000011e20000000000000000000000000000000000000000000000000000000000000cf60000000000000000000000000000000000000000000000000000000000000b3e00000000000000000000000000000000000000000000000000000000000006100000000000000000000000000000000000000000000000000000000000000b61000000000000000000000000000000000000000000000000000000000000074f000000000000000000000000000000000000000000000000000000000000079c00000000000000000000000000000000000000000000000000000000000006150000000000000000000000000000000000000000000000000000000000000a0400000000000000000000000000000000000000000000000000000000000010980000000000000000000000000000000000000000000000000000000000000dd40000000000000000000000000000000000000000000000000000000000000b45")),
			},
			{
				Address: common.BytesToAddress(funcs.Must(hexutil.Decode("0xe432150cce91c13a887f7D836923d5597adD8E31"))),
				Topics: []common.Hash{
					crypto.Keccak256Hash([]byte("Executed(bytes32)")),
					common.BytesToHash(funcs.Must(hexutil.Decode("0x5e9db070c46eaffae97c72d439404a8137df389dd075492f3b65021908dc1ab9"))),
				},
				Data: []byte{},
			},
		},
	}
}

// ProcessTransferKeyConfirmation votes on the correctness of an EVM chain key transfer
func (mgr Mgr) ProcessTransferKeyConfirmation(event *types.ConfirmKeyTransferStarted) error {
	if !slices.Any(event.Participants, func(v sdk.ValAddress) bool { return v.Equals(mgr.validator) }) {
		mgr.logger.Debug("ignoring key transfer confirmation poll: not a participant", "pollID", event.PollID)
		return nil
	}

	var txReceipt *geth.Receipt
	var err error

	if event.Chain.Equals(nexus.ChainName("filecoin")) && event.TxID == FilecoinTransferKeyTxID && mgr.axelarChainID == "axelar-dojo-1" {
		txReceipt = GetFilecoinTransferKeyTxReceipt()
	} else {
		txReceipt, err = mgr.GetTxReceiptIfFinalized(event.Chain, common.Hash(event.TxID), event.ConfirmationHeight)
		if err != nil {
			return err
		}
	}

	if txReceipt == nil {
		mgr.logger.Info(fmt.Sprintf("broadcasting empty vote for poll %s", event.PollID.String()))
		_, err := mgr.broadcaster.Broadcast(context.TODO(), voteTypes.NewVoteRequest(mgr.proxy, event.PollID, types.NewVoteEvents(event.Chain)))

		return err
	}

	var events []types.Event
	for i := len(txReceipt.Logs) - 1; i >= 0; i-- {
		log := txReceipt.Logs[i]

		if log.Topics[0] != MultisigTransferOperatorshipSig {
			continue
		}

		// Event is not emitted by the axelar gateway
		if log.Address != common.Address(event.GatewayAddress) {
			continue
		}

		transferOperatorshipEvent, err := DecodeMultisigOperatorshipTransferredEvent(log)
		if err != nil {
			mgr.logger.Debug(sdkerrors.Wrap(err, "failed decoding operatorship transferred event").Error())
			continue
		}

		if err := transferOperatorshipEvent.ValidateBasic(); err != nil {
			mgr.logger.Debug(sdkerrors.Wrap(err, "invalid event MultisigTransferOperatorship").Error())
			continue
		}

		events = append(events, types.Event{
			Chain: event.Chain,
			TxID:  event.TxID,
			Index: uint64(i),
			Event: &types.Event_MultisigOperatorshipTransferred{
				MultisigOperatorshipTransferred: &transferOperatorshipEvent,
			}})
		break
	}

	mgr.logger.Info(fmt.Sprintf("broadcasting vote %v for poll %s", events, event.PollID.String()))
	_, err = mgr.broadcaster.Broadcast(context.TODO(), voteTypes.NewVoteRequest(mgr.proxy, event.PollID, types.NewVoteEvents(event.Chain, events...)))

	return err
}

// ProcessGatewayTxConfirmation votes on the correctness of an EVM chain gateway's transactions
func (mgr Mgr) ProcessGatewayTxConfirmation(event *types.ConfirmGatewayTxStarted) error {
	if !slices.Any(event.Participants, func(v sdk.ValAddress) bool { return v.Equals(mgr.validator) }) {
		mgr.logger.Debug("ignoring gateway tx confirmation poll: not a participant", "pollID", event.PollID)
		return nil
	}

	txReceipt, err := mgr.GetTxReceiptIfFinalized(event.Chain, common.Hash(event.TxID), event.ConfirmationHeight)
	if err != nil {
		return err
	}
	if txReceipt == nil {
		mgr.logger.Info(fmt.Sprintf("broadcasting empty vote for poll %s", event.PollID.String()))
		_, err := mgr.broadcaster.Broadcast(context.TODO(), voteTypes.NewVoteRequest(mgr.proxy, event.PollID, types.NewVoteEvents(event.Chain)))

		return err
	}

	var events []types.Event
	for i, log := range txReceipt.Logs {
		if !bytes.Equal(event.GatewayAddress.Bytes(), log.Address.Bytes()) {
			continue
		}

		switch log.Topics[0] {
		case ContractCallSig:
			gatewayEvent, err := DecodeEventContractCall(log)
			if err != nil {
				mgr.logger.Debug(sdkerrors.Wrap(err, "decode event ContractCall failed").Error())
				continue
			}

			if err := gatewayEvent.ValidateBasic(); err != nil {
				mgr.logger.Debug(sdkerrors.Wrap(err, "invalid event ContractCall").Error())
				continue
			}

			events = append(events, types.Event{
				Chain: event.Chain,
				TxID:  event.TxID,
				Index: uint64(i),
				Event: &types.Event_ContractCall{
					ContractCall: &gatewayEvent,
				},
			})
		case ContractCallWithTokenSig:
			gatewayEvent, err := DecodeEventContractCallWithToken(log)
			if err != nil {
				mgr.logger.Debug(sdkerrors.Wrap(err, "decode event ContractCallWithToken failed").Error())
				continue
			}

			if err := gatewayEvent.ValidateBasic(); err != nil {
				mgr.logger.Debug(sdkerrors.Wrap(err, "invalid event ContractCallWithToken").Error())
				continue
			}

			events = append(events, types.Event{
				Chain: event.Chain,
				TxID:  event.TxID,
				Index: uint64(i),
				Event: &types.Event_ContractCallWithToken{
					ContractCallWithToken: &gatewayEvent,
				},
			})
		case TokenSentSig:
			gatewayEvent, err := DecodeEventTokenSent(log)
			if err != nil {
				mgr.logger.Debug(sdkerrors.Wrap(err, "decode event TokenSent failed").Error())
			}

			if err := gatewayEvent.ValidateBasic(); err != nil {
				mgr.logger.Debug(sdkerrors.Wrap(err, "invalid event TokenSent").Error())
				continue
			}

			events = append(events, types.Event{
				Chain: event.Chain,
				TxID:  event.TxID,
				Index: uint64(i),
				Event: &types.Event_TokenSent{
					TokenSent: &gatewayEvent,
				},
			})
		default:
		}
	}

	mgr.logger.Info(fmt.Sprintf("broadcasting vote %v for poll %s", events, event.PollID.String()))
	_, err = mgr.broadcaster.Broadcast(context.TODO(), voteTypes.NewVoteRequest(mgr.proxy, event.PollID, types.NewVoteEvents(event.Chain, events...)))

	return err
}

func DecodeEventTokenSent(log *geth.Log) (types.EventTokenSent, error) {
	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return types.EventTokenSent{}, err
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return types.EventTokenSent{}, err
	}

	arguments := abi.Arguments{
		{Type: stringType},
		{Type: stringType},
		{Type: stringType},
		{Type: uint256Type},
	}
	params, err := types.StrictDecode(arguments, log.Data)
	if err != nil {
		return types.EventTokenSent{}, err
	}

	return types.EventTokenSent{
		Sender:             types.Address(common.BytesToAddress(log.Topics[1].Bytes())),
		DestinationChain:   nexus.ChainName(params[0].(string)),
		DestinationAddress: params[1].(string),
		Symbol:             params[2].(string),
		Amount:             sdk.NewUintFromBigInt(params[3].(*big.Int)),
	}, nil
}

func DecodeEventContractCall(log *geth.Log) (types.EventContractCall, error) {
	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return types.EventContractCall{}, err
	}

	bytesType, err := abi.NewType("bytes", "bytes", nil)
	if err != nil {
		return types.EventContractCall{}, err
	}

	arguments := abi.Arguments{
		{Type: stringType},
		{Type: stringType},
		{Type: bytesType},
	}
	params, err := types.StrictDecode(arguments, log.Data)
	if err != nil {
		return types.EventContractCall{}, err
	}

	return types.EventContractCall{
		Sender:           types.Address(common.BytesToAddress(log.Topics[1].Bytes())),
		DestinationChain: nexus.ChainName(params[0].(string)),
		ContractAddress:  params[1].(string),
		PayloadHash:      types.Hash(common.BytesToHash(log.Topics[2].Bytes())),
	}, nil
}

func DecodeEventContractCallWithToken(log *geth.Log) (types.EventContractCallWithToken, error) {
	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return types.EventContractCallWithToken{}, err
	}

	bytesType, err := abi.NewType("bytes", "bytes", nil)
	if err != nil {
		return types.EventContractCallWithToken{}, err
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return types.EventContractCallWithToken{}, err
	}

	arguments := abi.Arguments{
		{Type: stringType},
		{Type: stringType},
		{Type: bytesType},
		{Type: stringType},
		{Type: uint256Type},
	}
	params, err := types.StrictDecode(arguments, log.Data)
	if err != nil {
		return types.EventContractCallWithToken{}, err
	}

	return types.EventContractCallWithToken{
		Sender:           types.Address(common.BytesToAddress(log.Topics[1].Bytes())),
		DestinationChain: nexus.ChainName(params[0].(string)),
		ContractAddress:  params[1].(string),
		PayloadHash:      types.Hash(common.BytesToHash(log.Topics[2].Bytes())),
		Symbol:           params[3].(string),
		Amount:           sdk.NewUintFromBigInt(params[4].(*big.Int)),
	}, nil
}

func (mgr Mgr) isTxReceiptFinalized(chain nexus.ChainName, txReceipt *geth.Receipt, confHeight uint64) (bool, error) {
	client, ok := mgr.rpcs[strings.ToLower(chain.String())]
	if !ok {
		return false, fmt.Errorf("rpc client not found for chain %s", chain.String())
	}

	if mgr.latestFinalizedBlockCache.Get(chain).Cmp(txReceipt.BlockNumber) >= 0 {
		return true, nil
	}

	latestFinalizedBlockNumber, err := client.LatestFinalizedBlockNumber(context.Background(), confHeight)
	if err != nil || latestFinalizedBlockNumber.Cmp(txReceipt.BlockNumber) < 0 {
		return false, err
	}

	mgr.latestFinalizedBlockCache.Set(chain, latestFinalizedBlockNumber)

	return true, nil
}

func (mgr Mgr) GetTxReceiptIfFinalized(chain nexus.ChainName, txID common.Hash, confHeight uint64) (*geth.Receipt, error) {
	client, ok := mgr.rpcs[strings.ToLower(chain.String())]
	if !ok {
		return nil, fmt.Errorf("rpc client not found for chain %s", chain.String())
	}

	txReceipt, err := client.TransactionReceipt(context.Background(), txID)
	keyvals := []interface{}{"chain", chain.String(), "tx_id", txID.Hex()}
	logger := mgr.logger.With(keyvals...)
	if err == ethereum.NotFound {
		logger.Debug(fmt.Sprintf("transaction receipt %s not found", txID.Hex()))
		return nil, nil
	}
	if err != nil {
		return nil, sdkerrors.Wrap(errors.With(err, keyvals...), "failed getting transaction receipt")
	}

	isFinalized, err := mgr.isTxReceiptFinalized(chain, txReceipt, confHeight)
	if err != nil {
		return nil, sdkerrors.Wrapf(errors.With(err, keyvals...), "cannot determine if the transaction %s is finalized", txID.Hex())
	}
	if !isFinalized {
		logger.Debug(fmt.Sprintf("transaction %s in block %s not finalized", txID.Hex(), txReceipt.BlockNumber.String()))

		return nil, nil
	}

	return txReceipt, nil
}

func DecodeERC20TransferEvent(log *geth.Log) (types.EventTransfer, error) {
	if len(log.Topics) != 3 || log.Topics[0] != ERC20TransferSig {
		return types.EventTransfer{}, fmt.Errorf("log is not an ERC20 transfer")
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return types.EventTransfer{}, err
	}

	to := common.BytesToAddress(log.Topics[2][:])

	arguments := abi.Arguments{
		{Type: uint256Type},
	}

	params, err := arguments.Unpack(log.Data)
	if err != nil {
		return types.EventTransfer{}, err
	}

	return types.EventTransfer{
		To:     types.Address(to),
		Amount: sdk.NewUintFromBigInt(params[0].(*big.Int)),
	}, nil
}

func DecodeERC20TokenDeploymentEvent(log *geth.Log) (types.EventTokenDeployed, error) {
	if len(log.Topics) != 1 || log.Topics[0] != ERC20TokenDeploymentSig {
		return types.EventTokenDeployed{}, fmt.Errorf("event is not for an ERC20 token deployment")
	}

	// Decode the data field
	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return types.EventTokenDeployed{}, err
	}
	addressType, err := abi.NewType("address", "address", nil)
	if err != nil {
		return types.EventTokenDeployed{}, err
	}

	arguments := abi.Arguments{{Type: stringType}, {Type: addressType}}
	params, err := types.StrictDecode(arguments, log.Data)
	if err != nil {
		return types.EventTokenDeployed{}, err
	}

	return types.EventTokenDeployed{
		Symbol:       params[0].(string),
		TokenAddress: types.Address(params[1].(common.Address)),
	}, nil
}

func DecodeMultisigOperatorshipTransferredEvent(log *geth.Log) (types.EventMultisigOperatorshipTransferred, error) {
	if len(log.Topics) != 1 || log.Topics[0] != MultisigTransferOperatorshipSig {
		return types.EventMultisigOperatorshipTransferred{}, fmt.Errorf("event is not OperatorshipTransferred")
	}

	newAddresses, newWeights, newThreshold, err := unpackMultisigTransferKeyEvent(log)
	if err != nil {
		return types.EventMultisigOperatorshipTransferred{}, err
	}

	event := types.EventMultisigOperatorshipTransferred{
		NewOperators: slices.Map(newAddresses, func(addr common.Address) types.Address { return types.Address(addr) }),
		NewWeights:   slices.Map(newWeights, sdk.NewUintFromBigInt),
		NewThreshold: sdk.NewUintFromBigInt(newThreshold),
	}

	return event, nil
}

func unpackMultisigTransferKeyEvent(log *geth.Log) ([]common.Address, []*big.Int, *big.Int, error) {
	bytesType := funcs.Must(abi.NewType("bytes", "bytes", nil))
	newOperatorsData, err := types.StrictDecode(abi.Arguments{{Type: bytesType}}, log.Data)
	if err != nil {
		return nil, nil, nil, err
	}

	addressesType := funcs.Must(abi.NewType("address[]", "address[]", nil))
	uint256ArrayType := funcs.Must(abi.NewType("uint256[]", "uint256[]", nil))
	uint256Type := funcs.Must(abi.NewType("uint256", "uint256", nil))

	arguments := abi.Arguments{{Type: addressesType}, {Type: uint256ArrayType}, {Type: uint256Type}}
	params, err := types.StrictDecode(arguments, newOperatorsData[0].([]byte))
	if err != nil {
		return nil, nil, nil, err
	}

	return params[0].([]common.Address), params[1].([]*big.Int), params[2].(*big.Int), nil
}
