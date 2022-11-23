package main

import (
	"github.com/boltdb/bolt"
	"log"
)

type BlockChainIterator struct {
	db *bolt.DB
	//游标，用于不断索引
	currentHshPointer []byte
}

func (blockChain *BlockChain) NewIterator() *BlockChainIterator {
	return &BlockChainIterator{
		blockChain.db,
		blockChain.tail,
	}
}

//迭代器属于区块链，next属于迭代器
//1.返回当前的区块
//2.指针前移
func (blockChainIterator *BlockChainIterator) Next() *Block {
	var block Block
	blockChainIterator.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket == nil {
			log.Panic("迭代器遍历时bucket不应该为空，请自查！")
		}
		blockTmp := bucket.Get(blockChainIterator.currentHshPointer)
		//解码
		block = Deserialize(blockTmp)
		//游标左移
		blockChainIterator.currentHshPointer = block.PreHash

		return nil
	})
	return &block
}
