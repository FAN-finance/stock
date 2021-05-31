package controls
import (
	"errors"
	"github.com/gin-gonic/gin"
	"stock/utils"
	"log"
	"fmt"
	"sync"
	"encoding/json"
	"stock/services"
	"strconv"
)

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
			snode.BigPrice =services.GetUnDecimalPrice(info.Price).String()
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
	sdata.BigPrice =services.GetUnDecimalPrice(sdata.Price).String()
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

