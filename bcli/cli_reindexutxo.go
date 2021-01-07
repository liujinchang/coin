package bcli

import (
	"coin"
	"fmt"
)

func (cli *CLI) reindexUTXO() {
	bc := coin.FindBlockchain(true)
	UTXOSet := coin.UTXOSet{Blockchain: bc}
	UTXOSet = UTXOSet.Init()
	defer UTXOSet.GetDB().Close()
	count := UTXOSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}
