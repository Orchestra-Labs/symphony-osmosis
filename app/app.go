package app

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibcwasmkeeper "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/keeper"
	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	ibc "github.com/cosmos/ibc-go/v7/modules/core"
	markettypes "github.com/osmosis-labs/osmosis/v23/x/market/types"
	oracletypes "github.com/osmosis-labs/osmosis/v23/x/oracle/types"

	//oracletypes "github.com/osmosis-labs/osmosis/v23/x/oracle/types"

	"github.com/osmosis-labs/osmosis/v23/ingest/sqs"
	"github.com/osmosis-labs/osmosis/v23/ingest/sqs/domain"

	"github.com/osmosis-labs/osmosis/osmoutils"

	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"

	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"

	"github.com/CosmWasm/wasmd/x/wasm"
	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/libs/bytes"
	tmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/libs/log"
	tmos "github.com/cometbft/cometbft/libs/os"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cast"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/x/crisis"

	protorevtypes "github.com/osmosis-labs/osmosis/v23/x/protorev/types"

	"github.com/osmosis-labs/osmosis/v23/app/keepers"
	"github.com/osmosis-labs/osmosis/v23/app/upgrades"
	v10 "github.com/osmosis-labs/osmosis/v23/app/upgrades/v10"
	v11 "github.com/osmosis-labs/osmosis/v23/app/upgrades/v11"
	v12 "github.com/osmosis-labs/osmosis/v23/app/upgrades/v12"
	v13 "github.com/osmosis-labs/osmosis/v23/app/upgrades/v13"
	v14 "github.com/osmosis-labs/osmosis/v23/app/upgrades/v14"
	v15 "github.com/osmosis-labs/osmosis/v23/app/upgrades/v15"
	v16 "github.com/osmosis-labs/osmosis/v23/app/upgrades/v16"
	v17 "github.com/osmosis-labs/osmosis/v23/app/upgrades/v17"
	v18 "github.com/osmosis-labs/osmosis/v23/app/upgrades/v18"
	v19 "github.com/osmosis-labs/osmosis/v23/app/upgrades/v19"
	v20 "github.com/osmosis-labs/osmosis/v23/app/upgrades/v20"
	v21 "github.com/osmosis-labs/osmosis/v23/app/upgrades/v21"
	v22 "github.com/osmosis-labs/osmosis/v23/app/upgrades/v22"
	v23 "github.com/osmosis-labs/osmosis/v23/app/upgrades/v23"
	v24 "github.com/osmosis-labs/osmosis/v23/app/upgrades/v24"
	v3 "github.com/osmosis-labs/osmosis/v23/app/upgrades/v3"
	v4 "github.com/osmosis-labs/osmosis/v23/app/upgrades/v4"
	v5 "github.com/osmosis-labs/osmosis/v23/app/upgrades/v5"
	v6 "github.com/osmosis-labs/osmosis/v23/app/upgrades/v6"
	v7 "github.com/osmosis-labs/osmosis/v23/app/upgrades/v7"
	v8 "github.com/osmosis-labs/osmosis/v23/app/upgrades/v8"
	v9 "github.com/osmosis-labs/osmosis/v23/app/upgrades/v9"
	_ "github.com/osmosis-labs/osmosis/v23/client/docs/statik"
	"github.com/osmosis-labs/osmosis/v23/ingest"
	"github.com/osmosis-labs/osmosis/v23/x/mint"
)

