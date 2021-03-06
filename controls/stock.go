package controls

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"log"
	"math"
	"stock/common"
	"stock/services"
	"stock/utils"
	"strconv"
	"sync"
	"time"
)

type coinvs struct {
	ID     int64
	Coin   float64
	VsCoin float64
}

func (cs coinvs) TableName() string {
	return "coins"
}

// @Tags default
// @Summary　获取币价换算，内部单节点
// @Description 获取币价换算，内部单节点
// @ID CoinPriceHandler
// @Accept  json
// @Produce  json
// @Param     coin   path    string     true        "目标币价" default(eth) Enums(btc,aed,ars,aud,bch,bdt,bhd,bits,bmd,bnb,brl,byn,cad,chf,clp,cny,czk,dkk,dot,eos,eth,eur,gbp,hkd,huf,idr,ils,inr,jpy,krw,kwd,link,lkr,ltc,mmk,mxn,myr,ngn,nok,nzd,php,pkr,pln,rub,sar,sats,sek,sgd,thb,try,twd,uah,usd,vef,vnd,xag,xau,xdr,xlm,xrp,yfi,zar)
// @Param     vs_coin   path    string     true        "vs币价" default(usd) Enums(btc,aed,ars,aud,bch,bdt,bhd,bits,bmd,bnb,brl,byn,cad,chf,clp,cny,czk,dkk,dot,eos,eth,eur,gbp,hkd,huf,idr,ils,inr,jpy,krw,kwd,link,lkr,ltc,mmk,mxn,myr,ngn,nok,nzd,php,pkr,pln,rub,sar,sats,sek,sgd,thb,try,twd,uah,usd,vef,vnd,xag,xau,xdr,xlm,xrp,yfi,zar)
// @Param     timestamp   query    int     false    "unix 秒数" default(1620383144)
// @Success 200 {object} services.CoinPriceView	"Coin Price"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} common.ApiErr "失败时，有相应测试日志输出"
// @Router /pub/internal/coin_price/{coin}/{vs_coin} [get]
func CoinPriceHandler(c *gin.Context) {
	coin := c.Param("coin")
	vsCoin := c.Param("vs_coin")
	timestampstr := c.Query("timestamp")
	timestamp, _ := strconv.Atoi(timestampstr)
	//err := utils.Orm.Where("code= ? and timestamp>= ? ", code,timestamp).Order("timestamp").First(info).Error

	coinField := fmt.Sprintf("cast(%s as decimal(19,6))  as coin", coin)
	if coin == "btc" {
		coinField = "1 as coin"
	}
	vsCoinField := fmt.Sprintf("cast(%s as decimal(19,6))  as vs_coin", vsCoin)
	if vsCoin == "btc" {
		vsCoinField = "1 as vs_coin"
	}
	cs := new(coinvs)
	err := utils.Orm.Order("id desc").Select("id", coinField, vsCoinField).First(cs).Error
	log.Println(cs)

	if err == nil {
		targetPrice := cs.VsCoin / cs.Coin
		log.Println(targetPrice)

		tPriceView := new(services.CoinPriceView)
		tPriceView.Coin = coin
		tPriceView.VsCoin = vsCoin
		tPriceView.Timestamp = int64(timestamp)
		tPriceView.Price = targetPrice
		tPriceView.Price = math.Trunc(tPriceView.Price*100) / 100
		tPriceView.BigPrice = services.GetUnDecimalPrice(float64(tPriceView.Price)).String()
		tPriceView.Sign = services.SignMsg(tPriceView.GetHash())
		c.JSON(200, tPriceView)
		return
	}
	if err != nil {
		common.ErrJson(c, err.Error())
		return
	}
}

