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

type coinvs struct{
	ID int64
	Coin float64
	VsCoin float64
}
func (cs coinvs) TableName()string{
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
// @Success 200 {object} services.CoinPriceView	"stock info"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} controls.ApiErr "失败时，有相应测试日志输出"
// @Router /pub/internal/coin_price/{coin}/{vs_coin} [get]
func CoinPriceHandler(c *gin.Context) {
	coin:=c.Param("coin")
	vsCoin:=c.Param("vs_coin")
	timestampstr:=c.Query("timestamp")
	timestamp,_:=strconv.Atoi(timestampstr)
	//err := utils.Orm.Where("code= ? and timestamp>= ? ", code,timestamp).Order("timestamp").First(info).Error

	coinField:=fmt.Sprintf("cast(%s as decimal(19,6))  as coin",coin)
	if coin=="btc"{
		coinField="1 as coin"
	}
	vsCoinField:=fmt.Sprintf("cast(%s as decimal(19,6))  as vs_coin",vsCoin)
	if vsCoin=="btc"{
		vsCoinField="1 as vs_coin"
	}
	cs:=new(coinvs)
	err:=utils.Orm.Order("id desc").Select("id",coinField,vsCoinField).Find(cs).Error
	log.Println(cs)

	if err == nil {
		targetPrice:=cs.VsCoin/cs.Coin
		log.Println(targetPrice)

		tPriceView:=new(services.CoinPriceView)
		tPriceView.Coin=coin
		tPriceView.VsCoin=vsCoin
		tPriceView.Timestamp=int64(timestamp)
		tPriceView.Price=targetPrice
		tPriceView.Price= math.Trunc(tPriceView.Price*100)/100
		tPriceView.BigPrice=services.GetUnDecimalPrice(float64(tPriceView.Price)).String()
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
// @Summary　获取币价换算，多节点签名版
// @Description 获取币价换算，多节点签名版
// @ID CoinPriceSignHandler
// @Accept  json
// @Produce  json
// @Param     coin   path    string     true        "目标币价" default(eth) Enums(btc,aed,ars,aud,bch,bdt,bhd,bits,bmd,bnb,brl,byn,cad,chf,clp,cny,czk,dkk,dot,eos,eth,eur,gbp,hkd,huf,idr,ils,inr,jpy,krw,kwd,link,lkr,ltc,mmk,mxn,myr,ngn,nok,nzd,php,pkr,pln,rub,sar,sats,sek,sgd,thb,try,twd,uah,usd,vef,vnd,xag,xau,xdr,xlm,xrp,yfi,zar)
// @Param     vs_coin   path    string     true        "vs币价" default(usd) Enums(btc,aed,ars,aud,bch,bdt,bhd,bits,bmd,bnb,brl,byn,cad,chf,clp,cny,czk,dkk,dot,eos,eth,eur,gbp,hkd,huf,idr,ils,inr,jpy,krw,kwd,link,lkr,ltc,mmk,mxn,myr,ngn,nok,nzd,php,pkr,pln,rub,sar,sats,sek,sgd,thb,try,twd,uah,usd,vef,vnd,xag,xau,xdr,xlm,xrp,yfi,zar)
// @Param     timestamp   query    int     false    "unix 秒数" default(1620383144)
// @Success 200 {object} services.CoinPriceView	"stock info"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} controls.ApiErr "失败时，有相应测试日志输出"
// @Router /pub/coin_price/{coin}/{vs_coin} [get]
func CoinPriceSignHandler(c *gin.Context) {
	coin:=c.Param("coin")
	vsCoin:=c.Param("vs_coin")
	timestampstr:=c.Query("timestamp")
	timestamp,_:=strconv.Atoi(timestampstr)

	resTokenView := new(services.DataCoinPriceView)
	resTokenView.Coin=coin
	resTokenView.VsCoin=vsCoin
	//var addres []*services.PriceView
	sc := sync.RWMutex{}
	wg := new(sync.WaitGroup)
	var porcNode = func(nodeUrl string) {
		defer wg.Done()
		reqUrl := fmt.Sprintf(nodeUrl+"/pub/internal/coin_price/%s/%s?timestamp=%d", coin,vsCoin,timestamp)
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
	resTokenView.Timestamp=int64(timestamp)
	resTokenView.Price = sumPrice / float64(len(resTokenView.Signs))
	resTokenView.BigPrice=services.GetUnDecimalPrice(float64(resTokenView.Price)).String()
	resTokenView.Sign=services.SignMsg(resTokenView.GetHash())
	c.JSON(200, resTokenView)
	return
END:
	if err == nil {
		ErrJson(c, err.Error())
	}
}

// @Tags default
// @Summary　获取美股价格:
// @Description 获取美股价格 苹果代码  AAPL  ,苹果代码 TSLA
// @ID StockInfoHandler
// @Accept  json
// @Produce  json
// @Param     code   query    string     true        "美股代码" default(AAPL)
// @Param     timestamp   query    int     false    "unix 秒数" default(1620383144)
// @Success 200 {object} services.StockNode	"stock info"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} controls.ApiErr "失败时，有相应测试日志输出"
// @Router /pub/stock/info [get]
func StockInfoHandler(c *gin.Context) {
	info := &services.ViewStock{}
	code:=c.Query("code")
	timestampstr:=c.Query("timestamp")
	timestamp,_:=strconv.Atoi(timestampstr)
	//err := utils.Orm.Where("code= ? and timestamp>= ? ", code,timestamp).Order("timestamp").First(info).Error
	err := utils.Orm.Where("code= ? and timestamp<= ? ", code,timestamp).Order("timestamp desc").First(info).Error
	if err == nil {
		var avgPrice float32
		//err = utils.Orm.Model(services.ViewStock{}).Select("avg(price)").Order("timestamp desc").Limit(2500).Where("code= ? and timestamp<= ? ", code,timestamp).Scan(&avgPrice).Error
		avgPrice,err:=getAvgPrice(code,timestamp)
		if err == nil {
			log.Println("avgPrice",avgPrice)
			info.Price=avgPrice

			snode:=new(services.StockNode)
			snode.Code=info.Code
			snode.Price=info.Price
			snode.BigPrice =services.GetUnDecimalPrice(float64(info.Price)).String()
			snode.Timestamp=info.Timestamp
			snode.SetSign()
			c.JSON(200, snode)
			return

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
		}
	}
	if err != nil {
		ErrJson(c,err.Error())
		return
	}
}
func getAvgPrice(code string,timestamp int)(avgPrice float32,err error) {
	err = utils.Orm.Model(services.ViewStock{}).Select("avg(price)").Order("timestamp desc").Limit(2500).Where("code= ? and timestamp<= ? ", code, timestamp).Scan(&avgPrice).Error
	return
}


// @Tags default
// @Summary　获取共识美股价格:
// @Description 获取共识美股价格 苹果代码  AAPL  ,苹果代码 TSLA
// @ID StockAggreHandler
// @Accept  json
// @Produce  json
// @Param     code   path    string     true        "美股代码" default(AAPL)
// @Param     timestamp   path    int     false    "unix 秒数" default(1620383144)
// @Success 200 {object} services.StockData	"stock info"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} controls.ApiErr "失败时，有相应测试日志输出"
// @Router /pub/stock/aggre_info/{code}/{timestamp} [get]
func StockAggreHandler(c *gin.Context) {
	code:=c.Param("code")
	timestampstr:=c.Param("timestamp")
	timestamp,_:=strconv.Atoi(timestampstr)
	sdata:=new(services.StockData)
	snodes:=[]services.StockNode{}
	var err error

	sc:=sync.RWMutex{}
	wg:= new(sync.WaitGroup)
	var porcNode=func(nodeUrl string) {
		defer wg.Done()
		reqUrl := fmt.Sprintf(nodeUrl+"/pub/stock/info?timestamp=%d&code=%s", timestamp, code)
		bs, err := utils.ReqResBody(reqUrl, "", "GET", nil, nil)
		if err == nil {
			snode := new(services.StockNode)
			json.Unmarshal(bs, snode)
			snode.Node=nodeUrl
			sc.Lock()
			snodes = append(snodes, *snode)
			sc.Unlock()
		}
	}
	for _, nurl := range utils.Nodes {
		wg.Add(1)
		go porcNode(nurl)
	}
	wg.Wait()

	sumPrice:=float32(0.0)
	sdata.Signs=snodes
	if len(sdata.Signs)==0{
		err=errors.New("数据不可用")
		goto END
	}
	if len(sdata.Signs)<len(utils.Nodes)/2+1{
		err=errors.New("节点不够用")
		goto END
	}

	for _, node := range snodes {
		sumPrice+=node.Price
	}
	sdata.Price=sumPrice/float32( len(snodes))
	sdata.BigPrice =services.GetUnDecimalPrice(float64(sdata.Price)).String()
	sdata.Timestamp=int64(timestamp)
	sdata.Code=code
	sdata.SetSign()
	sdata.IsMarketOpening =services.UsdStockTime()
	c.JSON(200,sdata)
	return

END:
	if err != nil {
		ErrJson(c,err.Error())
		return
	}
}

