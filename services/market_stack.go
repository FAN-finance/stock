package services


import (
	"encoding/json"
	"fmt"
	"log"

	"stock/utils"
)

type msData struct {
	Open float32 `json:"open"`
	High float64 `json:"high"`
	Low float64 `json:"low"`
	Close float32 `json:"close"`
	Volume float32 `json:"volume"`
	Date string `json:"date"`
	Symbol string `json:"symbol"`
}

type msResponse struct {
	Data []msData `json:"data"`
}

func GetMsData(code string )(apiResponse *msResponse,err error) {
	apiResponse=new(msResponse)
	//httpClient := http.Client{}

	//req, err := http.NewRequest("GET", "http://api.marketstack.com/v1/intraday", nil)
	//if err != nil {
	//	panic(err)
	//}
	//"http://api.marketstack.com/v1/intraday?access_key=a2223ab3c24359116ca94676453e5b1b&symbols=TSLA&interval=15min&limit=3"
	//q := req.URL.Query()
	//q.Add("access_key", "a2223ab3c24359116ca94676453e5b1b")
	//q.Add("symbols", code)
	//q.Add("interval", "1min")
	//q.Add("limit","4")
	//
	//req.URL.RawQuery = q.Encode()
	//
	//res, err := httpClient.Do(req)
	//if err != nil {
	//	panic(err)
	//}
	//defer res.Body.Close()
	//
	//var apiResponse Response
	//json.NewDecoder(res.Body).Decode(&apiResponse)
	urlstr:=fmt.Sprintf("http://api.marketstack.com/v1/intraday?access_key=dde1d6a489bb02c9bf0d67daf1b6821d&symbols=%s&interval=15min&limit=4",code)
	bs,err:=utils.ReqResBody(urlstr,"nil","GET",nil,nil)
	if err == nil {
		err=json.Unmarshal(bs,apiResponse)
		//for _, stockData := range apiResponse.Data {
		//	fmt.Println(fmt.Sprintf("Ticker %s has a day high of %v on %s",
		//		stockData.Symbol,
		//		stockData.High,
		//		stockData.Date))
		//}
	}
	if err != nil {
		log.Println(err)
	}
	return
}
func GetMsStatData(code string,dataType int) (float64,error){
	res,err:=GetMsData(code)
	if err != nil {
		return -1,err
	}
	max,min:=float64(0),float64(10000)
	if len(res.Data)==0{
		min=0
	}
	fmt.Println(fmt.Sprintf("Ticker %s datas: %d", code, len(res.Data)))
	for _, stockData := range res.Data {
		//fmt.Println(fmt.Sprintf("Ticker %s has a day high of %v on %s", stockData.Symbol, stockData.High, stockData.Date))
		if max<stockData.High{
			max=stockData.High
		}
		if min>stockData.Low{
			min=stockData.High
		}
	}
	if dataType==1{
		return max,nil
	}
	if dataType==2{
		return min,nil
	}
	return -1,err
}