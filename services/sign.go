package services

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"math"
	"math/big"
	"os"
)
import "github.com/ethereum/go-ethereum/common"

func GetUnDecimalPrice(price float64) *big.Int {
	pint := new(big.Int)
	//pfloat:= new(big.Float)
	//f32str := strconv.FormatFloat(float64(price), 'g', -1, 32)
	//f64, _ := strconv.ParseFloat(f32str, 64)
	//
	//pfloat.SetFloat64(f64)
	//pfloat=pfloat.Mul(pfloat,big.NewFloat(math.Pow10(18)))
	mint := int64(float64(price) * math.Pow10(4))
	pint.SetInt64(mint)
	pint = pint.Mul(pint, big.NewInt(int64(math.Pow10(14))))
	//pint,_=pfloat.Int(nil)
	return pint
}
func GetUnDecimalUsdPrice(price float64, decimal int) *big.Int {
	pint := new(big.Int)
	mint := int64(price * math.Pow10(decimal))
	pint.SetInt64(mint)
	//pint = pint.Mul(pint, big.NewInt(int64(math.Pow10(14))))
	//pint,_=pfloat.Int(nil)
	return pint
}

func GetStringsHash(items [][]byte) []byte {
	hash := crypto.Keccak256Hash(items...)
	prefixedHash := crypto.Keccak256(
		[]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%v", len(hash))),
		hash.Bytes(),
	)
	return prefixedHash
}
func (s *PriceView) GetHash() []byte {
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

func (s *DataPriceView) GetHash() []byte {
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
func (s *HLPriceView) GetHash() []byte {
	//msg:=fmt.Sprintf("%s,%d,%f",s.Code,s.Timestamp, s.Price)
	hash := crypto.Keccak256Hash(
		common.LeftPadBytes(big.NewInt(s.Timestamp).Bytes(), 32),
		common.LeftPadBytes(big.NewInt(int64(s.DataType)).Bytes(), 32),
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

func (s *HLDataPriceView) GetHash() []byte {
	//msg:=fmt.Sprintf("%s,%d,%f",s.Code,s.Timestamp, s.Price)
	hash := crypto.Keccak256Hash(
		common.LeftPadBytes(big.NewInt(s.Timestamp).Bytes(), 32),
		common.LeftPadBytes(big.NewInt(int64(s.DataType)).Bytes(), 32),
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
func (s *CoinPriceView) GetHash() []byte {
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

func (s *DataCoinPriceView) GetHash() []byte {
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
var stockAddres = map[string]string{
	"AAPL": "0xD87f6eCC45ABAD69446DA79a19D1E5Cf3B779098",
	"TSLA": "0x681E954a65573fC3152b909dDD75d510285eBB0D",
}

func (s *StockNode) GetHash() []byte {
	//Timestamp BigPrice Code
	//msg:=fmt.Sprintf("%s,%d,%f",s.Code,s.Timestamp, s.Price)
	pint := new(big.Int)
	pint.SetString(s.BigPrice, 10)
	s.Code = stockAddres[s.StockCode]
	hash := crypto.Keccak256Hash(
		common.LeftPadBytes(big.NewInt(s.Timestamp).Bytes(), 32),
		common.LeftPadBytes(big.NewInt(int64(s.DataType)).Bytes(), 32),
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
func (s *StockData) GetHash() []byte {
	//msg:=fmt.Sprintf("%s,%d,%f",s.Code,s.Timestamp, s.Price)
	s.Code = stockAddres[s.StockCode]
	/*hash := crypto.Keccak256Hash(
		common.LeftPadBytes(big.NewInt(s.Timestamp).Bytes(), 32),
		common.LeftPadBytes(big.NewInt(int64(s.DataType)).Bytes(), 32),
		[]byte(s.BigPrice),
		[]byte(s.Code),
	)*/

	uint256Ty, _ := abi.NewType("uint256", "", nil)
	addressTy, _ := abi.NewType("address", "", nil)
	arguments := abi.Arguments{
		{Type: uint256Ty},
		{Type: addressTy},
		{Type: uint256Ty},
		{Type: uint256Ty},
	}

	pint := new(big.Int)
	pint.SetString(s.BigPrice, 10)

	pack, _ := arguments.Pack(big.NewInt(int64(s.DataType)), s.Code, pint, big.NewInt(s.Timestamp))

	hash := crypto.Keccak256Hash(pack)
	// normally we sign prefixed hash
	// as in solidity with `ECDSA.toEthSignedMessageHash`
	prefixedHash := crypto.Keccak256(
		[]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%v", len(hash))),
		hash.Bytes(),
	)
	return prefixedHash
}
func SignMsg(message []byte) []byte {
	msg := crypto.Keccak256([]byte(message))
	sig, err := crypto.Sign(msg, EcKey)
	if err != nil {
		log.Printf("Sign error: %s", err)
		return nil
	}
	return sig
}

func Verify(hash, sig []byte, addre string) (bool, error) {
	pubKey, err := crypto.SigToPub(hash, sig)
	if err == nil {
		return addre == crypto.PubkeyToAddress(*pubKey).Hex(), nil
	}
	return false, err
}

var keyFpath = "asset/pkey"

func GenKeyFile() {
	fi, err := os.Stat(keyFpath)
	if err == nil && !fi.IsDir() {
		log.Println("key exists")
		return
	}
	os.MkdirAll("asset", os.ModePerm)
	key, _ := crypto.GenerateKey()
	//pubKey:=key.PublicKey
	//addre:=crypto.PubkeyToAddress(pubKey)
	//fileName:="asset/"+addre.Hex()
	err = crypto.SaveECDSA(keyFpath, key)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("gen keyfile:", keyFpath)
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
	log.Println("WalletAddre:", WalletAddre)
}
