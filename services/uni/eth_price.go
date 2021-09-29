package uni

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"log"
	"stock/utils"
	"strconv"
	"time"
)

func SubEthPrice(lastBlock int64,swapGraphApi string) {
	//utils.InitConf(nil)
	//utils.InitDb()
	//utils.InitEConn(false)

	utils.Orm.AutoMigrate(BlockPrice{})
	//log.Println("subNewHeader from block", lastBlock)

	payloadFmt := `{"query":"{\n  bundles(block:{number: %d}) {\n    id\n    ethPrice\n  }\n}\n","variables":null}`
	chanLog := make(chan *types.Header, 1000)
RETRYSUB:
	subcribe, err := utils.EthConn.SubscribeNewHead(context.Background(), chanLog)
	if err != nil {
		log.Println("EthConn subcribe err", err)
		time.Sleep(2 * time.Second)
		goto RETRYSUB
	}
	defer subcribe.Unsubscribe()
	if err == nil {
		//utils.EthConn.HeaderByNumber(context.Background(),100000)
		for {
			select {
			case err := <-subcribe.Err():
				log.Println(err)
				time.Sleep(2 * time.Second)
				if subcribe != nil {
					subcribe.Unsubscribe()
				}
				goto RETRYSUB
			case header := <-chanLog:
				if header.Number.Int64() > lastBlock {
					log.Println(header.Number.Int64(), time.Unix(int64(header.Time), 0))
					//time.Sleep(5*time.Second)
					go func() {
						stime := time.Now()
						time.Sleep(6 * time.Second)
						targetBlock := header.Number.Int64()
						payload := fmt.Sprintf(payloadFmt, targetBlock)
						log.Println(string(payload))
					RETRY:
						bs, err1 := utils.ReqResBody(swapGraphApi, "https://opensea.io/rankings?sortBy=seven_day_volume&category=art", "POST", nil, json.RawMessage(payload))
						if err1 != nil {
							log.Println("get price http err:", err1, string(bs))
							time.Sleep(10 * time.Second)
							log.Println("get price http err goto retry")
							goto RETRY
						}
						epayload := new(ethPricePayload)
						err = json.Unmarshal(bs, epayload)
						if err != nil {
							log.Println("get price json err:", err1, string(bs))
							time.Sleep(10 * time.Second)
							log.Println("get price json err goto retry")
							goto RETRY
						}
						if len(epayload.Data.Bundles) == 0 {
							log.Println("get price bundle err",string(bs))
							time.Sleep(10 * time.Second)
							log.Println("empty data goto retry")
							goto RETRY
						}
						price, _ := strconv.ParseFloat(epayload.Data.Bundles[0].EthPrice, 64)
						blockPrice := BlockPrice{ID: int(targetBlock), Price: price, BlockTime: header.Time}
						delay := int(time.Now().Sub(stime).Seconds())
						blockPrice.Delay = delay
						err = utils.Orm.Create(&blockPrice).Error
						if err != nil {
							log.Println(err)
						}
						//utils.JsonOutput(blockPrice)
					}()

				} else {
					log.Println("skip old", header.Number.Int64(), time.Unix(int64(header.Time), 0))
				}
			}
		}
	}
	//END:
	subcribe.Unsubscribe()
	if err != nil {
		log.Println("subNewEthPrice err:", err)
	}
}

type BlockPrice struct {
	ID        int
	Price     float64
	BlockTime uint64 `gorm:"index:,sort:desc"`
	CreatedAt time.Time
	Delay     int
}

//func (BlockPrice) TableName() string {
//	return "tmp_BlockPrice"
//}

func (bp BlockPrice) GetPrice() float64 {
	err := utils.Orm.Order("id desc").First(&bp).Error
	if err != nil {
		log.Println("BlockPrice,err", err)
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
