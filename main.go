package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	cors "github.com/itsjamie/gin-cors"
	"github.com/spf13/pflag"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"log"
	"os"
	"stock/controls"
	_ "stock/docs"
	"stock/services"
	"stock/utils"
	"strings"
	"sync"
	"time"
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
	log.SetFlags(log.LstdFlags)
	log.SetOutput(os.Stdout)

	var dbUrl, serverPort, env, infura, swapGraphApi string
	var job bool
	var nodes ,wlist []string
	pflag.StringVarP(&dbUrl, "db", "d", "root:password@tcp(localhost:3306)/mydb?loc=Local&parseTime=true&multiStatements=true", "mysql database url")
	pflag.StringVarP(&serverPort, "port", "p", "8001", "api　service port")
	//var keyfile,certFile string
	//pflag.StringVarP(&keyfile,"key","k","./asset/key.pem","pem encoded private key")
	//pflag.StringVarP(&certFile,"cert","c","./asset/cert.pem","pem encoded x509 cert")
	pflag.StringSliceVarP(&nodes, "nodes", "n", strings.Split("http://49.232.234.250:8001,http://localhost:8001,http://62.234.188.160:8001", ","), "所有节点列表,节点间用逗号分开")
	pflag.StringSliceVarP(&wlist, "wlist", "", strings.Split("0x4448993f493B1D8D9ED51F22F1d30b9B4377dFD2,0x0d93A21b4A971dF713CfC057e43F5D230E76261C,0x3054e19707447800f0666ba274A249fC9a67aA4a,0xa55203c75c95A95f5DdD7B58E877A9EBd85A1631", ","), "所有钱包地址白名单列表,节点间用逗号分开")
	pflag.StringVarP(&env, "env", "e", "debug", "环境名字debug prod test")
	pflag.StringVar(&infura, "infura", "27f0b03a4654478db14295fd1021e1b8", "infura的项目id,需要自行去https://infura.io申请")
	//https://api.thegraph.com/subgraphs/name/wxf4150/fanswap2 https://api.thegraph.com/subgraphs/name/ianlapham/uniswapv2
	pflag.StringVar(&swapGraphApi, "swapGraphApi", "https://api.thegraph.com/subgraphs/name/ianlapham/uniswapv2", "swap theGraphApi")
	pflag.BoolVarP(&job, "job", "j", true, "是否抓取数据")

	//pflag.BoolVarP(&job, "job", "j", false, "是否抓取数据")
	//nodes =  strings.Split("http://localhost:8001,http://localhost:8001,http://localhost:8001",",")
	pflag.Parse()
	utils.Nodes = nodes
	utils.SetWList(wlist)
	//utils.TwKey=twkey
	utils.InitDb(dbUrl)
	services.InitEConn(infura)
	services.SwapGraphApi = swapGraphApi

	//获取twelvedata最新数据
	//services.GetTwData("2021-07-26","",1000)
	//return
	//utils.Orm.AutoMigrate(services.MarketPrice{})
	////services.SubTwData()
	////services.GetTwData("","",3)
	//endTime:=time.Now().Truncate(time.Hour*24)
	//stime:=time.Date(2021,6,1,0,0,0,0,time.UTC)
	////i:=0
	//for stime.Before(endTime){
	//	nextDay:=stime.Add(5*time.Hour*24)
	//	services.GetTwData(stime.Format("2006-01-02"),nextDay.Format("2006-01-02"),5000)
	//	stime=nextDay
	//	//i++
	//	//if i>3{break}
	//	//time.Sleep(time.Second*63)
	//}
	//return
	//services.SetAllBullsFromTw(false)
	//return
	if job {
		//sync coingecko数据
		//go services.SyncCoinGeckoData()

		//go services.GetStocks()
		go services.SubEthPrice(0)
		go services.SubCoinsPrice()
		//coingecko bull
		//go services.SetAllBulls("btc3x")
		//go services.SetAllBulls("eth3x")
		go func() {
			//监听eth uniswap pair's token价格
			//tpc := services.TokenPairConf{PairAddre: "0x4d3c5db2c68f6859e0cd05d080979f597dd64bff", TokenAddre: "0x72e364f2abdc788b7e918bc238b21f109cd634d7", TokenDecimals: 18, ChainName: "eth"}
			tpc := services.TokenPairConf{PairAddre: "0xdfb8824b094f56b9216a015ff77bdb056923aaf6", TokenAddre: "0x011864d37035439e078d64630777ec518138af05", TokenDecimals: 18, ChainName: "eth"}
			services.SubPairlog(&tpc)
		}()

		////subcribe twelvedata data
		//go services.SubTwData()
		//
		////更新twelvedata数据源bull数据
		//go services.SetAllBullsFromTw(true)

		services.SetAllBullsFromTw(false)
		services.CronTwData()
		services.Cn.Start()

		//订阅coinmarketcap数据 Metaverse Index
		go services.SubCM()

		// token totalSupply daily data
		go services.TokenTotalSupplyDailyData()

		//股票时间间隔价格统计
		go services.SetStockStat()

	}

	services.InitNodeKey()
	//InitKey(keyfile,certFile)
	if env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	ReqHeader := []string{
		"Content-Type", "Origin", "Authorization", "Accept", "tokenId", "tokenid", "authorization", "ukey", "token", "cache-control", "x-requested-with"}
	router := gin.Default()
	router.Use(cors.Middleware(cors.Config{
		Origins:         "*",
		Methods:         "GET, PUT, POST, DELETE",
		RequestHeaders:  strings.Join(ReqHeader, ", "),
		ExposedHeaders:  "",
		MaxAge:          360000 * time.Second,
		Credentials:     true,
		ValidateHeaders: false,
	}))
	//router.Use(controls.TokenCheck())
	//domainDir:=router.Group("/nft")
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	api := router.Group("/pub")
	api.Use(controls.MiddleWareStat())
	go controls.SaveStat()
	controls.InitConfig()

	api.GET("/system/config", controls.ConfigHandler)
	api.GET("/stock/info/:code/:data_type/:timestamp", controls.StockInfoHandler)
	api.GET("/stock/market_status/:timestamp", controls.UsaMarketStatusHandler)
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
	api.GET("/internal/dex/pair/token_price/:pair/:token/:timestamp", controls.PairTokenPriceHandler)

	api.GET("/internal/dex/token_chain_price/:token/:timestamp", controls.TokenChainPriceHandler)

	api.GET("/internal/dex/token_info/:token/:timestamp", controls.TokenInfoHandler)
	api.GET("/internal/dex/pair/token_info/:pair/:token/:timestamp", controls.PairTokenInfoHandler)

	api.GET("/internal/coin_price/:coin/:vs_coin", controls.CoinPriceHandler)
	api.GET("/internal/dex/ftx_price/:coin_type/:timestamp", controls.FtxPriceHandler)
	api.GET("/coin_price/:coin/:vs_coin/:timestamp", controls.CoinPriceSignHandler)

	api.GET("/dex/token_price/:token/:data_type/:timestamp", controls.TokenPriceSignHandler)
	api.GET("/dex/pair/token_price/:pair/:token/:data_type/:timestamp", controls.PairTokenPriceSignHandler)

	api.GET("/dex/token_chain_price/:token/:data_type/:timestamp", controls.TokenChainPriceSignHandler)
	api.GET("/dex/ftx_price/:coin_type/:data_type/:timestamp", controls.FtxPriceSignHandler)
	api.GET("/dex/token_day_datas/:token/:days/:timestamp", controls.TokenDayDatasHandler)
	api.GET("/dex/ftx_chart_prices/:coin_type/:count/:interval/:timestamp", controls.FtxChartPricesHandler)
	api.GET("/dex/stock_chart_prices/:coin_type/:count/:interval/:timestamp", controls.StockChartPricesHandler)

	api.GET("/dex/token_chart_prices/:token/:count/:interval/:timestamp", controls.TokenDayPricesHandler)
	api.GET("/dex/pair/token_chart_prices/:pair/:token/:count/:interval/:timestamp", controls.PairTokenDayPricesHandler)

	api.GET("/dex/token/token_chart_supply/:token/:amount/:timestamp", controls.TokenChartSupplyHandler)
	api.GET("/dex/token/token_total_supply/:token/:timestamp", controls.GetTokenSupplyHandler)
	api.GET("/alert/ok", controls.OkHandler)
	api.GET("/alert/coindata", controls.CoinDataCheckHandler)
	api.GET("/alert/btc_sign_check", controls.BtcSignCheckHandler)
	api.GET("/db.sql.gz", controls.DbExportHandler)


	sys := router.Group("/sys")
	controls.InitJwt(sys)
	api.POST("/login", controls.AuthMiddleware.LoginHandler)
	api.GET("/pre_login", controls.ChallengeHandler)
	sys.GET("/hello", controls.HelloJwtHandler)
	//api.POST("/stock/sign_verify", VerifyInfoHandler)

	router.NoRoute(func(c *gin.Context) {
		controls.ErrJson(c, "none api router")
		//c.JSON(404,controls.ApiErr{Error:"none api router"})
	})
	go router.RunTLS(":8002", "./asset/tls.pem", "./asset/tls.pem")
	log.Fatal(router.Run(":" + serverPort))
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
	stat := NodeStat{}
	utils.Orm.Model(services.ViewStock{}).Count(&(stat.StockRows))
	utils.Orm.Model(services.ViewStock{}).Select("max(updated_at) ").Scan(&stat.StockUpdateAt)

	coinMaxid := 0
	utils.Orm.Model(services.Coin{}).Select("max(id)").Scan(&coinMaxid)
	stat.CionPricesUpdateAt = time.Unix(int64(coinMaxid), 0)

	utils.Orm.Model(services.BlockPrice{}).Select("max(created_at)").Scan(&stat.BlockPricesUpdateAt)

	stat.WalletAddre = services.WalletAddre
	stat.Uptime = services.Uptime
	c.JSON(200, stat)
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
	Uptime              time.Time
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
	sc := sync.RWMutex{}
	wg := new(sync.WaitGroup)
	var porcNode = func(nodeUrl string) {
		defer wg.Done()
		reqUrl := nodeUrl + "/pub/stock/stat"
		bs, err := utils.ReqResBody(reqUrl, "", "GET", nil, nil)
		if err == nil {
			stat := new(NodeStat)
			err = json.Unmarshal(bs, stat)
			if err == nil {
				log.Println(err)
			}
			stat.Node = nodeUrl
			sc.Lock()
			addres = append(addres, stat)
			sc.Unlock()
		}
	}
	for _, nurl := range utils.Nodes {
		wg.Add(1)
		go porcNode(nurl)
	}
	wg.Wait()
	c.JSON(200, addres)
	return

	if err != nil {
		controls.ErrJson(c, err.Error())
		return
	}
}

type VerObj struct {
	//stockInfo json: {"code":"AAPL","price":128.1,"name":"苹果","timestamp":1620292445,"UpdatedAt":"2021-05-06T17:14:05.878+08:00"}
	Data json.RawMessage `swaggertype:"object"`
	Sign []byte          `swaggertype:"string" format:"base64" example:"UhRVNsT8B5Za6oO3APH0T9ebPMKHxDDhkscYuILl7lDepDMzyBaQsEu9vwTRIfoYBS8udfEanI/DUAhwnIdFJf9woIv7Oo+OS6q3sF3B5Vx9NN2ipXJ4wjTf2ct7FbS1vXAvTXSmA2svj+LF8P1PIEClITBqu/EWZXTpHvAlbGAAeF+hHO7/FquLHVDavLC+OENyb0CP+NvH+ytZ69tav0DqbGp+NGGil/ImZpPsetbOxwuhC/U1CV6Ap8qgRWe8s6IpOawXDAavLMHUmXVvORDf/XVzaQUJ5ob+vTsSTZwQsvj/4jmsODFt8eKFYL/7vyN/i3HkiDwhq0w85kqHgg=="`
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
