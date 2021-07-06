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
// @Param     coin_type   path    string     true        "ftx类型" default(btc3x)  Enums(usd, btc3x, eth3x, vix3x, govt20x, gold10x, eur20x,ndx10x)
// @Param     data_type   path    int     true   "最高最低价１最高　２最低价" default(1) Enums(1,2)
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
	c.JSON(200, resTokenView)
	return
END:
	if err == nil {
		ErrJson(c, err.Error())
	}
}

var ftxAddres = map[string]string{
	"btc3x":   "0x0ce776b748e4935a67ef345aee09cf80a74f96c9",
	"eth3x":   "0x91dF141c33e43Fc97B0b6746A95f7bfc639D76bD",
	"vix3x":   "0x25CfA4eB34FE87794372c2Fac25fE1cEB1958183",
	"ust20x":  "0x36DeBA1578B11912F6a39f0E2060C5b15cF21c3c",
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
// @Param     coin_type   path    string     true        "ftx类型" default(btc3x)  Enums(usd, btc3x, eth3x, vix3x, gold10x, eur20x,ndx10x,govt20x)
// @Param     data_type   query    int     true   "最高最低价１最高　２最低价" default(1) Enums(1,2)
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
			vp := new(HLValuePair)
			err := utils.Orm.Raw(
				`SELECT max(bull) high,min(bull) low FROM coin_bull
WHERE
 timestamp >unix_timestamp()-3600 and coin_type=?;`, coin_type).Scan(vp).Error
			if err == nil {
				if vp.High==0{
					err = utils.Orm.Raw(
						`SELECT bull high,bull low FROM coin_bull
WHERE
 coin_type=? order by  timestamp desc limit 1;`, coin_type).Scan(vp).Error
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

	log.Println(*vp)
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

	//tPriceView.PriceUsd =  math.Trunc(tPriceView.PriceUsd*1000) / 1000
	tPriceView.PriceUsd, _ = decimal.NewFromFloat(tPriceView.PriceUsd).Round(18).Float64()
	tPriceView.BigPrice = services.GetUnDecimalPrice(tPriceView.PriceUsd).String()
	tPriceView.NodeAddress = services.WalletAddre
	if tPriceView.PriceUsd > 0.001 {
		tPriceView.Sign = services.SignMsg(tPriceView.GetHash())
	}
	c.JSON(200, tPriceView)
	return
}

// @Tags default
// @Summary　获取杠杆btc代币不同时间区间的价格图表信息
// @Description 获取杠杆btc代币不同时间区间的价格图表信息
// @ID FtxChartPricesHandler
// @Accept  json
// @Produce  json
// @Param     coin_type   path    string     true        "ftx类型" default(btc3x)  Enums(btc3x, eth3x, vix3x, govt20x, gold10x, eur20x,ndx10x,govt20x)
// @Param     count   path    int     true    "获取多少个数据点" default(10)
// @Param     interval   path    int     true    "数据间隔值,表示多少个15分钟, 如:1表示15分钟间隔 2表示30分钟间隔 3表示45分钟间隔 ,96表示1天间隔 ；" default(day) Enums(1,2,3,4,96)
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
	SetCacheResExpire(c, ckey, false, 600, proc, c.Query("debug") == "1")
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
