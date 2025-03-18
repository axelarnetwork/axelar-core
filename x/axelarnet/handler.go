package axelarnet

import (
	"encoding/hex"
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	"github.com/CosmWasm/wasmd/x/wasm"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/axelarnetwork/axelar-core/utils/events"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// NewProposalHandler returns the handler for the proposals of the axelarnet module
func NewProposalHandler(k keeper.Keeper, nexusK types.Nexus, accountK types.AccountKeeper) govv1beta1.Handler {
	return func(ctx sdk.Context, content govv1beta1.Content) error {
		switch c := content.(type) {
		case *types.CallContractsProposal:
			for _, contractCall := range c.ContractCalls {
				sender := nexus.CrossChainAddress{Chain: exported.Axelarnet, Address: accountK.GetModuleAddress(govtypes.ModuleName).String()}

				destChain, ok := nexusK.GetChain(ctx, contractCall.Chain)
				if !ok {
					// Try forwarding it to wasm router if destination chain is not registered
					// Wasm chain names are always lower case, so normalize it for consistency in core
					destChainName := nexus.ChainName(strings.ToLower(contractCall.Chain.String()))
					destChain = nexus.Chain{Name: destChainName, SupportsForeignAssets: false, KeyType: tss.None, Module: wasm.ModuleName}
				}
				recipient := nexus.CrossChainAddress{Chain: destChain, Address: contractCall.ContractAddress}

				// axelar gateway expects keccak256 hashes for payloads
				payloadHash := crypto.Keccak256(contractCall.Payload)
				msgID, txID, nonce := nexusK.GenerateMessageID(ctx)
				msg := nexus.NewGeneralMessage(msgID, sender, recipient, payloadHash, txID, nonce, nil)

				events.Emit(ctx, &types.ContractCallSubmitted{
					MessageID:        msg.ID,
					Sender:           msg.GetSourceAddress(),
					SourceChain:      msg.GetSourceChain(),
					DestinationChain: msg.GetDestinationChain(),
					ContractAddress:  msg.GetDestinationAddress(),
					PayloadHash:      msg.PayloadHash,
					Payload:          contractCall.Payload,
				})

				if err := nexusK.SetNewMessage(ctx, msg); err != nil {
					return errorsmod.Wrap(err, "failed to add general message")
				}

				k.Logger(ctx).Debug(fmt.Sprintf("successfully enqueued contract call for contract address %s on chain %s from sender %s with message id %s", recipient.Address, recipient.Chain.String(), sender.Address, msg.ID),
					types.AttributeKeyDestinationChain, recipient.Chain.String(),
					types.AttributeKeyDestinationAddress, recipient.Address,
					types.AttributeKeySourceAddress, sender.Address,
					types.AttributeKeyMessageID, msg.ID,
					types.AttributeKeyPayloadHash, hex.EncodeToString(payloadHash),
				)
			}

			return nil
		default:
			return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized axelarnet proposal content type: %T", c)
		}
	}
}
