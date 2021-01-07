package bcli

import (
	"coin"
	"fmt"
	"log"
	"utils"
)

func (cli *CLI) send(from, to string, amount int, mineNow bool) {
	if !coin.ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !coin.ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}
	bc := coin.FindBlockchain(true)
	UTXOSet := coin.UTXOSet{Blockchain: bc}
	UTXOSet.Init()
	defer bc.GetDB().Close()
	defer UTXOSet.GetDB().Close()
	wallets, err := coin.FindWallets()
	utils.ErrorLog(err)
	wallet := wallets.GetWallet(from)
	tx := coin.NewUTXOTransaction(&wallet, to, amount, &UTXOSet)
	if mineNow {
		cbTx := coin.NewCoinbaseTX(from, "")
		txs := []*coin.Transaction{cbTx, tx}
		newBlock := bc.MineBlock(txs)
		UTXOSet.Update(newBlock)
	} else {
		coin.SendTx("", tx)
	}
	fmt.Println("Success!")
}