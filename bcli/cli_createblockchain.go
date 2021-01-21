package bcli

import (
	"coin"
	"fmt"
	"log"
	"utils"
)

func (cli *CLI) createBlockchain(address string) {
	if !coin.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := coin.CreateBlockchain(address)
	db := utils.FindDB(stateFile)
	defer bc.GetDB().Close()
	utxoSet := coin.UTXOSet{bc,db}
	utxoSet.Reindex()
	fmt.Println("Done!")
}
