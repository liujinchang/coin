package coin

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strings"
	"utils"
)

// TXOutput represents a transaction output
type TXOutput struct {
	Value      int
	PubKeyHash []byte
}

// Lock signs the output
func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := utils.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

// IsLockedWithKey checks if the output can be used by the owner of the pubkey
func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

// NewTXOutput create a new TXOutput
func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock([]byte(address))
	return txo
}

// TXOutputs collects TXOutput
type TXOutputs struct {
	Outputs []TXOutput
}
// DeserializeOutputs deserializes TXOutputs
func DeserializeOutputs(data []byte) TXOutputs {
	var outputs TXOutputs
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	utils.ErrorLog(err)
	return outputs
}
// String returns a human-readable representation of a transaction
func (utxo TXOutput) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("--- TXOutput ---"))
	lines = append(lines, fmt.Sprintf("     Value : %d", utxo.Value))
	lines = append(lines, fmt.Sprintf("     PubKeyHash : %x", utxo.PubKeyHash))
	return strings.Join(lines, "\n")
}
