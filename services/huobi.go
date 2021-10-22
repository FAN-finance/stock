package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"stock/services/uni"
	"stock/utils"
	"strconv"
	"strings"
	"time"
)

func GetHuobiData(preDate,saveMP bool ) error {
	var items = map[string]string{
		"ETH": "ethusdt",
		"BTC": "btcusdt",
	}
	for key, value := range items {
		urlstr := fmt.Sprintf("https://api.huobi.pro/market/history/kline?period=1min&size=2&symbol=%s", value)
		bs, err := utils.ReqResBody(urlstr, "", "GET", nil, nil)
		if err != nil {
			log.Println("GetHuobiData err", err, string(bs))
			return err
		}
		res := new(resHuobiKline)
		err = json.Unmarshal(bs, res)
		if err != nil {
			return err
		}
		for _, data := range res.Data {
			btime := data.ID
			if preDate {
				if btime != time.Now().Truncate(time.Minute).Add(-1*time.Minute).Unix() {
					continue
				}
			}
			uprice:=new(uni.UniPrice)
			uprice.PairID=DATASOURCE_HUO_BI
			uprice.Symbol=key
			uprice.BlockTime=uint64(btime)
			uprice.Price=data.Close
			utils.Orm.Save(uprice)

			if saveMP {
				mprice := new(MarketPrice)
				mprice.ItemType = strings.ToLower(key)
				mprice.Price = data.Close
				mprice.Timestamp = int(btime)
				utils.Orm.Save(mprice)
			}
		}
	}
	return nil
}

type resHuobiKline struct {
	Ch   string `json:"ch" example:"market.btcusdt.kline.5min"`
	Data []struct {
		Amount float64 `json:"amount" example:"3.946282"`
		Close  float64 `json:"close" example:"49025.510000"`
		Count  int64   `json:"count" example:"196"`
		High   float64 `json:"high" example:"49056.380000"`
		ID     int64   `json:"id" example:"1629769200"`
		Low    float64 `json:"low" example:"49022.860000"`
		Open   float64 `json:"open" example:"49056.370000"`
		Vol    float64 `json:"vol" example:"193489.672757"`
	} `json:"data"`
	Status string `json:"status" example:"ok"`
	Ts     int64  `json:"ts" example:"1629769247172"`
}

func GetBinanceData(preDate,saveMP  bool) error {
	var items=map[string]string{
		"ETH":"ETHUSDT",
		"BTC":"BTCUSDT",
	}
	for key, value := range items {
		urlstr:=fmt.Sprintf("https://api3.binance.com/api/v3/klines?symbol=%s&limit=2&interval=1m",value)
		bs, err := utils.ReqResBody(urlstr, "", "GET", nil, nil)
		if err != nil {
			log.Println("GetBinanceData err",err,string(bs))
			return err
		}
		res:=[][]interface{}{}
		err=json.Unmarshal(bs,&res)
		if err != nil {
			log.Println("GetBinanceData json err",err)
			return err
		}
		if len(res)==0{return errors.New("GetBinanceData none data")}
		for _, data := range res {
			btime:=int64(data[0].(float64))/1000
			if preDate{
				if btime!=time.Now().Truncate(time.Minute).Add(-1*time.Minute).Unix(){
					continue
				}
			}
			uprice:=new(uni.UniPrice)
			uprice.PairID=DATASOURCE_BINANCE
			uprice.Symbol=key
			uprice.BlockTime=uint64(btime)
			uprice.Price,_=strconv.ParseFloat( data[4].(string),64)
			utils.Orm.Save(uprice)

			if saveMP {
				mprice := new(MarketPrice)
				mprice.ItemType = strings.ToLower(key)
				mprice.Price = uprice.Price
				mprice.Timestamp = int(btime)
				utils.Orm.Save(mprice)
			}
		}
	}
	return nil
}