const appName = "SymphonyApp"

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(keepers.AppModuleBasics...)

	// module account permissions
	maccPerms = moduleAccountPermissions

	// module accounts that are allowed to receive tokens.
	allowedReceivingModAcc = map[string]bool{
		protorevtypes.ModuleName:      true,
		oracletypes.ModuleName:        true,
		markettypes.ModuleName:        true,
		markettypes.ReserveModuleName: true,
	}

	// TODO: Refactor wasm items into a wasm.go file
	// WasmProposalsEnabled enables all x/wasm proposals when it's value is "true"
	// and EnableSpecificWasmProposals is empty. Otherwise, all x/wasm proposals
	// are disabled.
	WasmProposalsEnabled = "true"

	// EnableSpecificWasmProposals, if set, must be comma-separated list of values
	// that are all a subset of "EnableAllProposals", which takes precedence over
	// WasmProposalsEnabled.
	//
	// See: https://github.com/CosmWasm/wasmd/blob/02a54d33ff2c064f3539ae12d75d027d9c665f05/x/wasm/internal/types/proposal.go#L28-L34
	EnableSpecificWasmProposals = ""

	// EmptyWasmOpts defines a type alias for a list of wasm options.
	EmptyWasmOpts []wasmkeeper.Option

	_ runtime.AppI = (*SymphonyApp)(nil)

	Upgrades = []upgrades.Upgrade{v4.Upgrade, v5.Upgrade, v7.Upgrade, v9.Upgrade, v11.Upgrade, v12.Upgrade, v13.Upgrade, v14.Upgrade, v15.Upgrade, v16.Upgrade, v17.Upgrade, v18.Upgrade, v19.Upgrade, v20.Upgrade, v21.Upgrade, v22.Upgrade, v23.Upgrade, v24.Upgrade}
	Forks    = []upgrades.Fork{v3.Fork, v6.Fork, v8.Fork, v10.Fork}
)

// SymphonyApp extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type SymphonyApp struct {
	*baseapp.BaseApp
	keepers.AppKeepers

	cdc               *codec.LegacyAmino
	appCodec          codec.Codec
	interfaceRegistry types.InterfaceRegistry
	invCheckPeriod    uint

	mm           *module.Manager
	sm           *module.SimulationManager
	configurator module.Configurator
	homePath     string
}

// init sets DefaultNodeHome to default symphonyd install location.
func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, ".symphonyd")
}

// initReusablePackageInjections injects data available within symphony into the reusable packages.
// This is done to ensure they can be built without depending on at compilation time and thus imported by other chains
// This should always be called before any other function to avoid inconsistent data
func initReusablePackageInjections() {
	// Inject ClawbackVestingAccount account type into osmoutils
	osmoutils.OsmoUtilsExtraAccountTypes = map[reflect.Type]struct{}{
		reflect.TypeOf(&vestingtypes.ClawbackVestingAccount{}): {},
	}
}

// overrideWasmVariables overrides the wasm variables to:
//   - allow for larger wasm files
func overrideWasmVariables() {
	// Override Wasm size limitation from WASMD.
	wasmtypes.MaxWasmSize = 3 * 1024 * 1024
	wasmtypes.MaxProposalWasmSize = wasmtypes.MaxWasmSize
}

