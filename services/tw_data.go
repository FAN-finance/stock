package services

import (
	"encoding/json"
	"fmt"
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
var twSymbolMap =map[string]string{
	"vix":"vix",
	//"ust": "ust20x",
	"ndx": "ndx",
	"xau/usd":"gold",
	"eur/usd":"eur",
	"govt":"govt",
	"eth/usd":"eth",
	"btc/usd":"btc",
	"AAPL":"AAPL",
	"TSLA":"TSLA",
}

//subcribe twelvedata data
func SubTwData(){
	proc:=func()error{
		return GetTwData("","",10)
	}
	utils.IntervalSync("SubTwData", 60, proc)
}
func GetTwData(start_date ,end_date string ,limit int)error{
	//appkey="4e8a6b8b4afe47be815d9e3b4d8cf163"
	//appkey="21cad25580b74ba3a0a2ba9be29057bb"
	dataUrl:= fmt.Sprintf( "https://api.twelvedata.com/time_series?symbol=AAPL,TSLA,xau/usd,vix,ndx,eur/usd,govt,eth/usd,btc/usd&interval=1min&start_date=%s&end_date=%s&apikey=%s&source=docs&outputsize=%d",start_date ,end_date,"21cad25580b74ba3a0a2ba9be29057bb",limit)

	//非开盘时间，不请求股市数据, 减65用来确保收盘后，再请求一次．
	if !IsMarketTime(time.Now().Unix()-65) {
		dataUrl = fmt.Sprintf("https://api.twelvedata.com/time_series?symbol=xau/usd,eur/usd,eth/usd,btc/usd&interval=1min&start_date=%s&end_date=%s&apikey=%s&source=docs&outputsize=%d", start_date, end_date, "21cad25580b74ba3a0a2ba9be29057bb", limit)
	}
	//dataUrl="https://api.twelvedata.com/time_series?symbol=xau/usd,vix,ndx,eur/usd,govt&interval=1min&start_date=2021-06-24&end_date=2021-07-06&apikey=21cad25580b74ba3a0a2ba9be29057bb&source=docs&outputsize=2"
	bs, err := utils.ReqResBody(dataUrl, "", "GET", nil, nil)
	if err == nil {
		//log.Println(string(bs))
		res := map[string]json.RawMessage{}
		err = json.Unmarshal(bs, &res)
		if err == nil {
			for key, data := range res {
				//log.Println("key",key)
				twData:=new(resTw)
				err = json.Unmarshal(data, twData)
				if err == nil {
					for i:=len(twData.Values)-1;i>-1;i--{
						value:=twData.Values[i]
						//log.Printf("%v",value)
						mprice:=new(MarketPrice)
						mprice.ItemType= twSymbolMap[key]
						mprice.Price,_=strconv.ParseFloat(value.Close,64)
						ts,_:=time.ParseInLocation("2006-01-02 15:04:05",value.Datetime,time.UTC)
						mprice.Timestamp=int(ts.Unix())
						err=utils.Orm.Save(mprice).Error
						if err != nil {
							log.Println(err)
						}
						log.Println("process mm %v", mprice)
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

func GetTwHL(code string)(max ,min float64,err error){
	min=0
	dataUrl:=fmt.Sprintf("https://api.twelvedata.com/time_series?symbol=%s&interval=1min&apikey=21cad25580b74ba3a0a2ba9be29057bb&source=docs&outputsize=60",code)
	bs, err1 := utils.ReqResBody(dataUrl, "", "GET", nil, nil)
	err=err1
	if err == nil {
		twData:=new(resTw)
		err=json.Unmarshal(bs,twData)
		if err == nil {
			for _, value := range twData.Values {
				high,_:=strconv.ParseFloat(value.High,64)
				low,_:=strconv.ParseFloat(value.Low,64)
				if high>max{
					max=high
				}
				if min==0{
					min=low
				}
				if low<min{
					min=low
				}
			}
		}
	}
	return
}