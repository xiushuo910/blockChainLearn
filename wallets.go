package main

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"github.com/mr-tron/base58"
	"io/ioutil"
	"log"
	"os"
)

const walletFile = "wallet.dat"

//定义一个 Wallets结构，它保存所有的wallet以及它的地址

type Wallets struct {
	//map[地址]钱包
	WalletMap map[string]*Wallet
}

//创建方法
func NewWallets() *Wallets {
	var ws Wallets
	ws.WalletMap = make(map[string]*Wallet)
	ws.loadFile()
	return &ws
}

func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := wallet.NewAddress()

	//var wallets Wallets
	//wallets.WalletMap = make(map[string]*Wallet)
	ws.WalletMap[address] = wallet

	ws.saveToFile()
	return address

}

//保存方法，把新建的wallet添加进去
func (ws *Wallets) saveToFile() {
	var buffer bytes.Buffer

	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}

	ioutil.WriteFile(walletFile, buffer.Bytes(), 0600)
}

//读取文件方法，把所有的wallet读出来
func (ws *Wallets) loadFile() {
	//在读取之前，要确认文件是否存在,如果不存在，直接推测出
	_, err := os.Stat(walletFile)
	if os.IsNotExist(err) {
		ws.WalletMap = make(map[string]*Wallet)
		return
	}

	content, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}

	//解码
	gob.Register(elliptic.P256())

	decoder := gob.NewDecoder(bytes.NewReader(content))

	var wsLocal Wallets

	err = decoder.Decode(&wsLocal)
	if err != nil {
		log.Panic(err)
	}
	//ws = &wsLocal
	//对于结构来说，里面有map的，要指定复制，不要在最外层直接赋值
	ws.WalletMap = wsLocal.WalletMap
}

func (ws *Wallets) ListAddresses() []string {
	var addresses []string
	for address := range ws.WalletMap {
		addresses = append(addresses, address)
	}
	return addresses
}

//通过地址返回公钥的hash值
func GetPubKeyFromAddress(address string) []byte {
	//1.解码
	//2.截取出公钥哈希，取出version(1字节)，去除校验码(4字节)
	addressByte, err := base58.Decode(address) //25字节
	if err != nil {
		log.Panic(err)
	}
	len := len(addressByte)
	pubKeyHash := addressByte[1 : len-4]
	return pubKeyHash
}
