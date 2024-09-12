// Copyright (c) 2023-2024 Nibi, Inc.
package rpc

import (
	"fmt"
	"strconv"

	"cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	tmrpctypes "github.com/cometbft/cometbft/rpc/core/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/NibiruChain/nibiru/v2/eth"
	"github.com/NibiruChain/nibiru/v2/x/evm"
)

// ParsedTx is eth tx info parsed from ABCI events. Each `ParsedTx` corresponds
// to one eth tx msg ([evm.MsgEthereumTx]).
type ParsedTx struct {
	MsgIndex uint64

	// the following fields are parsed from events

	Hash common.Hash
	// -1 means uninitialized
	EthTxIndex int32
	GasUsed    uint64
	Failed     bool
}

func (p ParsedTx) FromEvent(e *evm.EventEthereumTx) (ParsedTx, error) {
	if e == nil {
		return ParsedTx{}, fmt.Errorf("nil eth tx event")
	}
	ethBlockTxIdx, err := strconv.ParseInt(e.Index, 10, 64)
	if err != nil {
		return ParsedTx{}, err
	}
	gasUsed, err := strconv.ParseUint(e.GasUsed, 10, 64)
	if err != nil {
		return ParsedTx{}, err
	}
	blockTxMsgIdx, err := strconv.ParseUint(e.Index, 10, 64)
	if err != nil {
		return ParsedTx{}, err
	}

	return ParsedTx{
		MsgIndex:   blockTxMsgIdx,
		Hash:       common.HexToHash(e.EthHash),
		EthTxIndex: int32(ethBlockTxIdx),
		GasUsed:    gasUsed,
		Failed:     len(e.EthTxFailed) > 0,
	}, nil
}

// NewParsedTx initialize a ParsedTx
func NewParsedTx(msgIndex uint64) ParsedTx {
	return ParsedTx{MsgIndex: msgIndex, EthTxIndex: -1}
}

// ParsedTxs is the tx infos parsed from eth tx events.
type ParsedTxs struct {
	// one item per message
	Txs []ParsedTx
	// map tx hash to msg index
	TxHashes map[common.Hash]int
}

// ParseTxResult: parses eth tx info from the ABCI events of Eth tx msgs
// ([evm.MsgEthereumTx]). It supports each [EventFormat].
func ParseTxResult(
	result *abci.ResponseDeliverTx, tx sdk.Tx,
) (*ParsedTxs, error) {
	// the index of current ethereum_tx event in format 1 or the second part of format 2
	// eventIndex := -1

	p := &ParsedTxs{
		TxHashes: make(map[common.Hash]int),
		Txs:      []ParsedTx{},
	}

	typeUrl := evm.TypeUrlEventEthereumTx
	for _, event := range result.Events {
		if event.Type != typeUrl {
			continue
		}

		typedProtoEvent, err := sdk.ParseTypedEvent(event)
		if err != nil {
			return nil, errors.Wrapf(
				err, "failed to parse event of type %s", typeUrl)
		}
		eventEthTx, ok := (typedProtoEvent).(*evm.EventEthereumTx)
		if !ok {
			return nil, errors.Wrapf(
				err, "failed to parse event of type %s", typeUrl)
		}

		parsedTx, err := ParsedTx{}.FromEvent(eventEthTx)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create parsed tx from event")
		}
		p.Txs = append(p.Txs, parsedTx)
		p.TxHashes[parsedTx.Hash] = int(parsedTx.MsgIndex)

		// if len(event.Attributes) == 2 {
		// 	// the first part of format 2
		// 	if err := p.newTx(event.Attributes); err != nil {
		// 		return nil, err
		// 	}
		// } else {
		// 	// format 1 or second part of format 2
		// 	eventIndex++
		// 	if format == eventFormat1 {
		// 		// append tx
		// 		if err := p.newTx(event.Attributes); err != nil {
		// 			return nil, err
		// 		}
		// 	} else {
		// 		// the second part of format 2, update tx fields
		// 		if err := p.updateTx(eventIndex, event.Attributes); err != nil {
		// 			return nil, err
		// 		}
		// 	}
		// }
	}

	// this could only happen if tx exceeds block gas limit
	if result.Code != 0 && tx != nil {
		for i := 0; i < len(p.Txs); i++ {
			p.Txs[i].Failed = true

			// replace gasUsed with gasLimit because that's what's actually deducted.
			gasLimit := tx.GetMsgs()[i].(*evm.MsgEthereumTx).GetGas()
			p.Txs[i].GasUsed = gasLimit
		}
	}
	return p, nil
}

