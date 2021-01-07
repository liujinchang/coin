package bcli

import (
	"coin"
	"flag"
	"fmt"
	"os"
	"strings"
	"utils"
)
// CLI responsible for processing command line arguments
type CLI struct{}
func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  createwallet - Generates a new key-pair and saves it into the wallet file")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  listaddresses - Lists all addresses from the wallet file")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  reindexutxo - Rebuilds the UTXO set")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT -mine - Send AMOUNT of coins from FROM address to TO. Mine on the same node, when -mine is set.")
	fmt.Println("  startnode -miner ADDRESS - Start a node with ID specified in NODE_ID env. var. -miner enables mining")
	fmt.Println("  showblock -hash HASH - Show block message")
	fmt.Println("  listunspent - Lists all unspend output")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

// Run parses command line arguments and processes commands
func (cli *CLI) Run() {
	prepareEnv()
	cli.validateArgs()
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	reindexUTXOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)
	showBlockCmd := flag.NewFlagSet("showblock", flag.ExitOnError)
	listunspentCmd := flag.NewFlagSet("listunspent", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")
	sendMine := sendCmd.Bool("mine", false, "Mine immediately on the same node")
	startNodeMiner := startNodeCmd.String("miner", "", "Enable mining mode and send reward to ADDRESS")
	showBlockHash := showBlockCmd.String("hash", "", "The hash is hash of a block")

	switch os.Args[1] {
		case "getbalance":
			err := getBalanceCmd.Parse(os.Args[2:])
			utils.ErrorLog(err)
		case "createblockchain":
			err := createBlockchainCmd.Parse(os.Args[2:])
			utils.ErrorLog(err)
		case "createwallet":
			err := createWalletCmd.Parse(os.Args[2:])
			utils.ErrorLog(err)
		case "listaddresses":
			err := listAddressesCmd.Parse(os.Args[2:])
			utils.ErrorLog(err)
		case "printchain":
			err := printChainCmd.Parse(os.Args[2:])
			utils.ErrorLog(err)
		case "reindexutxo":
			err := reindexUTXOCmd.Parse(os.Args[2:])
			utils.ErrorLog(err)
		case "send":
			err := sendCmd.Parse(os.Args[2:])
			utils.ErrorLog(err)
		case "startnode":
			err := startNodeCmd.Parse(os.Args[2:])
			utils.ErrorLog(err)
		case "showblock":
			err := showBlockCmd.Parse(os.Args[2:])
			utils.ErrorLog(err)
		case "listunspent":
			err := listunspentCmd.Parse(os.Args[2:])
			utils.ErrorLog(err)
		default:
			cli.printUsage()
			os.Exit(1)
	}
	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}
	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress)
	}
	if createWalletCmd.Parsed() {
		cli.createWallet()
	}
	if listAddressesCmd.Parsed() {
		cli.listAddresses()
	}
	if printChainCmd.Parsed() {
		cli.printChain()
	}
	if reindexUTXOCmd.Parsed() {
		cli.reindexUTXO()
	}
	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}
		cli.send(*sendFrom, *sendTo, *sendAmount, *sendMine)
	}
	if startNodeCmd.Parsed() {
		cli.startNode(*startNodeMiner)
	}
	if showBlockCmd.Parsed() {
		if *showBlockHash == "" {
			showBlockCmd.Usage()
			os.Exit(1)
		}
		cli.showBlock(*showBlockHash)
	}
	if listunspentCmd.Parsed() {
		cli.listUnspend()
	}
}
func prepareEnv() string {
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		fmt.Printf("NODE_ID env. var is not set!")
		os.Exit(1)
	}
	if !utils.FileExists(coin.Root) {
		err := os.Mkdir(coin.Root,0600)
		utils.ErrorLog(err)
	}
	//创建配置文件夹
	var builder strings.Builder
	builder.WriteString(coin.Root)
	builder.WriteString("/config")
	dir := builder.String()
	if !utils.FileExists(dir) {
		err := os.Mkdir(dir,0600)
		utils.ErrorLog(err)
	}
	//创建数据库文件夹(用于存放区块数据库，区块状态（UTXO）数据库)
	builder.Reset()
	builder.WriteString(coin.Root)
	builder.WriteString("/database")
	dir = builder.String()
	if !utils.FileExists(dir) {
		err := os.Mkdir(dir,0600)
		utils.ErrorLog(err)
	}
	//创建钱包文件夹
	builder.Reset()
	builder.WriteString(coin.Root)
	builder.WriteString("/wallet")
	dir = builder.String()
	if !utils.FileExists(dir) {
		err := os.Mkdir(dir,0600)
		utils.ErrorLog(err)
	}
	return nodeID
}
