package main

import (
	"fmt"
	"os"
	"strconv"
)

//这是一个用来接受命令行参数并且控制区块链操作的文件

type CLI struct {
	bc *BlockChain
}

const Usage = `
	printChain 				   "打印区块链"
	getBalance --address ADDRESS "获取指定地址的余额"
	send FROM TO AMOUNT MINER DATA   "由FROM转AMOUNT给TO，由MINER挖矿，同时写入DATA"
	newWallet 	"创建一个钱包（私钥、公钥对）"
	listAddresses "列举所有的钱包地址"
`

//接受参数的动作，我们放在一个函数中
func (cli *CLI) Run() {
	//1.得到所有的命令
	args := os.Args
	if len(args) < 2 {
		fmt.Printf(Usage)
		return
	}

	//2.分析命令
	cmd := args[1]
	switch cmd {
	//3.执行相应的命令
	case "printChain":
		//打印区块
		//fmt.Printf("打印区块")
		cli.FmtBlockChain()
	case "getBalance":
		//获取余额
		if len(args) == 4 && args[2] == "--address" {
			address := args[3]
			cli.GetBalance(address)
		} else {
			fmt.Println("获取余额参数使用不当，请自查！")
			fmt.Println(Usage)
		}
	case "send":
		fmt.Printf("转账开始...\n")
		if len(args) != 7 {
			fmt.Printf("参数个数错误，请检查！\n")
			fmt.Printf(Usage)
			return
		}
		//.block send FROM TO AMOUNT MINER DATA   "由FROM转AMOUNT给TO，由MINER挖矿，同时写入DATA"
		from := args[2]
		to := args[3]
		amount, _ := strconv.ParseFloat(args[4], 64)
		miner := args[5]
		data := args[6]
		cli.Send(from, to, amount, miner, data)
	case "newWallet":
		//fmt.Printf("创建一个新的钱包")
		cli.NewWallet()
	case "listAddresses":
		//打印区块
		//fmt.Printf("打印钱包地址")
		cli.listAddresses()
	default:
		fmt.Printf("出错了")
		fmt.Printf(Usage)
	}

}
