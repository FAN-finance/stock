package services

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"stock/contracts"

	//"bytes"
	//"errors"
	//"github.com/ethereum/go-ethereum"
	//"gorm.io/gorm"
	//"math/big"
	"context"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/robfig/cron/v3"
	"log"
	"stock/utils"
	"time"

	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"strconv"
)

func SubEthPrice(lastBlock int64) {
	//utils.InitConf(nil)
	//utils.InitDb()
	//utils.InitEConn(false)

	utils.Orm.AutoMigrate(BlockPrice{})
	//log.Println("subNewHeader from block", lastBlock)

	payloadFmt := `{"query":"{\n  bundles(block:{number: %d}) {\n    id\n    ethPrice\n  }\n}\n","variables":null}`

	chanLog := make(chan *types.Header, 1000)
RETRYSUB:
	subcribe, err := EthConn.SubscribeNewHead(context.Background(), chanLog)
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
						bs, err1 := utils.ReqResBody("https://api.thegraph.com/subgraphs/name/uniswap/uniswap-v2", "https://opensea.io/rankings?sortBy=seven_day_volume&category=art", "POST", nil, json.RawMessage(payload))
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
							log.Println(string(bs))
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

func InitBConn() {
	log.Println("init InitBscConn")
	ethUrl := "wss://bsc-ws-node.nariox.org:443"
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
	log.Println("init finish InitBscConn", ethUrl)
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

type TokenTotalSupply struct {
	ID          int `gorm:"primaryKey"`
	Token       string
	TotalSupply string
	DailyAmount string
	CreateAt    time.Time
}

var TokenAddressArr = []string{
	"0x011864d37035439e078d64630777ec518138af05",
}

var cronObj *cron.Cron = nil

func TokenTotalSupplyDailyData() {
	utils.Orm.AutoMigrate(TokenTotalSupply{})

	for _, address := range TokenAddressArr {
		contract, err := contracts.NewRei(common.HexToAddress(address), EthConn)
		if err == nil {
			lastInfo := &TokenTotalSupply{}
			err = utils.Orm.Model(TokenTotalSupply{}).Where("token=?", address).Order("create_at DESC").Limit(1).First(lastInfo).Error
			if err != nil || lastInfo.ID == 0 {
				res, err := contract.TotalSupply(nil)
				if err == nil {
					info := &TokenTotalSupply{}
					info.Token = address
					info.DailyAmount = "0"
					info.TotalSupply = res.String()
					info.CreateAt = time.Now()
					utils.Orm.Save(info)
				}
			}
		}
	}

	if cronObj == nil {
		location, _ := time.LoadLocation("UTC")
		cronObj = cron.New(cron.WithLocation(location))
	}

	cronObj.AddFunc("0 0 * * *", func() {
		for _, address := range TokenAddressArr {
			contract, err := contracts.NewRei(common.HexToAddress(address), EthConn)
			if err == nil {
				res, err := contract.TotalSupply(nil)
				if err == nil {
					currTotalSupply, _ := decimal.NewFromString(res.String())
					info := &TokenTotalSupply{}
					info.Token = address
					info.DailyAmount = "0"
					info.TotalSupply = res.String()
					info.CreateAt = time.Now()

					lastInfo := &TokenTotalSupply{}
					err = utils.Orm.Model(TokenTotalSupply{}).Where("token=?", info.Token).Order("create_at DESC").Limit(1).First(lastInfo).Error
					if err == nil && lastInfo.ID > 0 {
						lastTotalSupply, _ := decimal.NewFromString(lastInfo.TotalSupply)
						info.DailyAmount = currTotalSupply.Sub(lastTotalSupply).String()
					}
					utils.Orm.Save(info)
				}
			}
		}
	})
	cronObj.Start()
}

func TokenChartSupply(token string, amt int) (data []TokenTotalSupply, err error) {
	data = []TokenTotalSupply{}
	find := utils.Orm.Model(TokenTotalSupply{}).Where("token=?", token).Order("create_at DESC").Limit(amt).Find(&data)

	for find.Error == nil {
		return data, nil
	}
	return nil, find.Error
}

func GetTokenTotalSupply(token string) float64 {
	isExist := false
	for _, address := range TokenAddressArr {
		if address == token {
			isExist = true
			break
		}
	}
	if isExist {
		contract, err := contracts.NewRei(common.HexToAddress(token), EthConn)
		if err == nil {
			res, err := contract.TotalSupply(nil)
			if err == nil {
				currTotalSupply, _ := decimal.NewFromString(res.String())
				temp, _ := decimal.NewFromString("100000000000000000")
				currTotalSupply = currTotalSupply.Div(temp)
				totalSupply, _ := currTotalSupply.Float64()
				return totalSupply
			}
		}
	}
	return 0
}
