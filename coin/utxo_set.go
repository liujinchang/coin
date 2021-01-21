package coin

import (
	"encoding/hex"
	"log"
	"utils"

	"github.com/boltdb/bolt"
)

var utxoBucketName = []byte(chainstateBucket)
// UTXOSet represents UTXO set
type UTXOSet struct {
	Blockchain *Blockchain
	DB *bolt.DB
}

// FindSpendableOutputs finds and returns unspent outputs to reference in inputs
func (u UTXOSet) FindSpendableOutputs(pubkeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	err := u.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(utxoBucketName)
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
	err := u.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(utxoBucketName)
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
	err := u.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(utxoBucketName)
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			counter++
		}
		return nil
	})
	utils.ErrorLog(err)
	return counter
}
// Reindex rebuilds the UTXO set
func (u UTXOSet) Reindex() {
	log.Println("ReBuild the UTXO set!")
	err := u.DB.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(utxoBucketName)
		if err != nil && err != bolt.ErrBucketNotFound {
			log.Panic(err)
		}
		_, err = tx.CreateBucket(utxoBucketName)
		utils.ErrorLog(err)
		return nil
	})
	utils.ErrorLog(err)
	UTXO := u.Blockchain.FindUTXO()
	err = u.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(utxoBucketName)
		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			utils.ErrorLog(err)
			err = b.Put(key, utils.GobEncode(outs))
			utils.ErrorLog(err)
		}
		return nil
	})
}

// Update updates the UTXO set with transactions from the Block
// The Block is considered to be the tip of a blockchain
func (u UTXOSet) Update(block *Block) {
	err := u.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(utxoBucketName)
		utils.ErrorLog(err)
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
