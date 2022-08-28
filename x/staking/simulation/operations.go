package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Simulation operation weights constants
const (
	OpWeightMsgCreateValidator = "op_weight_msg_create_validator"
	OpWeightMsgEditValidator   = "op_weight_msg_edit_validator"
	OpWeightMsgDelegate        = "op_weight_msg_delegate"
	OpWeightMsgUndelegate      = "op_weight_msg_undelegate"
	OpWeightMsgBeginRedelegate = "op_weight_msg_begin_redelegate"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak types.AccountKeeper,
	bk types.BankKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgCreateValidator int
		weightMsgEditValidator   int
		weightMsgDelegate        int
		weightMsgUndelegate      int
		weightMsgBeginRedelegate int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgCreateValidator, &weightMsgCreateValidator, nil,
		func(_ *rand.Rand) {
			weightMsgCreateValidator = simappparams.DefaultWeightMsgCreateValidator
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgEditValidator, &weightMsgEditValidator, nil,
		func(_ *rand.Rand) {
			weightMsgEditValidator = simappparams.DefaultWeightMsgEditValidator
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgDelegate, &weightMsgDelegate, nil,
		func(_ *rand.Rand) {
			weightMsgDelegate = simappparams.DefaultWeightMsgDelegate
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgUndelegate, &weightMsgUndelegate, nil,
		func(_ *rand.Rand) {
			weightMsgUndelegate = simappparams.DefaultWeightMsgUndelegate
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgBeginRedelegate, &weightMsgBeginRedelegate, nil,
		func(_ *rand.Rand) {
			weightMsgBeginRedelegate = simappparams.DefaultWeightMsgBeginRedelegate
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgCreateValidator,
			SimulateMsgCreateValidator(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgEditValidator,
			SimulateMsgEditValidator(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgDelegate,
			SimulateMsgDelegate(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgUndelegate,
			SimulateMsgUndelegate(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgBeginRedelegate,
			SimulateMsgBeginRedelegate(ak, bk, k),
		),
	}
}

// SimulateMsgCreateValidator generates a MsgCreateValidator with random values
func SimulateMsgCreateValidator(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, sdk.Result, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		address := sdk.ValAddress(simAccount.Address)

		// ensure the validator doesn't exist already
		_, found := k.GetValidator(ctx, address)
		if found {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateValidator, "unable to find validator"), nil, sdk.Result{}, nil
		}

		denom := k.GetParams(ctx).BondDenom

		balance := bk.GetBalance(ctx, simAccount.Address, denom).Amount
		if !balance.IsPositive() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateValidator, "balance is negative"), nil, sdk.Result{}, nil
		}

		amount, err := simtypes.RandPositiveInt(r, balance)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateValidator, "unable to generate positive amount"), nil, sdk.Result{}, err
		}

		selfDelegation := sdk.NewCoin(denom, amount)

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		var fees sdk.Coins

		coins, hasNeg := spendable.SafeSub(sdk.Coins{selfDelegation})
		if !hasNeg {
			feeCoins := coins.FilterDenoms([]string{sdk.DefaultBondDenom})
			fees, err = simtypes.RandomFees(r, ctx, feeCoins)
			if err != nil {
				return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateValidator, "unable to generate fees"), nil, sdk.Result{}, err
			}
		}

		description := types.NewDescription(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
		)

		maxCommission := sdk.NewDecWithPrec(int64(simtypes.RandIntBetween(r, 0, 100)), 2)
		commission := types.NewCommissionRates(
			simtypes.RandomDecAmount(r, maxCommission),
			maxCommission,
			simtypes.RandomDecAmount(r, maxCommission),
		)

		msg, err := types.NewMsgCreateValidator(address, simAccount.ConsKey.PubKey(), selfDelegation, description, commission, sdk.OneInt())
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to create CreateValidator message"), nil, sdk.Result{}, err
		}

		txCtx := simulation.OperationInput{
			App:           app,
			TxGen:         simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:           nil,
			Msg:           msg,
			MsgType:       msg.Type(),
			Context:       ctx,
			SimAccount:    simAccount,
			AccountKeeper: ak,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTx(txCtx, fees)
	}
}

