这是一个区块链demo,采用的语言是go语言，只是用来学习！
参考自https://github.com/Jeiwan/blockchain_go,在此基础进行扩展！

实现的功能
1、创世区块的创建;
2、工作量证明
3、客户端命令
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
4、交易签名
5、打包交易生成新的区块链

后面要实现的功能
1、动态调整计算挖矿难度，保证10分钟来生成一个块
2、把交易放到内存池，内存池放到数据库中
3、增加挖矿奖励
4、增加矿工选择性打包交易
5、数据库文件分文件存储
6、区块链状态分文件存储
7、地址发现策略，把发现的地址存入数据库中
8、丰富命令