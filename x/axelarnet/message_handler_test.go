package axelarnet_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	captypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v4/modules/core/exported"
	"github.com/stretchr/testify/assert"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/axelarnet"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/mock"
	axelartestutils "github.com/axelarnetwork/axelar-core/x/axelarnet/types/testutils"
	evmKeeper "github.com/axelarnetwork/axelar-core/x/evm/keeper"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	evmtestutils "github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexustestutils "github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
	"github.com/axelarnetwork/utils/funcs"
	. "github.com/axelarnetwork/utils/test"
)

func TestHandleMessage(t *testing.T) {
	var (
		ctx      sdk.Context
		k        keeper.Keeper
		packet   ibcchanneltypes.Packet
		b        *mock.BankKeeperMock
		n        *mock.NexusMock
		channelK *mock.ChannelKeeperMock
		ibcK     keeper.IBCKeeper

		ics20Packet ibctransfertypes.FungibleTokenPacketData
		message     axelarnet.Message
	)

	sourceChannel := axelartestutils.RandomChannel()
	receiverChannel := axelartestutils.RandomChannel()

	srcChain := nexustestutils.RandomChain()
	destChain := nexustestutils.RandomChain()
	destAddress := evmtestutils.RandomAddress().Hex()
	payload := rand.BytesBetween(100, 500)

	givenPacketWithMessage := Given("a packet with general message", func() {
		message = axelarnet.Message{
			DestinationChain:   destChain.Name.String(),
			DestinationAddress: destAddress,
			Payload:            payload,
			Type:               nexus.TypeGeneralMessage,
		}

		ctx, k, channelK = setup()
		channelK.SendPacketFunc = func(sdk.Context, *captypes.Capability, ibcexported.PacketI) error { return nil }
		n = &mock.NexusMock{
			SetNewMessageFunc: func(sdk.Context, nexus.GeneralMessage) error { return nil },
			GetChainFunc: func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				switch chain {
				case srcChain.Name:
					return srcChain, true
				default:
					return nexus.Chain{}, false

				}
			},
			ValidateAddressFunc: func(ctx sdk.Context, address nexus.CrossChainAddress) error {
				switch address.Chain.Module {
				case evmtypes.ModuleName:
					return evmKeeper.NewAddressValidator()(ctx, address)
				default:
					return fmt.Errorf("module not found")
				}
			},
			GenerateMessageIDFunc: func(_ sdk.Context, bz []byte) string {
				hash := sha256.Sum256(bz)
				return fmt.Sprintf("%s-%d", hex.EncodeToString(hash[:]), 0)
			},
		}
		ibcK = keeper.NewIBCKeeper(k, &mock.IBCTransferKeeperMock{}, &mock.ChannelKeeperMock{})
	})

	ackError := func() func(t *testing.T) {
		return func(t *testing.T) {
			acknowledgement := axelarnet.OnRecvMessage(ctx, k, ibcK, n, b, packet)
			var ack ibcchanneltypes.Acknowledgement
			funcs.MustNoErr(ibctransfertypes.ModuleCdc.UnmarshalJSON(acknowledgement.Acknowledgement(), &ack))
			assert.False(t, ack.Success())
		}
	}

	givenPacketWithMessage.
		When("packet receiver is not Axelar gmp account", func() {
			ics20Packet = ibctransfertypes.NewFungibleTokenPacketData(
				rand.Denom(5, 10), strconv.FormatInt(rand.PosI64(), 10), rand.AccAddr().String(), rand.AccAddr().String(),
			)

			ics20Packet.Memo = string(rand.BytesBetween(100, 500))
			packet = axelartestutils.RandomPacket(ics20Packet, ibctransfertypes.PortID, sourceChannel, ibctransfertypes.PortID, receiverChannel)
		}).
		Then("should not handle message", func(t *testing.T) {
			acknowledgement := axelarnet.OnRecvMessage(ctx, k, ibcK, n, b, packet)
			assert.True(t, acknowledgement.Success())
		}).
		Run(t)

	whenPacketReceiverIsGMPAccount := givenPacketWithMessage.
		When("receiver is gmp account", func() {
			ics20Packet = ibctransfertypes.NewFungibleTokenPacketData(
				rand.Denom(5, 10), strconv.FormatInt(rand.PosI64(), 10), rand.AccAddr().String(), types.MessageSender.String(),
			)
			ics20Packet.Memo = string(funcs.Must(json.Marshal(message)))
			packet = axelartestutils.RandomPacket(ics20Packet, ibctransfertypes.PortID, sourceChannel, ibctransfertypes.PortID, receiverChannel)
		})

	whenPacketReceiverIsGMPAccount.
		When("payload is invalid", func() {
			ics20Packet.Memo = string(rand.BytesBetween(100, 500))
			packet = axelartestutils.RandomPacket(ics20Packet, ibctransfertypes.PortID, sourceChannel, ibctransfertypes.PortID, receiverChannel)
		}).
		Then("should return ack error", ackError()).
		Run(t)

	isIBCPathRegistered := func(isRegistered bool) func() {
		return func() {
			if isRegistered {
				funcs.MustNoErr(k.SetChainByIBCPath(ctx, types.NewIBCPath(ibctransfertypes.PortID, receiverChannel), srcChain.Name))
			}
		}
	}

	whenPacketReceiverIsGMPAccount.
		When("IBC path is not registered", isIBCPathRegistered(false)).
		Then("should return ack error", ackError()).
		Run(t)

	isChainActivated := func(c nexus.Chain, isActivated bool) func() {
		return func() {
			n.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool {
				switch chain.Name {
				case c.Name:
					return isActivated
				default:
					return true
				}
			}
		}
	}

	whenPacketReceiverIsGMPAccount.
		When("IBC path is registered", isIBCPathRegistered(true)).
		When("source chain is not activated", isChainActivated(srcChain, false)).
		Then("should return ack error", ackError()).
		Run(t)

	isChainFound := func(c nexus.Chain, isFound bool) func() {
		return func() {
			n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				switch chain {
				case srcChain.Name:
					return srcChain, true
				case c.Name:
					return c, isFound
				default:
					return nexus.Chain{}, false
				}
			}
		}
	}

	whenSourceChainIsValid := whenPacketReceiverIsGMPAccount.
		When("source chain is valid", func() {
			isIBCPathRegistered(true)()
			isChainActivated(srcChain, true)()
		})

	whenSourceChainIsValid.
		When("dest chain not found", isChainFound(destChain, false)).
		Then("should return ack error", ackError()).
		Run(t)

	whenSourceChainIsValid.
		When("dest chain is found", isChainFound(destChain, true)).
		When("dest chain is evm", func() { destChain.Module = evmtypes.ModuleName }).
		When("dest chain is not activated", isChainActivated(destChain, false)).
		Then("should return ack error", ackError()).
		Run(t)

	whenSourceChainIsValid.
		When("dest chain is found", isChainFound(destChain, true)).
		When("dest chain is evm", func() { destChain.Module = evmtypes.ModuleName }).
		When("dest chain is activated", isChainActivated(destChain, true)).
		When("dest address is not valid", func() {
			message.DestinationAddress = rand.AccAddr().String()
			ics20Packet.Memo = string(funcs.Must(json.Marshal(message)))
			packet = axelartestutils.RandomPacket(ics20Packet, ibctransfertypes.PortID, sourceChannel, ibctransfertypes.PortID, receiverChannel)
		}).
		Then("should return ack error", ackError()).
		Run(t)

	whenMessageIsValid := whenSourceChainIsValid.
		When("dest chain valid", func() {
			isChainFound(destChain, true)()
			destChain.Module = evmtypes.ModuleName
			isChainActivated(destChain, true)()
			message.DestinationAddress = evmtestutils.RandomAddress().Hex()
			ics20Packet.Memo = string(funcs.Must(json.Marshal(message)))
			packet = axelartestutils.RandomPacket(ics20Packet, ibctransfertypes.PortID, sourceChannel, ibctransfertypes.PortID, receiverChannel)
		})

	whenMessageIsValid.
		Then("should return ack success", func(t *testing.T) {
			assert.True(t, axelarnet.OnRecvMessage(ctx, k, ibcK, n, b, packet).Success())
		}).
		Run(t)
}