// NewSymphonyApp returns a reference to an initialized Symphony.
func NewSymphonyApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	appOpts servertypes.AppOptions,
	wasmOpts []wasmkeeper.Option,
	baseAppOptions ...func(*baseapp.BaseApp),
) *SymphonyApp {
	initReusablePackageInjections() // This should run before anything else to make sure the variables are properly initialized
	overrideWasmVariables()
	encodingConfig := GetEncodingConfig()
	appCodec := encodingConfig.Marshaler
	cdc := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry
	txConfig := encodingConfig.TxConfig

	bApp := baseapp.NewBaseApp(appName, logger, db, txConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	app := &SymphonyApp{
		AppKeepers:        keepers.AppKeepers{},
		BaseApp:           bApp,
		cdc:               cdc,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
	}

	app.homePath = homePath
	dataDir := filepath.Join(homePath, "data")
	wasmDir := filepath.Join(homePath, "wasm")
	ibcWasmConfig :=
		ibcwasmtypes.WasmConfig{
			DataDir:               filepath.Join(homePath, "ibc_08-wasm"),
			SupportedCapabilities: "iterator",
			ContractDebugMode:     false,
		}
	wasmConfig, err := wasm.ReadWasmConfig(appOpts)
	// Uncomment this for debugging contracts. In the future this could be made into a param passed by the tests
	// wasmConfig.ContractDebugMode = true
	if err != nil {
		panic(fmt.Sprintf("error while reading wasm config: %s", err))
	}
	app.InitSpecialKeepers(
		appCodec,
		bApp,
		wasmDir,
		cdc,
		invCheckPeriod,
		skipUpgradeHeights,
		homePath,
	)
	app.setupUpgradeStoreLoaders()
	app.InitNormalKeepers(
		appCodec,
		encodingConfig,
		bApp,
		maccPerms,
		dataDir,
		wasmDir,
		wasmConfig,
		wasmOpts,
		app.BlockedAddrs(),
		ibcWasmConfig,
	)

	// Initialize the ingest manager for propagating data to external sinks.
	app.IngestManager = ingest.NewIngestManager()

	sqsConfig := sqs.NewConfigFromOptions(appOpts)

	// Initialize the SQS ingester if it is enabled.
	if sqsConfig.IsEnabled {
		sqsKeepers := domain.SQSIngestKeepers{
			GammKeeper:         app.GAMMKeeper,
			CosmWasmPoolKeeper: app.CosmwasmPoolKeeper,
			BankKeeper:         app.BankKeeper,
			ProtorevKeeper:     app.ProtoRevKeeper,
			PoolManagerKeeper:  app.PoolManagerKeeper,
			ConcentratedKeeper: app.ConcentratedLiquidityKeeper,
		}

		sqsIngester, err := sqsConfig.Initialize(appCodec, sqsKeepers)
		if err != nil {
			panic(err)
		}

		// Set the sidecar query server ingester to the ingest manager.
		app.IngestManager.RegisterIngester(sqsIngester)
	}

	// TODO: There is a bug here, where we register the govRouter routes in InitNormalKeepers and then
	// call setupHooks afterwards. Therefore, if a gov proposal needs to call a method and that method calls a
	// hook, we will get a nil pointer dereference error due to the hooks in the keeper not being
	// setup yet. I will refrain from creating an issue in the sdk for now until after we unfork to 0.47,
	// because I believe the concept of Routes is going away.
	// https://github.com/osmosis-labs/osmosis/issues/6580
	app.SetupHooks()

	/****  Module Options ****/

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: All module / keeper changes should happen prior to this module.NewManager line being called.
	// However in the event any changes do need to happen after this call, ensure that that keeper
	// is only passed in its keeper form (not de-ref'd anywhere)
	//
	// Generally NewAppModule will require the keeper that module defines to be passed in as an exact struct,
	// but should take in every other keeper as long as it matches a certain interface. (So no need to be de-ref'd)
	//
	// Any time a module requires a keeper de-ref'd that's not its native one,
	// its code-smell and should probably change. We should get the staking keeper dependencies fixed.
	app.mm = module.NewManager(appModules(app, encodingConfig, skipGenesisInvariants)...)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	// NOTE: capability module's beginblocker must come before any modules using capabilities (e.g. IBC)

	// Tell the app's module manager how to set the order of BeginBlockers, which are run at the beginning of every block.
	app.mm.SetOrderBeginBlockers(orderBeginBlockers(app.mm.ModuleNames())...)

	// Tell the app's module manager how to set the order of EndBlockers, which are run at the end of every block.
	app.mm.SetOrderEndBlockers(OrderEndBlockers(app.mm.ModuleNames())...)

	app.mm.SetOrderInitGenesis(OrderInitGenesis(app.mm.ModuleNames())...)

	app.mm.RegisterInvariants(app.CrisisKeeper)

	app.configurator = module.NewConfigurator(app.AppCodec(), app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(app.configurator)

	app.setupUpgradeHandlers()

	app.sm = module.NewSimulationManager(
		auth.NewAppModule(appCodec, *app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
		bank.NewAppModule(appCodec, *app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		authzmodule.NewAppModule(appCodec, *app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(govtypes.ModuleName)),
		mint.NewAppModule(appCodec, *app.MintKeeper, app.AccountKeeper, app.BankKeeper),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName)),
		distr.NewAppModule(appCodec, *app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(distrtypes.ModuleName)),
		slashing.NewAppModule(appCodec, *app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(slashingtypes.ModuleName)),
		params.NewAppModule(*app.ParamsKeeper),
		evidence.NewAppModule(*app.EvidenceKeeper),
		wasm.NewAppModule(appCodec, app.WasmKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.MsgServiceRouter(), app.GetSubspace(wasmtypes.ModuleName)),
		ibc.NewAppModule(app.IBCKeeper),
		transfer.NewAppModule(*app.TransferKeeper),
	)

	app.sm.RegisterStoreDecoders()

	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.mm.Modules))

	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(err)
	}
	reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)

	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(app.GetKVStoreKey())
	app.MountTransientStores(app.GetTransientStoreKey())
	app.MountMemoryStores(app.GetMemoryStoreKey())

	anteHandler := NewAnteHandler(
		appOpts,
		wasmConfig,
		app.GetKey(wasmtypes.StoreKey),
		app.AccountKeeper,
		app.BankKeeper,
		app.TxFeesKeeper,
		app.GAMMKeeper,
		ante.DefaultSigVerificationGasConsumer,
		encodingConfig.TxConfig.SignModeHandler(),
		app.IBCKeeper,
	)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetAnteHandler(anteHandler)
	app.SetPostHandler(NewPostHandler(app.ProtoRevKeeper))
	app.SetEndBlocker(app.EndBlocker)

	// Register snapshot extensions to enable state-sync for wasm.
	if manager := app.SnapshotManager(); manager != nil {
		err := manager.RegisterExtensions(
			wasmkeeper.NewWasmSnapshotter(app.CommitMultiStore(), app.WasmKeeper),
		)
		if err != nil {
			panic(fmt.Errorf("failed to register snapshot extension: %s", err))
		}

		err = manager.RegisterExtensions(
			ibcwasmkeeper.NewWasmSnapshotter(app.CommitMultiStore(), app.IBCWasmClientKeeper),
		)
		if err != nil {
			panic(fmt.Errorf("failed to register snapshot extension: %s", err))
		}
	}

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}

		ctx := app.BaseApp.NewUncachedContext(true, tmproto.Header{})

		// Initialize pinned codes in wasmvm as they are not persisted there
		if err := app.WasmKeeper.InitializePinnedCodes(ctx); err != nil {
			tmos.Exit(fmt.Sprintf("failed initialize pinned codes %s", err))
		}

		if err := ibcwasmkeeper.InitializePinnedCodes(ctx, appCodec); err != nil {
			tmos.Exit(fmt.Sprintf("failed initialize pinned codes %s", err))
		}
	}

	return app
}

