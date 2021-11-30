package services

import (
	"encoding/json"
	"fmt"
	"github.com/robfig/cron/v3"
	"log"
	"stock/utils"
	"strconv"
	"time"
)

type resTw struct {
	Status string `json:"status" example:"ok"`
	Values []struct {
		Close    string `json:"close" example:"67.61850"`
		Datetime string `json:"datetime" example:"2021-06-25 14:30:00"`
		High     string `json:"high" example:"67.61850"`
		Low      string `json:"low" example:"67.39000"`
		Open     string `json:"open" example:"67.49000"`
		Volume   string `json:"volume" example:"543"`
	} `json:"values"`
}

var twSymbolMap = map[string]string{
	"vix": "vix",
	//"ust": "ust20x",
	"ndx":     "ndx",
	"xau/usd": "gold",
	"eur/usd": "eur",
	"govt":    "govt",
	"eth/usd": "eth",
	"btc/usd": "btc",
	"AAPL":    "AAPL",
	"TSLA":    "TSLA",
}
var twSymbolLocalMap = map[string]*time.Location{
	"vix": locUsaStock,
	//"ust": locUsaStock,
	"ndx":     locUsaStock,
	"govt":    locUsaStock,
	"AAPL":    locUsaStock,
	"TSLA":   locUsaStock,
	"eth/usd": time.UTC,
	"btc/usd": time.UTC,
	"xau/usd": time.UTC,
	"eur/usd": time.UTC,
}

//subcribe twelvedata data
func SyncCoinGeckoData() {
	proc := func() error {
		err := utils.Orm.Exec(
			`
insert into market_prices (item_type, price, timestamp, created_at)
  select *
  from (
         (select
            'btc' as item_type,
            usd   as price,
            coins.id,
            now()
          from coins
          where coins.id > unix_timestamp()-21
          order by coins.id
         )
         union all
         (select
            'eth'     as item_type,
            usd / eth as price,
            coins.id,
            now()
          from coins
          where coins.id > unix_timestamp()-21
          order by coins.id
         )
       ) aa;`,
		).Error
		return err
	}
	//err:=proc();
	//if err != nil {
	//	log.Println(err)
	//}
	utils.IntervalSync("SyncCoinGeckoData", 20, proc)

}

//subcribe twelvedata data
func SubTwData() {
	//GetTwData("","",10)
	//return
	//抓取失败次数:每失败一次,下次抓取相的就多抓取一行数据.
	fails := 0
	proc := func() error {
		err := GetTwData("", "", 10+fails)
		if err != nil {
			fails += 1
		} else {
			fails = 0
		}
		return err
	}
	utils.IntervalSync("SubTwData", 60, proc)
}

var Cn = cron.New(cron.WithSeconds(),
	cron.WithChain(
		cron.Recover(cron.DefaultLogger),
	))

