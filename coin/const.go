package coin

const (
	//每个块中最大交易数量
	MaxTransactionCount = 10
	//区块数据库文件
	dbFileName = "blockchain_%s_%s.db"
	//区块状态(UTXO)数据库文件
	StateFile = "chainstate_%s.db"
	blocksBucket = "blocks"
	//创世区块中用来放到pubkey的值
	genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
	//工作量证明用来调整目标值
	targetBits = 24
	//数据传输协议
	protocol = "tcp"
	nodeVersion = 1
	commandLength = 12
	//数据库保存区块链状态的桶：1、放置未花费的交易输出；2、放置区块链数据文件的序号
	chainstateBucket = "chainstate"
	//地址版本，生成的地址的第一位字符为1
	version = byte(0x00)
	//地址检验和的位数
	addressChecksumLen = 4
	//钱包文件名
	walletFileName = "wallet_%s.dat"
	//内存池数据库文件名
	mempoolFileName = "mempool_%s.db"
	//内存池数据库桶
	mempoolBucket = "mempool"
	//在一个数据库文件中包含块的数量
	blockCountInFile =10
)