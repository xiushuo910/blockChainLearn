package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"math/big"
)

const reward = 6.25

//1.定义交易结构
type Transaction struct {
	TXID      []byte     //交易ID
	TXInputs  []TXInput  //交易输入数组
	TXOutputs []TXOutput //交易输出的数组
}

//定义交易输入
type TXInput struct {
	TXid  []byte //引用的交易ID
	Index int64  //引用的output的索引值
	//Sig       string //解锁脚本，我们用地址来模拟	（签名、公钥）
	Signature []byte //真正的数字签名，由r，s拼成的[]byte
	PubKey    []byte //约定，这里的PubKey不存储原始的公钥，而是存储X与Y拼接的字符串，在校验段重新拆分
}

//定义交易输出
type TXOutput struct {
	Value float64 //转账金额
	//PubKeyHash string  //锁定脚本，我们用地址模拟（对方的公钥hash）
	PubKeyHash []byte //收款方的公钥的哈希
}

//由于现在存储的字段是地址的公钥哈希，所以无法直接创建TXOutput
//为了能够得到公钥哈希，我们需要处理一下，写一个Lock函数
func (output *TXOutput) Lock(address string) {
	pubKeyHash := GetPubKeyFromAddress(address)
	//真正的锁定动作
	output.PubKeyHash = pubKeyHash
}

//给TXOutput提供一个创建的方法，否则无法调用Lock
func NewTXOutput(value float64, address string) *TXOutput {
	output := TXOutput{
		Value: value,
	}
	output.Lock(address)
	return &output
}

//设置交易ID(对tx先编码再hash)
func (tx *Transaction) SetHash() {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(tx)
	if err != nil {
		log.Panic("编码出错！")
	}
	data := buffer.Bytes()
	hash := sha256.Sum256(data)
	tx.TXID = hash[:]
}

//实现一个函数，判断当前的交易是否为挖矿交易
func (tx *Transaction) IsCoinbase() bool {
	//1.交易input只有一个
	if len(tx.TXInputs) == 1 {
		input := tx.TXInputs[0]
		//2.交易id为空
		//3.交易的index为-1
		if len(input.TXid) == 0 && input.Index == -1 {
			return true
		}
	}
	return false
}

//2.提供创建交易的方法（铸币交易）
func NewCoinbaseTX(address string, data string) *Transaction {
	//铸币交易的特点
	//1.只有一个input
	//2.无需引用交易id
	//3.无需引用index
	//矿工由于挖矿时无需指定签名，所以这个PubKey字段可以由矿工自由填写数据，一般填写矿池名字
	//签名先填写为空，后面创建完整交易后，最后做一次签名即可
	input := TXInput{[]byte{}, -1, nil, []byte(data)}
	//output := TXOutput{reward, address}
	output := NewTXOutput(reward, address)
	//对于铸币交易，只有一个input,一个output
	tx := Transaction{[]byte{}, []TXInput{input}, []TXOutput{*output}}
	tx.SetHash()
	return &tx
}

//创建普通的转账交易
//1.找到最合理UTXO集合 map[string][]uint64
//2.将这些UTXO逐一转成inputs
//3.创建outputs
//4.如果有零钱，找零
func NewTransaction(from, to string, amount float64, bc *BlockChain) *Transaction {
	//1.创建交易之后要进行数字签名->所以需要私钥->打开钱包（NewWallets()）
	ws := NewWallets()
	//2.找到自己的钱包，根据地址返回自己的wallet
	wallet := ws.WalletMap[from]
	if wallet == nil {
		fmt.Printf("没有找到该地址的钱包，交易创建失败！\n")
		return nil
	}
	//3.得到对应的公钥，私钥
	pubKey := wallet.Pubkey
	privateKey := wallet.Private //

	pubKeyHash := HashPubKey(pubKey)

	//1.找到最合理UTXO集合 map[string][]uint64
	utxos, resValue := bc.FindNeedUTXOs(pubKeyHash, amount)

	fmt.Printf("resValue:%f\n", resValue) //6.25
	if resValue < amount {
		fmt.Printf("余额不足，交易失败！剩余金额为:%f\n", resValue)
		return nil
	}

	var inputs []TXInput
	var outputs []TXOutput

	//2.将这些UTXOs转化为inputs
	for id, indexArray := range utxos {
		for _, i := range indexArray {
			fmt.Printf("id:%x\n", id)
			input := TXInput{[]byte(id), int64(i), nil, pubKey}
			inputs = append(inputs, input)
		}
	}

	//3.创建outputs
	//output := TXOutput{amount, to}
	output := NewTXOutput(amount, to)
	outputs = append(outputs, *output)

	if resValue > amount {
		//找零
		//outputs = append(outputs, TXOutput{resValue - amount, from})
		output = NewTXOutput(resValue-amount, from)
		outputs = append(outputs, *output)
	}

	tx := Transaction{[]byte{}, inputs, outputs}
	tx.SetHash()

	//创建交易的最后进行签名
	bc.SignTransaction(&tx, privateKey)
	return &tx
}

