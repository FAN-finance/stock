package main

import (
	//"crypto"
	//"crypto/sha256"
	//"os"

	//"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	//"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	cors "github.com/itsjamie/gin-cors"
	"github.com/spf13/pflag"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"io/ioutil"
	"log"
	"stock/services"
	"strconv"
	"strings"
	"sync"
	"time"
	_ "stock/docs"
	"stock/utils"
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
	var dbUrl,serverPort,env string
	var job bool

	pflag.StringVarP(&dbUrl,"db","d","root:password@tcp(localhost:3306)/mydb?loc=Local&parseTime=true&multiStatements=true","mysql database url")
	pflag.StringVarP(&serverPort,"port","p","8001","api　service port")
	//var keyfile,certFile string
	//pflag.StringVarP(&keyfile,"key","k","./asset/key.pem","pem encoded private key")
	//pflag.StringVarP(&certFile,"cert","c","./asset/cert.pem","pem encoded x509 cert")
	pflag.StringSliceVarP(&Nodes,"nodes","n",strings.Split("http://localhost:8001,http://localhost:8001",","),"所有节点列表,节点间用逗号分开")
	pflag.StringVarP(&env,"env","e","debug","环境名字debug prod test")
	pflag.BoolVarP(&job,"job","j",true,"是否抓取数据")

	pflag.Parse()
	utils.InitDb(dbUrl)
	if job {
		go services.GetStocks()
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
	api.GET("/stock/info", StockInfoHandler)
	//api.GET("/stock/aggre_info", StockAggreHandler)
	api.GET("/stock/aggre_info/:code/:timestamp", StockAggreHandler)
	api.GET("/stock/stat", NodeStatHandler)
	api.GET("/stock/stats", NodeStatsHandler)
	//api.POST("/stock/sign_verify", VerifyInfoHandler)

	router.NoRoute(func(c *gin.Context){
		ErrJson(c,"none api router")
		//c.JSON(404,controls.ApiErr{Error:"none api router"})
	})
	log.Fatal(router.Run(":"+serverPort))
}
var Nodes []string
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
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
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
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
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
	for _, nurl := range Nodes {
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
	if len(sdata.Signs)<len(Nodes)/2+1{
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
	c.JSON(200,sdata)
	return

	END:
	if err != nil {
		ErrJson(c,err.Error())
		return
	}
}


// @Tags default
// @Summary　当前节点状态:记录数,钱包地址
// @Description 当前节点状态:记录数,钱包地址
// @ID NodeStatHandler
// @Accept  json
// @Produce  json
// @Success 200 {string} addr	"stock info"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
// @Router /pub/stock/stat [get]
func NodeStatHandler(c *gin.Context) {
	stat:=NodeStat{}
	utils.Orm.Model(services.ViewStock{}).Count(&(stat.Rows))
	stat.WalletAddre=services.WalletAddre
	c.JSON(200,stat)
}


type NodeStat struct {
	//节点名
	Node string
	//钱包地址
	WalletAddre string
	//数据库记录数
	Rows int64
}
// @Tags default
// @Summary　所有节点状态:记录数,钱包地址
// @Description 所有节点状态:记录数,钱包地址
// @ID NodeStatsHandler
// @Accept  json
// @Produce  json
// @Success 200 {array} NodeStat	"stock info"
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
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
	for _, nurl := range Nodes {
		wg.Add(1)
		go porcNode(nurl)
	}
	wg.Wait()
	c.JSON(200,addres)
	return

	if err != nil {
		ErrJson(c,err.Error())
		return
	}
}

// @Tags default
// @Summary　当前节点状态:记录数,钱包地址
// @Description 当前节点状态:记录数,钱包地址
// @ID NodeStatHandler
// @Accept  json
// @Produce  json
// @Success 200 {string} addr	"stock info"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
// @Router /pub/stock/stat [get]
func NodeAnyApiHandler(c *gin.Context) {
	stat:=NodeStat{}
	utils.Orm.Model(services.ViewStock{}).Count(&(stat.Rows))
	stat.WalletAddre=services.WalletAddre
	c.JSON(200,stat)
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
//// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
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


type ApiErr struct{
	Error string `json:"Error"`
}
type ApiOk struct{
	Msg string `json:"Msg" example:"ok"`
}

func ErrJson(c *gin.Context,msg string){
	if strings.HasPrefix( msg,"401"){
		c.JSON(400, ApiErr{msg})
		return
	}
	c.JSON(500, ApiErr{msg})
}
func OkJson(c *gin.Context,err error){
	if err!=nil{
		ErrJson(c,err.Error())
	}else{
		c.JSON(200, ApiOk{"ok"})
	}
}


var  Privkey *rsa.PrivateKey
var  LocalCert *x509.Certificate
func InitKey(keyFile,certFile string) {
	bs,err:=ioutil.ReadFile(keyFile)
	if err == nil {
	pblock,_:=pem.Decode(bs)

		priv,err1:=x509.ParsePKCS8PrivateKey(pblock.Bytes)
		err=err1
		if err == nil {
			Privkey =priv.(*rsa.PrivateKey)
		}
	}
	if err != nil {
		log.Fatalln("init pkey err",err)
	}
	log.Println("Privkey", Privkey.D)

	bs,err=ioutil.ReadFile(certFile)
	if err == nil {
		pblock,_:=pem.Decode(bs)

		c,err1:=x509.ParseCertificate(pblock.Bytes)
		err=err1
		if err == nil {
			LocalCert=c
		}
	}
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("init cert err",LocalCert.Subject)
}