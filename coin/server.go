package coin

import (
	"bytes"
	"config"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"utils"
)

var nodeAddress string
var miningAddress string
var knownNodes []string
var blocksInTransit = [][]byte{}
var mempool pool
var utxoSet *UTXOSet
var stateFile = fmt.Sprintf(config.Root+"/database/"+StateFile, nodeID)

type pool struct {
	//in memory
	pool 	map[string]Transaction
	//in db
	mp 		*Mempool
}
type addr struct {
	AddrList []string
}

type block struct {
	AddrFrom string
	Block    []byte
}

type getblocks struct {
	AddrFrom string
}

type getdata struct {
	AddrFrom string
	Type     string
	ID       []byte
}

type inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

type tx struct {
	AddFrom     string
	Transaction []byte
}

type verzion struct {
	Version    int
	BestHeight int
	AddrFrom   string
}

func commandToBytes(command string) []byte {
	var bytes [commandLength]byte
	for i, c := range command {
		bytes[i] = byte(c)
	}
	return bytes[:]
}

func bytesToCommand(bytes []byte) string {
	var command []byte
	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}
	return fmt.Sprintf("%s", command)
}

func requestBlocks() {
	for _, node := range knownNodes {
		sendGetBlocks(node)
	}
}

func sendAddr(address string) {
	nodes := addr{knownNodes}
	nodes.AddrList = append(nodes.AddrList, nodeAddress)
	payload := utils.GobEncode(nodes)
	request := append(commandToBytes("addr"), payload...)
	sendData(address, request)
}

func sendBlock(addr string, b *Block) {
	data := block{nodeAddress, utils.GobEncode(b)}
	payload := utils.GobEncode(data)
	request := append(commandToBytes("block"), payload...)
	sendData(addr, request)
}

func sendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		log.Printf("%s is not available\n", addr)
		var updatedNodes []string
		for _, node := range knownNodes {
			if node != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}
		knownNodes = updatedNodes
		return
	}
	defer conn.Close()
	_, err = io.Copy(conn, bytes.NewReader(data))
	utils.ErrorLog(err)
}

func sendInv(address, kind string, items [][]byte) {
	inventory := inv{nodeAddress, kind, items}
	payload := utils.GobEncode(inventory)
	request := append(commandToBytes("inv"), payload...)
	sendData(address, request)
}

func sendGetBlocks(address string) {
	payload := utils.GobEncode(getblocks{nodeAddress})
	request := append(commandToBytes("getblocks"), payload...)
	sendData(address, request)
}

func sendGetData(address, kind string, id []byte) {
	payload := utils.GobEncode(getdata{nodeAddress, kind, id})
	request := append(commandToBytes("getdata"), payload...)
	sendData(address, request)
}

func SendTx(addr string, tnx *Transaction) {
	if nodeAddress == "" {
		nodeAddress = fmt.Sprintf("localhost:%s", os.Getenv("NODE_ID"))
	}
	data := tx{nodeAddress, utils.GobEncode(tnx)}
	payload := utils.GobEncode(data)
	request := append(commandToBytes("tx"), payload...)
	if addr == "" {
		addr = knownNodes[0]
	}
	sendData(addr, request)
}

func sendVersion(addr string, bc *Blockchain) {
	var payload []byte
	//当本地无区块链时
	if bc.db == nil {
		payload = utils.GobEncode(verzion{nodeVersion, -1, nodeAddress})
	} else {
		bestHeight := bc.GetBestHeight()
		payload = utils.GobEncode(verzion{nodeVersion, bestHeight, nodeAddress})
	}
	request := append(commandToBytes("version"), payload...)
	sendData(addr, request)
}

func handleAddr(request []byte) {
	var buff bytes.Buffer
	var payload addr

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	utils.ErrorLog(err)
	knownNodes = append(knownNodes, payload.AddrList...)
	log.Printf("There are %d known nodes now!\n", len(knownNodes))
	requestBlocks()
}

