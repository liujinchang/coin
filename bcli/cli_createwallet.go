package bcli

import (
	"coin"
	"fmt"
)

func (cli *CLI) createWallet() {
	wallets, _ := coin.FindWallets()
	address := wallets.CreateWallet()
	wallets.SaveToFile()
	fmt.Printf("Your new address: %s\n", address)
}
