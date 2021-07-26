package controls

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"log"
	"stock/services"
	"stock/utils"
	"strconv"
	"sync"
)

// @Tags default
// @Summary　获取ftx token价格信息
// @Description 获取ftx token价格信息
// @ID FtxPriceSignHandler
// @Accept  json
// @Produce  json
// @Param     coin_type   path    string     true        "ftx类型" default(btc3x)  Enums(mvi2x,usd, btc3x, eth3x, vix3x, govt20x, gold10x, eur20x,ndx10x,mvi2s, btc3s, eth3s, vix3s, gold10s, eur20s,ndx10s,govt20s)
// @Param     data_type   path    int     true   "最高最低价１最高　２最低价 3平均价 4实时价" default(1) Enums(1,2,3,4)
// @Param     timestamp   path    int     false    "当前时间的unix秒数,该字段未使用，仅在云存储上用于标识" default(1620383144)
//@Param     debug   query    int     false    "调试" default(0)
// @Success 200 {object} services.HLDataPriceView	"token price info"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
// @Router /pub/dex/ftx_price/{coin_type}/{data_type}/{timestamp} [get]
func FtxPriceSignHandler(c *gin.Context) {
	coin_type := c.Param("coin_type")
	timestampstr := c.Param("timestamp")
	timestamp, _ := strconv.Atoi(timestampstr)
	dataTypeStr := c.Param("data_type")
	dataType, _ := strconv.Atoi(dataTypeStr)

	//isDisableSign:=false
	//if strings.HasPrefix(coin_type,"btc") ||strings.HasPrefix(coin_type,"eth"){
	//	isDisableSign=true
	//}

	//ckey:=fmt.Sprintf("TokenPriceSignHandler-%s",code)
	//var addres []*services.PriceView
	//proc:= func()(interface{},error) {}
	//SetCacheRes(c,ckey,false,proc,c.Query("debug")=="1")

	resTokenView := new(services.HLDataPriceView)
	avgNodesPrice := []*services.HLPriceView{}
	sc := sync.RWMutex{}
	wg := new(sync.WaitGroup)
	var porcNode = func(nodeUrl string) {
		defer wg.Done()
		reqUrl := fmt.Sprintf(nodeUrl+"/pub/internal/dex/ftx_price/%s/%d?data_type=%d", coin_type, timestamp, dataType)
		bs, err := utils.ReqResBody(reqUrl, "", "GET", nil, nil)
		if err == nil {
			token := new(services.HLPriceView)
			err = json.Unmarshal(bs, token)
			if err != nil {
				log.Println(err)
			}
			token.Node = nodeUrl
			sc.Lock()
			resTokenView.Signs = append(resTokenView.Signs, token)
			sc.Unlock()
		}
	}
	for _, nurl := range utils.Nodes {
		wg.Add(1)
		go porcNode(nurl)
	}
	wg.Wait()
	var err error

	if len(resTokenView.Signs) == 0 {
		err = errors.New("数据不可用")
		goto END
	}
	if len(resTokenView.Signs) < len(utils.Nodes)/2+1 {
		err = errors.New("节点不够用")
		goto END
	}

	porcNode = func(nodeUrl string) {
		defer wg.Done()
		reqUrl := fmt.Sprintf(nodeUrl + "/pub/internal/token_avgprice")
		bodyBs, _ := json.Marshal(resTokenView.Signs)
		bs, err := utils.ReqResBody(reqUrl, "", "POST", nil, bodyBs)
		if err == nil {
			snode := new(services.HLPriceView)
			json.Unmarshal(bs, snode)
			snode.Node = nodeUrl
			isMyData, _ := services.Verify(snode.GetHash(), snode.Sign, services.WalletAddre)
			if isMyData {
				//log.Println(myData,"124")
				resTokenView.PriceUsd = snode.PriceUsd
				resTokenView.BigPrice = snode.BigPrice
				resTokenView.Timestamp = snode.Timestamp
				resTokenView.Code = snode.Code
				resTokenView.DataType = dataType
				resTokenView.Sign = snode.Sign
			}

			//if isDisableSign{
			//	snode.Sign=nil
			//	resTokenView.Sign = nil
			//}
			sc.Lock()
			avgNodesPrice = append(avgNodesPrice, snode)
			sc.Unlock()
		}
	}
	for _, nurl := range utils.Nodes {
		wg.Add(1)
		go porcNode(nurl)
	}
	wg.Wait()
	if len(avgNodesPrice) == 0 {
		err = errors.New("数据不可用")
		goto END
	}
	if len(avgNodesPrice) < len(utils.Nodes)/2+1 {
		err = errors.New("节点不够用")
		goto END
	}
	resTokenView.AvgSigns = avgNodesPrice

	resTokenView.IsMarketOpening = true
	if coin_type == "ndx10x" || coin_type == "vix3x" || coin_type == "govt20x" {
		status, ts := services.IsWorkTime(0)
		if !status {
			resTokenView.IsMarketOpening = false
			//收盘没有签名时，选择第一个价格，方便应用显示价格
			if resTokenView.AvgSigns[0].Sign == nil {
				snode := resTokenView.AvgSigns[0]
				resTokenView.PriceUsd = snode.PriceUsd
				resTokenView.BigPrice = snode.BigPrice
				resTokenView.Timestamp = snode.Timestamp
				resTokenView.Code = snode.Code
				resTokenView.DataType = dataType
			}
		} else {
			resTokenView.MarketOpenTime = ts
		}
	}
	c.JSON(200, resTokenView)
	return
END:
	if err == nil {
		ErrJson(c, err.Error())
	}
}

