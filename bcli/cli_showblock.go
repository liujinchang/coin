package bcli

import (
	"coin"
	"encoding/hex"
	"fmt"
)

func (cli *CLI) showBlock(hash string) {
	bc := coin.FindBlockchain(true)
	defer bc.GetDB().Close()
	//把hash字符串解码为[]byte类型hash
	b, err := hex.DecodeString(hash)
	if err != nil {
		fmt.Println("The hash is a valid hash!")
	} else {
		block, _ := bc.GetBlock(b)
		fmt.Println(block)
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
	}
}
