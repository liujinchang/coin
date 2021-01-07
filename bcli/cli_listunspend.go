package bcli

import (
	"coin"
	"fmt"
	"utils"
)
//List all unspend output
func (cli *CLI) listUnspend() {
	bc := coin.FindBlockchain(false)
	defer bc.GetDB().Close()
	utxoSet := coin.UTXOSet{Blockchain: bc}
	utxoSet = utxoSet.Init()
	defer utxoSet.GetDB().Close()
	wallets, err := coin.FindWallets()
	utils.ErrorLog(err)
	addresses := wallets.GetAddresses()
	for _, address := range addresses {
		pubKeyHash := utils.Base58Decode([]byte(address))
		pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
		UTXOs := utxoSet.FindUTXO(pubKeyHash)
		for _, utxo := range UTXOs {
			fmt.Print(utxo)
		}
	}
}