func TestHandleMessageWithToken(t *testing.T) {
	var (
		ctx      sdk.Context
		k        keeper.Keeper
		packet   ibcchanneltypes.Packet
		b        *mock.BankKeeperMock
		n        *mock.NexusMock
		channelK *mock.ChannelKeeperMock
		ibcK     keeper.IBCKeeper

		denom       string
		amount      string
		ics20Packet ibctransfertypes.FungibleTokenPacketData
		message     axelarnet.Message
	)

	sourceChannel := axelartestutils.RandomChannel()
	receiverChannel := axelartestutils.RandomChannel()

	srcChain := nexustestutils.RandomChain()
	destChain := nexustestutils.RandomChain()
	destChain.Module = evmtypes.ModuleName
	destAddress := evmtestutils.RandomAddress().Hex()
	payload := rand.BytesBetween(100, 500)

	givenPacketWithMessageWithToken := Given("a packet with message with token", func() {
		message = axelarnet.Message{
			DestinationChain:   destChain.Name.String(),
			DestinationAddress: destAddress,
			Payload:            payload,
			Type:               nexus.TypeGeneralMessageWithToken,
		}

		// packet send to axelar gmp account
		denom = rand.Denom(5, 10)
		amount = strconv.FormatInt(rand.PosI64(), 10)
		ics20Packet = ibctransfertypes.NewFungibleTokenPacketData(
			denom, amount, rand.AccAddr().String(), types.MessageSender.String(),
		)
		ics20Packet.Memo = string(funcs.Must(json.Marshal(message)))
		packet = axelartestutils.RandomPacket(ics20Packet, ibctransfertypes.PortID, sourceChannel, ibctransfertypes.PortID, receiverChannel)

		ctx, k, channelK = setup()

		// path registered
		path := types.NewIBCPath(ibctransfertypes.PortID, receiverChannel)
		funcs.MustNoErr(k.SetChainByIBCPath(ctx, path, srcChain.Name))
		funcs.MustNoErr(k.SetCosmosChain(ctx, types.CosmosChain{
			Name:       srcChain.Name,
			IBCPath:    path,
			AddrPrefix: rand.StrBetween(5, 10),
		}))

		channelK.SendPacketFunc = func(sdk.Context, *captypes.Capability, ibcexported.PacketI) error { return nil }
		n = &mock.NexusMock{
			SetNewMessageFunc: func(sdk.Context, nexus.GeneralMessage) error { return nil },
			GetChainFunc: func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				switch chain {
				case srcChain.Name:
					return srcChain, true
				case destChain.Name:
					return destChain, true
				default:
					return nexus.Chain{}, false

				}
			},
			ValidateAddressFunc: func(ctx sdk.Context, address nexus.CrossChainAddress) error {
				switch address.Chain.Module {
				case evmtypes.ModuleName:
					return evmKeeper.NewAddressValidator()(ctx, address)
				default:
					panic("module not found")
				}
			},
			IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool { return true },
			GetChainByNativeAssetFunc: func(ctx sdk.Context, asset string) (nexus.Chain, bool) {
				return srcChain, true
			},
			GenerateMessageIDFunc: func(_ sdk.Context, bz []byte) string {
				hash := sha256.Sum256(bz)
				return fmt.Sprintf("%s-%d", hex.EncodeToString(hash[:]), 0)
			},
		}
		ibcK = keeper.NewIBCKeeper(k, &mock.IBCTransferKeeperMock{
			GetDenomTraceFunc: func(ctx sdk.Context, denomTraceHash tmbytes.HexBytes) (ibctransfertypes.DenomTrace, bool) {
				return ibctransfertypes.DenomTrace{
					Path:      fmt.Sprintf("%s/%s", ibctransfertypes.PortID, receiverChannel),
					BaseDenom: denom,
				}, true
			},
		}, &mock.ChannelKeeperMock{})
		b = &mock.BankKeeperMock{
			GetBalanceFunc: func(ctx sdk.Context, addr sdk.AccAddress, d string) sdk.Coin {
				if addr.Equals(types.MessageSender) {
					return sdk.NewCoin(d, funcs.MustOk(sdk.NewIntFromString(amount)))
				}
				return sdk.NewCoin(d, sdk.ZeroInt())
			},
			SendCoinsFunc: func(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
				return nil
			},
		}
	})

	ackError := func() func(t *testing.T) {
		return func(t *testing.T) {
			acknowledgement := axelarnet.OnRecvMessage(ctx, k, ibcK, n, b, packet)
			var ack ibcchanneltypes.Acknowledgement
			funcs.MustNoErr(ibctransfertypes.ModuleCdc.UnmarshalJSON(acknowledgement.Acknowledgement(), &ack))
			assert.False(t, ack.Success())
		}
	}

	isAssetRegistered := func(c nexus.Chain, isRegistered bool) func() {
		return func() {
			n.IsAssetRegisteredFunc = func(ctx sdk.Context, chain nexus.Chain, denom string) bool {
				switch chain {
				case c:
					return isRegistered
				default:
					return true
				}
			}
		}
	}

	givenPacketWithMessageWithToken.
		When("asset is not registered on source chain", isAssetRegistered(srcChain, false)).
		Then("should return ack error", ackError()).
		Run(t)

	givenPacketWithMessageWithToken.
		When("asset is registered on source chain", isAssetRegistered(srcChain, true)).
		When("asset is not registered on dest chain", isAssetRegistered(destChain, false)).
		Then("should return ack error", ackError()).
		Run(t)

	givenPacketWithMessageWithToken.
		When("asset is registered on source chain", isAssetRegistered(srcChain, true)).
		When("asset is registered on dest chain", isAssetRegistered(destChain, true)).
		Then("should return ack success", func(t *testing.T) {
			assert.True(t, axelarnet.OnRecvMessage(ctx, k, ibcK, n, b, packet).Success())
		}).
		Run(t)
}