//签名的具体实现,参数：私钥，inputs里面所有引用的交易的结构map[string]Transaction
func (tx *Transaction) Sign(privateKey *ecdsa.PrivateKey, prevTXs map[string]Transaction) {

	if tx.IsCoinbase() {
		return
	}

	//1.创建一个当前交易的copy:TrimmedCopy：要把Signature和PubKey字段设置为nil
	txCopy := tx.TrimmedCopy()
	//2.循环遍历txCopy的input索引的output的公钥哈希
	for i, input := range txCopy.TXInputs {
		prevTX := prevTXs[string(input.TXid)]
		if len(prevTX.TXID) == 0 {
			log.Panic("引用的交易无效")
		}
		//不要对input进行赋值，这是一个副本，要对txCopy.TXInputs[xx]进行操作，否则无法把pubKeyHash传进去
		txCopy.TXInputs[i].PubKey = prevTX.TXOutputs[input.Index].PubKeyHash

		//所需要的三个数据都具备了，开始做哈希处理
		//3.生成要签名的数据，要签名的数据一定是哈希值
		//a.我们对每一个input都要签名一次，签名的数据是由当前input引用的output的哈希+当前的outputs（都承载在当前这个txCopy里面）
		//b.要对这个凭借好的txCopy进行哈希处理，SetHash得到TXID，这个TXID就是我们要签名的数据
		txCopy.SetHash()
		//还原，以免影响后面input的签名
		txCopy.TXInputs[i].PubKey = nil
		signDataHash := txCopy.TXID
		//4.执行签名动作的到r，s字节流
		r, s, err := ecdsa.Sign(rand.Reader, privateKey, signDataHash)
		if err != nil {
			log.Panic(err)
		}

		//5.放到我们所签名的input的Signature中
		signature := append(r.Bytes(), s.Bytes()...)
		tx.TXInputs[i].Signature = signature
	}

}

//创建一个当前交易的copy
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, input := range tx.TXInputs {
		inputs = append(inputs, TXInput{input.TXid, input.Index, nil, nil})
	}
	for _, output := range tx.TXOutputs {
		outputs = append(outputs, output)
	}
	return Transaction{tx.TXID, inputs, outputs}
}

//校验
//所需要的数据：公钥、数据（txCopy，生成哈希），签名
//我们要对每一个签名过的input进行校验
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	//1.得到签名的数据
	txCopy := tx.TrimmedCopy()

	for i, input := range tx.TXInputs {
		prevTX := prevTXs[string(input.TXid)]
		if len(prevTX.TXID) == 0 {
			log.Panic("引用的交易无效!")
		}
		txCopy.TXInputs[i].PubKey = prevTX.TXOutputs[input.Index].PubKeyHash
		txCopy.SetHash()
		dataHash := txCopy.TXID
		//2.得到Signature，反推r，s
		signature := input.Signature //拆r,s

		//a.定义两个辅助的big.int
		r := big.Int{}
		s := big.Int{}
		//b.拆分我们signature，平均分，前半部分给r，后半部分给s
		r.SetBytes(signature[:len(signature)/2])
		s.SetBytes(signature[len(signature)/2:])

		//3.拆解PubKey，X,Y得到原生公钥
		pubKey := input.PubKey //拆X,Y

		//a.定义两个辅助的big.int
		X := big.Int{}
		Y := big.Int{}
		//b.拆分我们signature，平均分，前半部分给r，后半部分给s
		X.SetBytes(pubKey[:len(pubKey)/2])
		Y.SetBytes(pubKey[len(pubKey)/2:])
		pubKeyOrigin := ecdsa.PublicKey{elliptic.P256(), &X, &Y}

		//4.Verify
		if !ecdsa.Verify(&pubKeyOrigin, dataHash, &r, &s) {
			return false
		}
	}
	return true
}