var ftxAddres = map[string]string{
	"mvi2x":   "0x6b5ab672ac243193b006ea819a5eb08bcd518de7",
	"mvi2s":   "0xc7b86cc68c2b49f2609e9b5e12f0aa7be775bf1d",
	"btc3x":   "0x5190144c70f024bbccf9b41690e4ce3ccac31a68",
	"btc3s":   "0x66094a0624a4e8a8b9a7eff8dc0982706015340d",
	"eth3x":   "0x247913d11957f3561d4a14166ec478c3c70a9297",
	"eth3s":   "0xb1c1504c6f2646cad9ed291158b694723d38c394",
	"vix3x":   "0x25CfA4eB34FE87794372c2Fac25fE1cEB1958183",
	"govt20x": "0xab9016557b3fe80335415d60d33cf2be4b9ba461",
	"gold10x": "0x34d97B5F814Ca6E3230429DCfF42d169800cA697",
	"eur20x":  "0x2Be088a27150fc122233356dFBF3a0C01684329C",
	"ndx10x":  "0x9578BF55c12C66E222344c3244Db6eA8b2498aca",
	"usd":     "0x76417e660df3e5c90c0361674c192da152a806e4",
}

// @Tags default
// @Summary　获取ftx coin最近一小时最高最低价格信息,内部单节点
// @Description 获取ftx coin最近一小时最高最低价格信息
// @ID FtxPriceHandler
// @Accept  json
// @Produce  json
// @Param     coin_type   path    string     true        "ftx类型" default(btc3x)  Enums(mvi2x,usd, btc3x, eth3x, vix3x, gold10x, eur20x,ndx10x,govt20x,mvi2s, btc3s, eth3s, vix3s, gold10s, eur20s,ndx10s,govt20s)
// @Param     data_type   query    int     true   "最高最低价１最高　２最低价 3平均价 4实时价" default(1) Enums(1,2,3,4)
// @Param     timestamp   path    int     false    "当前时间的unix秒数,该字段未使用，仅在云存储上用于标识" default(1620383144)
// @Success 200 {object} services.HLPriceView	"Price View"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
// @Router /pub/internal/dex/ftx_price/{coin_type}/{timestamp} [get]
func FtxPriceHandler(c *gin.Context) {
	coin_type := c.Param("coin_type")
	timestampstr := c.Param("timestamp")
	timestamp, _ := strconv.Atoi(timestampstr)
	dataTypeStr := c.Query("data_type")
	dataType, _ := strconv.Atoi(dataTypeStr)

	//cb:=new(services.CoinBull)
	//err:=utils.Orm.Order("id desc").Where("coin_type=?",coin).First(cb).Error
	vp := new(HLValuePair)
	if coin_type == "usd" {
		vp.High = 1
		vp.Low = 1
	} else {
		intreval := "60s"
		count := 60
		//intreval:="hour"
		//count:=20
		//SetCacheRes(c,ckey,false,proc,c.Query("debug")=="1")
		ckey := fmt.Sprintf("FtxPriceHandler-%s-%s-%d", coin_type, intreval, count)
		proc := func() (interface{}, error) {
			lastPrice := 0.0
			err := utils.Orm.Raw(
				`SELECT bull FROM coin_bull
WHERE
 coin_type=? order by  timestamp desc limit 1;`, coin_type).Scan(&lastPrice).Error
			if err != nil {
				return nil, err
			}
			vp := new(HLValuePair)
			vp.Last=lastPrice
			err = utils.Orm.Raw(
				`SELECT max(bull) high,min(bull) low,avg(bull) avg FROM coin_bull
WHERE
 timestamp >unix_timestamp()-3600 and coin_type=?;`, coin_type).Scan(vp).Error
			if err == nil {
				if vp.High == 0 {
					vp.High = lastPrice
					vp.Low = lastPrice
					vp.Avg = lastPrice
				}
			}
			return vp, err
		}
		var res interface{}
		var err error
		if c.Query("debug") == "1" {
			res, err = proc()
		} else {
			log.Println("cache process", ckey)
			res, err = utils.CacheFromLru(1, ckey, int(100), proc)
		}
		if err == nil {
			vp = res.(*HLValuePair)
		}
		if err != nil {
			ErrJson(c, err.Error())
			return
		}
	}

	//price := services.BlockPrice{}.GetPrice()
	//fprice, _ := strconv.ParseFloat(res.DerivedETH, 64)
	//res.PriceUsd = fprice * price

	log.Println("FtxPriceHandler vp",*vp)
	tPriceView := new(services.HLPriceView)
	tPriceView.Code = ftxAddres[coin_type]
	//if code == "0xc011a73ee8576fb46f5e1c5751ca3b9fe0af2a6f" {
	//	tPriceView.Code = "0x6EBFD2E7678cFA9c8dA11b9dF00DB24a35ec7dD4"
	//}
	tPriceView.Timestamp = int64(timestamp)
	tPriceView.DataType = dataType
	if dataType == 1 {
		tPriceView.PriceUsd = vp.High
	}
	if dataType == 2 {
		tPriceView.PriceUsd = vp.Low
	}
	if dataType == 3 {
		tPriceView.PriceUsd = vp.Avg
	}
	if dataType == 4 {
		tPriceView.PriceUsd = vp.Last
	}

	//tPriceView.PriceUsd =  math.Trunc(tPriceView.PriceUsd*1000) / 1000
	tPriceView.PriceUsd, _ = decimal.NewFromFloat(tPriceView.PriceUsd).Round(18).Float64()
	tPriceView.BigPrice = services.GetUnDecimalPrice(tPriceView.PriceUsd).String()
	tPriceView.NodeAddress = services.WalletAddre
	if tPriceView.PriceUsd > 0.001 {
		if isStockFtx(coin_type) { //股票签名
			if services.IsSignTime(0) {
				tPriceView.Sign = services.SignMsg(tPriceView.GetHash())
			}
		} else {
			tPriceView.Sign = services.SignMsg(tPriceView.GetHash())
		}

	}
	c.JSON(200, tPriceView)
	return
}

