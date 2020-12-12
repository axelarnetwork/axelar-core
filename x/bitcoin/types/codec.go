package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgTrackAddress{}, "bitcoin/TrackAddress", nil)
	cdc.RegisterConcrete(MsgTrackPubKey{}, "bitcoin/MsgTrackPubKey", nil)
	cdc.RegisterConcrete(MsgVerifyTx{}, "bitcoin/VerifyTx", nil)
	cdc.RegisterConcrete(MsgRawTx{}, "bitcoin/RawTx", nil)
	cdc.RegisterConcrete(MsgWithdraw{}, "bitcoin/Withdraw", nil)
	cdc.RegisterConcrete(&MsgVoteVerifiedTx{}, "bitcoin/MsgVoteVerifiedTx", nil)
	cdc.RegisterConcrete(MsgTransferToNewMasterKey{}, "bitcoin/MsgTransferToNewMasterKey", nil)
	cdc.RegisterConcrete(MsgRawTxForMasterKey{}, "bitcoin/MsgRawTxForMasterKey", nil)
}

// ModuleCdc defines the module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
