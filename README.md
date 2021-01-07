这是一个区块链demo,采用的语言是go语言，只是用来学习！<br>
参考自https://github.com/Jeiwan/blockchain_go,在此基础进行扩展！<br>

实现的功能<br>
1、创世区块的创建;<br>
2、工作量证明<br>
3、客户端命令<br>
	createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS<br>
	createwallet - Generates a new key-pair and saves it into the wallet file<br>
	getbalance -address ADDRESS - Get balance of ADDRESS<br>
	listaddresses - Lists all addresses from the wallet file<br>
	printchain - Print all the blocks of the blockchain<br>
	reindexutxo - Rebuilds the UTXO set<br>
	send -from FROM -to TO -amount AMOUNT -mine - Send AMOUNT of coins from FROM address to TO. Mine on the same node, when -mine is set.<br>
	startnode -miner ADDRESS - Start a node with ID specified in NODE_ID env. var. -miner enables mining<br>
	showblock -hash HASH - Show block message<br>
	listunspent - Lists all unspend output<br>
4、交易签名<br>
5、打包交易生成新的区块链<br>

后面要实现的功能<br>
1、动态调整计算挖矿难度，保证10分钟来生成一个块<br>
2、把交易放到内存池，内存池放到数据库中<br>
3、增加挖矿奖励<br>
4、增加矿工选择性打包交易<br>
5、数据库文件分文件存储<br>
6、区块链状态分文件存储<br>
7、地址发现策略，把发现的地址存入数据库中<br>
8、丰富命令<br>
