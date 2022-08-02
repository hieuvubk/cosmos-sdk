package v4

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// If expedited, the deposit to enter voting period will be
// increased to 5000 OSMO. The proposal will have 24 hours to achieve
// a two-thirds majority of all staked OSMO voting power voting YES.
var minInitialDepositRatio = sdk.NewDec(25).Quo(sdk.NewDec(100))
var minExpeditedDeposit = sdk.NewCoins(sdk.NewCoin("osmo", sdk.NewInt(5000)))
var expeditedVotingPeriod = time.Duration(time.Hour * 24)
var expeditedThreshold = sdk.NewDec(2).Quo(sdk.NewDec(3))

// MigrateStore performs in-place store migrations for consensus version 3
// in the gov module.
// Please note that this is the first version that switches from using
// SDK versioning (v043 etc) for package names to consensus versioning
// of the gov module.
// The migration includes:
//
// - Setting the minimum deposit param in the paramstore.
func MigrateStore(ctx sdk.Context, paramstore types.ParamSubspace) error {
	migrateParamsStore(ctx, paramstore)
	return nil
}

func migrateParamsStore(ctx sdk.Context, paramstore types.ParamSubspace) {
	var depositParams types.DepositParams
	var votingParams types.VotingParams
	var tallyParams types.TallyParams

	//Set depositParams
	paramstore.Get(ctx, types.ParamStoreKeyDepositParams, &depositParams)
	depositParams.MinInitialDepositRatio = minInitialDepositRatio
	depositParams.MinExpeditedDeposit = minExpeditedDeposit
	paramstore.Set(ctx, types.ParamStoreKeyDepositParams, depositParams)

	//Set votingParams
	paramstore.Get(ctx, types.ParamStoreKeyVotingParams, &votingParams)
	votingParams.ExpeditedVotingPeriod = expeditedVotingPeriod
	paramstore.Set(ctx, types.ParamStoreKeyVotingParams, votingParams)

	//Set tallyParams
	paramstore.Get(ctx, types.ParamStoreKeyTallyParams, &votingParams)
	tallyParams.ExpeditedThreshold = expeditedThreshold
	paramstore.Set(ctx, types.ParamStoreKeyTallyParams, votingParams)
}