func handleBlock(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload block

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	utils.ErrorLog(err)
	blockData := payload.Block
	block := DeserializeBlock(blockData)
	log.Println("Recevied a new block!")
	if bc.db == nil {
		dbFile := fmt.Sprintf(config.Root + "/database/" + dbFileName, nodeID, utils.TransformFileIndex(bc.dbFileIndex))
		db, err := bolt.Open(dbFile, 0600, nil)
		utils.ErrorLog(err)
		err = db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucket([]byte(blocksBucket))
			utils.ErrorLog(err)
			return nil
		})
		utils.ErrorLog(err)
		bc.db = db
	}
	bc.AddBlock(block)
	//应该在每增加一个块时进行UTXOSet.update(block)
	if utxoSet == nil {
		utxoSet = &UTXOSet{bc,utils.FindDB(stateFile)}
	}
	utxoSet.Update(block)
	log.Printf("Added block %x\n", block.Hash)
	if len(blocksInTransit) > 0 {
		var bh []byte
		for _, blockHash := range blocksInTransit {
			//只请求本结点不存在的块
			if !bc.Exist(blockHash) {
				sendGetData(payload.AddrFrom, "block", blockHash)
				bh = blockHash
				break
			}
		}
		var updateBlockInTransit [][]byte
		if bh != nil {
			for _, blockHash := range blocksInTransit {
				if bytes.Compare(bh, blockHash) != 0 {
					updateBlockInTransit = append(updateBlockInTransit, blockHash)
				}
			}
		}
		blocksInTransit = updateBlockInTransit
	} else {
		//when receive new block, start mine!
		if len(miningAddress) > 0 {
		MineTransactions:
			var txs []*Transaction
			counter := 0
			for id := range mempool.pool {
				counter++
				if counter > MaxTransactionCount {
					break
				} else {
					tx := mempool.pool[id]
					if bc.VerifyTransaction(&tx) {
						txs = append(txs, &tx)
					}
				}
			}
			cbTx := NewCoinbaseTX(miningAddress, "")
			txs = append(txs, cbTx)
			newBlock := bc.MineBlock(txs)
			utxoSet.Update(newBlock)
			log.Println("New block is mined!")
			for _, tx := range txs {
				txID := hex.EncodeToString(tx.ID)
				delete(mempool.pool, txID)
			}
			for _, node := range knownNodes {
				if node != nodeAddress {
					sendInv(node, "block", [][]byte{newBlock.Hash})
				}
			}
			if len(mempool.pool) > 0 {
				goto MineTransactions
			}
		}
		//defer utxoSet.DB.Close()
	}
}

func handleInv(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload inv

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	utils.ErrorLog(err)
	log.Printf("Recevied inventory with %d %s\n", len(payload.Items), payload.Type)

	if payload.Type == "block" {
		//从连接结点拿到所有块，过滤到本地存在的块，即是本地缺失的块
		blocksInTransit = payload.Items
		newInTransit := [][]byte{}
		if bc.db != nil {
			for _, b := range blocksInTransit {
				if !bc.Exist(b) {
					newInTransit = append(newInTransit, b)
				}
			}
			blocksInTransit = newInTransit
		}
		//传送的块hash是从最后一个块到第一个块，把顺序倒过来，这样我们就可以从缺失块的位置一个一个块按顺序加回来
		for i, j := 0, len(blocksInTransit)-1; i < j; i, j = i+1, j-1 {
			blocksInTransit[i], blocksInTransit[j] = blocksInTransit[j], blocksInTransit[i]
		}
		blockHash := blocksInTransit[0]
		sendGetData(payload.AddrFrom, "block", blockHash)
		//去掉刚刚请求的块
		newInTransit = [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}

	if payload.Type == "tx" {
		txID := payload.Items[0]
		if mempool.pool[hex.EncodeToString(txID)].ID == nil {
			sendGetData(payload.AddrFrom, "tx", txID)
		}
	}
}

func handleGetBlocks(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload getblocks

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	utils.ErrorLog(err)
	blocks := bc.GetBlockHashes()
	sendInv(payload.AddrFrom, "block", blocks)
}

func handleGetData(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload getdata

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	utils.ErrorLog(err)
	if payload.Type == "block" {
		block, err := bc.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}
		sendBlock(payload.AddrFrom, &block)
	}
	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := mempool.pool[txID]
		SendTx(payload.AddrFrom, &tx)
		delete(mempool.pool, txID)
	}
}

