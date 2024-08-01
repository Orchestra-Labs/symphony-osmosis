package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	appparams "github.com/osmosis-labs/osmosis/v23/app/params"

	"github.com/osmosis-labs/osmosis/v23/x/market/types"
)

// Keeper of the market store
type Keeper struct {
	storeKey   storetypes.StoreKey
	cdc        codec.BinaryCodec
	paramSpace paramstypes.Subspace

	AccountKeeper types.AccountKeeper
	BankKeeper    types.BankKeeper
	OracleKeeper  types.OracleKeeper
}

// NewKeeper constructs a new keeper for oracle
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	paramstore paramstypes.Subspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	oracleKeeper types.OracleKeeper,
) Keeper {
	// ensure market module account is set
	if addr := accountKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// ensure reserve market module account is set
	if addr := accountKeeper.GetModuleAddress(types.ReserveModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ReserveModuleName))
	}

	// set KeyTable if it has not already been set
	if !paramstore.HasKeyTable() {
		paramstore = paramstore.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		paramSpace:    paramstore,
		AccountKeeper: accountKeeper,
		BankKeeper:    bankKeeper,
		OracleKeeper:  oracleKeeper,
	}
}

func (k Keeper) GetExchangePoolBalance(ctx sdk.Context) sdk.Coin {
	account := k.GetMarketAccount(ctx)
	if account == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	return k.BankKeeper.GetBalance(ctx, account.GetAddress(), appparams.BaseCoinUnit)
}

func (k Keeper) GetReservePoolBalance(ctx sdk.Context) sdk.Coin {
	account := k.GetReserveMarketAccount(ctx)
	if account == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	return k.BankKeeper.GetBalance(ctx, account.GetAddress(), appparams.BaseCoinUnit)
}
