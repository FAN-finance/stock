package controls

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"log"
	"math"
	"stock/services"
	"stock/utils"
	"strconv"
	"strings"
	"sync"
)

// @Tags default
// @Summary　获取token链上价格信息
// @Description 获取token链上价格信息，使用节点从eth或bsc合约事件监听到的价格变化数据；token信息要提前在节点配制才能被监听
// @ID TokenChainPriceSignHandler
// @Accept  json
// @Produce  json
// @Param     token   path    string     true        "token地址" default(0xc61624355667e4d5ca9cee25ad339c990a90eaea)
// @Param     data_type   path    int     true   "最高最低价１最高　２最低价" default(1) Enums(1,2)
// @Param     timestamp   path    int     false    "当前时间的unix秒数,该字段未使用，仅在云存储上用于标识" default(1620383144)
//@Param     debug   query    int     false    "调试" default(0)
// @Success 200 {object} services.HLDataPriceView	"token price info"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
// @Router /pub/dex/token_chain_price/{token}/{data_type}/{timestamp} [get]
func TokenChainPriceSignHandler(c *gin.Context) {
	tokenPriceSignProces(c, "/pub/internal/dex/token_chain_price/%s/%d?data_type=%d")
}

// @Tags default
// @Summary　获取token价格信息
// @Description 获取token价格信息
// @ID TokenPriceSignHandler
// @Accept  json
// @Produce  json
// @Param     token   path    string     true        "token地址" default(0x66a0f676479cee1d7373f3dc2e2952778bff5bd6)
// @Param     data_type   path    int     true   "最高最低价１最高　２最低价" default(1) Enums(1,2)
// @Param     timestamp   path    int     false    "当前时间的unix秒数,该字段未使用，仅在云存储上用于标识" default(1620383144)
//@Param     debug   query    int     false    "调试" default(0)
// @Success 200 {object} services.HLDataPriceView	"token price info"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
// @Router /pub/dex/token_price/{token}/{data_type}/{timestamp} [get]
func TokenPriceSignHandler(c *gin.Context) {
	tokenPriceSignProces(c, "/pub/internal/dex/token_price/%s/%d?data_type=%d")
}

