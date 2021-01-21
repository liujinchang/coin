package bcli

import (
	"coin"
	"fmt"
)

func (cli *CLI) printChain() {
	bc := coin.FindBlockchain(true)
	defer bc.GetDB().Close()
	bci := bc.Iterator()
	for {
		block := bci.Next()
		fmt.Println(block)
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}
