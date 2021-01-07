package bcli

import (
	"fmt"
	"log"
	"coin"
)

func (cli *CLI) createBlockchain(address string) {
	if !coin.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := coin.CreateBlockchain(address)
	defer bc.GetDB().Close()
	UTXOSet := coin.UTXOSet{Blockchain:bc}
	UTXOSet.Reindex()
	fmt.Println("Done!")
}