// SimulateMsgEditValidator generates a MsgEditValidator with random values
func SimulateMsgEditValidator(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, sdk.Result, error) {
		if len(k.GetAllValidators(ctx)) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgEditValidator, "number of validators equal zero"), nil, sdk.Result{}, nil
		}

		val, ok := keeper.RandomValidator(r, k, ctx)
		if !ok {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgEditValidator, "unable to pick a validator"), nil, sdk.Result{}, nil
		}

		address := val.GetOperator()

		newCommissionRate := simtypes.RandomDecAmount(r, val.Commission.MaxRate)

		if err := val.Commission.ValidateNewRate(newCommissionRate, ctx.BlockHeader().Time); err != nil {
			// skip as the commission is invalid
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgEditValidator, "invalid commission rate"), nil, sdk.Result{}, nil
		}

		simAccount, found := simtypes.FindAccount(accs, sdk.AccAddress(val.GetOperator()))
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgEditValidator, "unable to find account"), nil, sdk.Result{}, fmt.Errorf("validator %s not found", val.GetOperator())
		}

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		description := types.NewDescription(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
		)

		msg := types.NewMsgEditValidator(address, description, &newCommissionRate, nil)

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: spendable,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgDelegate generates a MsgDelegate with random values
func SimulateMsgDelegate(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, sdk.Result, error) {
		denom := k.GetParams(ctx).BondDenom

		if len(k.GetAllValidators(ctx)) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDelegate, "number of validators equal zero"), nil, sdk.Result{}, nil
		}

		simAccount, _ := simtypes.RandomAcc(r, accs)
		val, ok := keeper.RandomValidator(r, k, ctx)
		if !ok {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDelegate, "unable to pick a validator"), nil, sdk.Result{}, nil
		}

		if val.InvalidExRate() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDelegate, "validator's invalid echange rate"), nil, sdk.Result{}, nil
		}

		amount := bk.GetBalance(ctx, simAccount.Address, denom).Amount
		if !amount.IsPositive() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDelegate, "balance is negative"), nil, sdk.Result{}, nil
		}

		amount, err := simtypes.RandPositiveInt(r, amount)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDelegate, "unable to generate positive amount"), nil, sdk.Result{}, err
		}

		bondAmt := sdk.NewCoin(denom, amount)

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		var fees sdk.Coins

		coins, hasNeg := spendable.SafeSub(sdk.Coins{bondAmt})
		if !hasNeg {
			feeCoins := coins.FilterDenoms([]string{sdk.DefaultBondDenom})
			fees, err = simtypes.RandomFees(r, ctx, feeCoins)
			if err != nil {
				return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDelegate, "unable to generate fees"), nil, sdk.Result{}, err
			}
		}

		msg := types.NewMsgDelegate(simAccount.Address, val.GetOperator(), bondAmt)

		txCtx := simulation.OperationInput{
			App:           app,
			TxGen:         simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:           nil,
			Msg:           msg,
			MsgType:       msg.Type(),
			Context:       ctx,
			SimAccount:    simAccount,
			AccountKeeper: ak,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTx(txCtx, fees)
	}
}

