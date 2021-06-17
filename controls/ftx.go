package controls

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"stock/services"
	"strconv"
	"stock/utils"
)


// @Tags default
// @Summary　获取btc杠杆币价格，内部单节点模式 目前
// @Description 获取btc杠杆币价格，内部单节点模式
// @ID FtxPriceHandler
// @Accept  json
// @Produce  json
// @Param     coin_type   path    string     true        "ftx类型" default(btc3x)  Enums(btc3x, eth3x, vix3x, ust20x, gold10x, eur20x,ndx10x)
// @Param     timestamp   query    int     false    "unix 秒数" default(1620383144)
// @Success 200 {object} services.CoinPriceView	"Coin Price"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} controls.ApiErr "失败时，有相应测试日志输出"
// @Router /pub/internal/ftx_price/{coin_type} [get]
func FtxPriceHandler(c *gin.Context) {
	coin:=c.Param("coin_type")
	//vsCoin:=c.Param("vs_coin")
	//timestampstr:=c.Query("timestamp")
	//timestamp,_:=strconv.Atoi(timestampstr)

	cb:=new(services.CoinBull)
	err:=utils.Orm.Order("id desc").Where("coin_type=?",coin).First(cb).Error
	//log.Println(cs)

	if err == nil {
		//targetPrice:=cs.VsCoin/cs.Coin
		//log.Println(targetPrice)
		//
		//tPriceView:=new(services.CoinPriceView)
		//tPriceView.Coin=coin
		//tPriceView.VsCoin=vsCoin
		//tPriceView.Timestamp=int64(timestamp)
		//tPriceView.Price=targetPrice
		//tPriceView.Price= math.Trunc(tPriceView.Price*100)/100
		//tPriceView.BigPrice=services.GetUnDecimalPrice(float64(tPriceView.Price)).String()
		//tPriceView.Sign=services.SignMsg(tPriceView.GetHash())

		c.JSON(200,cb)
		return
	}
	if err != nil {
		ErrJson(c,err.Error())
		return
	}
}

// @Tags default
// @Summary　获取杠杆btc代币不同时间区间的价格图表信息
// @Description 获取杠杆btc代币不同时间区间的价格图表信息
// @ID FtxChartPricesHandler
// @Accept  json
// @Produce  json
// @Param     coin_type   path    string     true        "ftx类型" default(btc3x)  Enums(btc3x, eth3x, vix3x, ust20x, gold10x, eur20x,ndx10x)
// @Param     count   path    int     true    "获取多少个数据点" default(10)
// @Param     interval   path    int     true    "数据间隔值,表示多少个15分钟, 如:1表示15分钟间隔 2表示30分钟间隔 3表示45分钟间隔 ,96表示1天间隔 ；" default(day) Enums(1,2,3,4,96)
// @Param     timestamp   path    int     false    "当前时间的unix秒数,该字段未使用，仅在云存储上用于标识" default(1620383144)
// @Success 200 {array} services.BlockPrice	"stock info"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
// @Router /pub/dex/ftx_chart_prices/{coin_type}/{count}/{interval}/{timestamp} [get]
func FtxChartPricesHandler(c *gin.Context) {
	coin_type:=c.Param("coin_type")
	//code:="btc"
	interval_str:=c.Param("interval")
	interval,_:=strconv.Atoi(interval_str)
	count_str:=c.Param("count")
	count,_:=strconv.Atoi(count_str)


	//SetCacheRes(c,ckey,false,proc,c.Query("debug")=="1")
	ckey:=fmt.Sprintf("FtxChartPricesHandler-%s-%s",interval_str,count_str)
	proc:= func()(interface{},error) {
		items,err:=services.GetFtxTimesPrice(coin_type,interval,count)
		if err != nil {
			return nil,err
		}
		return items,err
	}
	SetCacheResExpire(c,ckey,false,600,proc,c.Query("debug")=="1")
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
