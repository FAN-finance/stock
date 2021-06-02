package services

import (
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"crypto/ecdsa"
	"math"
	"os"
	"math/big"
	"fmt"
)
import "github.com/ethereum/go-ethereum/common"

func GetUnDecimalPrice(price float64 )*big.Int{
	pint:= new(big.Int)
	//pfloat:= new(big.Float)
	//f32str := strconv.FormatFloat(float64(price), 'g', -1, 32)
	//f64, _ := strconv.ParseFloat(f32str, 64)
	//
	//pfloat.SetFloat64(f64)
	//pfloat=pfloat.Mul(pfloat,big.NewFloat(math.Pow10(18)))
	mint:=int64(float64(price)*math.Pow10(4))
	pint.SetInt64(mint)
	pint=pint.Mul(pint,big.NewInt(int64(math.Pow10(14))))
	//pint,_=pfloat.Int(nil)
	return pint
}

func GetStringsHash(items [][]byte)[]byte{
	hash:=crypto.Keccak256Hash(items...)
	prefixedHash := crypto.Keccak256(
		[]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%v", len(hash))),
		hash.Bytes(),
	)
	return prefixedHash
}
func (s *PriceView)GetHash()[]byte{
	//msg:=fmt.Sprintf("%s,%d,%f",s.Code,s.Timestamp, s.Price)
	hash := crypto.Keccak256Hash(
		common.LeftPadBytes(big.NewInt(s.Timestamp).Bytes(), 32),
		[]byte(s.Code),
		[]byte(s.BigPrice),
	)
	// normally we sign prefixed hash
	// as in solidity with `ECDSA.toEthSignedMessageHash`
	prefixedHash := crypto.Keccak256(
		[]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%v", len(hash))),
		hash.Bytes(),
	)
	return prefixedHash
}

func (s *DataPriceView)GetHash()[]byte{
	//msg:=fmt.Sprintf("%s,%d,%f",s.Code,s.Timestamp, s.Price)
	hash := crypto.Keccak256Hash(
		common.LeftPadBytes(big.NewInt(s.Timestamp).Bytes(), 32),
		[]byte(s.Code),
		[]byte(s.BigPrice),
	)
	// normally we sign prefixed hash
	// as in solidity with `ECDSA.toEthSignedMessageHash`
	prefixedHash := crypto.Keccak256(
		[]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%v", len(hash))),
		hash.Bytes(),
	)
	return prefixedHash
}
func (s *CoinPriceView)GetHash()[]byte{
	//msg:=fmt.Sprintf("%s,%d,%f",s.Code,s.Timestamp, s.Price)
	hash := crypto.Keccak256Hash(
		common.LeftPadBytes(big.NewInt(s.Timestamp).Bytes(), 32),
		[]byte(s.Coin),
		[]byte(s.VsCoin),
		[]byte(s.BigPrice),
	)
	// normally we sign prefixed hash
	// as in solidity with `ECDSA.toEthSignedMessageHash`
	prefixedHash := crypto.Keccak256(
		[]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%v", len(hash))),
		hash.Bytes(),
	)
	return prefixedHash
}

func (s *DataCoinPriceView)GetHash()[]byte{
	//msg:=fmt.Sprintf("%s,%d,%f",s.Code,s.Timestamp, s.Price)
	hash := crypto.Keccak256Hash(
		common.LeftPadBytes(big.NewInt(s.Timestamp).Bytes(), 32),
		[]byte(s.Coin),
		[]byte(s.VsCoin),
		[]byte(s.BigPrice),
	)
	// normally we sign prefixed hash
	// as in solidity with `ECDSA.toEthSignedMessageHash`
	prefixedHash := crypto.Keccak256(
		[]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%v", len(hash))),
		hash.Bytes(),
	)
	return prefixedHash
}

//代码 苹果代码 AAPL ,特斯拉代码 TSLA
var stockAddres=map[string]string{
	"AAPL":"0x732FbB5a1d79EbA483731d0eD35BD940947C6d96",
	"TSLA":"0x050F44a559B5A2B3741aD202f9bB0CD2d6e7dAF2",
}
func (s *StockNode)GetHash()[]byte{
	//Timestamp BigPrice Code
	//msg:=fmt.Sprintf("%s,%d,%f",s.Code,s.Timestamp, s.Price)
	pint:= new(big.Int)
	pint.SetString(s.BigPrice,10)
	s.Code=stockAddres[s.StockCode]
	hash := crypto.Keccak256Hash(
		common.LeftPadBytes(big.NewInt(s.Timestamp).Bytes(), 32),
		common.LeftPadBytes(pint.Bytes(), 32),
		[]byte(s.Code),
	)
	// normally we sign prefixed hash
	// as in solidity with `ECDSA.toEthSignedMessageHash`
	prefixedHash := crypto.Keccak256(
		[]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%v", len(hash))),
		hash.Bytes(),
	)
	return prefixedHash
}
func (s *StockData)GetHash()[]byte{
	//msg:=fmt.Sprintf("%s,%d,%f",s.Code,s.Timestamp, s.Price)
	s.Code=stockAddres[s.StockCode]
	hash := crypto.Keccak256Hash(
		common.LeftPadBytes(big.NewInt(s.Timestamp).Bytes(), 32),
		[]byte(s.BigPrice),
		[]byte(s.Code),
	)
	// normally we sign prefixed hash
	// as in solidity with `ECDSA.toEthSignedMessageHash`
	prefixedHash := crypto.Keccak256(
		[]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%v", len(hash))),
		hash.Bytes(),
	)
	return prefixedHash
}
func SignMsg(message []byte)[]byte{
	msg:= crypto.Keccak256([]byte(message))
	sig, err := crypto.Sign(msg, EcKey)
	if err != nil {
		log.Printf("Sign error: %s", err)
		return nil
	}
	return  sig
}

func Verify(hash, sig []byte,addre string)(bool,error){
	pubKey,err:=crypto.SigToPub(hash,sig)
	if err == nil{
		return addre==crypto.PubkeyToAddress(*pubKey).Hex(),nil
	}
	return  false ,err
}
var keyFpath ="asset/pkey"
func GenKeyFile(){
	fi,err:=os.Stat(keyFpath)
	if err == nil && !fi.IsDir() {
		log.Println("key exists")
		return
	}
	os.MkdirAll("asset",os.ModePerm)
	key,_:=crypto.GenerateKey()
	//pubKey:=key.PublicKey
	//addre:=crypto.PubkeyToAddress(pubKey)
	//fileName:="asset/"+addre.Hex()
	err=crypto.SaveECDSA(keyFpath,key)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("gen keyfile:",keyFpath)
}
var EcKey *ecdsa.PrivateKey
var WalletAddre string
func InitNodeKey() {
	GenKeyFile()
	//key, _ := crypto.HexToECDSA(testPrivHex)
	//addr := common.HexToAddress(testAddrHex)
	var err error
	EcKey, err = crypto.LoadECDSA(keyFpath)
	if err != nil {
		log.Fatalln(err)
	}
	pubKey := EcKey.PublicKey
	addre := crypto.PubkeyToAddress(pubKey)
	WalletAddre = addre.Hex()
	log.Println("WalletAddre:",WalletAddre)
}