// Copyright (c) 2023-2024 Nibi, Inc.
package statedb

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"

	"github.com/NibiruChain/nibiru/v2/x/evm"
)

// TxConfig encapsulates the readonly information of current tx for `StateDB`.
type TxConfig struct {
	BlockHash common.Hash // hash of current block
	TxHash    common.Hash // hash of current tx
	TxIndex   uint        // index of current eth transaction in the block
	LogIndex  uint        // index of next log within current block
}

// NewEmptyTxConfig construct an empty TxConfig,
// used in context where there's no transaction, e.g. `eth_call`/`eth_estimateGas`.
func NewEmptyTxConfig(bhash common.Hash) TxConfig {
	return TxConfig{
		BlockHash: bhash,
		TxHash:    common.Hash{},
		TxIndex:   0,
		LogIndex:  0,
	}
}

// EVMConfig encapsulates common parameters needed to create an EVM to execute a message
// It's mainly to reduce the number of method parameters
type EVMConfig struct {
	Params      evm.Params
	ChainConfig *params.ChainConfig
	CoinBase    common.Address
	BaseFee     *big.Int
}
