package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"github.com/mr-tron/base58"
	"golang.org/x/crypto/ripemd160"
	"log"
)

//这里的钱包是一结构，每个钱包保存了公钥、私钥对

type Wallet struct {
	//私钥
	Private *ecdsa.PrivateKey
	//约定，这里的PubKey不存储原始的公钥，而是存储X与Y拼接的字符串，在校验段重新拆分
	Pubkey []byte
}

//创建钱包
func NewWallet() *Wallet {
	//创建曲线
	curve := elliptic.P256()
	//生成私钥
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	//生成公钥
	pubKeyOrig := privateKey.PublicKey
	//拼接X.Y
	pubKey := append(pubKeyOrig.X.Bytes(), pubKeyOrig.Y.Bytes()...)

	return &Wallet{Private: privateKey, Pubkey: pubKey}
}

//生成地址
//1.pk---(RIPEMD160(pk))--->pkHash
//2.Version--pkHash(拼接为21bytes data)
//3.(21bytes data)---(sha256.Sum256(sha256.Sum256(21bytes data)))--->4字节校验码
//4.21bytes data--4字节校验码(拼接为25bytes data)
//5.address = base58(25bytes data)
func (w *Wallet) NewAddress() string {
	pubKey := w.Pubkey
	rip160HashValue := HashPubKey(pubKey)
	version := byte(00)
	payload := append([]byte{version}, rip160HashValue...)

	//checksum
	chekCode := CheckSum(payload)

	//25字节数据
	payload = append(payload, chekCode...)

	//go语言有一个库，叫做btcd，这是个go语言实现的比特币全节点源码
	address := base58.Encode(payload)

	return address
}

func HashPubKey(pubKey []byte) []byte {
	hash := sha256.Sum256(pubKey)

	//理解为编码器
	rip160hasher := ripemd160.New()
	_, err := rip160hasher.Write(hash[:])
	if err != nil {
		log.Panic(err)
	}

	//返回rip160的哈希结果
	return rip160hasher.Sum(nil)
}

func CheckSum(payload []byte) []byte {
	hash1 := sha256.Sum256(payload)
	hash2 := sha256.Sum256(hash1[:])
	chekCode := hash2[:4]
	return chekCode
}

func IsValidAddress(address string) bool {
	//1.解码
	addressByte, err := base58.Decode(address)
	if err != nil {
		log.Panic(err)
	}
	if len(addressByte) < 4 {
		return false
	}
	//2.取数据
	payload := addressByte[:len(addressByte)-4]
	//3.做CheckSum函数
	checksum1 := addressByte[len(addressByte)-4:]
	checksum2 := CheckSum(payload)
	//4.比较
	return bytes.Equal(checksum1, checksum2)
}