// SimulateMsgUndelegate generates a MsgUndelegate with random values
func SimulateMsgUndelegate(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, sdk.Result, error) {
		// get random validator
		validator, ok := keeper.RandomValidator(r, k, ctx)
		if !ok {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUndelegate, "validator is not ok"), nil, sdk.Result{}, nil
		}

		valAddr := validator.GetOperator()
		delegations := k.GetValidatorDelegations(ctx, validator.GetOperator())
		if delegations == nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUndelegate, "keeper does have any delegation entries"), nil, sdk.Result{}, nil
		}

		// get random delegator from validator
		delegation := delegations[r.Intn(len(delegations))]
		delAddr := delegation.GetDelegatorAddr()

		if k.HasMaxUnbondingDelegationEntries(ctx, delAddr, valAddr) {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUndelegate, "keeper does have a max unbonding delegation entries"), nil, sdk.Result{}, nil
		}

		totalBond := validator.TokensFromShares(delegation.GetShares()).TruncateInt()
		if !totalBond.IsPositive() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUndelegate, "total bond is negative"), nil, sdk.Result{}, nil
		}

		unbondAmt, err := simtypes.RandPositiveInt(r, totalBond)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUndelegate, "invalid unbond amount"), nil, sdk.Result{}, err
		}

		if unbondAmt.IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUndelegate, "unbond amount is zero"), nil, sdk.Result{}, nil
		}

		msg := types.NewMsgUndelegate(
			delAddr, valAddr, sdk.NewCoin(k.BondDenom(ctx), unbondAmt),
		)

		// need to retrieve the simulation account associated with delegation to retrieve PrivKey
		var simAccount simtypes.Account

		for _, simAcc := range accs {
			if simAcc.Address.Equals(delAddr) {
				simAccount = simAcc
				break
			}
		}
		// if simaccount.PrivKey == nil, delegation address does not exist in accs. Return error
		if simAccount.PrivKey == nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "account private key is nil"), nil, sdk.Result{}, fmt.Errorf("delegation addr: %s does not exist in simulation accounts", delAddr)
		}

		account := ak.GetAccount(ctx, delAddr)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: spendable,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgBeginRedelegate generates a MsgBeginRedelegate with random values
func SimulateMsgBeginRedelegate(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, sdk.Result, error) {
		// get random source validator
		srcVal, ok := keeper.RandomValidator(r, k, ctx)
		if !ok {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBeginRedelegate, "unable to pick validator"), nil, sdk.Result{}, nil
		}

		srcAddr := srcVal.GetOperator()
		delegations := k.GetValidatorDelegations(ctx, srcAddr)
		if delegations == nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBeginRedelegate, "keeper does have any delegation entries"), nil, sdk.Result{}, nil
		}

		// get random delegator from src validator
		delegation := delegations[r.Intn(len(delegations))]
		delAddr := delegation.GetDelegatorAddr()

		if k.HasReceivingRedelegation(ctx, delAddr, srcAddr) {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBeginRedelegate, "receveing redelegation is not allowed"), nil, sdk.Result{}, nil // skip
		}

		// get random destination validator
		destVal, ok := keeper.RandomValidator(r, k, ctx)
		if !ok {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBeginRedelegate, "unable to pick validator"), nil, sdk.Result{}, nil
		}

		destAddr := destVal.GetOperator()
		if srcAddr.Equals(destAddr) || destVal.InvalidExRate() || k.HasMaxRedelegationEntries(ctx, delAddr, srcAddr, destAddr) {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBeginRedelegate, "checks failed"), nil, sdk.Result{}, nil
		}

		totalBond := srcVal.TokensFromShares(delegation.GetShares()).TruncateInt()
		if !totalBond.IsPositive() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBeginRedelegate, "total bond is negative"), nil, sdk.Result{}, nil
		}

		redAmt, err := simtypes.RandPositiveInt(r, totalBond)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBeginRedelegate, "unable to generate positive amount"), nil, sdk.Result{}, err
		}

		if redAmt.IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBeginRedelegate, "amount is zero"), nil, sdk.Result{}, nil
		}

		// check if the shares truncate to zero
		shares, err := srcVal.SharesFromTokens(redAmt)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBeginRedelegate, "invalid shares"), nil, sdk.Result{}, err
		}

		if srcVal.TokensFromShares(shares).TruncateInt().IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBeginRedelegate, "shares truncate to zero"), nil, sdk.Result{}, nil // skip
		}

		// need to retrieve the simulation account associated with delegation to retrieve PrivKey
		var simAccount simtypes.Account

		for _, simAcc := range accs {
			if simAcc.Address.Equals(delAddr) {
				simAccount = simAcc
				break
			}
		}

		// if simaccount.PrivKey == nil, delegation address does not exist in accs. Return error
		if simAccount.PrivKey == nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBeginRedelegate, "account private key is nil"), nil, sdk.Result{}, fmt.Errorf("delegation addr: %s does not exist in simulation accounts", delAddr)
		}

		account := ak.GetAccount(ctx, delAddr)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		msg := types.NewMsgBeginRedelegate(
			delAddr, srcAddr, destAddr,
			sdk.NewCoin(k.BondDenom(ctx), redAmt),
		)

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: spendable,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