func TestHandleSendToken(t *testing.T) {
	var (
		ctx      sdk.Context
		k        keeper.Keeper
		packet   ibcchanneltypes.Packet
		b        *mock.BankKeeperMock
		n        *mock.NexusMock
		channelK *mock.ChannelKeeperMock
		ibcK     keeper.IBCKeeper

		denom       string
		amount      string
		ics20Packet ibctransfertypes.FungibleTokenPacketData
		message     axelarnet.Message
	)

	sourceChannel := axelartestutils.RandomChannel()
	receiverChannel := axelartestutils.RandomChannel()

	srcChain := nexustestutils.RandomChain()
	destChain := nexustestutils.RandomChain()
	destChain.Module = evmtypes.ModuleName
	destAddress := evmtestutils.RandomAddress().Hex()

	givenPacketWithSendToken := Given("a packet with send token", func() {
		message = axelarnet.Message{
			DestinationChain:   destChain.Name.String(),
			DestinationAddress: destAddress,
			Payload:            nil,
			Type:               nexus.TypeSendToken,
		}

		// packet send to axelar gmp account
		denom = rand.Denom(5, 10)
		amount = strconv.FormatInt(rand.PosI64(), 10)
		ics20Packet = ibctransfertypes.NewFungibleTokenPacketData(
			denom, amount, rand.AccAddr().String(), types.MessageSender.String(),
		)
		ics20Packet.Memo = string(funcs.Must(json.Marshal(message)))
		packet = axelartestutils.RandomPacket(ics20Packet, ibctransfertypes.PortID, sourceChannel, ibctransfertypes.PortID, receiverChannel)

		ctx, k, channelK = setup()

		// path registered
		path := types.NewIBCPath(ibctransfertypes.PortID, receiverChannel)
		funcs.MustNoErr(k.SetChainByIBCPath(ctx, path, srcChain.Name))
		funcs.MustNoErr(k.SetCosmosChain(ctx, types.CosmosChain{
			Name:       srcChain.Name,
			IBCPath:    path,
			AddrPrefix: rand.StrBetween(5, 10),
		}))

		channelK.SendPacketFunc = func(sdk.Context, *captypes.Capability, ibcexported.PacketI) error { return nil }
		n = &mock.NexusMock{
			SetNewMessageFunc: func(sdk.Context, nexus.GeneralMessage) error { return nil },
			GetChainFunc: func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				switch chain {
				case srcChain.Name:
					return srcChain, true
				case destChain.Name:
					return destChain, true
				default:
					return nexus.Chain{}, false

				}
			},
			ValidateAddressFunc: func(ctx sdk.Context, address nexus.CrossChainAddress) error {
				switch address.Chain.Module {
				case evmtypes.ModuleName:
					return evmKeeper.NewAddressValidator()(ctx, address)
				default:
					panic("module not found")
				}
			},
			IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool { return true },
			GetChainByNativeAssetFunc: func(ctx sdk.Context, asset string) (nexus.Chain, bool) {
				return srcChain, true
			},
			EnqueueTransferFunc: func(ctx sdk.Context, senderChain nexus.Chain, recipient nexus.CrossChainAddress, asset sdk.Coin) (nexus.TransferID, error) {
				return nexustestutils.RandomTransferID(), nil
			},
			GenerateMessageIDFunc: func(_ sdk.Context, bz []byte) string {
				hash := sha256.Sum256(bz)
				return fmt.Sprintf("%s-%d", hex.EncodeToString(hash[:]), 0)
			},
		}
		ibcK = keeper.NewIBCKeeper(k, &mock.IBCTransferKeeperMock{
			GetDenomTraceFunc: func(ctx sdk.Context, denomTraceHash tmbytes.HexBytes) (ibctransfertypes.DenomTrace, bool) {
				return ibctransfertypes.DenomTrace{
					Path:      fmt.Sprintf("%s/%s", ibctransfertypes.PortID, receiverChannel),
					BaseDenom: denom,
				}, true
			},
		}, &mock.ChannelKeeperMock{})
		b = &mock.BankKeeperMock{
			GetBalanceFunc: func(ctx sdk.Context, addr sdk.AccAddress, d string) sdk.Coin {
				if addr.Equals(types.MessageSender) {
					return sdk.NewCoin(d, funcs.MustOk(sdk.NewIntFromString(amount)))
				}
				return sdk.NewCoin(d, sdk.ZeroInt())
			},
			SendCoinsFunc: func(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
				return nil
			},
		}
	})

	ackError := func() func(t *testing.T) {
		return func(t *testing.T) {
			acknowledgement := axelarnet.OnRecvMessage(ctx, k, ibcK, n, b, packet)
			var ack ibcchanneltypes.Acknowledgement
			funcs.MustNoErr(ibctransfertypes.ModuleCdc.UnmarshalJSON(acknowledgement.Acknowledgement(), &ack))
			assert.False(t, ack.Success())
		}
	}

	isAssetRegistered := func(c nexus.Chain, isRegistered bool) func() {
		return func() {
			n.IsAssetRegisteredFunc = func(ctx sdk.Context, chain nexus.Chain, denom string) bool {
				switch chain {
				case c:
					return isRegistered
				default:
					return true
				}
			}
		}
	}

	givenPacketWithSendToken.
		When("asset is not registered on source chain", isAssetRegistered(srcChain, false)).
		Then("should return ack error", ackError()).
		Run(t)

	givenPacketWithSendToken.
		When("asset is registered on source chain", isAssetRegistered(srcChain, true)).
		When("asset is not registered on dest chain", isAssetRegistered(destChain, false)).
		Then("should return ack error", ackError()).
		Run(t)

	givenPacketWithSendToken.
		When("asset is registered on source chain", isAssetRegistered(srcChain, true)).
		When("asset is registered on dest chain", isAssetRegistered(destChain, true)).
		Then("should return ack success", func(t *testing.T) {
			assert.True(t, axelarnet.OnRecvMessage(ctx, k, ibcK, n, b, packet).Success())
		}).
		Run(t)
}
