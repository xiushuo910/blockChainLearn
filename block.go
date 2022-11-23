package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"time"
)

//定义区块
type Block struct {
	//版本号，前hash，默克尔树根，nonce，困难值，时间戳           当前hash，数据
	Version        uint64
	PreHash        []byte
	MerKerTreeRoot []byte
	Nonce          uint64
	Difficulty     uint64
	TimeStamp      uint64

	NowHash []byte
	//Data    []byte
	//真是的交易数组
	Transactions []*Transaction
}

//创建区块
func NewBlock(txs []*Transaction, preHsh []byte) *Block {
	block := Block{
		Version:        1,
		PreHash:        preHsh,
		MerKerTreeRoot: []byte{},
		Nonce:          0,
		Difficulty:     0,
		TimeStamp:      uint64(time.Now().Unix()),

		//Data: data,
		Transactions: txs,
	}
	//创建一个pow对象
	pow := newProofOfWork(&block)
	//查找随机数，不停的进行hash运算
	hash, nonce := pow.run()
	block.NowHash = hash
	block.Nonce = nonce
	block.MerKerTreeRoot = block.MakeMerkelTreeRoot()

	return &block
}

//添加区块
func (blockChain *BlockChain) AddBlock(txs []*Transaction) {
	for i, tx := range txs {
		//铸币交易不用验证
		if i > 0 {
			if !blockChain.VerifyTransaction(tx) {
				fmt.Printf("矿工发现无效失败！")
				return
			}
		}
	}

	//获取前区块hash
	db := blockChain.db
	lastHash := blockChain.tail
	db.Update(func(tx *bolt.Tx) error {
		//完成数据添加
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket == nil {
			log.Panic("bucket 不应该为空，请检查！")
		}
		block := NewBlock(txs, lastHash)

		//hash作为key，block的字节流作为value，尚未实现
		bucket.Put(block.NowHash, block.Serialize())
		bucket.Put([]byte("LastHashKey"), block.NowHash)
		lastHash = block.NowHash
		//更新一下内存中的区块链，指的是把最后的小尾巴tail更新一下
		blockChain.tail = block.NowHash
		return nil
	})
}

//uint64ToByte
func uint64ToByte(num uint64) []byte {
	var buffer bytes.Buffer
	err := binary.Write(&buffer, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}
	return buffer.Bytes()
}

//生成hash
func (block *Block) setHash() {

	tmp := [][]byte{
		uint64ToByte(block.Version),
		block.PreHash,
		block.MerKerTreeRoot,
		uint64ToByte(block.Nonce),
		uint64ToByte(block.Difficulty),
		uint64ToByte(block.TimeStamp),
		//block.Data,
	}
	blockInfo := bytes.Join(tmp, []byte{})
	hash := sha256.Sum256(blockInfo)
	block.NowHash = hash[:]
}

//模拟使用txs生成MerKerTreeRoot（简单拼接）
func (block *Block) MakeMerkelTreeRoot() []byte {
	var info []byte
	for _, tx := range block.Transactions {
		//将交易的哈希值拼接起来，再整体做hash
		info = append(info, tx.TXID...)
	}
	hash := sha256.Sum256(info)
	return hash[:]
}

//编码(序列化)
func (block *Block) Serialize() []byte {
	var buffer bytes.Buffer
	//1.定义一个编码器
	encoder := gob.NewEncoder(&buffer)
	//2.使用编码器编码
	err := encoder.Encode(&block)
	if err != nil {
		log.Panic("编码出错了！")
	}
	return buffer.Bytes()
}

//解码(反序列化)
func Deserialize(data []byte) Block {
	//1.定义一个解码器
	decoder := gob.NewDecoder(bytes.NewReader(data))
	var block Block
	//2.使用解码器解码
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic("解码器出错！")
	}

	return block
}
