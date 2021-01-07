package coin

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strconv"
	"strings"
	"time"
	"utils"
)

// Block represents a block in the blockchain
type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
	Height        int
}

// NewBlock creates and returns Block
func NewBlock(transactions []*Transaction, prevBlockHash []byte, height int) *Block {
	block := &Block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0, height}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{}, 0)
}

// HashTransactions returns a hash of the transactions in the block
func (b *Block) HashTransactions() []byte {
	var transactions [][]byte
	for _, tx := range b.Transactions {
		transactions = append(transactions, utils.GobEncode(tx))
	}
	mTree := NewMerkleTree(transactions)
	return mTree.RootNode.Data
}

// DeserializeBlock deserializes a block
func DeserializeBlock(d []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	utils.ErrorLog(err)
	return &block
}

// String returns a human-readable representation of a transaction
func (b Block) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Block %x\n", b.Hash))
	lines = append(lines, fmt.Sprintf("Height: %d\n", b.Height))
	lines = append(lines, fmt.Sprintf("Prev. block: %x\n", b.PrevBlockHash))
	pow := NewProofOfWork(&b)
	lines = append(lines, fmt.Sprintf("PoW: %s\n", strconv.FormatBool(pow.Validate())))
	return strings.Join(lines, "")
}