// @Tags Pair
// @Summary　从Pair获取token价格信息
// @Description 从Pair获取token价格信息
// @ID PairTokenPriceSignHandler
// @Accept  json
// @Produce  json
// @Param     pair   path    string     true        "token地址" default(0x4612b8de9fb6281f6d5aa29635cf5700148d1b67)
// @Param     token   path    string     true        "token地址" default(0x5df42c20d79fe40b51aba8fe5c8aa6531a3c453b)
// @Param     data_type   path    int     true   "最高最低价１最高　２最低价" default(1) Enums(1,2)
// @Param     timestamp   path    int     false    "当前时间的unix秒数,该字段未使用，仅在云存储上用于标识" default(1620383144)
// @Param     debug   query    int     false    "调试" default(0)
// @Success 200 {object} services.HLDataPriceView	"token price info"
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
// @Router /pub/dex/pair/token_price/{pair}/{token}/{data_type}/{timestamp} [get]
func PairTokenPriceSignHandler(c *gin.Context) {
	pair := c.Param("pair")
	token := c.Param("token")
	pair = strings.ToLower(pair)
	token = strings.ToLower(token)

	timestampstr := c.Param("timestamp")
	timestamp, _ := strconv.Atoi(timestampstr)

	dataTypeStr := c.Param("data_type")
	dataType, _ := strconv.Atoi(dataTypeStr)

	resTokenView := new(services.HLDataPriceView)
	avgNodesPrice := []*services.HLPriceView{}

	sc := sync.RWMutex{}
	wg := new(sync.WaitGroup)
	var porcNode = func(nodeUrl string) {
		defer wg.Done()
		reqUrl := fmt.Sprintf(nodeUrl+"/pub/internal/dex/pair/token_price/%s/%s/%d?data_type=%d", pair, token, timestamp, dataType)
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
			_ = json.Unmarshal(bs, snode)
			snode.Node = nodeUrl
			isMyData, _ := services.Verify(snode.GetHash(), snode.Sign, services.WalletAddre)
			if isMyData {
				resTokenView.PriceUsd = snode.PriceUsd
				resTokenView.BigPrice = snode.BigPrice
				resTokenView.Timestamp = snode.Timestamp
				resTokenView.Code = snode.Code
				resTokenView.DataType = dataType
				resTokenView.Sign = nil
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
	resTokenView.Signs = nil
	c.JSON(200, resTokenView)
	return

END:
	if err != nil {
		ErrJson(c, err.Error())
	}
}

func tokenPriceSignProces(c *gin.Context, providerUrl string) {
	code := c.Param("token")
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
		reqUrl := fmt.Sprintf(nodeUrl+providerUrl, code, timestamp, dataType)
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

type HLValuePair struct {
	High float64
	Low  float64
}

// @Tags default
// @Summary　获取token最近一小时最高最低价格信息,内部单节点模式
// @Description 获取token最近一小时最高最低价格信息；目前改为使用直接从链上监听的数据．
// @ID TokenChainPriceHandler
// @Accept  json
// @Produce  json
// @Param     token   path    string     true        "token地址" default(0x66a0f676479cee1d7373f3dc2e2952778bff5bd6)
// @Param     data_type   query    int     true   "最高最低价１最高　２最低价" default(1) Enums(1,2)
// @Param     timestamp   path    int     false    "当前时间的unix秒数,该字段未使用，仅在云存储上用于标识" default(1620383144)
// @Success 200 {object} services.HLPriceView	"Price View"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
// @Router /pub/internal/dex/token_chain_price/{token}/{timestamp} [get]
func TokenChainPriceHandler(c *gin.Context) {
	dataProc := func(code string) (interface{}, error) {
		vp := new(HLValuePair)
		err := utils.Orm.Raw(`
select *
from (
        select max(prices.token_price) high, min(prices.token_price) low
        from token_prices prices
        where prices.token_addre=? and prices.block_number >
              (select t.id from block_prices t where t.block_time > unix_timestamp() - 3600 limit 1)
    ) a
where a.high is not null`, code).First(vp).Error
		if err == nil {
			return vp, err
		}
		err = utils.Orm.Raw(`select prices.token_price high, prices.token_price low
			from token_prices prices where prices.token_addre=?
			order by id desc
			limit 1;`, code).First(vp).Error
		return vp, err

	}
	TokenChainPriceProcess(c, dataProc, "TokenChainPriceHandler")

}
func TokenChainPriceProcess(c *gin.Context, dataProc func(code string) (interface{}, error), processName string) {
	code := c.Param("token")
	timestampstr := c.Param("timestamp")
	timestamp, _ := strconv.Atoi(timestampstr)
	dataTypeStr := c.Query("data_type")
	dataType, _ := strconv.Atoi(dataTypeStr)

	//SetCacheRes(c,ckey,false,proc,c.Query("debug")=="1")
	ckey := fmt.Sprintf(processName+"-%s", code)

	proc := func() (interface{}, error) {
		return dataProc(code)
	}
	var res interface{}
	var err error
	if c.Query("debug") == "1" {
		res, err = proc()
	} else {
		log.Println("cache process", ckey)
		res, err = utils.CacheFromLru(1, ckey, int(30), proc)
	}
	//res, err := services.GetTokenInfo(code)
	if err == nil {
		//price := services.BlockPrice{}.GetPrice()
		//fprice, _ := strconv.ParseFloat(res.DerivedETH, 64)
		//res.PriceUsd = fprice * price

		vp := res.(*HLValuePair)
		log.Println(*vp)
		tPriceView := new(services.HLPriceView)
		tPriceView.Code = code
		if code == "0xc011a73ee8576fb46f5e1c5751ca3b9fe0af2a6f" {
			tPriceView.Code = "0x6EBFD2E7678cFA9c8dA11b9dF00DB24a35ec7dD4"
		}
		tPriceView.Timestamp = int64(timestamp)
		tPriceView.DataType = dataType
		if dataType == 1 {
			tPriceView.PriceUsd = vp.High
		}
		if dataType == 2 {
			tPriceView.PriceUsd = vp.Low
		}
		//tPriceView.PriceUsd = math.Trunc( tPriceView.PriceUsd*1000) / 1000
		tPriceView.PriceUsd, _ = decimal.NewFromFloat(tPriceView.PriceUsd).Round(18).Float64()
		tPriceView.BigPrice = services.GetUnDecimalPrice(tPriceView.PriceUsd).String()
		tPriceView.NodeAddress = services.WalletAddre
		if tPriceView.PriceUsd > 0.001 {
			tPriceView.Sign = services.SignMsg(tPriceView.GetHash())
		}
		c.JSON(200, tPriceView)
		return
	}
	if err != nil {
		ErrJson(c, err.Error())
		return
	}
}

// @Tags default
// @Summary　获取token最近一小时最高最低价格信息,内部单节点
// @Description 获取token最近一小时最高最低价格信息；．
// @ID TokenPriceHandler
// @Accept  json
// @Produce  json
// @Param     token   path    string     true        "token地址" default(0x66a0f676479cee1d7373f3dc2e2952778bff5bd6)
// @Param     data_type   query    int     true   "最高最低价１最高　２最低价" default(1) Enums(1,2)
// @Param     timestamp   path    int     false    "当前时间的unix秒数,该字段未使用，仅在云存储上用于标识" default(1620383144)
// @Success 200 {object} services.HLPriceView	"Price View"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
// @Router /pub/internal/dex/token_price/{token}/{timestamp} [get]
func TokenPriceHandler(c *gin.Context) {
	dataProc := func(code string) (interface{}, error) {
		intreval := "60s"
		count := 60
		items, err := services.GetTokenTimesPrice(code, intreval, count)
		if err != nil {
			return nil, err
		}
		vp := new(HLValuePair)
		for index, item := range items {
			vp.High = math.Max(vp.High, item.Price)
			if index == 0 {
				vp.Low = item.Price
			}
			vp.Low = math.Min(vp.Low, item.Price)
		}
		return vp, err
	}
	TokenChainPriceProcess(c, dataProc, "TokenPriceHandler")
}

// @Tags Pair
// @Summary　从Pair获取token最近一小时最高最低价格信息,内部单节点
// @Description 从Pair获取token最近一小时最高最低价格信息；．
// @ID PairTokenPriceHandler
// @Accept  json
// @Produce  json
// @Param     pair   path    string     true        "token地址" default(0x4612b8de9fb6281f6d5aa29635cf5700148d1b67)
// @Param     token   path    string     true        "token地址" default(0x5df42c20d79fe40b51aba8fe5c8aa6531a3c453b)
// @Param     data_type   query    int     true   "最高最低价１最高　２最低价" default(1) Enums(1,2)
// @Param     timestamp   path    int     false    "当前时间的unix秒数,该字段未使用，仅在云存储上用于标识" default(1620383144)
// @Success 200 {object} services.HLPriceView	"Price View"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
// @Router /pub/internal/dex/pair/token_price/{pair}/{token}/{timestamp} [get]
func PairTokenPriceHandler(c *gin.Context) {
	dataProc := func(pair, token string) (interface{}, error) {
		intreval := "60s"
		count := 60
		items, err := services.GetTokenTimesPriceFromPair(pair, token, intreval, count)
		if err != nil {
			return nil, err
		}
		vp := new(HLValuePair)
		for index, item := range items {
			vp.High = math.Max(vp.High, item.Price)
			if index == 0 {
				vp.Low = item.Price
			}
			vp.Low = math.Min(vp.Low, item.Price)
		}
		return vp, err
	}
	TokenChainPriceFromPairProcess(c, dataProc, "PairTokenPriceHandler")
}

func TokenChainPriceFromPairProcess(c *gin.Context, dataProc func(pair, token string) (interface{}, error), processName string) {
	pair := c.Param("pair")
	token := c.Param("token")
	pair = strings.ToLower(pair)
	token = strings.ToLower(token)
	timestampstr := c.Param("timestamp")
	timestamp, _ := strconv.Atoi(timestampstr)
	dataTypeStr := c.Query("data_type")
	dataType, _ := strconv.Atoi(dataTypeStr)

	//SetCacheRes(c,ckey,false,proc,c.Query("debug")=="1")
	ckey := fmt.Sprintf(processName+"%s-%s", pair, token)

	proc := func() (interface{}, error) {
		return dataProc(pair, token)
	}
	var res interface{}
	var err error
	if c.Query("debug") == "1" {
		res, err = proc()
	} else {
		log.Println("cache process", ckey)
		res, err = utils.CacheFromLru(1, ckey, int(30), proc)
	}
	//res, err := services.GetTokenInfo(code)
	if err == nil {
		//price := services.BlockPrice{}.GetPrice()
		//fprice, _ := strconv.ParseFloat(res.DerivedETH, 64)
		//res.PriceUsd = fprice * price

		vp := res.(*HLValuePair)
		log.Println(*vp)
		tPriceView := new(services.HLPriceView)
		tPriceView.Code = token
		tPriceView.Timestamp = int64(timestamp)
		tPriceView.DataType = dataType
		if dataType == 1 {
			tPriceView.PriceUsd = vp.High
		}
		if dataType == 2 {
			tPriceView.PriceUsd = vp.Low
		}
		//tPriceView.PriceUsd = math.Trunc( tPriceView.PriceUsd*1000) / 1000
		tPriceView.PriceUsd, _ = decimal.NewFromFloat(tPriceView.PriceUsd).Round(18).Float64()
		tPriceView.BigPrice = services.GetUnDecimalPrice(tPriceView.PriceUsd).String()
		tPriceView.NodeAddress = services.WalletAddre
		if tPriceView.PriceUsd > 0.001 {
			tPriceView.Sign = services.SignMsg(tPriceView.GetHash())
		}
		c.JSON(200, tPriceView)
		return
	}
	if err != nil {
		ErrJson(c, err.Error())
		return
	}
}

// @Tags default
// @Summary　获取token平均价格共识:
// @Description 获取token平均价格共识
// @ID TokenAvgHlPriceHandler
// @Accept  json
// @Produce  json
// @Param   nodePrices  body   []services.HLPriceView true       "节点价格列表"
// @Success 200 {object} services.HLPriceView	"stock info"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} controls.ApiErr "失败时，有相应测试日志输出"
// @Router /pub/internal/token_avgprice [post]
func TokenAvgHlPriceHandler(c *gin.Context) {
	nodePrices := []*services.HLPriceView{}
	err := c.BindJSON(&nodePrices)
	if err == nil {
		if len(nodePrices) == 0 {
			err = errors.New("数据不可用")
			ErrJson(c, err.Error())
			return
		}
		if len(nodePrices) < len(utils.Nodes)/2+1 {
			err = errors.New("节点不够用")
			ErrJson(c, err.Error())
			return
		}

		timestamp := nodePrices[0].Timestamp
		code := nodePrices[0].Code
		dataType := nodePrices[0].DataType

		sumPrice := decimal.NewFromFloat(0.0)
		for _, node := range nodePrices {
			sumPrice = sumPrice.Add(decimal.NewFromFloat(node.PriceUsd))
			//验证数据
			if timestamp != node.Timestamp || code != node.Code || dataType != node.DataType {
				err = errors.New("需要共识的数据不一致")
				break
			}
			//TODO 验证数据签名
			//services.Verify(node.GetHash(),node.Sign,"")
		}
		if err != nil {
			ErrJson(c, err.Error())
			return
		}
		sdata := new(services.HLPriceView)
		sdata.PriceUsd, _ = sumPrice.DivRound(decimal.NewFromInt(int64(len(nodePrices))), 18).Float64()
		//TODO 需要检测平均价格和当前自己的价格是否超出了千分之一的误差
		sdata.BigPrice = services.GetUnDecimalPrice(sdata.PriceUsd).String()
		sdata.Timestamp = int64(timestamp)
		sdata.Code = code
		sdata.DataType = dataType
		sdata.NodeAddress = services.WalletAddre
		if sdata.PriceUsd > 0.001 {
			sdata.Sign = services.SignMsg(sdata.GetHash())
		}
		c.JSON(200, sdata)
		return
	}
	if err != nil {
		ErrJson(c, err.Error())
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
	code := c.Param("token")
	//timestampstr:=c.Param("timestamp")
	//timestamp,_:=strconv.Atoi(timestampstr)

	//SetCacheRes(c,ckey,false,proc,c.Query("debug")=="1")
	ckey := fmt.Sprintf("TokenInfoHandler-%s", code)
	proc := func() (interface{}, error) {
		res, err := services.GetTokenInfo(code)
		if err == nil {
			price := services.BlockPrice{}.GetPrice()
			fprice, _ := decimal.NewFromString(res.DerivedETH)
			res.PriceUsd, _ = fprice.Mul(decimal.NewFromFloat(price)).Round(18).Float64()

			//ost, err1 := services.GetTokenInfosForStat(code, price)
			//if err1 != nil {
			//	log.Println("GetTokenInfosForStat err", err)
			//}
			//res.OneDayStat = ost
			return res, nil
		}
		return res, err
	}
	SetCacheRes(c, ckey, false, proc, c.Query("debug") == "1")

	//
	//res,err:=services.GetTokenInfo(code)
	//if err == nil {
	//	price:=services.BlockPrice{}.GetPrice()
	//	fprice,_:=strconv.ParseFloat(res.DerivedETH,64)
	//	res.PriceUsd=fprice*price
	//
	//	ost,err1:=services.GetTokenInfosForStat(code,price)
	//	if err1 != nil {
	//		log.Println("GetTokenInfosForStat err",err)
	//	}
	//	res.OneDayStat=ost
	//	c.JSON(200,res)
	//	return
	//}
	//if err != nil {
	//	ErrJson(c,err.Error())
	//	return
	//}
}

// @Tags Pair
// @Summary　从pair获取token信息,内部单节点
// @Description 内部单节点从pair获取token信息
// @ID PairTokenInfoHandler
// @Accept  json
// @Produce  json
// @Param     pair   path    string     true        "token地址" default(0x4612b8de9fb6281f6d5aa29635cf5700148d1b67)
// @Param     token   path    string     true        "token地址" default(0x5df42c20d79fe40b51aba8fe5c8aa6531a3c453b)
// @Param     timestamp   path    int     false    "当前时间的unix秒数,该字段未使用，仅在云存储上用于标识" default(1620383144)
// @Success 200 {object} services.TokenInfo	"stock info"
// @Header 200 {string} sign "签名信息"
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
// @Router /pub/internal/dex/pair/token_info/{pair}/{token}/{timestamp} [get]
func PairTokenInfoHandler(c *gin.Context) {
	pair := c.Param("pair")
	token := c.Param("token")
	pair = strings.ToLower(pair)
	token = strings.ToLower(token)
	ckey := fmt.Sprintf("PairTokenInfoHandler-%s-%s", pair, token)
	proc := func() (interface{}, error) {
		res, err := services.GetTokenInfoFromPair(pair, token)
		if err == nil {
			return res, nil
		}
		return res, err
	}
	SetCacheRes(c, ckey, false, proc, c.Query("debug") == "1")
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
	code := c.Param("token")
	interval := c.Param("interval")
	day_str := c.Param("count")
	count, _ := strconv.Atoi(day_str)

	//SetCacheRes(c,ckey,false,proc,c.Query("debug")=="1")
	ckey := fmt.Sprintf("TokenDayPricesHandler-%s-%s-%s", code, interval, day_str)
	proc := func() (interface{}, error) {
		items, err := services.GetTokenTimesPrice(code, interval, count)
		if err != nil {
			return nil, err
		}
		return items, err
	}
	SetCacheRes(c, ckey, false, proc, c.Query("debug") == "1")

	//
	////timestampstr:=c.Param("timestamp")
	////timestamp,_:=strconv.Atoi(timestampstr)
	//bs,err:=services.GetTokenTimesPrice(code,interval,count)
	//if err == nil {
	//	c.JSON(200,bs)
	//	return
	//}
	//if err != nil {
	//	ErrJson(c,err.Error())
	//	return
	//}
}

// @Tags Pair
// @Summary　从Pair获取token不同时间区间的价格图表信息
// @Description 从Pair获取token不同时间区间的价格图表信息
// @ID PairTokenDayPricesHandler
// @Accept  json
// @Produce  json
// @Param     pair   path    string     true        "token地址" default(0x4612b8de9fb6281f6d5aa29635cf5700148d1b67)
// @Param     token   path    string     true        "token地址" default(0x5df42c20d79fe40b51aba8fe5c8aa6531a3c453b)
// @Param     count   path    int     true    "获取多少个数据点" default(10)
// @Param     interval   path    string     true    "数据间隔 15minite hour day 1w(1周) 1m (1月) " default(day) Enums(15minite,hour,day,1w,1m)
// @Param     timestamp   path    int     false    "当前时间的unix秒数,该字段未使用，仅在云存储上用于标识" default(1620383144)
// @Success 200 {array} services.BlockPrice	"stock info"
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
// @Router /pub/dex/pair/token_chart_prices/{pair}/{token}/{count}/{interval}/{timestamp} [get]
func PairTokenDayPricesHandler(c *gin.Context) {
	pair := c.Param("pair")
	token := c.Param("token")

	pair = strings.ToLower(pair)
	token = strings.ToLower(token)

	interval := c.Param("interval")
	day_str := c.Param("count")
	count, _ := strconv.Atoi(day_str)

	ckey := fmt.Sprintf("PairTokenDayPricesHandler-%s-%s-%s-%s", pair, token, interval, day_str)
	proc := func() (interface{}, error) {
		items, err := services.GetTokenTimesPriceFromPair(pair, token, interval, count)
		if err != nil {
			return nil, err
		}
		return items, err
	}
	SetCacheRes(c, ckey, false, proc, c.Query("debug") == "1")
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
	code := c.Param("token")
	day_str := c.Param("days")

	//SetCacheRes(c,ckey,false,proc,c.Query("debug")=="1")
	ckey := fmt.Sprintf("TokenDayDatasHandler-%s-%s", code, day_str)
	proc := func() (interface{}, error) {
		days, _ := strconv.Atoi(day_str)
		//timestampstr:=c.Param("timestamp")
		//timestamp,_:=strconv.Atoi(timestampstr)
		bs, err := services.GetTokenDayData(code, days)
		if err == nil {
			return json.RawMessage(bs), nil
		}
		return nil, err
	}
	SetCacheRes(c, ckey, false, proc, c.Query("debug") == "1")

	//
	//days,_:=strconv.Atoi(day_str)
	////timestampstr:=c.Param("timestamp")
	////timestamp,_:=strconv.Atoi(timestampstr)
	//bs,err:=services.GetTokenDayData(code,days)
	//if err == nil {
	//	c.JSON(200,json.RawMessage(bs))
	//	return
	//}
	//if err != nil {
	//	ErrJson(c,err.Error())
	//	return
	//}
}

// @Tags default
// @Summary　获取lp价格信息
// @Description 获取lp价格信息
// @ID PairLpPriceSignHandler
// @Accept  json
// @Produce  json
// @Param     pair   path    string     true        "token地址" default(0x21b8065d10f73ee2e260e5b47d3344d3ced7596e)
// @Param     timestamp   path    int     false    "当前时间的unix秒数,该字段未使用，仅在云存储上用于标识" default(1620383144)
// @Success 200 {array} services.PriceView	"pair price view list"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
// @Router /pub/dex/lp_price/{pair}/{timestamp} [get]
func PairLpPriceSignHandler(c *gin.Context) {
	code := c.Param("pair")
	timestampstr := c.Param("timestamp")
	timestamp, _ := strconv.Atoi(timestampstr)

	//SetCacheRes(c,ckey,false,proc,c.Query("debug")=="1")
	ckey := fmt.Sprintf("PairLpPriceSignHandler-%s", code)
	proc := func() (interface{}, error) {
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
				if err != nil {
					log.Println(err)
				}
				token.Node = nodeUrl
				sc.Lock()
				resTokenView.Signs = append(resTokenView.Signs, token)
				sc.Unlock()
			} else {
				log.Println("PairLpPriceSignHandler", err)
			}

		}
		for _, nurl := range utils.Nodes {
			wg.Add(1)
			go porcNode(nurl)
		}

		wg.Wait()
		var err error
		sumPrice := decimal.NewFromFloat(0)
		if len(resTokenView.Signs) == 0 {
			err = errors.New("数据不可用")
			return nil, err
		}
		if len(resTokenView.Signs) < len(utils.Nodes)/2+1 {
			err = errors.New("节点不够用")
			return nil, err

		}
		for _, node := range resTokenView.Signs {
			sumPrice = sumPrice.Add(decimal.NewFromFloat(node.PriceUsd))
		}
		resTokenView.Timestamp = int64(timestamp)
		resTokenView.PriceUsd, _ = sumPrice.DivRound(decimal.NewFromInt(int64(len(resTokenView.Signs))), 18).Float64()
		resTokenView.BigPrice = services.GetUnDecimalPrice(resTokenView.PriceUsd).String()
		resTokenView.Sign = services.SignMsg(resTokenView.GetHash())
		return resTokenView, nil
	}
	SetCacheRes(c, ckey, false, proc, c.Query("debug") == "1")

	//	c.JSON(200, resTokenView)
	//	return
	//END:
	//	if err == nil {
	//		ErrJson(c, err.Error())
	//	}
}

// @Tags default
// @Summary　获取lp价格信息,内部单节点:
// @Description 内部单节点获取lp价格信息
// @ID PairLpPriceHandler
// @Accept  json
// @Produce  json
// @Param     pair   path    string     true        "pair地址" default(0x21b8065d10f73ee2e260e5b47d3344d3ced7596e)
// @Param     timestamp   path    int     false    "当前时间的unix秒数,该字段未使用，仅在云存储上用于标识" default(1620383144)
// @Success 200 {object} services.PriceView	"price vew"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
// @Router /pub/internal/dex/lp_price/{pair}/{timestamp} [get]
func PairLpPriceHandler(c *gin.Context) {
	code := c.Param("pair")
	timestampstr := c.Param("timestamp")
	timestamp, _ := strconv.Atoi(timestampstr)
	res, err := services.GetPairInfo(code)
	if err == nil {
		//price:=services.BlockPrice{}.GetPrice()
		supply, _ := strconv.ParseFloat(res.TotalSupply, 64)
		allUsd, _ := strconv.ParseFloat(res.ReserveUSD, 64)
		tPriceView := new(services.PriceView)
		tPriceView.Timestamp = int64(timestamp)
		//tPriceView.PriceUsd = allUsd / sulpply
		//tPriceView.PriceUsd = math.Trunc(tPriceView.PriceUsd*100) / 100
		tPriceView.PriceUsd, _ = decimal.NewFromFloat(allUsd).DivRound(decimal.NewFromFloat(supply), 18).Float64()
		tPriceView.BigPrice = services.GetUnDecimalPrice(tPriceView.PriceUsd).String()
		tPriceView.Sign = services.SignMsg(tPriceView.GetHash())
		c.JSON(200, tPriceView)
		return
	}
	if err != nil {
		ErrJson(c, err.Error())
		return
	}
}
