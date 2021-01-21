package coin

import (
	"config"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"os"
	"utils"
)

var mempoolFile = fmt.Sprintf(config.Root+"/database/"+mempoolFileName, nodeID)
var mempoolBucketName = []byte(mempoolBucket)
type Mempool struct {
	db 			*bolt.DB
}
func (m Mempool) AddTransactions(transactions []Transaction) {
	err := m.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(mempoolBucketName)
		for _, t := range transactions {
			err := b.Put(t.ID, utils.GobEncode(t))
			utils.ErrorLog(err)
		}
		return nil
	})
	utils.ErrorLog(err)
}
func (m Mempool) ReBuild() *Mempool {
	if !utils.FileExists(mempoolFile) {
		os.Create(mempoolFile)
		db, err := bolt.Open(mempoolFile,0600,nil)
		utils.ErrorLog(err)
		m.db = db
	} else {
		db, err := bolt.Open(mempoolFile,0600,nil)
		utils.ErrorLog(err)
		err = db.Update(func(tx *bolt.Tx) error {
			err := tx.DeleteBucket(mempoolBucketName)
			if err != nil && err != bolt.ErrBucketNotFound {
				log.Panic(err)
			}
			_, err = tx.CreateBucket(mempoolBucketName)
			utils.ErrorLog(err)
			return nil
		})
		utils.ErrorLog(err)
		m.db = db
	}
	return &m
}
func (m Mempool) FindTransactions(counter int) []Transaction {
	var transactions []Transaction
	bucketName := []byte("mempoolBucket")
	err := m.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket(bucketName)
		utils.ErrorLog(err)
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			transactionBytes := b.Get(k)
			transactions=append(transactions, DeserializeTransaction(transactionBytes))
			err = b.Delete(k)
			utils.ErrorLog(err)
			counter--
			if counter == 0 {
				break
			}
		}
		return nil
	})
	utils.ErrorLog(err)
	return transactions
}
func (m Mempool) FindTransaction(id []byte) Transaction {
	var transaction Transaction
	bucketName := []byte("mempoolBucket")
	err := m.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket(bucketName)
		utils.ErrorLog(err)
		transactionBytes := b.Get(id)
		transaction = DeserializeTransaction(transactionBytes)
		//b.Delete(id)
		return nil
	})
	utils.ErrorLog(err)
	return transaction
}
func (m Mempool) DeleteTransaction(id []byte) {
	err := m.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket(mempoolBucketName)
		utils.ErrorLog(err)
		b.Delete(id)
		return nil
	})
	utils.ErrorLog(err)
}
func (m Mempool) GetDB() *bolt.DB{
	return m.db
}
