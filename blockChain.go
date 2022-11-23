package main

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	_ "github.com/boltdb/bolt"
	"log"
)

//定义区块链
type BlockChain struct {
	//blocks []*Block
	//用bolt数据库改写
	db   *bolt.DB
	tail []byte //存储最后一个区块的哈希
}

const blockChainDb = "blockChain.db"
const blockBucket = "blockBucket"

//初始化区块链
func NewBlockChain() *BlockChain {
	//return &BlockChain{
	//	blocks: []*Block{genisisBlock},
	var lastHash []byte
	//1.打开数据库
	db, err := bolt.Open(blockChainDb, 0600, nil)
	//defer db.Close()
	if err != nil {
		log.Panic("打开数据库失败！")
	}
	//将要操作数据库（改写）
	db.Update(func(tx *bolt.Tx) error {
		//2.找到抽屉ducket（如果没有就创建）
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket == nil {
			bucket, err = tx.CreateBucket([]byte(blockBucket))
			if err != nil {
				log.Panic("创建bucket(b1)失败")
			}
			//创建一个创世区块，并作为第一个区块添加到区块链
			wallet := NewWallets()
			address := wallet.CreateWallet()
			genisisBlock := GenisisBlock(address)
			//3.写数据
			//hash作为key，block的字节流作为value
			bucket.Put(genisisBlock.NowHash, genisisBlock.Serialize())
			bucket.Put([]byte("LastHashKey"), genisisBlock.NowHash)
			lastHash = genisisBlock.NowHash
			fmt.Printf("使用了铸币交易")
		} else {
			lastHash = bucket.Get([]byte("LastHashKey"))
		}
		return nil
	})
	return &BlockChain{db, lastHash}
}

//创建创世区块
func GenisisBlock(address string) *Block {
	coinbase := NewCoinbaseTX(address, "我是第一个块")
	block := NewBlock([]*Transaction{coinbase}, []byte{})
	block.setHash()
	return block
}

func (blockChain *BlockChain) FindUTXO(senderPubKeyHash []byte) []TXOutput {
	var UTXO []TXOutput
	//我们定义一个map来保存消费过的output，key是这个output的交易id，value是这个交易中索引的数组

	txs := blockChain.FindUTXOTransaction(senderPubKeyHash)

	for _, tx := range txs {
		for _, output := range tx.TXOutputs {
			if bytes.Equal(senderPubKeyHash, output.PubKeyHash) {
				UTXO = append(UTXO, output)
			}
		}
	}

	return UTXO
}

func (blockChain *BlockChain) FindNeedUTXOs(senderPubKeyHash []byte, amount float64) (map[string][]uint64, float64) {
	//找到合理的utxos集合
	utxos := make(map[string][]uint64)
	//我们定义一个map来保存消费过的output，key是这个output的交易id，value是这个交易中索引的数组
	var calc float64

	txs := blockChain.FindUTXOTransaction(senderPubKeyHash)

	for _, tx := range txs {
		for i, output := range tx.TXOutputs {
			//两个[]byte比较
			if bytes.Equal(senderPubKeyHash, output.PubKeyHash) {
				if calc < amount {
					//1.把utxo加进来
					//array := utxos[string(tx.TXID)]
					//array = append(array, uint64(i))
					utxos[string(tx.TXID)] = append(utxos[string(tx.TXID)], uint64(i))
					//2.统计一下当前utxo得总额
					calc += output.Value

					//3.比较一下是否满足转账需求
					//	a.满足的话，直接返回 utxos，calc
					//  b.不满足继续统计
					if calc >= amount {
						return utxos, calc
					}
				} else {
					fmt.Printf("不满足转账金额。当前金额：%f,目前金额：%f\n", calc, amount)
				}
			}
		}
	}

	return utxos, calc
}

func (blockChain *BlockChain) FindUTXOTransaction(senderPubKeyHash []byte) []*Transaction {
	//var UTXO []TXOutput
	//存储所有包含utxo交易集合
	var txs []*Transaction
	//我们定义一个map来保存消费过的output，key是这个output的交易id，value是这个交易中索引的数组
	spendOutputs := make(map[string][]int64)
	//1.遍历区块
	//2.遍历交易
	//3.遍历output，找到和自己相关的utxo（在添加output之前检查一下是否已经消耗过）
	//4.遍历input，找到自己花费过的utxo的集合（把自己小号过的标识出来）

	//创建迭代器
	it := blockChain.NewIterator()
	for {
		//1.遍历区块
		block := it.Next()

		//2.遍历交易
		for _, tx := range block.Transactions {

			//3.遍历output
		OUTPUT:
			for i, output := range tx.TXOutputs {
				//这个output和我们目标的地址相同，满足条件，加到返回UTXO的数组中
				//在这里进行过滤，将所有消耗过的outputs和当前的所即将添加的output对比
				//如果相同，则跳过，否则添加
				if spendOutputs[string(tx.TXID)] != nil {
					for _, j := range spendOutputs[string(tx.TXID)] {
						if int64(i) == j {
							//当前准备添加output已经消耗过了，不再添加
							continue OUTPUT
						}
					}
				}
				if bytes.Equal(output.PubKeyHash, senderPubKeyHash) {
					//UTXO = append(UTXO, output)
					txs = append(txs, tx)
				}
			}
			//如果当前交易是挖矿交易的话，那么不做遍历，直接跳过
			if !tx.IsCoinbase() {
				//4.遍历input
				for _, input := range tx.TXInputs {
					//判断一下当前这个input和目标是否一致，如果相同就是已经被花费，加入map
					if bytes.Equal(HashPubKey(input.PubKey), senderPubKeyHash) {
						spendOutputs[string(input.TXid)] = append(spendOutputs[string(input.TXid)], input.Index)
					}
				}
			}
		}
		if len(block.PreHash) == 0 {
			break
			fmt.Printf("区块链遍历完成退出!")
		}

	}
	return txs
}

//根据id查找交易本身，需要遍历整个区块链
func (bc *BlockChain) FindTransactionByTXid(id []byte) (Transaction, error) {
	it := bc.NewIterator()
	//1.遍历区块链
	for {
		block := it.Next()
		//2.遍历交易
		for _, tx := range block.Transactions {
			//3.比较交易，找到了直接退出
			if bytes.Equal(tx.TXID, id) {
				return *tx, nil
			}
		}
		if len(block.PreHash) == 0 {
			break
			fmt.Printf("区块链遍历结束！\n")
		}

	}
	//4.如果没找到，返回空Transaction，同时返回错误状态
	return Transaction{}, errors.New("无效的交易id，请自查!")
}

func (bc *BlockChain) SignTransaction(tx *Transaction, privateKey *ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)
	//找到所有的input交易
	//1.根据inputs来找，有多少input，就遍历多少次
	//2.找到目标交易，（根据TXid来找）
	//3.添加到prevTXs
	for _, input := range tx.TXInputs {
		//根据TXid去找交易,需要遍历所有的区块链
		tx, err := bc.FindTransactionByTXid(input.TXid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[string(input.TXid)] = tx
	}

	tx.Sign(privateKey, prevTXs)
}

func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	prevTXs := make(map[string]Transaction)
	//找到所有的input交易
	//1.根据inputs来找，有多少input，就遍历多少次
	//2.找到目标交易，（根据TXid来找）
	//3.添加到prevTXs
	for _, input := range tx.TXInputs {
		//根据TXid去找交易,需要遍历所有的区块链
		tx, err := bc.FindTransactionByTXid(input.TXid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[string(input.TXid)] = tx
	}
	return tx.Verify(prevTXs)
}
