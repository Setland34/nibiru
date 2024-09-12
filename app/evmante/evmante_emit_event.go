// Copyright (c) 2023-2024 Nibi, Inc.
package evmante

import (
	"strconv"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/NibiruChain/nibiru/v2/x/evm"
)

// EthEmitEventDecorator emit events in ante handler in case of tx execution failed (out of block gas limit).
type EthEmitEventDecorator struct {
	evmKeeper EVMKeeper
}

// NewEthEmitEventDecorator creates a new EthEmitEventDecorator
func NewEthEmitEventDecorator(k EVMKeeper) EthEmitEventDecorator {
	return EthEmitEventDecorator{
		evmKeeper: k,
	}
}

// AnteHandle emits some basic events for the eth messages
func (eeed EthEmitEventDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	// After eth tx passed ante handler, the fee is deducted and nonce increased,
	// it shouldn't be ignored by json-rpc. We need to emit some events at the
	// very end of ante handler to be indexed by the consensus engine.
	blockTxIndex := eeed.evmKeeper.EVMState().BlockTxIndex.GetOr(ctx, 0)

	for msgIdx, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*evm.MsgEthereumTx)
		if !ok {
			return ctx, errorsmod.Wrapf(
				sdkerrors.ErrUnknownRequest,
				"invalid message type %T, expected %T",
				msg, (*evm.MsgEthereumTx)(nil),
			)
		}

		blockEthTxIndex := blockTxIndex + uint64(msgIdx)
		// emit ethereum tx hash as an event so that it can be indexed by
		// Tendermint for query purposes it's emitted in ante handler, so we can
		// query failed transaction (out of block gas limit).
		_ = ctx.EventManager().EmitTypedEvent(&evm.EventEthereumTx{
			// Amount:      "",
			EthHash: msgEthTx.Hash,
			Index:   strconv.FormatUint(blockEthTxIndex, 10),
			GasUsed: strconv.FormatUint(msgEthTx.GetGas(), 10),
			// Hash:        "",
			Recipient: msgEthTx.From,
			// EthTxFailed:  "",
		})
		// ctx.EventManager().EmitEvent(
		// 	sdk.NewEvent(
		// 		evm.TypeUrlEventEthereumTx,
		// 		sdk.NewAttribute(evm.AttributeKeyEthereumTxHash, msgEthTx.Hash),
		// 		sdk.NewAttribute(
		// 			evm.AttributeKeyTxIndex, strconv.FormatUint(blockTxIndex+uint64(msgIdx),
		// 				10,
		// 			),
		// 		), // #nosec G701
		// 	))
	}

	return next(ctx, tx, simulate)
}
