package coin

// MerkleTree represent a Merkle tree
/*
 * 梅克尔树
 *		一种二叉树结构，每个非叶子节点，有两个子节点，叶子节点为偶数个
 * 		1、一组数据块，当数据块的个数为奇数个时，把最后一个数据块复制一份，添加到这组数据块中，最终形成偶数个数据块
 * 		2、计算出每个数据块的hash值，形成偶数个hash值，做为梅克尔树叶子节点
 * 		3、非叶子节点由两个子节点构成，节点值为两个子节点值的计算出来的Hash值
 * 		梅克尔树用于快速比较出两组数据的不一致，并通过节点值快速定位到不一致的数据
 */

type MerkleTree struct {
	RootNode *MerkleNode
}

// NewMerkleTree creates a new Merkle tree from a sequence of data
func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleNode
	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}
	for _, datum := range data {
		node := NewMerkleNode(nil, nil, datum)
		nodes = append(nodes, *node)
	}
	for i := 0; i < len(data)/2; i++ {
		var newLevel []MerkleNode
		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
			newLevel = append(newLevel, *node)
		}
		nodes = newLevel
	}
	mTree := MerkleTree{&nodes[0]}
	return &mTree
}
