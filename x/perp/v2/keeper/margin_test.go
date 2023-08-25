package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/NibiruChain/collections"

	"github.com/NibiruChain/nibiru/x/common/asset"
	"github.com/NibiruChain/nibiru/x/common/denoms"
	"github.com/NibiruChain/nibiru/x/common/testutil"
	tutilaction "github.com/NibiruChain/nibiru/x/common/testutil/action"
	tutilassert "github.com/NibiruChain/nibiru/x/common/testutil/assertion"
	perpaction "github.com/NibiruChain/nibiru/x/perp/v2/integration/action"
	perpassert "github.com/NibiruChain/nibiru/x/perp/v2/integration/assertion"
	types "github.com/NibiruChain/nibiru/x/perp/v2/types"
)

func TestAddMargin(t *testing.T) {
	alice := testutil.AccAddress()
	pairBtcUsdc := asset.Registry.Pair(denoms.BTC, denoms.USDC)
	pairEthUsdc := asset.Registry.Pair(denoms.ETH, denoms.USDC)
	startBlockTime := time.Now()

	tc := tutilaction.TestCases{
		tutilaction.TC("existing long position, add margin").
			Given(
				perpaction.CreateCustomMarket(pairBtcUsdc),
				tutilaction.SetBlockNumber(1),
				tutilaction.SetBlockTime(startBlockTime),
				tutilaction.FundAccount(alice, sdk.NewCoins(sdk.NewCoin(denoms.USDC, sdk.NewInt(2020)))),
				perpaction.MarketOrder(alice, pairBtcUsdc, types.Direction_LONG, sdk.NewInt(1000), sdk.NewDec(10), sdk.ZeroDec()),
			).
			When(
				tutilaction.MoveToNextBlock(),
				perpaction.AddMargin(alice, pairBtcUsdc, sdk.NewInt(1000)),
			).
			Then(
				perpassert.PositionShouldBeEqual(alice, pairBtcUsdc, perpassert.Position_PositionShouldBeEqualTo(
					types.Position{
						Pair:                            pairBtcUsdc,
						TraderAddress:                   alice.String(),
						Size_:                           sdk.MustNewDecFromStr("9999.999900000001000000"),
						Margin:                          sdk.NewDec(2000),
						OpenNotional:                    sdk.NewDec(10000),
						LatestCumulativePremiumFraction: sdk.ZeroDec(),
						LastUpdatedBlockNumber:          2,
					},
				)),
				perpassert.PositionChangedEventShouldBeEqual(&types.PositionChangedEvent{
					FinalPosition: types.Position{
						Pair:                            pairBtcUsdc,
						TraderAddress:                   alice.String(),
						Size_:                           sdk.MustNewDecFromStr("9999.999900000001000000"),
						Margin:                          sdk.NewDec(2000),
						OpenNotional:                    sdk.NewDec(10000),
						LatestCumulativePremiumFraction: sdk.ZeroDec(),
						LastUpdatedBlockNumber:          2,
					},
					PositionNotional:  sdk.NewDec(10_000),
					RealizedPnl:       sdk.ZeroDec(),
					BadDebt:           sdk.NewCoin(denoms.USDC, sdk.ZeroInt()),
					FundingPayment:    sdk.ZeroDec(),
					TransactionFee:    sdk.NewCoin(denoms.USDC, sdk.ZeroInt()),
					BlockHeight:       2,
					MarginToUser:      sdk.NewInt(-1_000),
					ChangeReason:      types.ChangeReason_AddMargin,
					ExchangedNotional: sdk.MustNewDecFromStr("0"),
					ExchangedSize:     sdk.MustNewDecFromStr("0"),
				}),
				tutilassert.BalanceEqual(alice, denoms.USDC, sdk.ZeroInt()),
				tutilassert.ModuleBalanceEqual(types.PerpEFModuleAccount, denoms.USDC, sdk.NewInt(10)),
				tutilassert.ModuleBalanceEqual(types.FeePoolModuleAccount, denoms.USDC, sdk.NewInt(10)),
			),

		tutilaction.TC("existing short position, add margin").
			Given(
				perpaction.CreateCustomMarket(pairBtcUsdc),
				tutilaction.SetBlockNumber(1),
				tutilaction.SetBlockTime(startBlockTime),
				tutilaction.FundAccount(alice, sdk.NewCoins(sdk.NewCoin(denoms.USDC, sdk.NewInt(2020)))),
				perpaction.MarketOrder(alice, pairBtcUsdc, types.Direction_SHORT, sdk.NewInt(1000), sdk.NewDec(10), sdk.ZeroDec()),
			).
			When(
				tutilaction.MoveToNextBlock(),
				perpaction.AddMargin(alice, pairBtcUsdc, sdk.NewInt(1000)),
			).
			Then(
				perpassert.PositionShouldBeEqual(alice, pairBtcUsdc, perpassert.Position_PositionShouldBeEqualTo(
					types.Position{
						Pair:                            pairBtcUsdc,
						TraderAddress:                   alice.String(),
						Size_:                           sdk.MustNewDecFromStr("-10000.000100000001000000"),
						Margin:                          sdk.NewDec(2000),
						OpenNotional:                    sdk.NewDec(10000),
						LatestCumulativePremiumFraction: sdk.ZeroDec(),
						LastUpdatedBlockNumber:          2,
					},
				)),
				perpassert.PositionChangedEventShouldBeEqual(&types.PositionChangedEvent{
					FinalPosition: types.Position{
						Pair:                            pairBtcUsdc,
						TraderAddress:                   alice.String(),
						Size_:                           sdk.MustNewDecFromStr("-10000.000100000001000000"),
						Margin:                          sdk.NewDec(2000),
						OpenNotional:                    sdk.NewDec(10000),
						LatestCumulativePremiumFraction: sdk.ZeroDec(),
						LastUpdatedBlockNumber:          2,
					},
					PositionNotional:  sdk.NewDec(10_000),
					RealizedPnl:       sdk.ZeroDec(),
					BadDebt:           sdk.NewCoin(denoms.USDC, sdk.ZeroInt()),
					FundingPayment:    sdk.ZeroDec(),
					TransactionFee:    sdk.NewCoin(denoms.USDC, sdk.ZeroInt()),
					BlockHeight:       2,
					MarginToUser:      sdk.NewInt(-1000),
					ChangeReason:      types.ChangeReason_AddMargin,
					ExchangedNotional: sdk.MustNewDecFromStr("0"),
					ExchangedSize:     sdk.MustNewDecFromStr("0"),
				}),
				tutilassert.BalanceEqual(alice, denoms.USDC, sdk.ZeroInt()),
				tutilassert.ModuleBalanceEqual(types.PerpEFModuleAccount, denoms.USDC, sdk.NewInt(10)),
				tutilassert.ModuleBalanceEqual(types.FeePoolModuleAccount, denoms.USDC, sdk.NewInt(10)),
			),

		tutilaction.TC("Testing fails").
			Given(
				perpaction.CreateCustomMarket(pairBtcUsdc),
				perpaction.CreateCustomMarket(pairEthUsdc),

				tutilaction.SetBlockNumber(1),
				tutilaction.SetBlockTime(startBlockTime),
				tutilaction.FundAccount(alice, sdk.NewCoins(sdk.NewCoin(denoms.USDC, sdk.NewInt(1020)))),
				perpaction.MarketOrder(alice, pairBtcUsdc, types.Direction_LONG, sdk.NewInt(1000), sdk.NewDec(10), sdk.ZeroDec()),
			).
			When(
				tutilaction.MoveToNextBlock(),
				perpaction.AddMarginFail(alice, asset.MustNewPair("luna:usdt"), sdk.NewInt(1000), types.ErrPairNotFound),
				perpaction.AddMarginFail(alice, pairEthUsdc, sdk.NewInt(1000), collections.ErrNotFound),
				perpaction.AddMarginFail(alice, pairBtcUsdc, sdk.NewInt(1000), sdkerrors.ErrInsufficientFunds),

				perpaction.RemoveMarginFail(alice, asset.MustNewPair("luna:usdt"), sdk.NewInt(1000), types.ErrPairNotFound),
				perpaction.RemoveMarginFail(alice, pairEthUsdc, sdk.NewInt(1000), collections.ErrNotFound),
				perpaction.RemoveMarginFail(alice, pairBtcUsdc, sdk.NewInt(2000), types.ErrBadDebt),
				perpaction.RemoveMarginFail(alice, pairBtcUsdc, sdk.NewInt(900), types.ErrMarginRatioTooLow),
			),
	}

	tutilaction.NewTestSuite(t).WithTestCases(tc...).Run()
}

