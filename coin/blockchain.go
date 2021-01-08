package coin

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"utils"

	"github.com/boltdb/bolt"
)
var dbFile = fmt.Sprintf(Root+"/database/"+dbFileName, nodeID)
// Blockchain implements interactions with a DB
type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

// CreateBlockchain creates a new blockchain DB
func CreateBlockchain(address string) *Blockchain {
	if utils.FileExists(dbFile) {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}
	var tip []byte
	cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
	genesis := NewGenesisBlock(cbtx)
	db, err := bolt.Open(dbFile, 0600, nil)
	utils.ErrorLog(err)
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(blocksBucket))
		utils.ErrorLog(err)
		err = b.Put(genesis.Hash, utils.GobEncode(genesis))
		utils.ErrorLog(err)
		err = b.Put([]byte("l"), genesis.Hash)
		utils.ErrorLog(err)
		tip = genesis.Hash
		return nil
	})
	utils.ErrorLog(err)
	bc := Blockchain{tip, db}
	return &bc
}

// FindBlockchain finds a Blockchain with genesis Block
// 本地不存在区块链时，是否退出程序
func FindBlockchain(isExit bool) *Blockchain {
	var bc Blockchain
	if utils.FileExists(dbFile) == false {
		if isExit {
			fmt.Println("No existing blockchain found. Create one first.")
			os.Exit(1)
		} else {
			bc = Blockchain{[]byte{},nil}
		}
	} else {
		var tip []byte
		db, err := bolt.Open(dbFile, 0600, nil)
		utils.ErrorLog(err)
		err = db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(blocksBucket))
			tip = b.Get([]byte("l"))
			return nil
		})
		utils.ErrorLog(err)
		if tip != nil {
			bc = Blockchain{tip, db}
		} else {
			bc = Blockchain{[]byte{},nil}
		}
	}
	return &bc
}

// AddBlock saves the block into the blockchain
func (bc *Blockchain) AddBlock(block *Block) {
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		blockInDb := b.Get(block.Hash)
		if blockInDb != nil {
			return nil
		}
		blockData := utils.GobEncode(block)
		err := b.Put(block.Hash, blockData)
		utils.ErrorLog(err)
		lastHash := b.Get([]byte("l"))
		if lastHash == nil {
			bc.SetLeader(b, block.Hash)
		} else {
			lastBlockData := b.Get(lastHash)
			lastBlock := DeserializeBlock(lastBlockData)
			if block.Height > lastBlock.Height {
				bc.SetLeader(b, block.Hash)
			}
		}
		return nil
	})
	utils.ErrorLog(err)
}
func (bc *Blockchain) SetLeader(bucket *bolt.Bucket,blockhash []byte) {
	err := bucket.Put([]byte("l"),blockhash)
	utils.ErrorLog(err)
	bc.tip = blockhash
}
// FindTransaction finds a transaction by its ID
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()
	for {
		block := bci.Next()
		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return Transaction{}, errors.New("Transaction is not found")
}

// FindUTXO finds all unspent transaction outputs and returns transactions with spent outputs removed
func (bc *Blockchain) FindUTXO() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs)
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()
	for {
		block := bci.Next()
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)
		Outputs:
			for outIdx, out := range tx.Vout {
				// Was the output spent?
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return UTXO
}

// Iterator returns a BlockchainIterat
func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{bc.tip, bc.db}
}
// GetBestHeight returns the height of the latest block
func (bc *Blockchain) GetBestHeight() int {
	var lastBlock Block
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash := b.Get([]byte("l"))
		blockData := b.Get(lastHash)
		lastBlock = *DeserializeBlock(blockData)
		return nil
	})
	utils.ErrorLog(err)
	return lastBlock.Height
}

// GetBlock finds a block by its hash and returns it
func (bc *Blockchain) GetBlock(blockHash []byte) (Block, error) {
	var block Block
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		blockData := b.Get(blockHash)
		if blockData == nil {
			return errors.New("Block is not found.")
		}
		block = *DeserializeBlock(blockData)
		return nil
	})
	utils.ErrorLog(err)
	return block, nil
}
// Exist whether a block exist or not
func (bc *Blockchain) Exist(blockHash []byte) bool {
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		blockData := b.Get(blockHash)
		if blockData == nil {
			return errors.New("Block is not found.")
		}
		return nil
	})
	if err == nil {
		return true
	} else {
		return false
	}
}
// GetBlockHashes returns a list of hashes of all the blocks in the chain
func (bc *Blockchain) GetBlockHashes() [][]byte {
	var blocks [][]byte
	bci := bc.Iterator()
	for {
		block := bci.Next()
		blocks = append(blocks, block.Hash)
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return blocks
}

// MineBlock mines a new block with the provided transactions
func (bc *Blockchain) MineBlock(transactions []*Transaction) *Block {
	var lastHash []byte
	var lastHeight int

	for _, tx := range transactions {
		// TODO: ignore transaction if it's not valid
		if bc.VerifyTransaction(tx) != true {
			log.Panic("ERROR: Invalid transaction")
		}
	}
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		blockData := b.Get(lastHash)
		block := DeserializeBlock(blockData)
		lastHeight = block.Height
		return nil
	})
	utils.ErrorLog(err)
	newBlock := NewBlock(transactions, lastHash, lastHeight+1)
	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, utils.GobEncode(newBlock))
		utils.ErrorLog(err)
		err = b.Put([]byte("l"), newBlock.Hash)
		utils.ErrorLog(err)
		bc.tip = newBlock.Hash
		return nil
	})
	utils.ErrorLog(err)
	return newBlock
}

// SignTransaction signs inputs of a Transaction
func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)
	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		utils.ErrorLog(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	tx.Sign(privKey, prevTXs)
}

// VerifyTransaction verifies transaction input signatures
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	prevTXs := make(map[string]Transaction)
	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		utils.ErrorLog(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	return tx.Verify(prevTXs)
}
//GetDB get database of blockchain
func (bc *Blockchain) GetDB() *bolt.DB{
	return bc.db
}