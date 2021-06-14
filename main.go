package main

import (

	"encoding/json"
	"github.com/gin-gonic/gin"
	cors "github.com/itsjamie/gin-cors"
	"github.com/spf13/pflag"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"log"
	"stock/services"
	"strings"
	"sync"
	"time"
	_ "stock/docs"
	"stock/utils"
	"stock/controls"
)

// @title stock-info-api
// @version 1.0
// @description stock-info-api接口文档.
//@termsOfService https://rrl360.com/index.html

// @contact.name 伍晓飞
// @contact.email wuxiaofei@rechaintech.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

//host 192.168.122.1:8080
// @BasePath /
func main() {
	var dbUrl,serverPort,env,infura,swapGraphApi string
	var job bool

	var nodes []string
	pflag.StringVarP(&dbUrl,"db","d","root:password@tcp(localhost:3306)/mydb?loc=Local&parseTime=true&multiStatements=true","mysql database url")
	pflag.StringVarP(&serverPort,"port","p","8001","api　service port")
	//var keyfile,certFile string
	//pflag.StringVarP(&keyfile,"key","k","./asset/key.pem","pem encoded private key")
	//pflag.StringVarP(&certFile,"cert","c","./asset/cert.pem","pem encoded x509 cert")
	pflag.StringSliceVarP(&nodes,"nodes","n",strings.Split("http://localhost:8001,http://localhost:8001",","),"所有节点列表,节点间用逗号分开")
	pflag.StringVarP(&env,"env","e","debug","环境名字debug prod test")
	pflag.StringVar(&infura,"infura","infura_proj_id","infura的项目id,需要自行去https://infura.io申请")
	//https://api.thegraph.com/subgraphs/name/wxf4150/fanswap2 https://api.thegraph.com/subgraphs/name/ianlapham/uniswapv2
	pflag.StringVar(&swapGraphApi,"swapGraphApi","https://api.thegraph.com/subgraphs/name/ianlapham/uniswapv2","swap theGraphApi")
	pflag.BoolVarP(&job,"job","j",true,"是否抓取数据")


	pflag.Parse()
	utils.Nodes=nodes
	utils.InitDb(dbUrl)
	services.InitEConn(infura)
	services.SwapGraphApi=swapGraphApi
	if job {
		//go services.GetStocks()
		go services.SubEthPrice(0)
		go services.SubCoinsPrice()
	}

	services.InitNodeKey()
	//InitKey(keyfile,certFile)
	if env=="prod"{
		gin.SetMode(gin.ReleaseMode)
	}
	log.SetFlags(log.LstdFlags)
	ReqHeader:=[]string{
		"Content-Type","Origin","Authorization", "Accept", "tokenId", "tokenid", "authorization","ukey","token","cache-control", "x-requested-with"}
	router := gin.Default()
	router.Use(cors.Middleware(cors.Config{
		Origins:        "*",
		Methods:        "GET, PUT, POST, DELETE",
		RequestHeaders: strings.Join(ReqHeader,", "),
		ExposedHeaders: "",
		MaxAge: 360000 * time.Second,
		Credentials: true,
		ValidateHeaders: false,
	}))
	//router.Use(controls.TokenCheck())
	//domainDir:=router.Group("/nft")
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	api:=router.Group("/pub")
	api.GET("/stock/info", controls.StockInfoHandler)
	api.POST("/internal/stock_avgprice", controls.StockAvgPriceHandler)
	api.POST("/internal/token_avgprice", controls.TokenAvgHlPriceHandler)
	//api.GET("/stock/aggre_info", StockAggreHandler)
	api.GET("/stock/aggre_info/:code/:data_type/:timestamp", controls.StockAggreHandler)
	api.GET("/stock/stat", NodeStatHandler)
	api.GET("/stock/stats", NodeStatsHandler)
	api.GET("/stock/any_api", controls.NodeAnyApiHandler)
	api.GET("/stock/any_apis", controls.NodeAnyApisHandler)
	api.GET("/internal/dex/lp_price/:pair/:timestamp", controls.PairLpPriceHandler)
	api.GET("/dex/lp_price/:pair/:timestamp", controls.PairLpPriceSignHandler)
	api.GET("/internal/dex/token_price/:token/:timestamp", controls.TokenPriceHandler)
	api.GET("/internal/dex/token_info/:token/:timestamp", controls.TokenInfoHandler)
	api.GET("/internal/coin_price/:coin/:vs_coin", controls.CoinPriceHandler)
	api.GET("/coin_price/:coin/:vs_coin", controls.CoinPriceSignHandler)
	api.GET("/dex/token_price/:token/:data_type/:timestamp", controls.TokenPriceSignHandler)
	api.GET("/dex/token_day_datas/:token/:days/:timestamp", controls.TokenDayDatasHandler)
	api.GET("/dex/token_chart_prices/:token/:count/:interval/:timestamp", controls.TokenDayPricesHandler)
	//api.POST("/stock/sign_verify", VerifyInfoHandler)

	router.NoRoute(func(c *gin.Context){
		controls.ErrJson(c,"none api router")
		//c.JSON(404,controls.ApiErr{Error:"none api router"})
	})
	go router.RunTLS(":8002","./asset/tls.pem","./asset/tls.pem")
	log.Fatal(router.Run(":"+serverPort))
}