//股票类ftx判断
func isStockFtx(code string) bool {
	if code == "ndx10x" || code == "vix3x" || code == "govt20x" {
		return true
	}
	if code == ftxAddres["ndx10x"] || code == ftxAddres["vix3x"] || code == ftxAddres["govt20x"] {
		return true
	}
	return false
}

// @Tags default
// @Summary　获取杠杆btc代币不同时间区间的价格图表信息
// @Description 获取杠杆btc代币不同时间区间的价格图表信息
// @ID FtxChartPricesHandler
// @Accept  json
// @Produce  json
// @Param     coin_type   path    string     true        "ftx类型" default(btc3x)  Enums(mvi2x,btc3x, eth3x, vix3x, gold10x, eur20x,ndx10x,govt20x,mvi2s, btc3s, eth3s, vix3s, gold10s, eur20s,ndx10s,govt20s)
// @Param     count   path    int     true    "获取多少个数据点" default(10)
// @Param     interval   path    int     true    "数据间隔值,表示多少个15分钟, 如:1表示15分钟间隔 2表示30分钟间隔 3表示45分钟间隔 ,96表示1天间隔 ；" default(1) Enums(1,2,3,4,96)
// @Param     timestamp   path    int     false    "当前时间的unix秒数,该字段未使用，仅在云存储上用于标识" default(1620383144)
// @Success 200 {array} services.BlockPrice	"stock info"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
// @Router /pub/dex/ftx_chart_prices/{coin_type}/{count}/{interval}/{timestamp} [get]
func FtxChartPricesHandler(c *gin.Context) {
	coin_type := c.Param("coin_type")
	//code:="btc"
	interval_str := c.Param("interval")
	interval, _ := strconv.Atoi(interval_str)
	count_str := c.Param("count")
	count, _ := strconv.Atoi(count_str)

	//SetCacheRes(c,ckey,false,proc,c.Query("debug")=="1")
	ckey := fmt.Sprintf("FtxChartPricesHandler-%s-%s-%s", coin_type, interval_str, count_str)
	proc := func() (interface{}, error) {
		items, err := services.GetFtxTimesPrice(coin_type, interval, count)
		if err != nil {
			return nil, err
		}
		return items, err
	}
	SetCacheResExpire(c, ckey, false, 200, proc, c.Query("debug") == "1")
	//
	////timestampstr:=c.Param("timestamp")
	////timestamp,_:=strconv.Atoi(timestampstr)
	//bs,err:=services.GetFtxTimesPrice(interval,count)
	//if err == nil {
	//	c.JSON(200,bs)
	//	return
	//}
	//if err != nil {
	//	ErrJson(c,err.Error())
	//	return
	//}
}

