package main

import (
	"fmt"
	"time"
)

//打印
func (cli *CLI) FmtBlockChain() {
	//创建迭代器
	it := cli.bc.NewIterator()
	//调用迭代器，返回每一个区块数据
	for {
		block := it.Next()

		fmt.Println("========================================")
		fmt.Printf("版本号：%d\n", block.Version)
		fmt.Printf("前区块哈希值：%x\n", block.PreHash)
		fmt.Printf("默克尔树根：%x\n", block.MerKerTreeRoot)
		fmt.Printf("当前区块哈希值：%x\n", block.NowHash)
		fmt.Printf("区块数据：%s\n", block.Transactions[0].TXInputs[0].PubKey)
		timeFormat := time.Unix(int64(block.TimeStamp), 0).Format("2006-01-02 15:04:05")
		fmt.Printf("时间戳：%s\n", timeFormat)

		if len(block.PreHash) == 0 {
			fmt.Printf("区块链遍历结束")
			break
		}
	}
}

//获取地址的余额
func (cli *CLI) GetBalance(address string) {

	//1.校验地址
	if !IsValidAddress(address) {
		fmt.Printf("地址无效：%s\n", address)
		return
	}
	//2.生成公钥哈希
	pubKeyHash := GetPubKeyFromAddress(address)

	utxos := cli.bc.FindUTXO(pubKeyHash)
	total := 0.0
	for _, utxo := range utxos {
		total += utxo.Value
	}
	fmt.Printf("\"%s\"余额为：%f\n", address, total)
}

//发送交易
func (cli *CLI) Send(from, to string, amount float64, miner, data string) {

	//1.校验地址
	if !IsValidAddress(from) {
		fmt.Printf("from地址无效：%s\n", from)
		return
	}
	if !IsValidAddress(to) {
		fmt.Printf("to地址无效：%s\n", to)
		return
	}
	if !IsValidAddress(miner) {
		fmt.Printf("miner地址无效：%s\n", miner)
		return
	}

	//1.创建挖矿交易
	coinbase := NewCoinbaseTX(miner, data)
	//2.创建一个普遍交易
	tx := NewTransaction(from, to, amount, cli.bc)
	//3.添加到区块
	cli.bc.AddBlock([]*Transaction{coinbase, tx})
	fmt.Printf("转账结束！\n")
}

//创建一个新的钱包

func (cli *CLI) NewWallet() {
	ws := NewWallets()
	address := ws.CreateWallet()
	fmt.Printf("地址：%s\n", address)
}

func (cli *CLI) listAddresses() {
	ws := NewWallets()
	addresses := ws.ListAddresses()
	for _, address := range addresses {
		fmt.Printf("地址：%s\n", address)
	}
}
