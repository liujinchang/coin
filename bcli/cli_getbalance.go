package bcli

import (
	"coin"
	"fmt"
	"log"
	"utils"
)

func (cli *CLI) getBalance(address string) {
	if !coin.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := coin.FindBlockchain(false)
	db := bc.GetDB()
	defer db.Close()
	if db == nil {
		fmt.Printf("Balance of '%s': %d\n", address, 0)
	} else {
		UTXOSet := coin.UTXOSet{Blockchain: bc}
		UTXOSet = UTXOSet.Init()
		defer UTXOSet.GetDB().Close()
		balance := 0
		pubKeyHash := utils.Base58Decode([]byte(address))
		pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
		UTXOs := UTXOSet.FindUTXO(pubKeyHash)
		for _, out := range UTXOs {
			balance += out.Value
		}
		fmt.Printf("Balance of '%s': %d\n", address, balance)
	}
}
