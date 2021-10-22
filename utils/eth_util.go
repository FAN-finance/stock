package utils

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"time"
)

var EthConn *ethclient.Client
var EthAuth *bind.TransactOpts

func InitEConn(infura string) {
	log.Println("init InitEConn")
	ethUrl := fmt.Sprintf("wss://mainnet.infura.io/ws/v3/%s", infura)
	for {
		conn1, err := ethclient.Dial(ethUrl)
		if err != nil {
			log.Printf("Failed to connect to the bsc Ethereum client: %v", err)
			time.Sleep(1 * time.Second)
		} else {
			EthConn = conn1
			break
		}
	}
	//conn.SendTransaction()
	log.Println("init finish InitEConn", ethUrl)
}

//ethUrl := fmt.Sprintf("wss://mainnet.infura.io/ws/v3/%s", infura)
//bsc ethUrl := "wss://bsc-ws-node.nariox.org:443"
//polygon ethUrl := "wss://rpc-mainnet.matic.network:443

//bsc wss endpoint: https://account.getblock.io/
//https://docs.polygon.technology/docs/develop/network-details/network
var EthUrlMap =map[string]string{
	"eth":"wss://mainnet.infura.io/ws/v3/%s",
	"bsc":"wss://bsc.getblock.io/mainnet/?api_key=25c285b0-35ec-4455-8902-3187daf08750",
	//"bsc":"wss://bsc-ws-node.nariox.org",
	"polygon":"wss://rpc-mainnet.matic.quiknode.pro",
}
func GetEthConn(chainName string ) *ethclient.Client {
	ethUrl:=EthUrlMap[chainName]
	if chainName=="eth"{
		ethUrl=fmt.Sprintf(ethUrl,InfuraID)
	}
	log.Println("get ethconn",ethUrl)
	for {
		conn1, err := ethclient.Dial(ethUrl)
		if err != nil {
			log.Printf("Failed to connect to the Ethereum client: url %s, err: %v", ethUrl, err)
			time.Sleep(1 * time.Second)
		} else {
			log.Println("GetEthConn ok", ethUrl)
			return conn1
		}
	}
	//conn.SendTransaction()
}
func EthBlockTime( block uint64,cli *ethclient.Client) (uint64){
	header, err := cli.HeaderByNumber(context.Background(), big.NewInt(int64(block)))
	if err != nil {
		log.Println("EthBlockTime",err)
	}
	return  header.Time

}
func EthLastBlock( cli *ethclient.Client) (int,error){
	lastBlock, err := cli.BlockNumber(context.Background())
	return int(lastBlock),err
}

func InitEConnLocal() {
	conn1, err := ethclient.Dial("ws://localhost:8546")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
		if err == nil {

		}
	}
	EthConn = conn1
	//conn.SendTransaction()
}