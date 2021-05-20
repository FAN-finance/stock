package services

import (
	"stock/utils"
	"time"
	"log"
	"context"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"strconv"
	"fmt"
	"encoding/json"
)

func SubEthPrice(lastBlock int64) {
	//utils.InitConf(nil)
	//utils.InitDb()
	//utils.InitEConn(false)

	utils.Orm.AutoMigrate(BlockPrice{})
	//log.Println("subNewHeader from block", lastBlock)

	payloadFmt := `{"query":"{\n  bundles(block:{number: %d}) {\n    id\n    ethPrice\n  }\n}\n","variables":null}`

	chanLog := make(chan *types.Header, 1000)
	subcribe, err := EthConn.SubscribeNewHead(context.Background(), chanLog)
	defer subcribe.Unsubscribe()
	if err == nil {
		//utils.EthConn.HeaderByNumber(context.Background(),100000)
		for {
			select {
			case err := <-subcribe.Err():
				time.Sleep(2 * time.Second)
				log.Println(err)
				goto END
			case header := <-chanLog:
				if header.Number.Int64() > lastBlock {
					log.Println(header.Number.Int64(), time.Unix(int64(header.Time), 0))
					//time.Sleep(5*time.Second)
					targetBlock:=header.Number.Int64()-2
					payload := fmt.Sprintf(payloadFmt, targetBlock)
					log.Println(string(payload))
				RETRY:
					bs, err1 := utils.ReqResBody("https://api.thegraph.com/subgraphs/name/uniswap/uniswap-v2", "https://opensea.io/rankings?sortBy=seven_day_volume&category=art", "POST",nil, json.RawMessage(payload))
					if err1 != nil {
						log.Println("get price err:", err1,string(bs))
						continue
					}
					epayload := new(ethPricePayload)
					err = json.Unmarshal(bs, epayload)
					if err != nil {
						log.Println("get price json err:", err1, string(bs))
						continue
					}
					if len(epayload.Data.Bundles)==0{
						log.Println(string(bs))
						time.Sleep(10*time.Second)
						log.Println("empty data goto retry")
						goto RETRY
					}
					price,_:=strconv.ParseFloat(epayload.Data.Bundles[0].EthPrice,64)
					blockPrice:= BlockPrice{ID: int(targetBlock),Price: price,BlockTime:header.Time }
					err=utils.Orm.Create(&blockPrice).Error
					if err != nil {
						log.Println(err)
					}
					//utils.JsonOutput(blockPrice)
				} else {
					log.Println("skip old", header.Number.Int64(), time.Unix(int64(header.Time), 0))
				}
			}
		}
	}
END:
	subcribe.Unsubscribe()
	if err != nil {
		log.Println("subNewEthPrice err:", err)
	}
}

type BlockPrice struct {
	ID int
	Price float64
	BlockTime uint64 `gorm:"index:,sort:desc"`
	CreatedAt time.Time
}
func (bp BlockPrice )GetPrice() float64{
	err:=utils.Orm.Order("id desc").First(&bp).Error
	if err != nil {
		log.Println("BlockPrice,err",err)
	}
	return bp.Price
}

type ethPricePayload struct {
	Data struct {
		Bundles []struct {
			EthPrice string `json:"ethPrice"`
			ID       string `json:"id"`
		} `json:"bundles"`
	} `json:"data"`
}


var EthConn *ethclient.Client
var EthAuth *bind.TransactOpts
func InitEConn( infura string ) {
	conn1, err := ethclient.Dial(fmt.Sprintf("wss://mainnet.infura.io/ws/v3/%s",infura))
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
		if err == nil {

		}
	}
	EthConn = conn1
	//conn.SendTransaction()
}