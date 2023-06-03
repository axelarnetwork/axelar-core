package evm

import (
	"bytes"
	"context"
	goerrors "errors"
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

	"github.com/axelarnetwork/axelar-core/sdk-utils/broadcast"
	"github.com/axelarnetwork/axelar-core/utils/errors"
	"github.com/axelarnetwork/axelar-core/vald/evm/rpc"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	voteTypes "github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/log"
	rs "github.com/axelarnetwork/utils/monads/results"
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

// NotFinalized is returned when a transaction is not finalized
var NotFinalized = goerrors.New("not finalized")

// Mgr manages all communication with Ethereum
type Mgr struct {
	rpcs                      map[string]rpc.Client
	broadcaster               broadcast.Broadcaster
	validator                 sdk.ValAddress
	proxy                     sdk.AccAddress
	latestFinalizedBlockCache LatestFinalizedBlockCache
	axelarChainID             string
}

// NewMgr returns a new Mgr instance
func NewMgr(rpcs map[string]rpc.Client, broadcaster broadcast.Broadcaster, valAddr sdk.ValAddress, proxy sdk.AccAddress, latestFinalizedBlockCache LatestFinalizedBlockCache, axelarChainID string) *Mgr {
	return &Mgr{
		rpcs:                      rpcs,
		proxy:                     proxy,
		broadcaster:               broadcaster,
		validator:                 valAddr,
		latestFinalizedBlockCache: latestFinalizedBlockCache,
		axelarChainID:             axelarChainID,
	}
}

func (mgr Mgr) logger(keyvals ...any) log.Logger {
	keyvals = append([]any{"listener", "evm"}, keyvals...)
	return log.WithKeyVals(keyvals...)
}

// ProcessNewChain notifies the operator that vald needs to be restarted/udpated for a new chain
func (mgr Mgr) ProcessNewChain(event *types.ChainAdded) (err error) {
	mgr.logger().Info(fmt.Sprintf("VALD needs to be updated and restarted for new chain %s", event.Chain.String()))
	return nil
}

// ProcessDepositConfirmation votes on the correctness of an EVM chain token deposit
func (mgr Mgr) ProcessDepositConfirmation(event *types.ConfirmDepositStarted) error {
	if !mgr.isParticipantOf(event.Participants) {
		mgr.logger("pollID", event.PollID).Debug("ignoring deposit confirmation poll: not a participant")
		return nil
	}

	txReceipt, err := mgr.GetTxReceiptIfFinalized(event.Chain, common.Hash(event.TxID), event.ConfirmationHeight)
	if err != nil {
		return err
	}
	if txReceipt == nil {
		mgr.logger().Infof("broadcasting empty vote for poll %s", event.PollID.String())
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
			mgr.logger().Debug(sdkerrors.Wrap(err, "decode event Transfer failed").Error())
			continue
		}

		if erc20Event.To != event.DepositAddress {
			continue
		}

		if err := erc20Event.ValidateBasic(); err != nil {
			mgr.logger().Debug(sdkerrors.Wrap(err, "invalid event Transfer").Error())
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

	mgr.logger().Infof("broadcasting vote %v for poll %s", events, event.PollID.String())
	_, err = mgr.broadcaster.Broadcast(context.TODO(), voteTypes.NewVoteRequest(mgr.proxy, event.PollID, types.NewVoteEvents(event.Chain, events...)))

	return err
}

// ProcessTokenConfirmation votes on the correctness of an EVM chain token deployment
func (mgr Mgr) ProcessTokenConfirmation(event *types.ConfirmTokenStarted) error {
	if !mgr.isParticipantOf(event.Participants) {
		mgr.logger("pollID", event.PollID).Debug("ignoring token confirmation poll: not a participant")
		return nil
	}

	txReceipt, err := mgr.GetTxReceiptIfFinalized(event.Chain, common.Hash(event.TxID), event.ConfirmationHeight)
	if err != nil {
		return err
	}
	if txReceipt == nil {
		mgr.logger().Infof("broadcasting empty vote for poll %s", event.PollID.String())
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
			mgr.logger().Debug(sdkerrors.Wrap(err, "decode event TokenDeployed failed").Error())
			continue
		}

		if erc20Event.TokenAddress != event.TokenAddress || erc20Event.Symbol != event.TokenDetails.Symbol {
			continue
		}

		if err := erc20Event.ValidateBasic(); err != nil {
			mgr.logger().Debug(sdkerrors.Wrap(err, "invalid event ERC20TokenDeployment").Error())
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

	mgr.logger().Infof("broadcasting vote %v for poll %s", events, event.PollID.String())
	_, err = mgr.broadcaster.Broadcast(context.TODO(), voteTypes.NewVoteRequest(mgr.proxy, event.PollID, types.NewVoteEvents(event.Chain, events...)))

	return err
}

// This is a temporary workaround to allow a Filecoin key transfer tx to be processed.
// This workaround allows validators to avoid having to sync an archival node for Filecoin to process this tx
// The tx receipt is hardcoded below and can be verified against your own archival node or the explorer
// https://filfox.info/en/tx/0xcdb332629b739b752ae291f988a58c59bc963c8aea4f7008b8537c4579ade271
var (
	// FilecoinTransferKeyTxID is the tx hash of the Filecoin key transfer
	FilecoinTransferKeyTxID = types.Hash(common.BytesToHash(funcs.Must(
		hexutil.Decode("0xcdb332629b739b752ae291f988a58c59bc963c8aea4f7008b8537c4579ade271"),
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
				Data: funcs.Must(hexutil.Decode("0x000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000009c0000000000000000000000000000000000000000000000000000000000001ecd8000000000000000000000000000000000000000000000000000000000000004a000000000000000000000000030fe8ea0cbda2850a6ed251685e3cc5a98537b5000000000000000000000000083495480fffeeb3d858e490455d9621e8e2ec060000000000000000000000000a90ca81b1fe573a1f18cf09c6e31b4f7083eaaf0000000000000000000000000b08d5afa0f2f8f2d872b84cda0607f82e1cc5f50000000000000000000000000ea62310ffd1a681cff5540ce6cd9edbe58af8f300000000000000000000000015c326450952bfa002ca419cc73eded137a2145d00000000000000000000000017b7d57e4e268145e63e0f613faa3307ebec22020000000000000000000000001836588dbb9aa8281ea3ed97be5ea839855342080000000000000000000000001840be2565e01bfb3bf27b275eaf2a58e2b32c150000000000000000000000001891fdecf6d53a5ba27e6509d982a7a2641d95d90000000000000000000000001a06bc7488c8f68c297d0d29d171e4c3c5de29ce0000000000000000000000001a6b335475c4f25b87fc5d15189f0b521441a7190000000000000000000000001e64330cbbea099617272f480fd70fa9563321b90000000000000000000000001ff303d0dc1166b48147b09f363b71cfc8398c370000000000000000000000002091c7adc89bca324b84cb3c3e8183adbb8e68e90000000000000000000000002c2a44a59e69a1a4e2dabf5e6aaf3867ad0079270000000000000000000000002d360e4d60880359498d92b2cb47c5cb26bf50c80000000000000000000000002d76c90b4106564eef63ed71e4b4cc3c4d31f83f0000000000000000000000002d7c74b583850adeb8b8f3a8eb7c24d3d44a6a0d0000000000000000000000002fc248f0eca6f2c6fbe20009a5181642b07383ce0000000000000000000000003908c326456518b241c9675de19c4e5c18eaf15e000000000000000000000000437998809fb9d3a903b879ddd662b7a0e4f4b8f600000000000000000000000049bddbf7a6b8ff08b74b0119b6a076083144c54f0000000000000000000000004adc52763f84ee4a7256cc7d17c88b96e2a49f3b0000000000000000000000004c00a1b271f0adaed06c6d1d7d1ddcb8cfd083140000000000000000000000004c43a134eb1120c6153203605d54f686a433e1bc0000000000000000000000004ffb9cdfda2e84488ea5a9c18b0088aad05a4a44000000000000000000000000518f45b72872f55f62fba0c8b6afcdf1f18f20b70000000000000000000000005b0ea5b74d013fe3179cfa271aaf62b79a14e8b70000000000000000000000005bd77555748090e41b950d9ebad5ddc75c464bf20000000000000000000000005d7698e5726c287cc5961d0c53762183894c245c00000000000000000000000061b3d381e64aa96f0ad435f7b71d192cf0078d9200000000000000000000000063872a167e35ddd4a94cbdc5c02130ed60e02c0300000000000000000000000066a85597f6651c4e4743289c66d495b06608ad040000000000000000000000006a5bcef90a6f4493a26d019d1276d62827c4faad0000000000000000000000006cc1b81768ec94519aeacd0ea015ebe04fe1e26c0000000000000000000000007368e6189005b7d0b777d2137b0889be2c97aabf00000000000000000000000073da12afc8fa2dc273f983556865212d447ac4350000000000000000000000007572c1824c74ebb403f5de962c61a88df3ddc86a00000000000000000000000076ae621774d93ae150e4ae5bdc2024938037c89a00000000000000000000000077c19dfcfbd83dbf809565c05c27729dcf22377600000000000000000000000078e77eb2342e9fddc819196f27bd8fe02d1bf4470000000000000000000000007c484a8c342f18231500f95f52a6e7c7e2bbd58100000000000000000000000080ed109bc404d943356c76df6c473b182af64e2c0000000000000000000000008348ee1378c260a545e02eab02e47887440e6729000000000000000000000000856860525316a761fb292376600eaf8154c8f06d000000000000000000000000890b440b2588d8c296f0c94819b907eebf181a58000000000000000000000000913470552a33889274d80cc833f77ac8c02a764e0000000000000000000000009216598829eb776488056eba4f4d6c1fc9fb3f6200000000000000000000000096e2d80e681f47a104fd16d74068a93766dd89660000000000000000000000009cee4234c180f9fbfb4297c17eb8c21674a873e5000000000000000000000000a1b08a8fc6bee2d32e158f1a9ec2a2408b5b50cb000000000000000000000000a7b5bbcf9d056ca004c98f725c460c011ca919b4000000000000000000000000ade93d08ea7724fa4870b755a960b3f78cc99d73000000000000000000000000b3ff4adf350f97a1e984114428d8608b731dfe68000000000000000000000000bee340c603362d804f881112df2ca0eee8a1816e000000000000000000000000c4d5669812bf923319a568f9b97b905f7f558d5a000000000000000000000000c9f77e301c871cbff2db986312941623f2313845000000000000000000000000d0678601e5a7263b37bc664758672ce235acd013000000000000000000000000d078334539033b21a0aa6c8404c8b1886d7c681d000000000000000000000000d73ec07910b9f194c3650936c71762855b9f7411000000000000000000000000d93c976fb3eef0b6a4b55fc9d577146924e8e3c0000000000000000000000000dc131f62d737c27ce2c4454f4f7c95e16108680d000000000000000000000000dd17d7832402175244938945d054f87ad73a126b000000000000000000000000e04792278a980e331c874b9f332e6f223b2882cb000000000000000000000000e35f71101580a6fffe3676058998ea2fb1009481000000000000000000000000e5c24155fb8a31237a102c29398fbdd735dd99d8000000000000000000000000f02650188e96702d27cad6f284a5af0e142cde75000000000000000000000000f029be4b0430acb9d4e8046cb4aea3ba4c4687fe000000000000000000000000f58c477ca48c9154d9adfe3f22e1185a614b8144000000000000000000000000f64f3efdc2b028311e9e995889d522f41d3ec40b000000000000000000000000fb8cac70a2acef84e34c85cdcfb799b43b5c12ee000000000000000000000000fe8ed66988052d316c4aebdfd9537df94890e5ad000000000000000000000000ff6e4973b3c8f27120ca5c03472d3712102d61e7000000000000000000000000000000000000000000000000000000000000004a000000000000000000000000000000000000000000000000000000000000062b00000000000000000000000000000000000000000000000000000000000008c20000000000000000000000000000000000000000000000000000000000000a2000000000000000000000000000000000000000000000000000000000000005ae00000000000000000000000000000000000000000000000000000000000007580000000000000000000000000000000000000000000000000000000000000d8f00000000000000000000000000000000000000000000000000000000000006cb0000000000000000000000000000000000000000000000000000000000000d8200000000000000000000000000000000000000000000000000000000000007590000000000000000000000000000000000000000000000000000000000000a790000000000000000000000000000000000000000000000000000000000000f38000000000000000000000000000000000000000000000000000000000000061d0000000000000000000000000000000000000000000000000000000000000b3b0000000000000000000000000000000000000000000000000000000000000e7b0000000000000000000000000000000000000000000000000000000000000d680000000000000000000000000000000000000000000000000000000000000e6b00000000000000000000000000000000000000000000000000000000000010ee00000000000000000000000000000000000000000000000000000000000008a80000000000000000000000000000000000000000000000000000000000000a820000000000000000000000000000000000000000000000000000000000000e2c00000000000000000000000000000000000000000000000000000000000000c20000000000000000000000000000000000000000000000000000000000000ef30000000000000000000000000000000000000000000000000000000000000db0000000000000000000000000000000000000000000000000000000000000071a0000000000000000000000000000000000000000000000000000000000000ccc000000000000000000000000000000000000000000000000000000000000076b0000000000000000000000000000000000000000000000000000000000000afd000000000000000000000000000000000000000000000000000000000000084900000000000000000000000000000000000000000000000000000000000010820000000000000000000000000000000000000000000000000000000000000cc20000000000000000000000000000000000000000000000000000000000000d1e0000000000000000000000000000000000000000000000000000000000000eec0000000000000000000000000000000000000000000000000000000000000b9100000000000000000000000000000000000000000000000000000000000013370000000000000000000000000000000000000000000000000000000000000b520000000000000000000000000000000000000000000000000000000000000cc80000000000000000000000000000000000000000000000000000000000000abb0000000000000000000000000000000000000000000000000000000000000c5a0000000000000000000000000000000000000000000000000000000000000eb10000000000000000000000000000000000000000000000000000000000000eae00000000000000000000000000000000000000000000000000000000000005310000000000000000000000000000000000000000000000000000000000000dca0000000000000000000000000000000000000000000000000000000000000a3c00000000000000000000000000000000000000000000000000000000000000eb00000000000000000000000000000000000000000000000000000000000007570000000000000000000000000000000000000000000000000000000000000e62000000000000000000000000000000000000000000000000000000000000074f0000000000000000000000000000000000000000000000000000000000000754000000000000000000000000000000000000000000000000000000000000104500000000000000000000000000000000000000000000000000000000000006180000000000000000000000000000000000000000000000000000000000000d2700000000000000000000000000000000000000000000000000000000000017d200000000000000000000000000000000000000000000000000000000000007160000000000000000000000000000000000000000000000000000000000000c8d000000000000000000000000000000000000000000000000000000000000090d0000000000000000000000000000000000000000000000000000000000000dc200000000000000000000000000000000000000000000000000000000000008f90000000000000000000000000000000000000000000000000000000000000b420000000000000000000000000000000000000000000000000000000000000cf300000000000000000000000000000000000000000000000000000000000014ad0000000000000000000000000000000000000000000000000000000000000fd00000000000000000000000000000000000000000000000000000000000000b460000000000000000000000000000000000000000000000000000000000000775000000000000000000000000000000000000000000000000000000000000053d00000000000000000000000000000000000000000000000000000000000007d20000000000000000000000000000000000000000000000000000000000000ff00000000000000000000000000000000000000000000000000000000000000a8700000000000000000000000000000000000000000000000000000000000000a100000000000000000000000000000000000000000000000000000000000000c20000000000000000000000000000000000000000000000000000000000000b4800000000000000000000000000000000000000000000000000000000000008020000000000000000000000000000000000000000000000000000000000001ba30000000000000000000000000000000000000000000000000000000000000cc80000000000000000000000000000000000000000000000000000000000000820")),
			},
			{
				Address: common.BytesToAddress(funcs.Must(hexutil.Decode("0xe432150cce91c13a887f7D836923d5597adD8E31"))),
				Topics: []common.Hash{
					crypto.Keccak256Hash([]byte("OperatorshipTransferred(bytes)")),
				},
				Data: funcs.Must(hexutil.Decode("0x00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000001320000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000009c0000000000000000000000000000000000000000000000000000000000001ecd8000000000000000000000000000000000000000000000000000000000000004a000000000000000000000000030fe8ea0cbda2850a6ed251685e3cc5a98537b5000000000000000000000000083495480fffeeb3d858e490455d9621e8e2ec060000000000000000000000000a90ca81b1fe573a1f18cf09c6e31b4f7083eaaf0000000000000000000000000b08d5afa0f2f8f2d872b84cda0607f82e1cc5f50000000000000000000000000ea62310ffd1a681cff5540ce6cd9edbe58af8f300000000000000000000000015c326450952bfa002ca419cc73eded137a2145d00000000000000000000000017b7d57e4e268145e63e0f613faa3307ebec22020000000000000000000000001836588dbb9aa8281ea3ed97be5ea839855342080000000000000000000000001840be2565e01bfb3bf27b275eaf2a58e2b32c150000000000000000000000001891fdecf6d53a5ba27e6509d982a7a2641d95d90000000000000000000000001a06bc7488c8f68c297d0d29d171e4c3c5de29ce0000000000000000000000001a6b335475c4f25b87fc5d15189f0b521441a7190000000000000000000000001e64330cbbea099617272f480fd70fa9563321b90000000000000000000000001ff303d0dc1166b48147b09f363b71cfc8398c370000000000000000000000002091c7adc89bca324b84cb3c3e8183adbb8e68e90000000000000000000000002c2a44a59e69a1a4e2dabf5e6aaf3867ad0079270000000000000000000000002d360e4d60880359498d92b2cb47c5cb26bf50c80000000000000000000000002d76c90b4106564eef63ed71e4b4cc3c4d31f83f0000000000000000000000002d7c74b583850adeb8b8f3a8eb7c24d3d44a6a0d0000000000000000000000002fc248f0eca6f2c6fbe20009a5181642b07383ce0000000000000000000000003908c326456518b241c9675de19c4e5c18eaf15e000000000000000000000000437998809fb9d3a903b879ddd662b7a0e4f4b8f600000000000000000000000049bddbf7a6b8ff08b74b0119b6a076083144c54f0000000000000000000000004adc52763f84ee4a7256cc7d17c88b96e2a49f3b0000000000000000000000004c00a1b271f0adaed06c6d1d7d1ddcb8cfd083140000000000000000000000004c43a134eb1120c6153203605d54f686a433e1bc0000000000000000000000004ffb9cdfda2e84488ea5a9c18b0088aad05a4a44000000000000000000000000518f45b72872f55f62fba0c8b6afcdf1f18f20b70000000000000000000000005b0ea5b74d013fe3179cfa271aaf62b79a14e8b70000000000000000000000005bd77555748090e41b950d9ebad5ddc75c464bf20000000000000000000000005d7698e5726c287cc5961d0c53762183894c245c00000000000000000000000061b3d381e64aa96f0ad435f7b71d192cf0078d9200000000000000000000000063872a167e35ddd4a94cbdc5c02130ed60e02c0300000000000000000000000066a85597f6651c4e4743289c66d495b06608ad040000000000000000000000006a5bcef90a6f4493a26d019d1276d62827c4faad0000000000000000000000006cc1b81768ec94519aeacd0ea015ebe04fe1e26c0000000000000000000000007368e6189005b7d0b777d2137b0889be2c97aabf00000000000000000000000073da12afc8fa2dc273f983556865212d447ac4350000000000000000000000007572c1824c74ebb403f5de962c61a88df3ddc86a00000000000000000000000076ae621774d93ae150e4ae5bdc2024938037c89a00000000000000000000000077c19dfcfbd83dbf809565c05c27729dcf22377600000000000000000000000078e77eb2342e9fddc819196f27bd8fe02d1bf4470000000000000000000000007c484a8c342f18231500f95f52a6e7c7e2bbd58100000000000000000000000080ed109bc404d943356c76df6c473b182af64e2c0000000000000000000000008348ee1378c260a545e02eab02e47887440e6729000000000000000000000000856860525316a761fb292376600eaf8154c8f06d000000000000000000000000890b440b2588d8c296f0c94819b907eebf181a58000000000000000000000000913470552a33889274d80cc833f77ac8c02a764e0000000000000000000000009216598829eb776488056eba4f4d6c1fc9fb3f6200000000000000000000000096e2d80e681f47a104fd16d74068a93766dd89660000000000000000000000009cee4234c180f9fbfb4297c17eb8c21674a873e5000000000000000000000000a1b08a8fc6bee2d32e158f1a9ec2a2408b5b50cb000000000000000000000000a7b5bbcf9d056ca004c98f725c460c011ca919b4000000000000000000000000ade93d08ea7724fa4870b755a960b3f78cc99d73000000000000000000000000b3ff4adf350f97a1e984114428d8608b731dfe68000000000000000000000000bee340c603362d804f881112df2ca0eee8a1816e000000000000000000000000c4d5669812bf923319a568f9b97b905f7f558d5a000000000000000000000000c9f77e301c871cbff2db986312941623f2313845000000000000000000000000d0678601e5a7263b37bc664758672ce235acd013000000000000000000000000d078334539033b21a0aa6c8404c8b1886d7c681d000000000000000000000000d73ec07910b9f194c3650936c71762855b9f7411000000000000000000000000d93c976fb3eef0b6a4b55fc9d577146924e8e3c0000000000000000000000000dc131f62d737c27ce2c4454f4f7c95e16108680d000000000000000000000000dd17d7832402175244938945d054f87ad73a126b000000000000000000000000e04792278a980e331c874b9f332e6f223b2882cb000000000000000000000000e35f71101580a6fffe3676058998ea2fb1009481000000000000000000000000e5c24155fb8a31237a102c29398fbdd735dd99d8000000000000000000000000f02650188e96702d27cad6f284a5af0e142cde75000000000000000000000000f029be4b0430acb9d4e8046cb4aea3ba4c4687fe000000000000000000000000f58c477ca48c9154d9adfe3f22e1185a614b8144000000000000000000000000f64f3efdc2b028311e9e995889d522f41d3ec40b000000000000000000000000fb8cac70a2acef84e34c85cdcfb799b43b5c12ee000000000000000000000000fe8ed66988052d316c4aebdfd9537df94890e5ad000000000000000000000000ff6e4973b3c8f27120ca5c03472d3712102d61e7000000000000000000000000000000000000000000000000000000000000004a000000000000000000000000000000000000000000000000000000000000062b00000000000000000000000000000000000000000000000000000000000008c20000000000000000000000000000000000000000000000000000000000000a2000000000000000000000000000000000000000000000000000000000000005ae00000000000000000000000000000000000000000000000000000000000007580000000000000000000000000000000000000000000000000000000000000d8f00000000000000000000000000000000000000000000000000000000000006cb0000000000000000000000000000000000000000000000000000000000000d8200000000000000000000000000000000000000000000000000000000000007590000000000000000000000000000000000000000000000000000000000000a790000000000000000000000000000000000000000000000000000000000000f38000000000000000000000000000000000000000000000000000000000000061d0000000000000000000000000000000000000000000000000000000000000b3b0000000000000000000000000000000000000000000000000000000000000e7b0000000000000000000000000000000000000000000000000000000000000d680000000000000000000000000000000000000000000000000000000000000e6b00000000000000000000000000000000000000000000000000000000000010ee00000000000000000000000000000000000000000000000000000000000008a80000000000000000000000000000000000000000000000000000000000000a820000000000000000000000000000000000000000000000000000000000000e2c00000000000000000000000000000000000000000000000000000000000000c20000000000000000000000000000000000000000000000000000000000000ef30000000000000000000000000000000000000000000000000000000000000db0000000000000000000000000000000000000000000000000000000000000071a0000000000000000000000000000000000000000000000000000000000000ccc000000000000000000000000000000000000000000000000000000000000076b0000000000000000000000000000000000000000000000000000000000000afd000000000000000000000000000000000000000000000000000000000000084900000000000000000000000000000000000000000000000000000000000010820000000000000000000000000000000000000000000000000000000000000cc20000000000000000000000000000000000000000000000000000000000000d1e0000000000000000000000000000000000000000000000000000000000000eec0000000000000000000000000000000000000000000000000000000000000b9100000000000000000000000000000000000000000000000000000000000013370000000000000000000000000000000000000000000000000000000000000b520000000000000000000000000000000000000000000000000000000000000cc80000000000000000000000000000000000000000000000000000000000000abb0000000000000000000000000000000000000000000000000000000000000c5a0000000000000000000000000000000000000000000000000000000000000eb10000000000000000000000000000000000000000000000000000000000000eae00000000000000000000000000000000000000000000000000000000000005310000000000000000000000000000000000000000000000000000000000000dca0000000000000000000000000000000000000000000000000000000000000a3c00000000000000000000000000000000000000000000000000000000000000eb00000000000000000000000000000000000000000000000000000000000007570000000000000000000000000000000000000000000000000000000000000e62000000000000000000000000000000000000000000000000000000000000074f0000000000000000000000000000000000000000000000000000000000000754000000000000000000000000000000000000000000000000000000000000104500000000000000000000000000000000000000000000000000000000000006180000000000000000000000000000000000000000000000000000000000000d2700000000000000000000000000000000000000000000000000000000000017d200000000000000000000000000000000000000000000000000000000000007160000000000000000000000000000000000000000000000000000000000000c8d000000000000000000000000000000000000000000000000000000000000090d0000000000000000000000000000000000000000000000000000000000000dc200000000000000000000000000000000000000000000000000000000000008f90000000000000000000000000000000000000000000000000000000000000b420000000000000000000000000000000000000000000000000000000000000cf300000000000000000000000000000000000000000000000000000000000014ad0000000000000000000000000000000000000000000000000000000000000fd00000000000000000000000000000000000000000000000000000000000000b460000000000000000000000000000000000000000000000000000000000000775000000000000000000000000000000000000000000000000000000000000053d00000000000000000000000000000000000000000000000000000000000007d20000000000000000000000000000000000000000000000000000000000000ff00000000000000000000000000000000000000000000000000000000000000a8700000000000000000000000000000000000000000000000000000000000000a100000000000000000000000000000000000000000000000000000000000000c20000000000000000000000000000000000000000000000000000000000000b4800000000000000000000000000000000000000000000000000000000000008020000000000000000000000000000000000000000000000000000000000001ba30000000000000000000000000000000000000000000000000000000000000cc80000000000000000000000000000000000000000000000000000000000000820")),
			},
			{
				Address: common.BytesToAddress(funcs.Must(hexutil.Decode("0xe432150cce91c13a887f7D836923d5597adD8E31"))),
				Topics: []common.Hash{
					crypto.Keccak256Hash([]byte("Executed(bytes32)")),
					common.BytesToHash(funcs.Must(hexutil.Decode("0x7a88bc0511dd1d299cc36f3be87daf8e945fd97cb116c06972a2157b211bf992"))),
				},
				Data: []byte{},
			},
		},
	}
}

// ProcessTransferKeyConfirmation votes on the correctness of an EVM chain key transfer
func (mgr Mgr) ProcessTransferKeyConfirmation(event *types.ConfirmKeyTransferStarted) error {
	if !mgr.isParticipantOf(event.Participants) {
		mgr.logger("pollID", event.PollID).Debug("ignoring key transfer confirmation poll: not a participant")
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
		mgr.logger().Infof("broadcasting empty vote for poll %s", event.PollID.String())
		_, err := mgr.broadcaster.Broadcast(context.TODO(), voteTypes.NewVoteRequest(mgr.proxy, event.PollID, types.NewVoteEvents(event.Chain)))

		return err
	}

	var events []types.Event
	for i := len(txReceipt.Logs) - 1; i >= 0; i-- {
		txlog := txReceipt.Logs[i]

		if txlog.Topics[0] != MultisigTransferOperatorshipSig {
			continue
		}

		// Event is not emitted by the axelar gateway
		if txlog.Address != common.Address(event.GatewayAddress) {
			continue
		}

		transferOperatorshipEvent, err := DecodeMultisigOperatorshipTransferredEvent(txlog)
		if err != nil {
			mgr.logger().Debug(sdkerrors.Wrap(err, "failed decoding operatorship transferred event").Error())
			continue
		}

		if err := transferOperatorshipEvent.ValidateBasic(); err != nil {
			mgr.logger().Debug(sdkerrors.Wrap(err, "invalid event MultisigTransferOperatorship").Error())
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

	mgr.logger().Infof("broadcasting vote %v for poll %s", events, event.PollID.String())
	_, err = mgr.broadcaster.Broadcast(context.TODO(), voteTypes.NewVoteRequest(mgr.proxy, event.PollID, types.NewVoteEvents(event.Chain, events...)))

	return err
}

// ProcessGatewayTxConfirmation votes on the correctness of an EVM chain gateway's transactions
func (mgr Mgr) ProcessGatewayTxConfirmation(event *types.ConfirmGatewayTxStarted) error {
	if !mgr.isParticipantOf(event.Participants) {
		mgr.logger("pollID", event.PollID).Debug("ignoring gateway tx confirmation poll: not a participant")
		return nil
	}

	txReceipt, err := mgr.GetTxReceiptIfFinalized(event.Chain, common.Hash(event.TxID), event.ConfirmationHeight)
	if err != nil {
		return err
	}
	if txReceipt == nil {
		mgr.logger().Infof("broadcasting empty vote for poll %s", event.PollID.String())
		_, err := mgr.broadcaster.Broadcast(context.TODO(), voteTypes.NewVoteRequest(mgr.proxy, event.PollID, types.NewVoteEvents(event.Chain)))

		return err
	}

	events := mgr.processGatewayTxLogs(event.Chain, event.GatewayAddress, txReceipt.Logs)
	mgr.logger().Infof("broadcasting vote %v for poll %s", events, event.PollID.String())
	_, err = mgr.broadcaster.Broadcast(context.TODO(), voteTypes.NewVoteRequest(mgr.proxy, event.PollID, types.NewVoteEvents(event.Chain, events...)))

	return err
}

// ProcessGatewayTxsConfirmation votes on the correctness of an EVM chain multiple gateway transactions
func (mgr Mgr) ProcessGatewayTxsConfirmation(event *types.ConfirmGatewayTxsStarted) error {
	if !mgr.isParticipantOf(event.Participants) {
		pollIDs := slices.Map(event.PollMappings, func(m types.PollMapping) vote.PollID { return m.PollID })
		mgr.logger("poll_ids", pollIDs).Debug("ignoring gateway txs confirmation poll: not a participant")
		return nil
	}

	txIDs := slices.Map(event.PollMappings, func(poll types.PollMapping) common.Hash { return common.Hash(poll.TxID) })
	txReceipts, err := mgr.GetTxReceiptsIfFinalized(event.Chain, txIDs, event.ConfirmationHeight)
	if err != nil {
		return err
	}

	var votes []sdk.Msg
	for i, result := range txReceipts {
		pollID := event.PollMappings[i].PollID
		txID := event.PollMappings[i].TxID

		logger := mgr.logger("chain", event.Chain, "poll_id", pollID.String(), "tx_id", txID.Hex())

		// only broadcast empty votes if the tx is not found or not finalized
		switch result.Err() {
		case nil:
			events := mgr.processGatewayTxLogs(event.Chain, event.GatewayAddress, result.Ok().Logs)
			logger.Infof("broadcasting vote %v", events)
			votes = append(votes, voteTypes.NewVoteRequest(mgr.proxy, pollID, types.NewVoteEvents(event.Chain, events...)))
		case NotFinalized, ethereum.NotFound:
			logger.Infof("broadcasting empty vote due to error: %s", result.Err().Error())
			votes = append(votes, voteTypes.NewVoteRequest(mgr.proxy, pollID, types.NewVoteEvents(event.Chain)))
		default:
			logger.Errorf("failed to get tx receipt: %s", result.Err().Error())
		}

	}

	_, err = mgr.broadcaster.Broadcast(context.TODO(), votes...)
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
	if err != nil {
		return false, err
	}

	mgr.latestFinalizedBlockCache.Set(chain, latestFinalizedBlockNumber)

	if latestFinalizedBlockNumber.Cmp(txReceipt.BlockNumber) < 0 {
		return false, nil
	}

	return true, nil
}

func (mgr Mgr) GetTxReceiptIfFinalized(chain nexus.ChainName, txID common.Hash, confHeight uint64) (*geth.Receipt, error) {
	client, ok := mgr.rpcs[strings.ToLower(chain.String())]
	if !ok {
		return nil, fmt.Errorf("rpc client not found for chain %s", chain.String())
	}

	txReceipt, err := client.TransactionReceipt(context.Background(), txID)
	keyvals := []interface{}{"chain", chain.String(), "tx_id", txID.Hex()}
	logger := mgr.logger(keyvals...)
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

// GetTxReceiptsIfFinalized retrieves receipts for provided transaction IDs, only if they're finalized.
func (mgr Mgr) GetTxReceiptsIfFinalized(chain nexus.ChainName, txIDs []common.Hash, confHeight uint64) ([]rs.Result[*geth.Receipt], error) {
	client, ok := mgr.rpcs[strings.ToLower(chain.String())]
	if !ok {
		return nil, fmt.Errorf("rpc client not found for chain %s", chain.String())
	}

	results, err := client.TransactionReceipts(context.Background(), txIDs)
	if err != nil {
		return nil, sdkerrors.Wrapf(errors.With(err, "chain", chain.String(), "tx_ids", txIDs),
			"cannot get transaction receipts")
	}

	isFinalized := func(receipt *geth.Receipt) rs.Result[*geth.Receipt] {
		isFinalized, err := mgr.isTxReceiptFinalized(chain, receipt, confHeight)
		if err != nil {
			return rs.FromErr[*geth.Receipt](sdkerrors.Wrapf(errors.With(err, "chain", chain.String()),
				"cannot determine if the transaction %s is finalized", receipt.TxHash.Hex()),
			)
		}

		if !isFinalized {
			return rs.FromErr[*geth.Receipt](NotFinalized)
		}

		return rs.FromOk(receipt)
	}

	return slices.Map(results, func(r rpc.Result) rs.Result[*geth.Receipt] {
		return rs.Pipe(rs.Result[*geth.Receipt](r), isFinalized)
	}), nil
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

// extract receipt processing from ProcessGatewayTxConfirmation, so that it can be used in ProcessGatewayTxsConfirmation
func (mgr Mgr) processGatewayTxLogs(chain nexus.ChainName, gatewayAddress types.Address, logs []*geth.Log) []types.Event {
	var events []types.Event
	for i, txlog := range logs {
		if !bytes.Equal(gatewayAddress.Bytes(), txlog.Address.Bytes()) {
			continue
		}

		switch txlog.Topics[0] {
		case ContractCallSig:
			gatewayEvent, err := DecodeEventContractCall(txlog)
			if err != nil {
				mgr.logger().Debug(sdkerrors.Wrap(err, "decode event ContractCall failed").Error())
				continue
			}

			if err := gatewayEvent.ValidateBasic(); err != nil {
				mgr.logger().Debug(sdkerrors.Wrap(err, "invalid event ContractCall").Error())
				continue
			}

			events = append(events, types.Event{
				Chain: chain,
				TxID:  types.Hash(txlog.TxHash),
				Index: uint64(i),
				Event: &types.Event_ContractCall{
					ContractCall: &gatewayEvent,
				},
			})
		case ContractCallWithTokenSig:
			gatewayEvent, err := DecodeEventContractCallWithToken(txlog)
			if err != nil {
				mgr.logger().Debug(sdkerrors.Wrap(err, "decode event ContractCallWithToken failed").Error())
				continue
			}

			if err := gatewayEvent.ValidateBasic(); err != nil {
				mgr.logger().Debug(sdkerrors.Wrap(err, "invalid event ContractCallWithToken").Error())
				continue
			}

			events = append(events, types.Event{
				Chain: chain,
				TxID:  types.Hash(txlog.TxHash),
				Index: uint64(i),
				Event: &types.Event_ContractCallWithToken{
					ContractCallWithToken: &gatewayEvent,
				},
			})
		case TokenSentSig:
			gatewayEvent, err := DecodeEventTokenSent(txlog)
			if err != nil {
				mgr.logger().Debug(sdkerrors.Wrap(err, "decode event TokenSent failed").Error())
			}

			if err := gatewayEvent.ValidateBasic(); err != nil {
				mgr.logger().Debug(sdkerrors.Wrap(err, "invalid event TokenSent").Error())
				continue
			}

			events = append(events, types.Event{
				Chain: chain,
				TxID:  types.Hash(txlog.TxHash),
				Index: uint64(i),
				Event: &types.Event_TokenSent{
					TokenSent: &gatewayEvent,
				},
			})
		default:
		}
	}

	return events
}

// isParticipantOf checks if the validator is in the poll participants list
func (mgr Mgr) isParticipantOf(participants []sdk.ValAddress) bool {
	return slices.Any(participants, func(v sdk.ValAddress) bool { return v.Equals(mgr.validator) })
}
