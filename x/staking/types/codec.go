package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"
)

// RegisterLegacyAminoCodec registers the necessary x/staking interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateValidator{}, "cosmos-sdk/MsgCreateValidator", nil)
	cdc.RegisterConcrete(&MsgEditValidator{}, "cosmos-sdk/MsgEditValidator", nil)
	cdc.RegisterConcrete(&MsgDelegate{}, "cosmos-sdk/MsgDelegate", nil)
	cdc.RegisterConcrete(&MsgUndelegate{}, "cosmos-sdk/MsgUndelegate", nil)
	cdc.RegisterConcrete(&MsgBeginRedelegate{}, "cosmos-sdk/MsgBeginRedelegate", nil)
}

// RegisterInterfaces registers the x/staking interfaces types with the interface registry
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateValidator{},
		&MsgEditValidator{},
		&MsgDelegate{},
		&MsgUndelegate{},
		&MsgBeginRedelegate{},
	)
	registry.RegisterImplementations(
		(*authz.Authorization)(nil),
		&StakeAuthorization{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	sdk.RegisterLegacyAminoCodec(amino)

	// Register all Amino interfaces and concrete types on the authz Amino codec so that this can later be
	// used to properly serialize MsgGrant and MsgExec instances
	RegisterLegacyAminoCodec(authzcodec.Amino)
}
