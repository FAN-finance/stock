package services

import (
	"encoding/json"
	"log"
	"strconv"
	"time"
	"fmt"
	"stock/utils"
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
	"vix":"vix3x",
	//"ust": "ust20x",
	"ndx": "ndx10x",
	"xau/usd":"gold10x",
	"eur/usd":"eur20x",
	"govt":"govt20x",
	"eth/usd":"eth3x",
	"btc/usd":"btc3x",
}

//subcribe twelvedata data
func SubTwData(){
	proc:=func()error{
		return GetTwData("","",11)
	}
	utils.IntervalSync("SubTwData", 60*10, proc)
}
func GetTwData(start_date ,end_date string ,limit int)error{
	//appkey="4e8a6b8b4afe47be815d9e3b4d8cf163"
	//appkey="21cad25580b74ba3a0a2ba9be29057bb"
	dataUrl:= fmt.Sprintf( "https://api.twelvedata.com/time_series?symbol=xau/usd,vix,ndx,eur/usd,govt,eth/usd,btc/usd&interval=1min&start_date=%s&end_date=%s&apikey=21cad25580b74ba3a0a2ba9be29057bb&source=docs&outputsize=%d",start_date ,end_date,limit)
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
}