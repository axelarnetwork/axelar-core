package types

import (
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgTrackAddress{}, "btcbridge/TrackAddress", nil)
	cdc.RegisterConcrete(MsgVerifyTx{}, "btcbridge/VerifyTx", nil)
	cdc.RegisterConcrete(MsgWithdraw{}, "btcbridge/Withdraw", nil)
	cdc.RegisterConcrete(MsgRawTx{}, "btcbridge/RawTx", nil)
	cdc.RegisterConcrete(MsgTrackAddressFromPubKey{}, "btcbridge/MsgTrackAddressFromPubKey", nil)

	cdc.RegisterInterface((*btcutil.Address)(nil), nil)
	cdc.RegisterConcrete(btcutil.AddressPubKey{}, "btcutil/AddressPubkey", nil)
	cdc.RegisterConcrete(btcutil.AddressPubKeyHash{}, "btcutil/AddressPubkeyHash", nil)
	cdc.RegisterConcrete(btcutil.AddressWitnessPubKeyHash{}, "btcutil/AddressWitnessPubKeyHash", nil)
	cdc.RegisterConcrete(btcutil.AddressScriptHash{}, "btcutil/AddressScriptHash", nil)
	cdc.RegisterConcrete(btcutil.AddressWitnessScriptHash{}, "btcutil/AddressWitnessScriptHash", nil)
}

// ModuleCdc defines the module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
