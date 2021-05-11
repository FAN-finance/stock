package services

import (

	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"crypto/ecdsa"
	"os"
)
//import "github.com/ethereum/go-ethereum/common"

func SignMsg(message string )[]byte{
	msg := crypto.Keccak256([]byte(message))
	sig, err := crypto.Sign(msg, EcKey)
	if err != nil {
		log.Printf("Sign error: %s", err)
		return nil
	}
	return  sig
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