package coin

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"utils"

	"github.com/boltdb/bolt"
)

var bucketName = []byte(utxoBucket)
var file = fmt.Sprintf(Root+"/database/"+StateFile, os.Getenv("NODE_ID"))
// UTXOSet represents UTXO set
type UTXOSet struct {
	Blockchain *Blockchain
	Db *bolt.DB
}

// FindSpendableOutputs finds and returns unspent outputs to reference in inputs
func (u UTXOSet) FindSpendableOutputs(pubkeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	err := u.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs := DeserializeOutputs(v)
			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubkeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				}
			}
		}
		return nil
	})
	utils.ErrorLog(err)
	return accumulated, unspentOutputs
}

// FindUTXO finds UTXO for a public key hash
func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	err := u.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := DeserializeOutputs(v)
			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}
		return nil
	})
	utils.ErrorLog(err)
	return UTXOs
}

// CountTransactions returns the number of transactions in the UTXO set
func (u UTXOSet) CountTransactions() int {
	counter := 0
	err := u.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			counter++
		}
		return nil
	})
	utils.ErrorLog(err)
	return counter
}
func (u UTXOSet) Init() UTXOSet{
	if !utils.FileExists(file) {
		u.Reindex()
	} else {
		db, err := bolt.Open(file,0600,nil)
		utils.ErrorLog(err)
		u.Db = db
	}
	return u
}
// Reindex rebuilds the UTXO set
func (u UTXOSet) Reindex() {
	fmt.Println("ReBuild the UTXO set!")
	db, err := bolt.Open(file,0600,nil)
	utils.ErrorLog(err)
	err = db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucketName)
		if err != nil && err != bolt.ErrBucketNotFound {
			log.Panic(err)
		}
		_, err = tx.CreateBucket(bucketName)
		utils.ErrorLog(err)
		return nil
	})
	utils.ErrorLog(err)
	UTXO := u.Blockchain.FindUTXO()
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			utils.ErrorLog(err)
			err = b.Put(key, utils.GobEncode(outs))
			utils.ErrorLog(err)
		}
		return nil
	})
	u.Db = db
}

// Update updates the UTXO set with transactions from the Block
// The Block is considered to be the tip of a blockchain
func (u UTXOSet) Update(block *Block) {
	err := u.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		for _, tx := range block.Transactions {
			if tx.IsCoinbase() == false {
				for _, vin := range tx.Vin {
					updatedOuts := TXOutputs{}
					outsBytes := b.Get(vin.Txid)
					outs := DeserializeOutputs(outsBytes)
					for outIdx, out := range outs.Outputs {
						if outIdx != vin.Vout {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}
					if len(updatedOuts.Outputs) == 0 {
						err := b.Delete(vin.Txid)
						utils.ErrorLog(err)
					} else {
						err := b.Put(vin.Txid, utils.GobEncode(updatedOuts))
						utils.ErrorLog(err)
					}
				}
			}
			newOutputs := TXOutputs{}
			for _, out := range tx.Vout {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}
			err := b.Put(tx.ID, utils.GobEncode(newOutputs))
			utils.ErrorLog(err)
		}
		return nil
	})
	utils.ErrorLog(err)
}
func (u UTXOSet) GetDB() *bolt.DB{
	return u.Db
}