// ParseTxIndexerResult parse tm tx result to a format compatible with the custom tx indexer.
func ParseTxIndexerResult(
	txResult *tmrpctypes.ResultTx, tx sdk.Tx, getter func(*ParsedTxs) *ParsedTx,
) (*eth.TxResult, error) {
	txs, err := ParseTxResult(&txResult.TxResult, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tx events: block %d, index %d, %v", txResult.Height, txResult.Index, err)
	}

	parsedTx := getter(txs)
	if parsedTx == nil {
		return nil, fmt.Errorf("ethereum tx not found in msgs: block %d, index %d", txResult.Height, txResult.Index)
	}
	index := uint32(parsedTx.MsgIndex) // #nosec G701
	return &eth.TxResult{
		Height:            txResult.Height,
		TxIndex:           txResult.Index,
		MsgIndex:          index,
		EthTxIndex:        parsedTx.EthTxIndex,
		Failed:            parsedTx.Failed,
		GasUsed:           parsedTx.GasUsed,
		CumulativeGasUsed: txs.AccumulativeGasUsed(parsedTx.MsgIndex),
	}, nil
}

// newTx parse a new tx from events, called during parsing.
func (p *ParsedTxs) newTx(attrs []abci.EventAttribute) error {
	msgIndex := len(p.Txs)
	tx := NewParsedTx(uint64(msgIndex))
	if err := fillTxAttributes(&tx, attrs); err != nil {
		return err
	}
	p.Txs = append(p.Txs, tx)
	p.TxHashes[tx.Hash] = msgIndex
	return nil
}

// updateTx updates an exiting tx from events, called during parsing.
// In event format 2, we update the tx with the attributes of the second `ethereum_tx` event,
func (p *ParsedTxs) updateTx(eventIndex int, attrs []abci.EventAttribute) error {
	tx := NewParsedTx(uint64(eventIndex))
	if err := fillTxAttributes(&tx, attrs); err != nil {
		return err
	}
	if tx.Hash != p.Txs[eventIndex].Hash {
		// if hash is different, index the new one too
		p.TxHashes[tx.Hash] = eventIndex
	}
	// override the tx because the second event is more trustworthy
	p.Txs[eventIndex] = tx
	return nil
}

// GetTxByHash find ParsedTx by tx hash, returns nil if not exists.
func (p *ParsedTxs) GetTxByHash(hash common.Hash) *ParsedTx {
	if idx, ok := p.TxHashes[hash]; ok {
		return &p.Txs[idx]
	}
	return nil
}

// GetTxByMsgIndex returns ParsedTx by msg index
func (p *ParsedTxs) GetTxByMsgIndex(i int) *ParsedTx {
	if i < 0 || i >= len(p.Txs) {
		return nil
	}
	return &p.Txs[i]
}

// GetTxByTxIndex returns ParsedTx by tx index
func (p *ParsedTxs) GetTxByTxIndex(txIndex int) *ParsedTx {
	if len(p.Txs) == 0 {
		return nil
	}
	// assuming the `EthTxIndex` increase continuously,
	// convert TxIndex to MsgIndex by subtract the begin TxIndex.
	msgIndex := txIndex - int(p.Txs[0].EthTxIndex)
	// GetTxByMsgIndex will check the bound
	return p.GetTxByMsgIndex(msgIndex)
}

// AccumulativeGasUsed calculates the accumulated gas used within the batch of txs
func (p *ParsedTxs) AccumulativeGasUsed(msgIndex uint64) (result uint64) {
	for i := uint64(0); i <= msgIndex; i++ {
		result += p.Txs[i].GasUsed
	}
	return result
}

// fillTxAttribute parse attributes by name, less efficient than hardcode the
// index, but more stable against event format changes.
func fillTxAttribute(tx *ParsedTx, key string, value string) error {
	switch key {
	case evm.AttributeKeyEthereumTxHash:
		tx.Hash = common.HexToHash(value)
	case evm.AttributeKeyTxIndex:
		txIndex, err := strconv.ParseUint(value, 10, 31) // #nosec G701
		if err != nil {
			return err
		}
		tx.EthTxIndex = int32(txIndex) // #nosec G701
	case evm.AttributeKeyTxGasUsed:
		gasUsed, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		tx.GasUsed = gasUsed
	case evm.AttributeKeyEthereumTxFailed:
		tx.Failed = len(value) > 0
	}
	return nil
}

func fillTxAttributes(tx *ParsedTx, attrs []abci.EventAttribute) error {
	for _, attr := range attrs {
		if err := fillTxAttribute(tx, attr.Key, attr.Value); err != nil {
			return err
		}
	}
	return nil
}
