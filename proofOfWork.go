package main

import (
	"bytes"
	"crypto/sha256"
	"math/big"
)

//定义ProofOfWork
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

//创建pow的函数
func newProofOfWork(block *Block) *ProofOfWork {
	pow := ProofOfWork{
		block: block,
	}
	//targetStr为指定的难度值
	targetStr := "0000100000000000000000000000000000000000000000000000000000000000"
	tmpInt := big.Int{}
	tmpInt.SetString(targetStr, 16)
	pow.target = &tmpInt
	return &pow
}

//提供计算不断计算hash,返回hash和nonce
func (pow *ProofOfWork) run() ([]byte, uint64) {
	//拼装数据（区块的数据，nonce）
	var nonce uint64
	var hash [32]byte
	block := pow.block

	for {
		//拼装数据并hash
		tmp := [][]byte{
			uint64ToByte(block.Version),
			block.PreHash,
			block.MerKerTreeRoot,
			uint64ToByte(nonce),
			uint64ToByte(block.Difficulty),
			uint64ToByte(block.TimeStamp),
			//block.Data,
		}
		blockInfo := bytes.Join(tmp, []byte{})
		hash = sha256.Sum256(blockInfo)

		//与pow中的target 比较
		tmpInt := big.Int{}
		tmpInt.SetBytes(hash[:])
		if tmpInt.Cmp(pow.target) == -1 {
			//tmpInt < pow.target //招到了
			break
		} else {
			nonce++
		}
	}

	return hash[:], nonce
}