// @Tags default
// @Summary　获取股票不同时间区间的价格图表信息
// @Description 获取股票不同时间区间的价格图表信息
// @ID StockChartPricesHandler
// @Accept  json
// @Produce  json
// @Param     coin_type   path    string     true        "股票类型" default(AAPL)  Enums(AAPL,TSLA)
// @Param     count   path    int     true    "获取多少个数据点" default(100)
// @Param     interval   path    int     true    "数据间隔值,表示多少个15分钟, 如:1表示15分钟间隔 ４表示60分钟间隔 ,96表示1天间隔 ；" default(4) Enums(1,4,96)
// @Param     timestamp   path    int     false    "当前时间的unix秒数,该字段未使用，仅在云存储上用于标识" default(1620383144)
// @Success 200 {array} services.BlockPrice	"stock info"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
// @Router /pub/dex/stock_chart_prices/{coin_type}/{count}/{interval}/{timestamp} [get]
func StockChartPricesHandler(c *gin.Context) {
	coin_type := c.Param("coin_type")
	//code:="btc"
	interval_str := c.Param("interval")
	interval, _ := strconv.Atoi(interval_str)
	count_str := c.Param("count")
	count, _ := strconv.Atoi(count_str)

	//SetCacheRes(c,ckey,false,proc,c.Query("debug")=="1")
	ckey := fmt.Sprintf("FtxChartPricesHandler-%s-%s-%s", coin_type, interval_str, count_str)
	proc := func() (interface{}, error) {
		items, err := services.GetStockTimesPrice(coin_type, interval, count)
		if err != nil {
			return nil, err
		}
		return items, err
	}
	SetCacheResExpire(c, ckey, false, 200, proc, c.Query("debug") == "1")
	//
	////timestampstr:=c.Param("timestamp")
	////timestamp,_:=strconv.Atoi(timestampstr)
	//bs,err:=services.GetFtxTimesPrice(interval,count)
	//if err == nil {
	//	c.JSON(200,bs)
	//	return
	//}
	//if err != nil {
	//	ErrJson(c,err.Error())
	//	return
	//}
}