// @Tags default
// @Summary　获取币价换算，多节点签名版
// @Description 获取币价换算，多节点签名版
// @ID CoinPriceSignHandler
// @Accept  json
// @Produce  json
// @Param     coin   path    string     true        "目标币价" default(eth) Enums(btc,aed,ars,aud,bch,bdt,bhd,bits,bmd,bnb,brl,byn,cad,chf,clp,cny,czk,dkk,dot,eos,eth,eur,gbp,hkd,huf,idr,ils,inr,jpy,krw,kwd,link,lkr,ltc,mmk,mxn,myr,ngn,nok,nzd,php,pkr,pln,rub,sar,sats,sek,sgd,thb,try,twd,uah,usd,vef,vnd,xag,xau,xdr,xlm,xrp,yfi,zar)
// @Param     vs_coin   path    string     true        "vs币价" default(usd) Enums(btc,aed,ars,aud,bch,bdt,bhd,bits,bmd,bnb,brl,byn,cad,chf,clp,cny,czk,dkk,dot,eos,eth,eur,gbp,hkd,huf,idr,ils,inr,jpy,krw,kwd,link,lkr,ltc,mmk,mxn,myr,ngn,nok,nzd,php,pkr,pln,rub,sar,sats,sek,sgd,thb,try,twd,uah,usd,vef,vnd,xag,xau,xdr,xlm,xrp,yfi,zar)
// @Param     timestamp   path    int     false    "unix 秒数" default(1620383144)
// @Success 200 {object} services.CoinPriceView	"CoinPriceView"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} common.ApiErr "失败时，有相应测试日志输出"
// @Router /pub/coin_price/{coin}/{vs_coin}/{timestamp} [get]
func CoinPriceSignHandler(c *gin.Context) {
	coin := c.Param("coin")
	vsCoin := c.Param("vs_coin")
	timestampstr := c.Param("timestamp")
	timestamp, _ := strconv.Atoi(timestampstr)

	resTokenView := new(services.DataCoinPriceView)
	resTokenView.Coin = coin
	resTokenView.VsCoin = vsCoin
	//var addres []*services.PriceView
	sc := sync.RWMutex{}
	wg := new(sync.WaitGroup)
	var porcNode = func(nodeUrl string) {
		defer wg.Done()
		reqUrl := fmt.Sprintf(nodeUrl+"/pub/internal/coin_price/%s/%s?timestamp=%d", coin, vsCoin, timestamp)
		bs, err := utils.ReqResBody(reqUrl, "", "GET", nil, nil)
		if err == nil {
			token := new(services.CoinPriceView)
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
		sumPrice += node.Price
	}
	resTokenView.Timestamp = int64(timestamp)
	resTokenView.Price = sumPrice / float64(len(resTokenView.Signs))
	resTokenView.BigPrice = services.GetUnDecimalPrice(float64(resTokenView.Price)).String()
	resTokenView.Sign = services.SignMsg(resTokenView.GetHash())
	c.JSON(200, resTokenView)
	return
END:
	if err == nil {
		common.ErrJson(c, err.Error())
	}
}

// @Tags default
// @Summary　获取美股价格:
// @Description 获取美股价格 苹果代码  AAPL  ,苹果代码 TSLA
// @ID StockInfoHandler
// @Accept  json
// @Produce  json
// @Param     code   path    string     true        "美股代码" default(AAPL) Enums(AAPL,TSLA,USD)
// @Param     data_type   path    int     true   "最高最低价１最高　２最低价" default(1) Enums(1,2)
// @Param     timestamp   path    int     false    "unix 秒数" default(1620383144)
// @Success 200 {object} services.StockNode	"stock info"
// @Header 200 {string} sign "签名信息"
// @Failure 500 {object} common.ApiErr "失败时，有相应测试日志输出"
// @Router /pub/stock/info/{code}/{data_type}/{timestamp}  [get]
func StockInfoHandler(c *gin.Context) {
	//info := &services.ViewStock{}
	code := c.Param("code")
	timestampstr := c.Param("timestamp")
	timestamp, _ := strconv.Atoi(timestampstr)
	dataTypeStr := c.Param("data_type")
	dataType, _ := strconv.Atoi(dataTypeStr)

	////err := utils.Orm.Where("code= ? and timestamp>= ? ", code,timestamp).Order("timestamp").First(info).Error
	//err := utils.Orm.Where("code= ? and timestamp<= ? ", code,timestamp).Order("timestamp desc").First(info).Error
	//if err == nil {
	//	//var avgPrice float64
	//	//err = utils.Orm.Model(services.ViewStock{}).Select("avg(price)").Order("timestamp desc").Limit(2500).Where("code= ? and timestamp<= ? ", code,timestamp).Scan(&avgPrice).Error
	//	//avgPrice,err:=getAvgPrice(code,timestamp)
	//	if err == nil {
	//	}
	//}
	var err error
	avgPrice := 0.0
	if code == "USD" {
		avgPrice = 1.0
	} else {
		//avgPrice, err = services.GetMsStatData(code, dataType)
		hl, err1 := getTwHLFromCache(code)
		err = err1
		if err == nil {
			if dataType == 1 {
				avgPrice = hl.High
			} else {
				avgPrice = hl.Low
			}
		}
		log.Println("stockPrice", hl, avgPrice)
	}
	if err == nil {
		//info.Price=avgPrice
		snode := new(services.StockNodeRaw)
		snode.StockCode = code
		snode.DataType = dataType
		snode.NodeAddress = services.WalletAddre
		snode.Price = (math.Trunc(float64(avgPrice)*1000) / 1000)
		snode.BigPrice = services.GetUnDecimalPrice(float64(snode.Price)).String()
		snode.Timestamp = int64(timestamp)
		snode.SetCode()
		snode.SetSign()
		//if services.IsSignTime(0) {
		//	snode.SetSign()
		//}
		c.JSON(200, snode)
		return
	}

	//bs, _ := json.Marshal(snode)
	////md5str:=crypto.SHA256.New()
	//hashbs := sha256.Sum256(bs)
	////log.Println(hashbs, len(hashbs))
	//sign, signErr := Privkey.Sign(rand.Reader, hashbs[0:32], crypto.SHA256)
	//if signErr == nil {
	//	signStr := base64.StdEncoding.EncodeToString(sign)
	//	//c.Header("sign", signStr)
	//	//log.Println(signStr)
	//	snode.Sign=[]byte(signStr)
	//	c.JSON(200, snode)
	//	return
	//} else {
	//	log.Println(signErr)
	//}

	if err != nil {
		common.ErrJson(c, err.Error())
		return
	}
}
func getTwHLFromCache(code string) (*HLValuePair, error) {
	ttl := services.GetUsdStockCacheTime()
	//var addres []*services.PriceView
	proc := func() (interface{}, error) {
		hl := HLValuePair{}
		hight, low, err := services.GetTwHL(code)
		if err == nil {
			hl.High = hight
			hl.Low = low
		}
		return &hl, err
	}
	ckey := fmt.Sprintf("getTwHLFromCache-%s-%d", code, time.Now().Unix()+ttl)
	log.Println(ckey, ttl)
	res, err := utils.CacheFromLruWithFixKey(ckey, proc)
	return res.(*HLValuePair), err
}

type resMarketStatus struct {
	IsOpening bool
	OpenTime  int64
}

// @Tags default
// @Summary　获取美股市场开盘状态:
// @Description 获取美股市场开盘状态,支持节假日,夏令时
// @ID UsaMarketStatusHandler
// @Accept  json
// @Produce  json
// @Param     timestamp   path    int     false    "unix 秒数； 0表示当前时间" default(0)
// @Success 200 {object} resMarketStatus	"status"
// @Failure 500 {object} common.ApiErr "失败时，有相应测试日志输出"
// @Router /pub/stock/market_status/{timestamp} [get]
func UsaMarketStatusHandler(c *gin.Context) {
	timestampstr := c.Param("timestamp")
	timestamp, _ := strconv.Atoi(timestampstr)
	if timestamp == 0 {
		timestamp = int(time.Now().Unix())
	}
	resMarketStatus := new(resMarketStatus)

	resMarketStatus.IsOpening, resMarketStatus.OpenTime = services.IsWorkTime(int64(timestamp))
	c.JSON(200, resMarketStatus)
	return
}
func init() {
	services.InitCalendar()
}

func getAvgPrice(code string, timestamp int) (avgPrice float64, err error) {
	err = utils.Orm.Model(services.ViewStock{}).Select("avg(price)").Order("timestamp desc").Limit(2500).Where("code= ? and timestamp<= ? ", code, timestamp).Scan(&avgPrice).Error
	if err == nil {
		avgPrice = (math.Trunc(avgPrice*1000) / 1000)
	}
	return
}

// @Tags default
// @Summary　获取共识美股价格:
// @Description 获取共识美股价格 苹果代码  AAPL  ,苹果代码 TSLA
// @ID StockAggreHandler
// @Accept  json
// @Produce  json
// @Param     code   path    string     true        "美股代码" default(AAPL)  Enums(USD,AAPL,TSLA)
// @Param     data_type   path    int     true   "最高最低价１最高　２最低价" default(1) Enums(1,2)
// @Param     timestamp   path    int     false    "unix 秒数" default(1620383144)
// @Success 200 {object} services.StockData	"stock info list"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} common.ApiErr "失败时，有相应测试日志输出"
// @Router /pub/stock/aggre_info/{code}/{data_type}/{timestamp} [get]
func StockAggreHandler(c *gin.Context) {
	code := c.Param("code")
	timestampstr := c.Param("timestamp")
	timestamp, _ := strconv.Atoi(timestampstr)
	dataTypeStr := c.Param("data_type")
	dataType, _ := strconv.Atoi(dataTypeStr)

	//节点数据
	nodesPirce := []*services.StockNodeRaw{}
	//节点间平均值数据
	avgNodesPrice := []services.StockNode{}
	sdata := new(services.StockData)
	var err error
	sc := sync.RWMutex{}
	wg := new(sync.WaitGroup)
	var porcNode = func(nodeUrl string) {
		defer wg.Done()
		reqUrl := fmt.Sprintf(nodeUrl+"/pub/stock/info/%s/%d/%d", code, dataType, timestamp)
		bs, err := utils.ReqResBody(reqUrl, "", "GET", nil, nil)
		if err == nil {
			snode := new(services.StockNodeRaw)
			json.Unmarshal(bs, snode)
			snode.Node = nodeUrl
			sc.Lock()
			nodesPirce = append(nodesPirce, snode)
			sc.Unlock()
		}
	}
	for _, nurl := range utils.Nodes {
		wg.Add(1)
		go porcNode(nurl)
	}
	wg.Wait()
	if len(nodesPirce) == 0 {
		err = errors.New("数据不可用")
		goto END
	}
	if len(nodesPirce) < len(utils.Nodes)/2+1 {
		err = errors.New("节点不够用")
		goto END
	}

	//var err error
	//sc:=sync.RWMutex{}
	//wg:= new(sync.WaitGroup)
	porcNode = func(nodeUrl string) {
		defer wg.Done()
		reqUrl := fmt.Sprintf(nodeUrl + "/pub/internal/stock_avgprice")
		bodyBs, _ := json.Marshal(nodesPirce)
		bs, err := utils.ReqResBody(reqUrl, "", "POST", nil, bodyBs)
		if err == nil {
			snode := new(services.StockNode)
			json.Unmarshal(bs, snode)
			snode.Node = nodeUrl

			isMyData, _ := services.Verify(snode.GetHash(), snode.Sign, services.WalletAddre)
			if isMyData {
				//log.Println(myData,"124")
				sdata.Price = snode.Price
				sdata.BigPrice = snode.BigPrice
				sdata.Timestamp = snode.Timestamp
				sdata.StockCode = snode.StockCode
				sdata.Code = snode.Code
				sdata.DataType = dataType
				sdata.Sign = snode.Sign
			}
			sc.Lock()
			avgNodesPrice = append(avgNodesPrice, *snode)
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

	sdata.Signs = nodesPirce
	sdata.AvgSigns = avgNodesPrice
	sdata.Clean()
	//收盘没有签名时，选择第一个价格，方便应用显示价格
	if sdata.AvgSigns[0].Sign == nil {
		snode := sdata.AvgSigns[0]
		sdata.Price = snode.Price
		sdata.BigPrice = snode.BigPrice
		sdata.Timestamp = snode.Timestamp
		sdata.StockCode = snode.StockCode
		sdata.Code = snode.Code
		sdata.DataType = dataType
	}

	//if len(sdata.Signs)==0{
	//	err=errors.New("数据不可用")
	//	goto END
	//}
	//if len(sdata.Signs)<len(utils.Nodes)/2+1{
	//	err=errors.New("节点不够用")
	//	goto END
	//}
	//
	//for _, node := range nodesPirce {
	//	sumPrice+=node.Price
	//}
	//sdata.Price=sumPrice/float64(len(nodesPirce))
	//
	//sdata.Price= (math.Trunc(float64( sdata.Price)*1000)/1000)
	//sdata.BigPrice =services.GetUnDecimalUsdPrice(float64(sdata.Price),3).String()
	//sdata.Timestamp=int64(timestamp)
	//sdata.StockCode=code
	//sdata.DataType = dataType
	//sdata.SetSign()
	// sdata.IsMarketOpening = services.UsdStockTime()
	sdata.IsMarketOpening, sdata.MarketOpenTime = services.IsWorkTime(int64(timestamp))
	if len(avgNodesPrice) == 0 || avgNodesPrice[0].Sign == nil {
		sdata.IsMarketOpening = false
	} else {
		sdata.IsMarketOpening = true
	}

	c.JSON(200, sdata)
	return

END:
	if err != nil {
		common.ErrJson(c, err.Error())
		return
	}
}

// @Tags default
// @Summary　获取股票平均价格共识:
// @Description 获取股票平均价格共识 苹果代码  AAPL  ,苹果代码 TSLA
// @ID StockAvgPriceHandler
// @Accept  json
// @Produce  json
// @Param   nodePrices  body   []services.StockNode true       "节点价格列表"
// @Success 200 {object} services.StockNode	"stock info"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} common.ApiErr "失败时，有相应测试日志输出"
// @Router /pub/internal/stock_avgprice [post]
func StockAvgPriceHandler(c *gin.Context) {
	nodePrices := []*services.StockNodeRaw{}
	err := c.BindJSON(&nodePrices)
	if err == nil {
		if len(nodePrices) == 0 {
			err = errors.New("数据不可用")
			common.ErrJson(c, err.Error())
			return
		}
		if len(nodePrices) < len(utils.Nodes)/2+1 {
			err = errors.New("节点不够用")
			common.ErrJson(c, err.Error())
			return
		}

		timestamp := nodePrices[0].Timestamp
		code := nodePrices[0].Code
		stockCode := nodePrices[0].StockCode
		dataType := nodePrices[0].DataType

		sumPrice := decimal.NewFromFloat(0.0)
		for _, node := range nodePrices {

			//验证数据
			if timestamp != node.Timestamp || code != node.Code || dataType != node.DataType {
				err = errors.New("需要共识的数据不一致")
				break
			}

			//验证数据
			if timestamp != node.Timestamp || code != node.Code || dataType != node.DataType {
				err = errors.New("需要共识的数据不一致")
				break
			}
			// 验证数据签名
			if node.Sign == nil {
				err = errors.New("miss node.sign")
				break
			}
			//白名单验证
			if !utils.IsInWL(node.NodeAddress) {
				err = errors.New("addre not int whiteList")
				break
			}

			ok, err1 := services.Verify(node.GetHash(), node.Sign, node.NodeAddress)
			if !ok {
				log.Println("Verify err", node.NodeAddress, err1)
				err = errors.New("Verify err " + node.NodeAddress + " " + err1.Error())
				break
			}
			sumPrice = sumPrice.Add(decimal.NewFromFloat(node.Price))

		}
		if err != nil {
			log.Println( "StockAvgPriceHandler check",err)
			common.ErrJson(c, err.Error())
			return
		}
		sdata := new(services.StockNode)
		sdata.Price, _ = sumPrice.DivRound(decimal.NewFromInt(int64(len(nodePrices))), 18).Float64()
		//sdata.Price = (math.Trunc(float64(sdata.Price)*1000) / 1000)
		sdata.BigPrice = services.GetUnDecimalPrice(sdata.Price).String()
		sdata.Timestamp = int64(timestamp)
		sdata.StockCode = stockCode
		sdata.SetCode()
		sdata.DataType = dataType
		sdata.NodeAddress = services.WalletAddre

		signAble,msg:=IsSignAble(sdata.Code, sdata.Price)
		sdata.Msg=msg
		if signAble && services.IsSignTime(0) {
			sdata.SetSign()
		}else{
			sdata.Msg="not in market time"
		}
		c.JSON(200, sdata)
		return
	}
	if err != nil {
		common.ErrJson(c, err.Error())
		return
	}
}
