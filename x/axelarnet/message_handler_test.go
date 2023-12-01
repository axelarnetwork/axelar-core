package axelarnet_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	captypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v4/modules/core/exported"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/stretchr/testify/assert"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/axelarnet"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
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
		r        axelarnet.RateLimiter

		ics20Packet ibctransfertypes.FungibleTokenPacketData
		message     axelarnet.Message
		genMsg      nexus.GeneralMessage
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
		funcs.MustNoErr(k.SetCosmosChain(ctx, types.CosmosChain{
			Name:       srcChain.Name,
			IBCPath:    axelartestutils.RandomIBCPath(),
			AddrPrefix: "cosmos",
		}))
		channelK.SendPacketFunc = func(sdk.Context, *captypes.Capability, ibcexported.PacketI) error { return nil }
		n = &mock.NexusMock{
			SetNewMessageFunc: func(ctx sdk.Context, msg nexus.GeneralMessage) error {
				genMsg = msg
				return nil
			},
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
				case exported.ModuleName:
					return keeper.NewAddressValidator(k)(ctx, address)
				default:
					return fmt.Errorf("module not found")
				}
			},
			GenerateMessageIDFunc: func(ctx sdk.Context) (string, []byte, uint64) {
				hash := sha256.Sum256(ctx.TxBytes())
				return fmt.Sprintf("%s-%d", hex.EncodeToString(hash[:]), 0), hash[:], 0
			},
			RateLimitTransferFunc: func(ctx sdk.Context, chain nexus.ChainName, asset sdk.Coin, direction nexus.TransferDirection) error {
				return nil
			},
			GetChainByNativeAssetFunc: func(ctx sdk.Context, asset string) (nexus.Chain, bool) {
				return srcChain, true
			},
		}
		ibcK = keeper.NewIBCKeeper(k, &mock.IBCTransferKeeperMock{
			GetDenomTraceFunc: func(ctx sdk.Context, denomTraceHash tmbytes.HexBytes) (ibctransfertypes.DenomTrace, bool) {
				return ibctransfertypes.DenomTrace{
					Path:      fmt.Sprintf("%s/%s", ibctransfertypes.PortID, receiverChannel),
					BaseDenom: rand.Denom(5, 10),
				}, true
			},
		})

		r = axelarnet.NewRateLimiter(&k, n)
		b = &mock.BankKeeperMock{
			SendCoinsFunc: func(sdk.Context, sdk.AccAddress, sdk.AccAddress, sdk.Coins) error { return nil },
		}
	})

	whenRateLimitIsSet := func(randDenom bool) func() {
		return func() {
			token := sdk.NewCoin(ics20Packet.GetDenom(), funcs.MustOk(sdk.NewIntFromString(ics20Packet.Amount)))
			if randDenom {
				token.Denom = rand.Denom(10, 20)
			}

			n.RateLimitTransferFunc = func(ctx sdk.Context, chain nexus.ChainName, asset sdk.Coin, direction nexus.TransferDirection) error {
				if direction == nexus.Incoming && asset.Equal(token) {
					return fmt.Errorf("rate limit exceeded")
				}

				return nil
			}
		}
	}

	ackError := func() func(t *testing.T) {
		return func(t *testing.T) {
			acknowledgement := axelarnet.OnRecvMessage(ctx, k, ibcK, n, b, r, packet)
			var ack ibcchanneltypes.Acknowledgement
			funcs.MustNoErr(ibctransfertypes.ModuleCdc.UnmarshalJSON(acknowledgement.Acknowledgement(), &ack))
			assert.False(t, ack.Success())
		}
	}

	nonGMPPacket := func() {
		ics20Packet = ibctransfertypes.NewFungibleTokenPacketData(
			rand.Denom(5, 10), strconv.FormatInt(rand.PosI64(), 10), rand.AccAddr().String(), rand.AccAddr().String(),
		)

		ics20Packet.Memo = string(rand.BytesBetween(100, 500))
		packet = axelartestutils.RandomPacket(ics20Packet, ibctransfertypes.PortID, sourceChannel, ibctransfertypes.PortID, receiverChannel)
	}

	givenPacketWithMessage.
		When("packet receiver is not Axelar gmp account", nonGMPPacket).
		Then("should not handle message", func(t *testing.T) {
			acknowledgement := axelarnet.OnRecvMessage(ctx, k, ibcK, n, b, r, packet)
			assert.True(t, acknowledgement.Success())
		}).
		Run(t)

	whenPacketReceiverIsGMPAccount := givenPacketWithMessage.
		When("receiver is gmp account", func() {
			ics20Packet = ibctransfertypes.NewFungibleTokenPacketData(
				rand.Denom(5, 10), strconv.FormatInt(rand.PosI64(), 10), rand.AccAddr().String(), types.AxelarGMPAccount.String(),
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
				case exported.Axelarnet.Name:
					return exported.Axelarnet, true
				case c.Name:
					return c, isFound
				default:
					return nexus.Chain{}, false
				}
			}
		}
	}

	givenPacketWithMessage.
		When("packet receiver is not Axelar gmp account", nonGMPPacket).
		When("source chain is valid", func() {
			isIBCPathRegistered(true)()
			isChainActivated(srcChain, true)()
		}).
		When("rate limit is set", whenRateLimitIsSet(false)).
		Then("should fail due to ibc transfer rate limit", ackError()).
		Run(t)

	givenPacketWithMessage.
		When("packet receiver is not Axelar gmp account", nonGMPPacket).
		When("source chain is valid", func() {
			isIBCPathRegistered(true)()
			isChainActivated(srcChain, true)()
		}).
		When("rate limit is set on another asset", whenRateLimitIsSet(true)).
		Then("should not handle message", func(t *testing.T) {
			acknowledgement := axelarnet.OnRecvMessage(ctx, k, ibcK, n, b, r, packet)
			assert.True(t, acknowledgement.Success())
		}).
		Run(t)

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
		When("rate limit is set", whenRateLimitIsSet(false)).
		Then("should return ack error", ackError()).
		Run(t)

	whenMessageIsValid.
		When("rate limit on another asset is set", whenRateLimitIsSet(true)).
		Then("should return ack success", func(t *testing.T) {
			assert.True(t, axelarnet.OnRecvMessage(ctx, k, ibcK, n, b, r, packet).Success())
			assert.Equal(t, genMsg.Status, nexus.Approved)
		}).
		Run(t)

	setFee := func(amount sdk.Int, recipient sdk.AccAddress) {
		fee := axelarnet.Fee{
			Amount:    amount.String(),
			Recipient: recipient.String(),
		}
		message.Fee = &fee
		ics20Packet.Memo = string(funcs.Must(json.Marshal(message)))
		packet = axelartestutils.RandomPacket(ics20Packet, ibctransfertypes.PortID, sourceChannel, ibctransfertypes.PortID, receiverChannel)
	}

	whenMessageIsValid.
		When("fee is negative", func() {
			setFee(sdk.NewInt(-1000), rand.AccAddr())
		}).
		Then("should return ack error", ackError()).
		Run(t)

	whenMessageIsValid.
		When("fee is zero", func() {
			setFee(sdk.ZeroInt(), rand.AccAddr())
		}).
		Then("should return ack error", ackError()).
		Run(t)

	whenMessageIsValid.
		When("fee is greater than transfer amount", func() {
			feeAmount := funcs.MustOk(sdk.NewIntFromString(ics20Packet.Amount)).Add(sdk.OneInt())
			setFee(feeAmount, rand.AccAddr())
		}).
		Then("should return ack error", ackError()).
		Run(t)

	whenMessageIsValid.
		When("fee overflows", func() {
			fee := axelarnet.Fee{
				Amount:    math.BigPow(2, 256).String(),
				Recipient: rand.AccAddr().String(),
			}
			message.Fee = &fee
			ics20Packet.Memo = string(funcs.Must(json.Marshal(message)))
			packet = axelartestutils.RandomPacket(ics20Packet, ibctransfertypes.PortID, sourceChannel, ibctransfertypes.PortID, receiverChannel)

		}).
		Then("should return ack error", ackError()).
		Run(t)

	// Fee related tests
	isAssetRegistered := func(isRegistered bool) func() {
		return func() {
			n.IsAssetRegisteredFunc = func(ctx sdk.Context, chain nexus.Chain, denom string) bool {
				return isRegistered
			}
		}
	}

	whenMessageIsValid.
		When("fee denom is not registered", isAssetRegistered(false)).
		When("message with fee", func() {
			setFee(funcs.MustOk(sdk.NewIntFromString(ics20Packet.Amount)), rand.AccAddr())
		}).
		Then("should return ack error", ackError()).
		Run(t)

	whenMessageIsValid.
		When("fee denom is registered", isAssetRegistered(true)).
		When("message with fee", func() {
			setFee(funcs.MustOk(sdk.NewIntFromString(ics20Packet.Amount)), rand.AccAddr())
		}).
		Then("should return ack success", func(t *testing.T) {
			assert.True(t, axelarnet.OnRecvMessage(ctx, k, ibcK, n, b, r, packet).Success())
			assert.Equal(t, genMsg.Status, nexus.Approved)
		}).
		Run(t)

	whenMessageIsValid.
		When("receiver is in uppercase", func() {
			ics20Packet.Receiver = strings.ToUpper(ics20Packet.Receiver)
			packet = axelartestutils.RandomPacket(ics20Packet, ibctransfertypes.PortID, sourceChannel, ibctransfertypes.PortID, receiverChannel)
		}).
		Then("should return ack error", func(t *testing.T) { ackError() }).
		Run(t)

	whenMessageIsValid.
		When("dest chain is cosmos", func() {
			funcs.MustNoErr(k.SetCosmosChain(ctx, types.CosmosChain{
				Name:       destChain.Name,
				IBCPath:    types.NewIBCPath(ibctransfertypes.PortID, axelartestutils.RandomChannel()),
				AddrPrefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
			}))
			message.DestinationAddress = rand.AccAddr().String()
			ics20Packet.Memo = string(funcs.Must(json.Marshal(message)))
			packet = axelartestutils.RandomPacket(ics20Packet, ibctransfertypes.PortID, sourceChannel, ibctransfertypes.PortID, receiverChannel)

			destChain.Module = exported.ModuleName
			isChainFound(destChain, true)()
		}).
		When("fee denom is registered", isAssetRegistered(true)).
		When("message with fee", func() {
			setFee(funcs.MustOk(sdk.NewIntFromString(ics20Packet.Amount)), rand.AccAddr())
		}).
		Then("should return ack success", func(t *testing.T) {
			assert.True(t, axelarnet.OnRecvMessage(ctx, k, ibcK, n, b, r, packet).Success())
			assert.Equal(t, genMsg.Status, nexus.Approved)
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
		r        axelarnet.RateLimiter

		denom       string
		amount      string
		ics20Packet ibctransfertypes.FungibleTokenPacketData
		message     axelarnet.Message
		genMsg      nexus.GeneralMessage
		feeAmount   sdk.Int
	)

	sourceChannel := axelartestutils.RandomChannel()
	receiverChannel := axelartestutils.RandomChannel()

	srcChain := nexustestutils.RandomChain()
	destChain := nexustestutils.RandomChain()
	destChain.Module = evmtypes.ModuleName
	destAddress := evmtestutils.RandomAddress().Hex()
	payload := rand.BytesBetween(100, 500)
	feeAmount = sdk.ZeroInt()

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
			denom, amount, rand.AccAddr().String(), types.AxelarGMPAccount.String(),
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
			SetNewMessageFunc: func(ctx sdk.Context, msg nexus.GeneralMessage) error {
				genMsg = msg
				return nil
			},
			GetChainFunc: func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				switch chain {
				case srcChain.Name:
					return srcChain, true
				case destChain.Name:
					return destChain, true
				case exported.Axelarnet.Name:
					return exported.Axelarnet, true
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
			GenerateMessageIDFunc: func(ctx sdk.Context) (string, []byte, uint64) {
				hash := sha256.Sum256(ctx.TxBytes())
				return fmt.Sprintf("%s-%d", hex.EncodeToString(hash[:]), 0), hash[:], 0
			},
			RateLimitTransferFunc: func(ctx sdk.Context, chain nexus.ChainName, asset sdk.Coin, direction nexus.TransferDirection) error {
				return nil
			},
		}
		ibcK = keeper.NewIBCKeeper(k, &mock.IBCTransferKeeperMock{
			GetDenomTraceFunc: func(ctx sdk.Context, denomTraceHash tmbytes.HexBytes) (ibctransfertypes.DenomTrace, bool) {
				return ibctransfertypes.DenomTrace{
					Path:      fmt.Sprintf("%s/%s", ibctransfertypes.PortID, receiverChannel),
					BaseDenom: denom,
				}, true
			},
		})
		b = &mock.BankKeeperMock{
			SpendableBalanceFunc: func(ctx sdk.Context, addr sdk.AccAddress, d string) sdk.Coin {
				if addr.Equals(types.AxelarGMPAccount) {
					return sdk.NewCoin(d, funcs.MustOk(sdk.NewIntFromString(amount)).Sub(feeAmount))
				}
				return sdk.NewCoin(d, sdk.ZeroInt())
			},
			SendCoinsFunc: func(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
				return nil
			},
		}
		r = axelarnet.NewRateLimiter(&k, n)
	})

	ackError := func() func(t *testing.T) {
		return func(t *testing.T) {
			acknowledgement := axelarnet.OnRecvMessage(ctx, k, ibcK, n, b, r, packet)
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

	whenRateLimitIsSet := func(randDenom bool) func() {
		return func() {
			token := sdk.NewCoin(denom, funcs.MustOk(sdk.NewIntFromString(amount)))
			if randDenom {
				token.Denom = rand.Denom(10, 20)
			}

			n.RateLimitTransferFunc = func(ctx sdk.Context, chain nexus.ChainName, asset sdk.Coin, direction nexus.TransferDirection) error {
				if direction == nexus.Incoming && asset.Equal(token) {
					return fmt.Errorf("rate limit exceeded")
				}

				return nil
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
		When("rate limit is set", whenRateLimitIsSet(false)).
		Then("should return ack error", ackError()).
		Run(t)

	givenPacketWithMessageWithToken.
		When("asset is registered on source chain", isAssetRegistered(srcChain, true)).
		When("asset is registered on dest chain", isAssetRegistered(destChain, true)).
		When("rate limit on another asset is set", whenRateLimitIsSet(true)).
		Then("should return ack success", func(t *testing.T) {
			assert.True(t, axelarnet.OnRecvMessage(ctx, k, ibcK, n, b, r, packet).Success())
			assert.Equal(t, genMsg.Status, nexus.Approved)
		}).
		Run(t)

	setFee := func(amount sdk.Int, recipient sdk.AccAddress) {
		fee := axelarnet.Fee{
			Amount:    amount.String(),
			Recipient: recipient.String(),
		}
		message.Fee = &fee
		ics20Packet.Memo = string(funcs.Must(json.Marshal(message)))
		packet = axelartestutils.RandomPacket(ics20Packet, ibctransfertypes.PortID, sourceChannel, ibctransfertypes.PortID, receiverChannel)
	}

	givenPacketWithMessageWithToken.
		When("asset is registered on source chain", isAssetRegistered(srcChain, true)).
		When("asset is registered on dest chain", isAssetRegistered(destChain, true)).
		When("fee is equal to transfer amount", func() {
			feeAmount = funcs.MustOk(sdk.NewIntFromString(ics20Packet.Amount))
			setFee(feeAmount, rand.AccAddr())
		}).
		Then("should return ack error", ackError()).
		Run(t)

	givenPacketWithMessageWithToken.
		When("asset is registered on source chain", isAssetRegistered(srcChain, true)).
		When("asset is registered on dest chain", isAssetRegistered(destChain, true)).
		When("fee is valid", func() {
			feeAmount = funcs.MustOk(sdk.NewIntFromString(ics20Packet.Amount)).Sub(sdk.OneInt())
			setFee(feeAmount, rand.AccAddr())
		}).
		Then("should return ack success", func(t *testing.T) {
			assert.True(t, axelarnet.OnRecvMessage(ctx, k, ibcK, n, b, r, packet).Success())
			assert.Equal(t, genMsg.Status, nexus.Approved)
			assert.True(t, genMsg.Asset.Amount.Equal(sdk.OneInt()))
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
		r        axelarnet.RateLimiter

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
			denom, amount, rand.AccAddr().String(), types.AxelarGMPAccount.String(),
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
			GenerateMessageIDFunc: func(ctx sdk.Context) (string, []byte, uint64) {
				hash := sha256.Sum256(ctx.TxBytes())
				return fmt.Sprintf("%s-%d", hex.EncodeToString(hash[:]), 0), hash[:], 0
			},
			RateLimitTransferFunc: func(ctx sdk.Context, chain nexus.ChainName, asset sdk.Coin, direction nexus.TransferDirection) error {
				return nil
			},
		}
		ibcK = keeper.NewIBCKeeper(k, &mock.IBCTransferKeeperMock{
			GetDenomTraceFunc: func(ctx sdk.Context, denomTraceHash tmbytes.HexBytes) (ibctransfertypes.DenomTrace, bool) {
				return ibctransfertypes.DenomTrace{
					Path:      fmt.Sprintf("%s/%s", ibctransfertypes.PortID, receiverChannel),
					BaseDenom: denom,
				}, true
			},
		})
		b = &mock.BankKeeperMock{
			SpendableBalanceFunc: func(ctx sdk.Context, addr sdk.AccAddress, d string) sdk.Coin {
				if addr.Equals(types.AxelarGMPAccount) {
					return sdk.NewCoin(d, funcs.MustOk(sdk.NewIntFromString(amount)))
				}
				return sdk.NewCoin(d, sdk.ZeroInt())
			},
			SendCoinsFunc: func(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
				return nil
			},
		}
		r = axelarnet.NewRateLimiter(&k, n)
	})

	ackError := func() func(t *testing.T) {
		return func(t *testing.T) {
			acknowledgement := axelarnet.OnRecvMessage(ctx, k, ibcK, n, b, r, packet)
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

	whenEnqueueTransferFailed := func() {
		n.EnqueueTransferFunc = func(ctx sdk.Context, senderChain nexus.Chain, recipient nexus.CrossChainAddress, asset sdk.Coin) (nexus.TransferID, error) {
			return 0, fmt.Errorf("enqueue transfer failed")
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
		When("enqueue transfer failed", whenEnqueueTransferFailed).
		Then("should return ack error", ackError()).
		Run(t)

	givenPacketWithSendToken.
		When("asset is registered on source chain", isAssetRegistered(srcChain, true)).
		When("asset is registered on dest chain", isAssetRegistered(destChain, true)).
		Then("should return ack success", func(t *testing.T) {
			assert.True(t, axelarnet.OnRecvMessage(ctx, k, ibcK, n, b, r, packet).Success())
		}).
		Run(t)
}