// InitSymphonyAppForTestnet is broken down into two sections:
// Required Changes: Changes that, if not made, will cause the testnet to halt or panic
// Optional Changes: Changes to customize the testnet to one's liking (lower vote times, fund accounts, etc)
func InitSymphonyAppForTestnet(app *SymphonyApp, newValAddr bytes.HexBytes, newValPubKey crypto.PubKey, newOperatorAddress, upgradeToTrigger string) *SymphonyApp {
	//
	// Required Changes:
	//

	ctx := app.BaseApp.NewUncachedContext(true, tmproto.Header{})
	pubkey := &ed25519.PubKey{Key: newValPubKey.Bytes()}
	pubkeyAny, err := types.NewAnyWithValue(pubkey)
	if err != nil {
		tmos.Exit(err.Error())
	}

	// STAKING
	//

	// Create Validator struct for our new validator.
	_, bz, err := bech32.DecodeAndConvert(newOperatorAddress)
	if err != nil {
		tmos.Exit(err.Error())
	}
	bech32Addr, err := bech32.ConvertAndEncode("symphonyvaloper", bz)
	if err != nil {
		tmos.Exit(err.Error())
	}
	newVal := stakingtypes.Validator{
		OperatorAddress: bech32Addr,
		ConsensusPubkey: pubkeyAny,
		Jailed:          false,
		Status:          stakingtypes.Bonded,
		Tokens:          sdk.NewInt(900000000000000),
		DelegatorShares: sdk.MustNewDecFromStr("10000000"),
		Description: stakingtypes.Description{
			Moniker: "Testnet Validator",
		},
		Commission: stakingtypes.Commission{
			CommissionRates: stakingtypes.CommissionRates{
				Rate:          sdk.MustNewDecFromStr("0.05"),
				MaxRate:       sdk.MustNewDecFromStr("0.1"),
				MaxChangeRate: sdk.MustNewDecFromStr("0.05"),
			},
		},
		MinSelfDelegation: sdk.OneInt(),
	}

	// Remove all validators from power store
	stakingKey := app.GetKey(stakingtypes.ModuleName)
	stakingStore := ctx.KVStore(stakingKey)
	iterator := app.StakingKeeper.ValidatorsPowerStoreIterator(ctx)
	for ; iterator.Valid(); iterator.Next() {
		stakingStore.Delete(iterator.Key())
	}
	iterator.Close()

	// Remove all valdiators from last validators store
	iterator = app.StakingKeeper.LastValidatorsIterator(ctx)
	for ; iterator.Valid(); iterator.Next() {
		stakingStore.Delete(iterator.Key())
	}
	iterator.Close()

	// Remove all validators from validators store
	iterator = sdk.KVStorePrefixIterator(stakingStore, stakingtypes.ValidatorsKey)
	for ; iterator.Valid(); iterator.Next() {
		stakingStore.Delete(iterator.Key())
	}
	iterator.Close()

	// Remove all validators from unbonding queue
	iterator = sdk.KVStorePrefixIterator(stakingStore, stakingtypes.ValidatorQueueKey)
	for ; iterator.Valid(); iterator.Next() {
		stakingStore.Delete(iterator.Key())
	}
	iterator.Close()

	// Add our validator to power and last validators store
	app.StakingKeeper.SetValidator(ctx, newVal)
	err = app.StakingKeeper.SetValidatorByConsAddr(ctx, newVal)
	if err != nil {
		tmos.Exit(err.Error())
	}
	app.StakingKeeper.SetValidatorByPowerIndex(ctx, newVal)
	app.StakingKeeper.SetLastValidatorPower(ctx, newVal.GetOperator(), 0)
	if err := app.StakingKeeper.Hooks().AfterValidatorCreated(ctx, newVal.GetOperator()); err != nil {
		panic(err)
	}

	// DISTRIBUTION
	//

	// Initialize records for this validator across all distribution stores
	app.DistrKeeper.SetValidatorHistoricalRewards(ctx, newVal.GetOperator(), 0, distrtypes.NewValidatorHistoricalRewards(sdk.DecCoins{}, 1))
	app.DistrKeeper.SetValidatorCurrentRewards(ctx, newVal.GetOperator(), distrtypes.NewValidatorCurrentRewards(sdk.DecCoins{}, 1))
	app.DistrKeeper.SetValidatorAccumulatedCommission(ctx, newVal.GetOperator(), distrtypes.InitialValidatorAccumulatedCommission())
	app.DistrKeeper.SetValidatorOutstandingRewards(ctx, newVal.GetOperator(), distrtypes.ValidatorOutstandingRewards{Rewards: sdk.DecCoins{}})

	// SLASHING
	//

	// Set validator signing info for our new validator.
	newConsAddr := sdk.ConsAddress(newValAddr.Bytes())
	newValidatorSigningInfo := slashingtypes.ValidatorSigningInfo{
		Address:     newConsAddr.String(),
		StartHeight: app.LastBlockHeight() - 1,
		Tombstoned:  false,
	}
	app.SlashingKeeper.SetValidatorSigningInfo(ctx, newConsAddr, newValidatorSigningInfo)

	//
	// Optional Changes:
	//

	// GOV
	//

	newExpeditedVotingPeriod := time.Minute
	newVotingPeriod := time.Minute * 2

	govParams := app.GovKeeper.GetParams(ctx)
	govParams.ExpeditedVotingPeriod = &newExpeditedVotingPeriod
	govParams.VotingPeriod = &newVotingPeriod
	govParams.MinDeposit = sdk.NewCoins(sdk.NewInt64Coin("note", 100000000))
	govParams.ExpeditedMinDeposit = sdk.NewCoins(sdk.NewInt64Coin("note", 150000000))

	err = app.GovKeeper.SetParams(ctx, govParams)
	if err != nil {
		tmos.Exit(err.Error())
	}

	// EPOCHS
	//

	dayEpochInfo := app.EpochsKeeper.GetEpochInfo(ctx, "day")
	dayEpochInfo.Duration = time.Hour * 6
	// Prevents epochs from running back to back
	dayEpochInfo.CurrentEpochStartTime = time.Now().UTC()
	// If you want epoch to run a minute after starting the chain, uncomment the line below and comment the line above
	// dayEpochInfo.CurrentEpochStartTime = time.Now().UTC().Add(-dayEpochInfo.Duration).Add(time.Minute)
	dayEpochInfo.CurrentEpochStartHeight = app.LastBlockHeight()
	app.EpochsKeeper.DeleteEpochInfo(ctx, "day")
	err = app.EpochsKeeper.AddEpochInfo(ctx, dayEpochInfo)
	if err != nil {
		tmos.Exit(err.Error())
	}

	weekEpochInfo := app.EpochsKeeper.GetEpochInfo(ctx, "week")
	weekEpochInfo.Duration = time.Hour * 12
	// Prevents epochs from running back to back
	weekEpochInfo.CurrentEpochStartTime = time.Now().UTC()
	weekEpochInfo.CurrentEpochStartHeight = app.LastBlockHeight()
	app.EpochsKeeper.DeleteEpochInfo(ctx, "week")
	err = app.EpochsKeeper.AddEpochInfo(ctx, weekEpochInfo)
	if err != nil {
		tmos.Exit(err.Error())
	}

	// BANK
	//
	//
	//defaultCoins := sdk.NewCoins(
	//	sdk.NewInt64Coin("ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7", 1000000000000), // DAI
	//	sdk.NewInt64Coin("note", 1000000000000))
	//
	//localSymphonyAccounts := []sdk.AccAddress{
	//	sdk.MustAccAddressFromBech32("symphony1p7mp7r9f9f6sf2c95ht42ncm6ga96ha8xghdeg"),
	//	sdk.MustAccAddressFromBech32("symphony1c605nvcw94rvvehrcdfj85qe09ulseyt0efhk7"),
	//	sdk.MustAccAddressFromBech32("symphony1jpr5824frn5472qm73ckfe2c3rh6vrn4lvlgj7"),
	//	sdk.MustAccAddressFromBech32("symphony1amr6zrvs0hymf62qd5mwvshx94ul8cgfu9jtxn"),
	//	sdk.MustAccAddressFromBech32("symphony1egts9ayaqr6t54ahs62awmz5smuf764uu5f5xv"),
	//	sdk.MustAccAddressFromBech32("symphony1450weujlqvtd0d5z59v388jmzwyk3e6qhlj5r5"),
	//	sdk.MustAccAddressFromBech32("symphony12mdnm5yv5dfz37qsu0eu60x8qwxxl0x7sqqzn0"),
	//	sdk.MustAccAddressFromBech32("symphony1ar8mfrrtkwlm62wgu88d0cfleng5gl8y062gsn"),
	//	sdk.MustAccAddressFromBech32("symphony1kvgujs5yg9h6l6e265smwx99fmnnmc4af5v0ah"),
	//	sdk.MustAccAddressFromBech32("symphony1ww5e3y7ptw8h3lc0cumxe5lmcu3m53dn7qyn4k"),
	//	sdk.MustAccAddressFromBech32("symphony1tsehv6f0v7ce4gy7574thxnp6v8jx7jm4evkpe"),
	//	sdk.MustAccAddressFromBech32("symphony1fg5d24fgmxgux2p8e6xm8vjdjza8xy3ju6ta6m")}
	//
	//// Fund localsymphony accounts
	//for _, account := range localSymphonyAccounts {
	//	err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, defaultCoins)
	//	if err != nil {
	//		tmos.Exit(err.Error())
	//	}
	//	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, account, defaultCoins)
	//	if err != nil {
	//		tmos.Exit(err.Error())
	//	}
	//}
	//
	//// Fund edgenet faucet
	//faucetCoins := sdk.NewCoins(
	//	sdk.NewInt64Coin("ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7", 1000000000000000), // DAI
	//	sdk.NewInt64Coin("note", 1000000000000000))
	//err = app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, faucetCoins)
	//if err != nil {
	//	tmos.Exit(err.Error())
	//}
	//err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, sdk.MustAccAddressFromBech32("melody1rqgf207csps822qwmd3k2n6k6k4e99w502e79t"), faucetCoins)
	//if err != nil {
	//	tmos.Exit(err.Error())
	//}
	//
	//// Mars bank account
	//marsCoins := sdk.NewCoins(
	//	sdk.NewInt64Coin("note", 10000000000000),
	//	sdk.NewInt64Coin("ibc/903A61A498756EA560B85A85132D3AEE21B5DEDD41213725D22ABF276EA6945E", 400000000000),
	//	sdk.NewInt64Coin("ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858", 3000000000000),
	//	sdk.NewInt64Coin("ibc/C140AFD542AE77BD7DCC83F13FDD8C5E5BB8C4929785E6EC2F4C636F98F17901", 200000000000),
	//	sdk.NewInt64Coin("ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2", 700000000000),
	//	sdk.NewInt64Coin("ibc/D1542AA8762DB13087D8364F3EA6509FD6F009A34F00426AF9E4F9FA85CBBF1F", 2000000000),
	//	sdk.NewInt64Coin("ibc/EA1D43981D5C9A1C4AAEA9C23BB1D4FA126BA9BC7020A25E0AE4AA841EA25DC5", 3000000000000000000))
	//err = app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, marsCoins)
	//if err != nil {
	//	tmos.Exit(err.Error())
	//}
	//err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, sdk.MustAccAddressFromBech32("melody1ev02crc36675xd8s029qh7wg3wjtfk37jr004z"), marsCoins)
	//if err != nil {
	//	tmos.Exit(err.Error())
	//}

	// UPGRADE
	//

	if upgradeToTrigger != "" {
		upgradePlan := upgradetypes.Plan{
			Name:   upgradeToTrigger,
			Height: app.LastBlockHeight() + 10,
		}
		err = app.UpgradeKeeper.ScheduleUpgrade(ctx, upgradePlan)
		if err != nil {
			panic(err)
		}
	}

	return app
}