// @Tags default
// @Summary　当前节点状态:记录数,钱包地址
// @Description 当前节点状态:记录数,钱包地址
// @ID NodeStatHandler
// @Accept  json
// @Produce  json
// @Success 200 {object} NodeStat	"node stat"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} controls.ApiErr "失败时，有相应测试日志输出"
// @Router /pub/stock/stat [get]
func NodeStatHandler(c *gin.Context) {
	stat:=NodeStat{}
	utils.Orm.Model(services.ViewStock{}).Count(&(stat.StockRows))
	utils.Orm.Model(services.ViewStock{}).Select("max(updated_at) ").Scan(&stat.StockUpdateAt)

	coinMaxid:=0
	utils.Orm.Model(services.Coin{}).Select("max(id)").Scan(&coinMaxid)
	stat.CionPricesUpdateAt=time.Unix(int64(coinMaxid),0)

	utils.Orm.Model(services.BlockPrice{}).Select("max(created_at)").Scan(&stat.BlockPricesUpdateAt)

	stat.WalletAddre=services.WalletAddre
	c.JSON(200,stat)
}


type NodeStat struct {
	//节点名
	Node string
	//钱包地址
	WalletAddre string
	//股票信息数据库记录数
	StockRows int64
	//股票信息最后更新时间
	StockUpdateAt time.Time
	//币价换算信息最后更新时间
	CionPricesUpdateAt time.Time
	//eth价格信息最后更新时间
	BlockPricesUpdateAt time.Time

}
// @Tags default
// @Summary　所有节点状态:记录数,钱包地址
// @Description 所有节点状态:记录数,钱包地址
// @ID NodeStatsHandler
// @Accept  json
// @Produce  json
// @Success 200 {array} NodeStat	"Node Stat list"
// @Failure 500 {object} controls.ApiErr "失败时，有相应测试日志输出"
// @Router /pub/stock/stats [get]
func NodeStatsHandler(c *gin.Context) {
	var err error

	var addres []*NodeStat
	sc:=sync.RWMutex{}
	wg:= new(sync.WaitGroup)
	var porcNode=func(nodeUrl string) {
		defer wg.Done()
		reqUrl := nodeUrl+"/pub/stock/stat"
		bs, err := utils.ReqResBody(reqUrl, "", "GET", nil, nil)
		if err == nil {
			stat:=new(NodeStat)
			err=json.Unmarshal(bs,stat)
			if err == nil {
				log.Println(err)
			}
			stat.Node=nodeUrl
			sc.Lock()
			addres=append(addres,stat)
			sc.Unlock()
		}
	}
	for _, nurl := range utils.Nodes {
		wg.Add(1)
		go porcNode(nurl)
	}
	wg.Wait()
	c.JSON(200,addres)
	return

	if err != nil {
		controls.ErrJson(c,err.Error())
		return
	}
}




type VerObj struct {
	//stockInfo json: {"code":"AAPL","price":128.1,"name":"苹果","timestamp":1620292445,"UpdatedAt":"2021-05-06T17:14:05.878+08:00"}
	Data json.RawMessage `swaggertype:"object"`
	Sign []byte `swaggertype:"string" format:"base64" example:"UhRVNsT8B5Za6oO3APH0T9ebPMKHxDDhkscYuILl7lDepDMzyBaQsEu9vwTRIfoYBS8udfEanI/DUAhwnIdFJf9woIv7Oo+OS6q3sF3B5Vx9NN2ipXJ4wjTf2ct7FbS1vXAvTXSmA2svj+LF8P1PIEClITBqu/EWZXTpHvAlbGAAeF+hHO7/FquLHVDavLC+OENyb0CP+NvH+ytZ69tav0DqbGp+NGGil/ImZpPsetbOxwuhC/U1CV6Ap8qgRWe8s6IpOawXDAavLMHUmXVvORDf/XVzaQUJ5ob+vTsSTZwQsvj/4jmsODFt8eKFYL/7vyN/i3HkiDwhq0w85kqHgg=="`
}
//// @Tags default
//// @Summary　签名验证:
//// @Description 签名验证
//// @ID VerifyInfoHandler
//// @Accept  json
//// @Produce  json
//// @Param     verObj   body    VerObj     true        "需要验证的对象" default(AAPL)
//// @Success 200 {object} ApiOk	"ok info"
//// @Failure 500 {object} controls.ApiErr "失败时，有相应测试日志输出"
//// @Router /pub/stock/sign_verify [post]
//func VerifyInfoHandler(c *gin.Context) {
//	vobj:=new(VerObj)
//	err:=c.Bind(vobj)
//	if err == nil {
//		hashbs:=sha256.Sum256(vobj.Data)
//		err=rsa.VerifyPKCS1v15(LocalCert.PublicKey.(*rsa.PublicKey),crypto.SHA256,hashbs[0:32],vobj.Sign,)
//		if err == nil {
//			c.JSON(200, ApiOk{"ok"})
//			return
//		}
//	}
//	if err != nil {
//		ErrJson(c,err.Error())
//		return
//	}
//}


