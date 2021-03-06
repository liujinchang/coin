package coin

import (
	"bytes"
	"config"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"utils"

	"github.com/boltdb/bolt"
)

// Blockchain implements interactions with a DB
var indexDb *bolt.DB
type Blockchain struct {
	tip 			[]byte
	db  			*bolt.DB
	dbFileIndex		int
}
func getDBfileIndex() []byte {
	var fileIndex []byte
	var err error
	indexDb, err = bolt.Open(fmt.Sprintf(config.Root+"/database/index_%s.db", nodeID), 0600, nil)
	utils.ErrorLog(err)
	err = indexDb.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("index"))
		utils.ErrorLog(err)
		fileIndex = b.Get([]byte("fileIndex"))
		if fileIndex == nil {
			fileIndex = []byte("000000")
			b.Put([]byte("fileIndex"), fileIndex)
		}
		return nil
	})
	utils.ErrorLog(err)
	return fileIndex
}
// CreateBlockchain creates a new blockchain DB
func CreateBlockchain(address string) *Blockchain {
	var dbFileIndex string = string(getDBfileIndex())
	var dbFile = fmt.Sprintf(config.Root + "/database/" + dbFileName, nodeID, dbFileIndex)
	if utils.FileExists(dbFile) {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	} else {
		var tip []byte
		cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(cbtx)
		db, err := bolt.Open(dbFile,0600,nil)
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
		index, err := strconv.Atoi(dbFileIndex)
		utils.ErrorLog(err)
		bc := Blockchain{tip,db,index}
		bc.createIndex(genesis)
		return &bc
	}
	return nil
}

// FindBlockchain finds a Blockchain with genesis Block
// 本地不存在区块链时，是否退出程序
func FindBlockchain(isExit bool) *Blockchain {
	var bc Blockchain
	var dbFileIndex string = string(getDBfileIndex())
	var dbFile = fmt.Sprintf(config.Root + "/database/" + dbFileName, nodeID, dbFileIndex)
	if !utils.FileExists(dbFile) {
		if isExit {
			fmt.Println("No existing blockchain found. Create one first.")
			os.Exit(1)
		} else {
			bc = Blockchain{[]byte{},nil,0}
		}
	} else {
		var tip []byte
		db, err := bolt.Open(dbFile,0600,nil)
		utils.ErrorLog(err)
		err = db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(blocksBucket))
			tip = b.Get([]byte("l"))
			return nil
		})
		utils.ErrorLog(err)
		if tip != nil {
			index, err := strconv.Atoi(dbFileIndex)
			utils.ErrorLog(err)
			bc = Blockchain{tip,db,index}
		} else {
			bc = Blockchain{[]byte{},nil,0}
		}
	}
	return &bc
}

// AddBlock saves the block into the blockchain
func (bc *Blockchain) AddBlock(block *Block) {
	//There are 10 blocks in the db file
	if (block.Height+1) % blockCountInFile == 0 {
		bc.db.Close()
		bc.dbFileIndex++
		var dbFileIndex = utils.TransformFileIndex(bc.dbFileIndex)
		var dbFile = fmt.Sprintf(config.Root + "/database/" + dbFileName, nodeID, dbFileIndex)
		db, err := bolt.Open(dbFile,0600,nil)
		utils.ErrorLog(err)
		bc.db = db
		err = indexDb.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte("index"))
			utils.ErrorLog(err)
			b.Put([]byte("fileIndex"),[]byte(dbFileIndex))
			return nil
		})
		utils.ErrorLog(err)
	}
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(blocksBucket))
		utils.ErrorLog(err)
		blockInDb := b.Get(block.Hash)
		if blockInDb != nil {
			return nil
		}
		blockData := utils.GobEncode(block)
		err = b.Put(block.Hash, blockData)
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
	bc.createIndex(block)
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
	return &BlockchainIterator{bc.tip,bc.db,bc.dbFileIndex}
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
	var fileIndex string
	indexDb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("index"))
		fileIndex = string(b.Get(blockHash))
		return nil
	})
	var db *bolt.DB
	var err error
	if fileIndex != utils.TransformFileIndex(bc.dbFileIndex) {
		var dbFile = fmt.Sprintf(config.Root + "/database/" + dbFileName, nodeID, fileIndex)
		db, err = bolt.Open(dbFile, 0600, nil)
		utils.ErrorLog(err)
		defer db.Close()
	} else {
		db = bc.db
	}
	var block Block
	err = db.View(func(tx *bolt.Tx) error {
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
	if (newBlock.Height+1) % blockCountInFile == 0 {
		defer bc.db.Close()
		bc.dbFileIndex++
		var dbFileIndex = utils.TransformFileIndex(bc.dbFileIndex)
		var dbFile = fmt.Sprintf(config.Root + "/database/" + dbFileName, nodeID, dbFileIndex)
		db, err := bolt.Open(dbFile,0600,nil)
		utils.ErrorLog(err)
		bc.db = db
		err = indexDb.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte("index"))
			utils.ErrorLog(err)
			b.Put([]byte("fileIndex"),[]byte(dbFileIndex))
			return nil
		})
		utils.ErrorLog(err)
	}

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(blocksBucket))
		utils.ErrorLog(err)
		err = b.Put(newBlock.Hash, utils.GobEncode(newBlock))
		utils.ErrorLog(err)
		err = b.Put([]byte("l"), newBlock.Hash)
		utils.ErrorLog(err)
		bc.tip = newBlock.Hash
		return nil
	})
	utils.ErrorLog(err)
	bc.createIndex(newBlock)
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
func (bc *Blockchain) createIndex(block *Block){
	err := indexDb.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("index"))
		utils.ErrorLog(err)
		b.Put(block.Hash,[]byte(utils.TransformFileIndex(bc.dbFileIndex)))
		return nil
	})
	utils.ErrorLog(err)
}