func TestRemoveMargin(t *testing.T) {
	alice := testutil.AccAddress()
	pairBtcUsdc := asset.Registry.Pair(denoms.BTC, denoms.USDC)
	startBlockTime := time.Now()

	tc := tutilaction.TestCases{
		tutilaction.TC("existing long position, remove margin").
			Given(
				perpaction.CreateCustomMarket(pairBtcUsdc),
				tutilaction.SetBlockNumber(1),
				tutilaction.SetBlockTime(startBlockTime),
				tutilaction.FundAccount(alice, sdk.NewCoins(sdk.NewCoin(denoms.USDC, sdk.NewInt(1002)))),
				perpaction.MarketOrder(alice, pairBtcUsdc, types.Direction_LONG, sdk.NewInt(1000), sdk.OneDec(), sdk.ZeroDec()),
			).
			When(
				tutilaction.MoveToNextBlock(),
				perpaction.RemoveMargin(alice, pairBtcUsdc, sdk.NewInt(500)),
			).
			Then(
				perpassert.PositionShouldBeEqual(alice, pairBtcUsdc, perpassert.Position_PositionShouldBeEqualTo(types.Position{
					Pair:                            pairBtcUsdc,
					TraderAddress:                   alice.String(),
					Size_:                           sdk.MustNewDecFromStr("999.999999000000001000"),
					Margin:                          sdk.NewDec(500),
					OpenNotional:                    sdk.NewDec(1000),
					LatestCumulativePremiumFraction: sdk.ZeroDec(),
					LastUpdatedBlockNumber:          2,
				})),
				perpassert.PositionChangedEventShouldBeEqual(&types.PositionChangedEvent{
					FinalPosition: types.Position{
						Pair:                            pairBtcUsdc,
						TraderAddress:                   alice.String(),
						Size_:                           sdk.MustNewDecFromStr("999.999999000000001000"),
						Margin:                          sdk.NewDec(500),
						OpenNotional:                    sdk.NewDec(1000),
						LatestCumulativePremiumFraction: sdk.ZeroDec(),
						LastUpdatedBlockNumber:          2,
					},
					PositionNotional:  sdk.NewDec(1000),
					RealizedPnl:       sdk.ZeroDec(),
					BadDebt:           sdk.NewCoin(denoms.USDC, sdk.ZeroInt()),
					FundingPayment:    sdk.ZeroDec(),
					TransactionFee:    sdk.NewCoin(denoms.USDC, sdk.ZeroInt()),
					BlockHeight:       2,
					MarginToUser:      sdk.NewInt(500),
					ChangeReason:      types.ChangeReason_RemoveMargin,
					ExchangedNotional: sdk.MustNewDecFromStr("0"),
					ExchangedSize:     sdk.MustNewDecFromStr("0"),
				}),
				tutilassert.BalanceEqual(alice, denoms.USDC, sdk.NewInt(500)),
				tutilassert.ModuleBalanceEqual(types.PerpEFModuleAccount, denoms.USDC, sdk.OneInt()),
				tutilassert.ModuleBalanceEqual(types.FeePoolModuleAccount, denoms.USDC, sdk.OneInt()),
			),

		tutilaction.TC("existing long position, remove almost all margin fails").
			Given(
				perpaction.CreateCustomMarket(pairBtcUsdc),
				tutilaction.SetBlockNumber(1),
				tutilaction.SetBlockTime(startBlockTime),
				tutilaction.FundAccount(alice, sdk.NewCoins(sdk.NewCoin(denoms.USDC, sdk.NewInt(1002)))),
				perpaction.MarketOrder(alice, pairBtcUsdc, types.Direction_LONG, sdk.NewInt(1000), sdk.OneDec(), sdk.ZeroDec()),
				tutilaction.MoveToNextBlock(),
			).
			When(
				perpaction.RemoveMarginFail(alice, pairBtcUsdc, sdk.NewInt(999), types.ErrMarginRatioTooLow),
			).
			Then(
				perpassert.PositionShouldBeEqual(alice, pairBtcUsdc, perpassert.Position_PositionShouldBeEqualTo(types.Position{
					Pair:                            pairBtcUsdc,
					TraderAddress:                   alice.String(),
					Size_:                           sdk.MustNewDecFromStr("999.999999000000001000"),
					Margin:                          sdk.NewDec(1000),
					OpenNotional:                    sdk.NewDec(1000),
					LatestCumulativePremiumFraction: sdk.ZeroDec(),
					LastUpdatedBlockNumber:          1,
				})),
				tutilassert.BalanceEqual(alice, denoms.USDC, sdk.ZeroInt()),
				tutilassert.ModuleBalanceEqual(types.PerpEFModuleAccount, denoms.USDC, sdk.OneInt()),
				tutilassert.ModuleBalanceEqual(types.FeePoolModuleAccount, denoms.USDC, sdk.OneInt()),
			),

		tutilaction.TC("existing short position, remove margin").
			Given(
				perpaction.CreateCustomMarket(pairBtcUsdc),
				tutilaction.SetBlockNumber(1),
				tutilaction.SetBlockTime(startBlockTime),
				tutilaction.FundAccount(alice, sdk.NewCoins(sdk.NewCoin(denoms.USDC, sdk.NewInt(1002)))),
				perpaction.MarketOrder(alice, pairBtcUsdc, types.Direction_SHORT, sdk.NewInt(1000), sdk.OneDec(), sdk.ZeroDec()),
			).
			When(
				tutilaction.MoveToNextBlock(),
				perpaction.RemoveMargin(alice, pairBtcUsdc, sdk.NewInt(500)),
			).
			Then(
				perpassert.PositionShouldBeEqual(alice, pairBtcUsdc, perpassert.Position_PositionShouldBeEqualTo(types.Position{
					Pair:                            pairBtcUsdc,
					TraderAddress:                   alice.String(),
					Size_:                           sdk.MustNewDecFromStr("-1000.000001000000001000"),
					Margin:                          sdk.NewDec(500),
					OpenNotional:                    sdk.NewDec(1000),
					LatestCumulativePremiumFraction: sdk.ZeroDec(),
					LastUpdatedBlockNumber:          2,
				})),
				perpassert.PositionChangedEventShouldBeEqual(&types.PositionChangedEvent{
					FinalPosition: types.Position{
						Pair:                            pairBtcUsdc,
						TraderAddress:                   alice.String(),
						Size_:                           sdk.MustNewDecFromStr("-1000.000001000000001000"),
						Margin:                          sdk.NewDec(500),
						OpenNotional:                    sdk.NewDec(1000),
						LatestCumulativePremiumFraction: sdk.ZeroDec(),
						LastUpdatedBlockNumber:          2,
					},
					PositionNotional:  sdk.NewDec(1000),
					RealizedPnl:       sdk.ZeroDec(),
					BadDebt:           sdk.NewCoin(denoms.USDC, sdk.ZeroInt()),
					FundingPayment:    sdk.ZeroDec(),
					TransactionFee:    sdk.NewCoin(denoms.USDC, sdk.ZeroInt()),
					BlockHeight:       2,
					MarginToUser:      sdk.NewInt(500),
					ChangeReason:      types.ChangeReason_RemoveMargin,
					ExchangedNotional: sdk.MustNewDecFromStr("0"),
					ExchangedSize:     sdk.MustNewDecFromStr("0"),
				}),
				tutilassert.BalanceEqual(alice, denoms.USDC, sdk.NewInt(500)),
				tutilassert.ModuleBalanceEqual(types.PerpEFModuleAccount, denoms.USDC, sdk.OneInt()),
				tutilassert.ModuleBalanceEqual(types.FeePoolModuleAccount, denoms.USDC, sdk.OneInt()),
			),

		tutilaction.TC("existing short position, remove almost all margin fails").
			Given(
				perpaction.CreateCustomMarket(pairBtcUsdc),
				tutilaction.SetBlockNumber(1),
				tutilaction.SetBlockTime(startBlockTime),
				tutilaction.FundAccount(alice, sdk.NewCoins(sdk.NewCoin(denoms.USDC, sdk.NewInt(1002)))),
				perpaction.MarketOrder(alice, pairBtcUsdc, types.Direction_SHORT, sdk.NewInt(1000), sdk.OneDec(), sdk.ZeroDec()),
				tutilaction.MoveToNextBlock(),
			).
			When(
				perpaction.RemoveMarginFail(alice, pairBtcUsdc, sdk.NewInt(999), types.ErrMarginRatioTooLow),
			).
			Then(
				perpassert.PositionShouldBeEqual(alice, pairBtcUsdc, perpassert.Position_PositionShouldBeEqualTo(types.Position{
					Pair:                            pairBtcUsdc,
					TraderAddress:                   alice.String(),
					Size_:                           sdk.MustNewDecFromStr("-1000.000001000000001000"),
					Margin:                          sdk.NewDec(1000),
					OpenNotional:                    sdk.NewDec(1000),
					LatestCumulativePremiumFraction: sdk.ZeroDec(),
					LastUpdatedBlockNumber:          1,
				})),
				tutilassert.BalanceEqual(alice, denoms.USDC, sdk.ZeroInt()),
				tutilassert.ModuleBalanceEqual(types.PerpEFModuleAccount, denoms.USDC, sdk.OneInt()),
				tutilassert.ModuleBalanceEqual(types.FeePoolModuleAccount, denoms.USDC, sdk.OneInt()),
			),
	}

	tutilaction.NewTestSuite(t).WithTestCases(tc...).Run()
}
