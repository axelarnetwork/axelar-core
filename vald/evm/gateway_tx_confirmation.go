package evm

import (
	"bytes"

	errorsmod "cosmossdk.io/errors"
	geth "github.com/ethereum/go-ethereum/core/types"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// processGatewayTxLogs extracts events from gateway transaction logs
func (mgr Mgr) processGatewayTxLogs(chain nexus.ChainName, gatewayAddress types.Address, logs []*geth.Log) []types.Event {
	var events []types.Event
	for i, txlog := range logs {
		if !bytes.Equal(gatewayAddress.Bytes(), txlog.Address.Bytes()) {
			continue
		}

		if len(txlog.Topics) == 0 {
			continue
		}

		switch txlog.Topics[0] {
		case ContractCallSig:
			gatewayEvent, err := DecodeEventContractCall(txlog)
			if err != nil {
				mgr.logger().Debug(errorsmod.Wrap(err, "decode event ContractCall failed").Error())
				continue
			}

			if err := gatewayEvent.ValidateBasic(); err != nil {
				mgr.logger().Debug(errorsmod.Wrap(err, "invalid event ContractCall").Error())
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
				mgr.logger().Debug(errorsmod.Wrap(err, "decode event ContractCallWithToken failed").Error())
				continue
			}

			if err := gatewayEvent.ValidateBasic(); err != nil {
				mgr.logger().Debug(errorsmod.Wrap(err, "invalid event ContractCallWithToken").Error())
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
				mgr.logger().Debug(errorsmod.Wrap(err, "decode event TokenSent failed").Error())
			}

			if err := gatewayEvent.ValidateBasic(); err != nil {
				mgr.logger().Debug(errorsmod.Wrap(err, "invalid event TokenSent").Error())
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
