package coin

import (
	"bytes"
	"config"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"strconv"
	"utils"
)

// BlockchainIterator is used to iterate over blockchain blocks
type BlockchainIterator struct {
	currentHash 	[]byte
	db          	*bolt.DB
	dbFileIndex		int
}
// Next returns next block starting from the tip
func (i *BlockchainIterator) Next() *Block {
	var block *Block
	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		if encodedBlock != nil {
			block = DeserializeBlock(encodedBlock)
		} else {
			block = nil
		}
		return nil
	})
	utils.ErrorLog(err)
	//when current file do not find block, find block in prev file
	if block == nil {
		i.dbFileIndex--
		var bt bytes.Buffer
		dbFileIndex := strconv.Itoa(i.dbFileIndex)
		for i := 0; i < 6 - len(dbFileIndex); i++ {
			bt.WriteString("0")
		}
		bt.WriteString(dbFileIndex)
		var dbFile = fmt.Sprintf(config.Root + "/database/" + dbFileName, nodeID, bt.String())
		if !utils.FileExists(dbFile) {
			log.Panic(dbFile+" is not exist, Don't traverse blocks!")
		}
		db, err := bolt.Open(dbFile,0600,nil)
		utils.ErrorLog(err)
		i.db = db
		err = i.db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(blocksBucket))
			encodedBlock := b.Get(i.currentHash)
			if encodedBlock != nil {
				block = DeserializeBlock(encodedBlock)
			} else {
				block = nil
			}
			return nil
		})
		utils.ErrorLog(err)
	}
	i.currentHash = block.PrevBlockHash
	return block
}