func handleTx(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload tx
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	utils.ErrorLog(err)
	txData := payload.Transaction
	tx := DeserializeTransaction(txData)
	mempool.pool[hex.EncodeToString(tx.ID)] = tx
	//内存与数据库动态交换交易 保证内存中存放到的交易在10到15之间
	changeTransactions()
	if nodeAddress == knownNodes[0] {
		for _, node := range knownNodes {
			if node != nodeAddress && node != payload.AddFrom {
				sendInv(node, "tx", [][]byte{tx.ID})
			}
		}
	}
}

func handleVersion(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload verzion

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	utils.ErrorLog(err)
	var myBestHeight int
	//当本地无区块链时
	if bc.db == nil {
		myBestHeight = -1
	} else {
		myBestHeight = bc.GetBestHeight()
	}
	foreignerBestHeight := payload.BestHeight
	if myBestHeight < foreignerBestHeight {
		sendGetBlocks(payload.AddrFrom)
	} else if myBestHeight > foreignerBestHeight {
		sendVersion(payload.AddrFrom, bc)
	}
	// sendAddr(payload.AddrFrom)
	if !nodeIsKnown(payload.AddrFrom) {
		knownNodes = append(knownNodes, payload.AddrFrom)
	}
}

func handleConnection(conn net.Conn, bc *Blockchain) {
	request, err := ioutil.ReadAll(conn)
	utils.ErrorLog(err)
	command := bytesToCommand(request[:commandLength])
	log.Printf("Received %s command\n", command)
	switch command {
		case "addr":
			handleAddr(request)
		case "block":
			handleBlock(request, bc)
		case "inv":
			handleInv(request, bc)
		case "getblocks":
			handleGetBlocks(request, bc)
		case "getdata":
			handleGetData(request, bc)
		case "tx":
			handleTx(request, bc)
		case "version":
			handleVersion(request, bc)
		default:
			log.Println("Unknown command!")
	}
	conn.Close()
}

// StartServer starts a node
func StartServer(minerAddress string) {
	mempool = pool{make(map[string]Transaction),Mempool{}.ReBuild()}
	knownNodes = strings.Split(config.GetConfig("known_nodes"),",")
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	miningAddress = minerAddress
	ln, err := net.Listen(protocol, nodeAddress)
	utils.ErrorLog(err)
	defer ln.Close()
	bc := FindBlockchain(false)
	if nodeAddress != knownNodes[0] {
		sendVersion(knownNodes[0], bc)
	}
	for {
		conn, err := ln.Accept()
		utils.ErrorLog(err)
		go handleConnection(conn, bc)
	}
}

func nodeIsKnown(addr string) bool {
	for _, node := range knownNodes {
		if node == addr {
			return true
		}
	}
	return false
}
/*
 * 内存与数据库交换交易 保证内存中交易数量保持在上限值下限值之间
 * 		上下限的阀值，表示在内存中存放交易的数量
 * 			当内存中存储的交易大于上限值时，把len(mempool.pool)-上限值+int((上限值-下限值)/2)个数据从内存放入到数据库中
 * 			当内存中存储的交易小于下限值时，把下限值-len(mempool.pool)+int((上限值-下限值)/2)个数据从数据库中加载进内存
 */
func changeTransactions(){
	length := len(mempool.pool)
	upperLimitTransactionInMemory, err := strconv.Atoi(config.GetConfig(config.UpperLimitTransactionInMemory))
	utils.ErrorLog(err)
	lowerLimitTransactionInMemory, err := strconv.Atoi(config.GetConfig(config.LowerLimitTransactionInMemory))
	utils.ErrorLog(err)
	half := int((upperLimitTransactionInMemory-lowerLimitTransactionInMemory)/2)
	if length > upperLimitTransactionInMemory {
		var transactions []Transaction
		counter := len(mempool.pool)-upperLimitTransactionInMemory+half
		for id := range mempool.pool {
			transactions = append(transactions, mempool.pool[id])
			counter--
			if counter == 0 {
				break
			}
		}
		mempool.mp.AddTransactions(transactions)
	} else if length < lowerLimitTransactionInMemory {
		counter := lowerLimitTransactionInMemory-len(mempool.pool)+half
		transactions := mempool.mp.FindTransactions(counter)
		for _, t := range transactions {
			mempool.pool[hex.EncodeToString(t.ID)] = t
		}
	}
}