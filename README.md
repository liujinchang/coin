这是一个区块链demo,采用的语言是go语言，只是用来学习！<br>
参考自https://github.com/Jeiwan/blockchain_go,在此基础进行扩展！<br>

实现的功能<br>

已完成的工作：<br>
1、创世区块的创建;<br>
2、工作量证明<br>
3、客户端命令<br>
```
	createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS
	createwallet - Generates a new key-pair and saves it into the wallet file
	getbalance -address ADDRESS - Get balance of ADDRESS
	listaddresses - Lists all addresses from the wallet file
	printchain - Print all the blocks of the blockchain
	reindexutxo - Rebuilds the UTXO set
	send -from FROM -to TO -amount AMOUNT -mine - Send AMOUNT of coins from FROM address to TO. Mine on the same node, when -mine is set.
	startnode -miner ADDRESS - Start a node with ID specified in NODE_ID env. var. -miner enables mining
	showblock -hash HASH - Show block message
	listunspent - Lists all unspend output
```
4、交易签名<br>
5、打包交易生成新的区块链<br>
6、增加配置文件，从配置文件读取相关信息<br>
7、把交易放到内存池，内存池放到数据库中，内存中只存一定数量的交易，其它的存入数据库中<br>
8、增加梅克尔树的树根hash到区块中<br>
9、控制区块的大小，区块中最多存放100个交易<br>
10、每个块生成后，更新未花费的交易输出<br>
11、每个块生成后，就开始在本地挖矿<br>
12、数据库文件分文件存储<br>


未完成的工作：<br>
1、动态调整计算挖矿难度，保证10分钟来生成一个块<br>
2、增加交易费用<br>
3、增加矿工选择性打包交易<br>
4、地址发现策略，把发现的地址存入数据库中<br>
5、丰富命令<br>
6、控制挖矿的开始时间，当获长的链时就开始挖矿，当收到新块时，校验通过则停止当前的挖矿，开发挖下一下区块<br>