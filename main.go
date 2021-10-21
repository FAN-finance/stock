package main

import (
	"github.com/gin-gonic/gin"
	cors "github.com/itsjamie/gin-cors"
	"github.com/spf13/pflag"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"log"
	"os"
	"stock/common"
	"stock/controls"
	_ "stock/docs"
	"stock/services"
	"stock/services/uni"
	"stock/sys"
	"stock/utils"
	"strings"
	"time"
)

// @title oracle-api
// @version 1.0
// @description oracle-api接口文档.
//@termsOfService https://test.com/index.html

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
	var nodes ,wlist,adminList []string
	pflag.StringVarP(&dbUrl, "db", "d", "root:password@tcp(localhost:3306)/mydb?loc=Local&parseTime=true&multiStatements=true", "mysql database url")
	pflag.StringVarP(&serverPort, "port", "p", "8001", "api　service port")
	//var keyfile,certFile string
	//pflag.StringVarP(&keyfile,"key","k","./asset/key.pem","pem encoded private key")
	//pflag.StringVarP(&certFile,"cert","c","./asset/cert.pem","pem encoded x509 cert")
	pflag.StringSliceVarP(&nodes, "nodes", "n", strings.Split("http://49.232.234.250:8001,http://localhost:8001,http://62.234.188.160:8001", ","), "所有节点列表,节点间用逗号分开")
	pflag.StringSliceVarP(&wlist, "wlist", "", strings.Split("0x4448993f493B1D8D9ED51F22F1d30b9B4377dFD2,0x0d93A21b4A971dF713CfC057e43F5D230E76261C,0x3054e19707447800f0666ba274A249fC9a67aA4a,0xa55203c75c95A95f5DdD7B58E877A9EBd85A1631", ","), "所有钱包地址白名单列表,节点间用逗号分开")
	pflag.StringSliceVarP(&adminList, "alist", "", strings.Split("0x24C93Aaec52539a60240BCd2E972AB672D33eD79", ","), "管理员钱包地址列表,用逗号分开")
	pflag.StringVarP(&env, "env", "e", "prod", "环境名字debug prod test")
	pflag.StringVar(&infura, "infura", "27f0b03a4654478db14295fd1021e1b8", "infura的项目id,需要自行去https://infura.io申请")
	//https://api.thegraph.com/subgraphs/name/wxf4150/fanswap2 https://api.thegraph.com/subgraphs/name/ianlapham/uniswapv2
	pflag.StringVar(&swapGraphApi, "swapGraphApi", "https://api.thegraph.com/subgraphs/name/ianlapham/uniswapv2", "swap theGraphApi")
	pflag.BoolVarP(&job, "job", "j", true, "是否抓取数据")

	//pflag.BoolVarP(&job, "job", "j", false, "是否抓取数据")
	//nodes =  strings.Split("http://localhost:8001,http://localhost:8001,http://localhost:8001",",")
	pflag.Parse()
	utils.Nodes = nodes
	utils.InfuraID = infura
	utils.SetWList(wlist)
	utils.InitDb(dbUrl)
	utils.InitEConn(infura)
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
		chainPairConfs := map[string][]uni.SubPairConfig{
			"eth": []uni.SubPairConfig{{"uniswap", "0xdfb8824b094f56b9216a015ff77bdb056923aaf6","REI","eth"},},
			"bsc": []uni.SubPairConfig{
				{"pancake", "0x8c83e7aef5116be215223d3688a2f5dc4c7f241b","REI","bsc"},
				{"pancake", "0x7C613ccf0656B509fEE51900d55b39308b1FC00d","zUSD","bsc"},
				//	//load wault-usd price for  wault rei-wusd;
				{"wault", "0x6102d8a7c963f78d46a35a6218b0db4845d1612f","WUSD","bsc"},
				{"wault", "0x4b31d95654300cbe8ce3fe2b2ec5c6d2929ae7a6","REI","bsc"},
				{"baby", "0xd02f44fa87f365cd160a033007ef80c311b7f5d9","REI","bsc"},
				{"baby", "0x659951a7f393496232a4e8c308bb4d1ad6400b59","zUSD","bsc"},
			},
			"polygon":[]uni.SubPairConfig{
				{"wault","0xf7bc741b2086ca344e78225d06224ffdcd86d110","WMATIC","polygon"},
			},
		}

		for chainName, subPireConfs := range chainPairConfs {
			go uni.SubPair(chainName, subPireConfs, false, "891eeaa3c7f945b880608e1cc9976284")
		}

		//sync coingecko数据
		//go services.SyncCoinGeckoData()

		//go services.GetStocks()

		go uni.SubEthPrice(0,services.SwapGraphApi)

		go services.SubCoinsPrice()
		//coingecko bull
		//go services.SetAllBulls("btc3x")
		//go services.SetAllBulls("eth3x")

		//go func() {
		//	//监听eth uniswap pair's token价格
		//	//tpc := services.TokenPairConf{PairAddre: "0x4d3c5db2c68f6859e0cd05d080979f597dd64bff", TokenAddre: "0x72e364f2abdc788b7e918bc238b21f109cd634d7", TokenDecimals: 18, ChainName: "eth"}
		//	tpc := uni.TokenPairConf{PairAddre: "0xdfb8824b094f56b9216a015ff77bdb056923aaf6", TokenAddre: "0x011864d37035439e078d64630777ec518138af05", TokenDecimals: 18, ChainName: "eth"}
		//	uni.SubPairlog(&tpc)
		//}()


		////subcribe twelvedata data
		//go services.SubTwData()
		//
		////更新twelvedata数据源bull数据
		//go services.SetAllBullsFromTw(true)

		services.SetAllBullsFromTw(true)
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
	services.SetAList(adminList)

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
	api.Use(common.MiddleWareStat())
	go common.SaveStat()
	sys.InitDicConfig()
	go sys.InitDicConfigWatch()


	api.GET("/stock/info/:code/:data_type/:timestamp", controls.StockInfoHandler)
	api.GET("/stock/market_status/:timestamp", controls.UsaMarketStatusHandler)
	api.POST("/internal/stock_avgprice", controls.StockAvgPriceHandler)

	api.POST("/internal/token_avgprice", controls.TokenAvgHlPriceHandler)

	//api.GET("/stock/aggre_info", StockAggreHandler)
	api.GET("/stock/aggre_info/:code/:data_type/:timestamp", controls.StockAggreHandler)
	api.GET("/stock/stat", controls.NodeStatHandler)
	api.GET("/stock/stats", controls.NodeStatsHandler)
	api.GET("/stock/any_api", controls.NodeAnyApiHandler)
	api.GET("/stock/any_apis", controls.NodeAnyApisHandler)
	api.GET("/internal/dex/lp_price/:pair/:timestamp", controls.PairLpPriceHandler)
	api.GET("/dex/lp_price/:pair/:timestamp", controls.PairLpPriceSignHandler)

	api.GET("/internal/dex/token_price/:token/:timestamp", controls.TokenPriceHandler)
	api.GET("/internal/dex/pair/token_price/:pair/:token/:timestamp", controls.PairTokenPriceHandler)

	api.GET("/internal/dex/token_chain_price/:token/:timestamp", controls.TokenChainPriceHandler3)

	api.GET("/internal/dex/token_info/:token/:timestamp", controls.TokenInfoHandler)
	api.GET("/internal/dex/pair/token_info/:pair/:token/:timestamp", controls.PairTokenInfoHandler)

	api.GET("/internal/coin_price/:coin/:vs_coin", controls.CoinPriceHandler)
	api.GET("/internal/dex/ftx_price/:coin_type/:timestamp", controls.FtxPriceHandler)
	api.GET("/coin_price/:coin/:vs_coin/:timestamp", controls.CoinPriceSignHandler)

	api.GET("/dex/token_price/:token/:data_type/:timestamp", controls.TokenPriceSignHandler)
	api.GET("/dex/pair/token_price/:pair/:token/:da:qta_type/:timestamp", controls.PairTokenPriceSignHandler)

	api.GET("/dex/token_chain_price/:token/:data_type/:timestamp", controls.TokenChainPriceSignHandler)
	api.GET("/dex/ftx_price/:coin_type/:data_type/:timestamp", controls.FtxPriceSignHandler)
	api.GET("/dex/token_day_datas/:token/:days/:timestamp", controls.TokenDayDatasHandler)
	api.GET("/dex/ftx_chart_prices/:coin_type/:count/:interval/:timestamp", controls.FtxChartPricesHandler)
	api.GET("/dex/uni_chart_prices/:coin_type/:count/:interval/:timestamp", controls.UniChainChartPricesHandler)
	api.GET("/dex/stock_chart_prices/:coin_type/:count/:interval/:timestamp", controls.StockChartPricesHandler)

	api.GET("/dex/token_chart_prices/:token/:count/:interval/:timestamp", controls.TokenDayPricesHandler)
	api.GET("/dex/pair/token_chart_prices/:pair/:token/:count/:interval/:timestamp", controls.PairTokenDayPricesHandler)

	api.GET("/dex/token/token_chart_supply/:token/:amount/:timestamp", controls.TokenChartSupplyHandler)
	api.GET("/dex/token/token_total_supply/:token/:timestamp", controls.GetTokenSupplyHandler)
	api.GET("/alert/ok", controls.OkHandler)
	api.GET("/alert/coindata", controls.CoinDataCheckHandler)
	api.GET("/alert/btc_sign_check", controls.BtcSignCheckHandler)
	api.GET("/alert/ftx_price_check", controls.FtxPriceCheck)
	api.GET("/db.sql.gz", controls.DbExportHandler)


	api.GET("dic_config", sys.ConfigHandler)
	api.POST("dic_config", sys.ConfigUpdateHandler)
	admin := router.Group("/sys")
	sys.InitJwt(admin)
	api.POST("/login", sys.AuthMiddleware.LoginHandler)
	api.GET("/pre_login", sys.ChallengeHandler)
	admin.GET("/hello", sys.HelloJwtHandler)
	admin.GET("/user_info", sys.UserInfoHandler)
	api.GET("/ftxs", sys.FtxListHandler)
	admin.POST("dic_config", sys.ConfigUpdateHandler)

	//api.POST("/stock/sign_verify", VerifyInfoHandler)

	router.NoRoute(func(c *gin.Context) {
		common.ErrJson(c, "none api router")
		//c.JSON(404,controls.ApiErr{Error:"none api router"})
	})
	go router.RunTLS(":8002", "./asset/tls.pem", "./asset/tls.pem")
	log.Fatal(router.Run(":" + serverPort))
}

