package bcli

import (
	"coin"
	"fmt"
	"log"
	"os"
)

func (cli *CLI) startNode(minerAddress string) {
	fmt.Printf("Starting node %s\n", os.Getenv("NODE_ID"))
	if len(minerAddress) > 0 {
		if coin.ValidateAddress(minerAddress) {
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
		} else {
			log.Panic("Wrong miner address!")
		}
	}
	coin.StartServer(minerAddress)
}
