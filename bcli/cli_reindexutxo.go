package bcli

import (
	"coin"
	"fmt"
	"utils"
)

func (cli *CLI) reindexUTXO() {
	bc := coin.FindBlockchain(true)
	db := utils.FindDB(stateFile)
	utxoSet := coin.UTXOSet{bc,db}
	defer bc.GetDB().Close()
	defer utxoSet.DB.Close()
	utxoSet.Reindex()
	count := utxoSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}