func CronTwData() {
	fails := 15
	proc := func() error {
		log.Println("fails", fails)
		err := GetTwData("", "", 2+fails)
		if err != nil {
			fails += 1
		} else {
			fails = 0
		}
		return err
	}
	Cn.AddFunc("20 * * * * *", func() {
		proc()
	})
	//Cn.Start()
}
func GetTwData(start_date, end_date string, limit int) error {
	GetHuobiData(true)
	return
	//使用分symbol访问方法，减少twdata 压力
	//[]string{"AAPL","TSLA","vix","ndx","xau/usd","eur/usd","govt","eth/usd","btc/usd"}
	stockSymbol:=[]string{"AAPL","TSLA","vix","ndx","govt"}
	otherSymbol:=[]string{"xau/usd","eur/usd"} //,"eth/usd","btc/usd"
	if IsMarketTime(time.Now().Unix() - 65) { //减65用来确保收盘后，再请求一次．
		otherSymbol=append(otherSymbol,stockSymbol...)
	}
	for _, symbol := range otherSymbol {
		dataUrl:= fmt.Sprintf("https://api.twelvedata.com/time_series?symbol=%s&interval=1min&start_date=%s&end_date=%s&apikey=%s&source=docs&outputsize=%d", symbol, start_date, end_date, "21cad25580b74ba3a0a2ba9be29057bb", limit)
		bs, err := utils.ReqResBody(dataUrl, "", "GET", nil, nil)
		if err == nil {
			twData := new(resTw)
			err = json.Unmarshal(bs, twData)
			if err == nil {
				for i := len(twData.Values) - 1; i > -1; i-- {
					value := twData.Values[i]
					//log.Printf("%v",value)
					mprice := new(MarketPrice)
					mprice.ItemType = twSymbolMap[symbol]
					mprice.Price, _ = strconv.ParseFloat(value.Close, 64)
					ts, _ := time.ParseInLocation("2006-01-02 15:04:05", value.Datetime, twSymbolLocalMap[symbol])
					mprice.Timestamp = int(ts.Unix())
					dberr := utils.Orm.Save(mprice).Error
					if dberr != nil {
						log.Println(dberr)
					}
					//log.Println("process mm %v", mprice)
				}
			}
		}
	}
	return nil
}
func GetTwDataBack(start_date, end_date string, limit int) error {
	//appkey="4e8a6b8b4afe47be815d9e3b4d8cf163"
	//appkey="21cad25580b74ba3a0a2ba9be29057bb"
	//dataUrl := fmt.Sprintf("https://api.twelvedata.com/time_series?symbol=AAPL,TSLA,xau/usd,vix,ndx,eur/usd,govt,eth/usd,btc/usd&interval=1min&start_date=%s&end_date=%s&apikey=%s&source=docs&outputsize=%d", start_date, end_date, "21cad25580b74ba3a0a2ba9be29057bb", limit)
	dataUrl := fmt.Sprintf("https://api.twelvedata.com/time_series?symbol=AAPL,TSLA,xau/usd,vix,ndx,eur/usd,govt&interval=1min&start_date=%s&end_date=%s&apikey=%s&source=docs&outputsize=%d", start_date, end_date, "21cad25580b74ba3a0a2ba9be29057bb", limit)

	//非开盘时间，不请求股市数据, 减65用来确保收盘后，再请求一次．
	if !IsMarketTime(time.Now().Unix() - 65) {
		dataUrl = fmt.Sprintf("https://api.twelvedata.com/time_series?symbol=xau/usd,eur/usd,eth/usd,btc/usd&interval=1min&start_date=%s&end_date=%s&apikey=%s&source=docs&outputsize=%d", start_date, end_date, "21cad25580b74ba3a0a2ba9be29057bb", limit)
	}
	//dataUrl="https://api.twelvedata.com/time_series?symbol=AAPL,TSLA,xau/usd,vix,ndx,eur/usd,govt,eth/usd,btc/usd&interval=1min&start_date=2021-07-26%2019:00:00&end_date=2021-07-27&apikey=21cad25580b74ba3a0a2ba9be29057bb&source=docs&outputsize=5000"
	//dataUrl="https://api.twelvedata.com/time_series?symbol=AAPL,TSLA&interval=1min&apikey=21cad25580b74ba3a0a2ba9be29057bb&source=docs&outputsize="+strconv.Itoa(limit)

	bs, err := utils.ReqResBody(dataUrl, "", "GET", nil, nil)
	if err == nil {
		//log.Println(string(bs))
		res := map[string]json.RawMessage{}
		err = json.Unmarshal(bs, &res)
		if err == nil {
			for key, data := range res {
				//log.Println("key",key)
				twData := new(resTw)
				err = json.Unmarshal(data, twData)
				if err == nil {
					for i := len(twData.Values) - 1; i > -1; i-- {
						value := twData.Values[i]
						//log.Printf("%v",value)
						mprice := new(MarketPrice)
						mprice.ItemType = twSymbolMap[key]
						mprice.Price, _ = strconv.ParseFloat(value.Close, 64)
						ts, _ := time.ParseInLocation("2006-01-02 15:04:05", value.Datetime, twSymbolLocalMap[key])
						mprice.Timestamp = int(ts.Unix())
						dberr := utils.Orm.Save(mprice).Error
						if dberr != nil {
							log.Println(dberr)
						}
						//log.Println("process mm %v", mprice)
					}
				}
			}
		}
	}
	if err != nil {
		log.Println(err)
	}
	return err
}

func GetTwHL(code string) (max, min float64, err error) {
	min = 0
	dataUrl := fmt.Sprintf("https://api.twelvedata.com/time_series?symbol=%s&interval=1min&apikey=21cad25580b74ba3a0a2ba9be29057bb&source=docs&outputsize=60", code)
	bs, err1 := utils.ReqResBody(dataUrl, "", "GET", nil, nil)
	err = err1
	if err == nil {
		twData := new(resTw)
		err = json.Unmarshal(bs, twData)
		if err == nil {
			for _, value := range twData.Values {
				high, _ := strconv.ParseFloat(value.High, 64)
				low, _ := strconv.ParseFloat(value.Low, 64)
				if high > max {
					max = high
				}
				if min == 0 {
					min = low
				}
				if low < min {
					min = low
				}
			}
		}
	}
	return
}