// MakeCodecs returns the application codec and a legacy Amino codec.
func MakeCodecs() (codec.Codec, *codec.LegacyAmino) {
	config := MakeEncodingConfig()
	return config.Marshaler, config.Amino
}

func (app *SymphonyApp) GetBaseApp() *baseapp.BaseApp {
	return app.BaseApp
}

// Name returns the name of the App.
func (app *SymphonyApp) Name() string { return app.BaseApp.Name() }

// BeginBlocker application updates every begin block.
func (app *SymphonyApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	BeginBlockForks(ctx, app)
	return app.mm.BeginBlock(ctx, req)
}

// EndBlocker application updates every end block.
func (app *SymphonyApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	// Process the block and ingest data into various sinks.
	app.IngestManager.ProcessBlock(ctx)
	return app.mm.EndBlock(ctx, req)
}

// InitChainer application update at chain initialization.
func (app *SymphonyApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}

	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())

	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height.
func (app *SymphonyApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *SymphonyApp) LegacyAmino() *codec.LegacyAmino {
	return app.cdc
}

// AppCodec returns Symphony' app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *SymphonyApp) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns Symphony' InterfaceRegistry.
func (app *SymphonyApp) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

func (app *SymphonyApp) ModuleManager() module.Manager {
	return *app.mm
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *SymphonyApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register legacy and grpc-gateway routes for all modules.
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register node gRPC service for grpc-gateway.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register swagger API from root so that other applications can override easily
	if apiConfig.Swagger {
		RegisterSwaggerAPI(clientCtx, apiSvr.Router)
	}
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *SymphonyApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *SymphonyApp) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(
		clientCtx,
		app.BaseApp.GRPCQueryRouter(),
		app.interfaceRegistry,
		app.Query,
	)
}

