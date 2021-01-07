package bcli

import (
	"coin"
	"fmt"
	"utils"
)

func (cli *CLI) listAddresses() {
	wallets, err := coin.FindWallets()
	utils.ErrorLog(err)
	addresses := wallets.GetAddresses()
	for _, address := range addresses {
		fmt.Println(address)
	}
}
