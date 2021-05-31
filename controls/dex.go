package controls
import (
	"errors"
	"github.com/gin-gonic/gin"
	"math"
	"stock/utils"
	"log"
	"fmt"
	"sync"
	"encoding/json"
	"stock/services"
	"strconv"
)

// @Tags default
// @Summary　获取token价格信息,内部单节点
// @Description 获取token价格信息
// @ID TokenPriceSignHandler
// @Accept  json
// @Produce  json
// @Param     token   path    string     true        "token地址" default(0x66a0f676479cee1d7373f3dc2e2952778bff5bd6)
// @Param     timestamp   path    int     false    "当前时间的unix秒数,该字段未使用，仅在云存储上用于标识" default(1620383144)
// @Success 200 {object} services.TokenInfo	"stock info"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
// @Router /pub/dex/token_price/{token}/{timestamp} [get]
func TokenPriceSignHandler(c *gin.Context) {
	code := c.Param("token")
	timestampstr := c.Param("timestamp")
	timestamp, _ := strconv.Atoi(timestampstr)
	resTokenView := new(services.DataPriceView)
	//var addres []*services.PriceView
	sc := sync.RWMutex{}
	wg := new(sync.WaitGroup)
	var porcNode = func(nodeUrl string) {
		defer wg.Done()
		reqUrl := fmt.Sprintf(nodeUrl+"/pub/internal/dex/token_price/%s/%d", code, timestamp)
		bs, err := utils.ReqResBody(reqUrl, "", "GET", nil, nil)
		if err == nil {
			token := new(services.PriceView)
			err = json.Unmarshal(bs, token)
			if err == nil {
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
	sumPrice := float64(0.0)
	if len(resTokenView.Signs) == 0 {
		err = errors.New("数据不可用")
		goto END
	}
	if len(resTokenView.Signs) < len(utils.Nodes)/2+1 {
		err = errors.New("节点不够用")
		goto END
	}

	for _, node := range resTokenView.Signs {
		sumPrice += node.PriceUsd
	}
	resTokenView.Timestamp=int64(timestamp)
	resTokenView.PriceUsd = sumPrice / float64(len(resTokenView.Signs))
	resTokenView.BigPrice=services.GetUnDecimalPrice(float64(resTokenView.PriceUsd)).String()
	resTokenView.Sign=services.SignMsg(resTokenView.GetHash())
	c.JSON(200, resTokenView)
	return
END:
	if err == nil {
		ErrJson(c, err.Error())
	}
}

// @Tags default
// @Summary　获取token价格信息,内部单节点
// @Description 内部单节点获取token信息,含pair的lp Token内容
// @ID TokenPriceHandler
// @Accept  json
// @Produce  json
// @Param     token   path    string     true        "token地址" default(0x66a0f676479cee1d7373f3dc2e2952778bff5bd6)
// @Param     timestamp   path    int     false    "当前时间的unix秒数,该字段未使用，仅在云存储上用于标识" default(1620383144)
// @Success 200 {object} services.PriceView	"stock info"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
// @Router /pub/internal/dex/token_price/{token}/{timestamp} [get]
func TokenPriceHandler(c *gin.Context) {
	code:=c.Param("token")
	timestampstr:=c.Param("timestamp")
	timestamp,_:=strconv.Atoi(timestampstr)
	res,err:=services.GetTokenInfo(code)
	if err == nil {
		price:=services.BlockPrice{}.GetPrice()
		fprice,_:=strconv.ParseFloat(res.DerivedETH,64)
		res.PriceUsd=fprice*price

		tPriceView:=new(services.PriceView)
		tPriceView.Timestamp=int64(timestamp)
		tPriceView.PriceUsd=res.PriceUsd
		tPriceView.PriceUsd= math.Trunc(tPriceView.PriceUsd*100)/100
		tPriceView.BigPrice=services.GetUnDecimalPrice(float64(res.PriceUsd)).String()
		tPriceView.Sign=services.SignMsg(tPriceView.GetHash())
		c.JSON(200,tPriceView)
		return
	}
	if err != nil {
		ErrJson(c,err.Error())
		return
	}
}

// @Tags default
// @Summary　获取token信息,内部单节点
// @Description 内部单节点获取token信息
// @ID TokenInfoHandler
// @Accept  json
// @Produce  json
// @Param     token   path    string     true        "token地址" default(0x66a0f676479cee1d7373f3dc2e2952778bff5bd6)
// @Param     timestamp   path    int     false    "当前时间的unix秒数,该字段未使用，仅在云存储上用于标识" default(1620383144)
// @Success 200 {object} services.TokenInfo	"stock info"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
// @Router /pub/internal/dex/token_info/{token}/{timestamp} [get]
func TokenInfoHandler(c *gin.Context) {
	code:=c.Param("token")
	//timestampstr:=c.Param("timestamp")
	//timestamp,_:=strconv.Atoi(timestampstr)
	res,err:=services.GetTokenInfo(code)
	if err == nil {
		price:=services.BlockPrice{}.GetPrice()
		fprice,_:=strconv.ParseFloat(res.DerivedETH,64)
		res.PriceUsd=fprice*price

		ost,err1:=services.GetTokenInfosForStat(code,price)
		if err1 != nil {
			log.Println("GetTokenInfosForStat err",err)
		}
		res.OneDayStat=ost
		c.JSON(200,res)
		return
	}
	if err != nil {
		ErrJson(c,err.Error())
		return
	}
}


// @Tags default
// @Summary　获取token不同时间区间的价格图表信息
// @Description 获取token不同时间区间的价格图表信息
// @ID TokenDayPricesHandler
// @Accept  json
// @Produce  json
// @Param     token   path    string     true        "token地址" default(0x66a0f676479cee1d7373f3dc2e2952778bff5bd6)
// @Param     count   path    int     true    "获取多少个数据点" default(10)
// @Param     interval   path    string     true    "数据间隔 15minite hour day 1w(1周) 1m (1月) " default(day) Enums(15minite,hour,day,1w,1m)
// @Param     timestamp   path    int     false    "当前时间的unix秒数,该字段未使用，仅在云存储上用于标识" default(1620383144)
// @Success 200 {array} services.BlockPrice	"stock info"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
// @Router /pub/dex/token_chart_prices/{token}/{count}/{interval}/{timestamp} [get]
func TokenDayPricesHandler(c *gin.Context) {
	code:=c.Param("token")
	interval:=c.Param("interval")
	day_str:=c.Param("count")
	count,_:=strconv.Atoi(day_str)
	//timestampstr:=c.Param("timestamp")
	//timestamp,_:=strconv.Atoi(timestampstr)
	bs,err:=services.GetTokenTimesPrice(code,interval,count)
	if err == nil {
		c.JSON(200,bs)
		return
	}
	if err != nil {
		ErrJson(c,err.Error())
		return
	}
}

// @Tags default
// @Summary　获取token相应天数的统计图表信息
// @Description 获取token相应天数的统计图表信息
// @ID TokenDayDatasHandler
// @Accept  json
// @Produce  json
// @Param     token   path    string     true        "token地址" default(0x66a0f676479cee1d7373f3dc2e2952778bff5bd6)
// @Param     days   path    int     true    "获取最近多少天的数据" default(14)
// @Param     timestamp   path    int     false    "当前时间的unix秒数,该字段未使用，仅在云存储上用于标识" default(1620383144)
// @Success 200 {array} services.TokenDayData	"stock info"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
// @Router /pub/dex/token_day_datas/{token}/{days}/{timestamp} [get]
func TokenDayDatasHandler(c *gin.Context) {
	code:=c.Param("token")
	day_str:=c.Param("days")
	days,_:=strconv.Atoi(day_str)
	//timestampstr:=c.Param("timestamp")
	//timestamp,_:=strconv.Atoi(timestampstr)
	bs,err:=services.GetTokenDayData(code,days)
	if err == nil {
		c.JSON(200,json.RawMessage(bs))
		return
	}
	if err != nil {
		ErrJson(c,err.Error())
		return
	}
}

// @Tags default
// @Summary　获取lp价格信息
// @Description 获取lp价格信息
// @ID PairLpPriceSignHandler
// @Accept  json
// @Produce  json
// @Param     pair   path    string     true        "token地址" default(0x21b8065d10f73ee2e260e5b47d3344d3ced7596e)
// @Param     timestamp   path    int     false    "当前时间的unix秒数,该字段未使用，仅在云存储上用于标识" default(1620383144)
// @Success 200 {object} services.PairInfo	"pair info"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
// @Router /pub/dex/lp_price/{pair}/{timestamp} [get]
func PairLpPriceSignHandler(c *gin.Context) {
	code := c.Param("pair")
	timestampstr := c.Param("timestamp")
	timestamp, _ := strconv.Atoi(timestampstr)
	resTokenView := new(services.DataPriceView)
	//var addres []*services.PriceView
	sc := sync.RWMutex{}
	wg := new(sync.WaitGroup)
	var porcNode = func(nodeUrl string) {
		defer wg.Done()
		reqUrl := fmt.Sprintf(nodeUrl+"/pub/internal/dex/lp_price/%s/%d", code, timestamp)
		bs, err := utils.ReqResBody(reqUrl, "", "GET", nil, nil)
		if err == nil {
			token := new(services.PriceView)
			err = json.Unmarshal(bs, token)
			if err == nil {
				log.Println(err)
			}
			token.Node = nodeUrl
			sc.Lock()
			resTokenView.Signs = append(resTokenView.Signs, token)
			sc.Unlock()
		}else{
			log.Println("PairLpPriceSignHandler",err)
		}

	}
	for _, nurl := range utils.Nodes {
		wg.Add(1)
		go porcNode(nurl)
	}

	wg.Wait()
	var err error
	sumPrice := float64(0.0)
	if len(resTokenView.Signs) == 0 {
		err = errors.New("数据不可用")
		goto END
	}
	if len(resTokenView.Signs) < len(utils.Nodes)/2+1 {
		err = errors.New("节点不够用")
		goto END
	}
	for _, node := range resTokenView.Signs {
		sumPrice += node.PriceUsd
	}
	resTokenView.Timestamp=int64(timestamp)
	resTokenView.PriceUsd = sumPrice / float64(len(resTokenView.Signs))
	resTokenView.BigPrice=services.GetUnDecimalPrice(float64(resTokenView.PriceUsd)).String()
	resTokenView.Sign=services.SignMsg(resTokenView.GetHash())
	c.JSON(200, resTokenView)
	return
END:
	if err == nil {
		ErrJson(c, err.Error())
	}
}

// @Tags default
// @Summary　获取lp价格信息,内部单节点:
// @Description 内部单节点获取lp价格信息
// @ID PairLpPriceHandler
// @Accept  json
// @Produce  json
// @Param     pair   path    string     true        "pair地址" default(0x21b8065d10f73ee2e260e5b47d3344d3ced7596e)
// @Param     timestamp   path    int     false    "当前时间的unix秒数,该字段未使用，仅在云存储上用于标识" default(1620383144)
// @Success 200 {object} services.PairInfo	"stock info"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
// @Router /pub/internal/dex/lp_price/{pair}/{timestamp} [get]
func PairLpPriceHandler(c *gin.Context) {
	code:=c.Param("pair")
	timestampstr:=c.Param("timestamp")
	timestamp,_:=strconv.Atoi(timestampstr)
	res,err:=services.GetPairInfo(code)
	if err == nil {
		//price:=services.BlockPrice{}.GetPrice()
		sulpply,_:=strconv.ParseFloat(res.TotalSupply,64)
		allUsd,_:=strconv.ParseFloat(res.ReserveUSD,64)
		tPriceView:=new(services.PriceView)
		tPriceView.Timestamp=int64(timestamp)
		tPriceView.PriceUsd=allUsd/sulpply
		tPriceView.PriceUsd= math.Trunc(tPriceView.PriceUsd*100)/100
		tPriceView.BigPrice=services.GetUnDecimalPrice(float64(tPriceView.PriceUsd)).String()
		tPriceView.Sign=services.SignMsg(tPriceView.GetHash())
		c.JSON(200,tPriceView)
		return
	}
	if err != nil {
		ErrJson(c,err.Error())
		return
	}
}