// RegisterNodeService registers the node gRPC Query service.
func (app *SymphonyApp) RegisterNodeService(clientCtx client.Context) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter())
}

// SimulationManager implements the SimulationApp interface
func (app *SymphonyApp) SimulationManager() *module.SimulationManager {
	return app.sm
}

// ChainID gets chainID from private fields of BaseApp
// Should be removed once SDK 0.50.x will be adopted
func (app *SymphonyApp) ChainID() string {
	field := reflect.ValueOf(app.BaseApp).Elem().FieldByName("chainID")
	return field.String()
}

// configure store loader that checks if version == upgradeHeight and applies store upgrades
func (app *SymphonyApp) setupUpgradeStoreLoaders() {
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Sprintf("failed to read upgrade info from disk %s", err))
	}

	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	currentHeight := app.CommitMultiStore().LastCommitID().Version

	if upgradeInfo.Height == currentHeight+1 {
		app.customPreUpgradeHandler(upgradeInfo)
	}

	for _, upgrade := range Upgrades {
		if upgradeInfo.Name == upgrade.UpgradeName {
			storeUpgrades := upgrade.StoreUpgrades
			app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
		}
	}
}

func (app *SymphonyApp) customPreUpgradeHandler(upgradeInfo upgradetypes.Plan) {
	switch upgradeInfo.Name {
	case "v16":
		// v16 upgrade handler
		// remove the wasm cache for cosmwasm cherry https://github.com/CosmWasm/advisories/blob/main/CWAs/CWA-2023-002.md#wasm-module-cache-issue
		err := os.RemoveAll(app.homePath + "/wasm/wasm/cache")
		if err != nil {
			panic(err)
		}
	}
}

func (app *SymphonyApp) setupUpgradeHandlers() {
	for _, upgrade := range Upgrades {
		app.UpgradeKeeper.SetUpgradeHandler(
			upgrade.UpgradeName,
			upgrade.CreateUpgradeHandler(
				app.mm,
				app.configurator,
				app.BaseApp,
				&app.AppKeepers,
			),
		)
	}
}

// RegisterSwaggerAPI registers swagger route with API Server.
func RegisterSwaggerAPI(ctx client.Context, rtr *mux.Router) {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	rtr.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticServer))
	rtr.PathPrefix("/swagger/").Handler(staticServer)
}

// GetMaccPerms returns a copy of the module account permissions.
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		dupMaccPerms[k] = v
	}

	return dupMaccPerms
}
