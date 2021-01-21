package coin

import (
	"bytes"
	"crypto/sha256"
	"math"
	"math/big"
	"utils"
)

var maxNonce = math.MaxInt64

// ProofOfWork represents a proof-of-work
type ProofOfWork struct {
	block  *Block
	target *big.Int
}
// NewProofOfWork builds and returns a ProofOfWork
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	pow := &ProofOfWork{b, target}
	return pow
}
func (pow *ProofOfWork) prepareData(nonce int, isValid bool) []byte {
	var data []byte
	/*
	 * when valid, renew create hash of transactions in the block!
	 * when mine, use hash created
	 */
	if isValid {
		data = bytes.Join(
			[][]byte{
				pow.block.PrevBlockHash,
				pow.block.HashTransactions(),
				utils.IntToHex(pow.block.Timestamp),
				utils.IntToHex(int64(targetBits)),
				utils.IntToHex(int64(nonce)),
			},
			[]byte{},
		)
	} else {
		data = bytes.Join(
			[][]byte{
				pow.block.PrevBlockHash,
				pow.block.MerkleTreeRootHash,
				utils.IntToHex(pow.block.Timestamp),
				utils.IntToHex(int64(targetBits)),
				utils.IntToHex(int64(nonce)),
			},
			[]byte{},
		)
	}
	return data
}

// Run performs a proof-of-work
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0
	for nonce < maxNonce {
		data := pow.prepareData(nonce,false)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])
		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	return nonce, hash[:]
}
// Validate validates block's PoW
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int
	data := pow.prepareData(pow.block.Nonce,true)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	isValid := hashInt.Cmp(pow.target) == -1
	return isValid
}
