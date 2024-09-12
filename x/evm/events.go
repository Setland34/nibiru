// Copyright (c) 2023-2024 Nibi, Inc.
package evm

import (
	"fmt"
	"reflect"
	"strings"
)

// Event attribute keys
const (
	AttributeKeyRecipient = "recipient"
	// EventEthereumTx.EthHash == "eth_hash"
	AttributeKeyEthereumTxHash = "eth_hash"
	// EventEthereumTx.Index == "index"
	AttributeKeyTxIndex   = "index"
	AttributeKeyTxGasUsed = "txGasUsed"
	AttributeKeyTxLog     = "txLog"
	// EventEthereumTx.EthTxFailed == "eth_tx_failed"
	// tx failed in eth vm execution
	AttributeKeyEthereumTxFailed = "eth_tx_failed"
	// JSON name of EventBlockBloom.Bloom
	AttributeKeyEthereumBloom = "bloom"
)

// Evm event protobuf type URLs
var (
	//  proto.MessageName(new(EventBlockBloom))
	TypeUrlEventBlockBloom = "eth.evm.v1.EventBlockBloom"

	//  proto.MessageName(new(EventEthereumTx))
	TypeUrlEventEthereumTx = "eth.evm.v1.EventEthereumTx"
	//  proto.MessageName(new(EventTxLog))
	TypeUrlEventTxLog = "eth.evm.v1.EventTxLog"
)

// case evm.AttributeKeyEthereumTxHash:
// case evm.AttributeKeyTxIndex:
// case evm.AttributeKeyTxGasUsed:
// case evm.AttributeKeyEthereumTxFailed:

func init() {
	eventEthTx := new(EventEthereumTx)
	// TypeUrlEventBlockBloom = proto.MessageName(&EventBlockBloom{})

	fmt.Printf("TODO: UD-DEBUG: TypeUrlEventBlockBloom: %v\n", TypeUrlEventBlockBloom)
	fmt.Printf("TODO: UD-DEBUG: TypeUrlEventEthereumTx: %v\n", TypeUrlEventEthereumTx)
	fmt.Printf("TODO: UD-DEBUG: TypeUrlEventTxLog: %v\n", TypeUrlEventTxLog)

	getJSONFieldName(eventEthTx, "EthHash")
}

// getJSONFieldName gets the JSON tag for a given field name in a struct
func getJSONFieldName(theStruct any, fieldName string) (jsonName string, ok bool) {
	val := reflect.ValueOf(theStruct)
	typ := val.Type()

	// Ensure the value is a struct
	if typ.Kind() != reflect.Struct {
		return "", false
	}

	// Find the field by name
	field, ok := typ.FieldByName(fieldName)
	if !ok {
		return "", false
	}

	// Get the "json" tag
	jsonTag := field.Tag.Get("json")

	// If there's no json tag, return the field name
	if jsonTag == "" {
		return fieldName, true
	}

	// Split the json tag to remove ",omitempty" or other options
	tagParts := strings.Split(jsonTag, ",")
	return tagParts[0], true
}
