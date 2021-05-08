package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"github.com/gin-gonic/gin"
	cors "github.com/itsjamie/gin-cors"
	"github.com/spf13/pflag"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"io/ioutil"
	"log"
	"stock/services"
	"strings"
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
	var dbUrl,serverPort,keyfile,certFile,env string
	pflag.StringVarP(&dbUrl,"db","d","root:password@tcp(localhost:3306)/mydb?loc=Local&parseTime=true&multiStatements=true","mysql database url")
	pflag.StringVarP(&serverPort,"port","p","8001","api　service port")
	pflag.StringVarP(&keyfile,"key","k","./asset/key.pem","pem encoded private key")
	pflag.StringVarP(&certFile,"cert","c","./asset/cert.pem","pem encoded x509 cert")
	pflag.StringVarP(&env,"env","e","debug","环境名字debug prod test")
	pflag.Parse()

	utils.InitDb(dbUrl)
	go services.GetStocks()

	InitKey(keyfile,certFile)
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
	api.POST("/stock/sign_verify", VerifyInfoHandler)

	router.NoRoute(func(c *gin.Context){
		ErrJson(c,"none api router")
		//c.JSON(404,controls.ApiErr{Error:"none api router"})
	})
	log.Fatal(router.Run(":"+serverPort))
}

// @Tags default
// @Summary　获取美股价格:
// @Description 获取美股价格 苹果代码  AAPL  ,苹果代码 TSLA
// @ID StockInfoHandler
// @Accept  json
// @Produce  json
// @Param     code   query    string     true        "美股代码" default(AAPL)
// @Param     timestamp   query    int     false    "unix 秒数" default(1620383144)
// @Success 200 {object} services.ViewStock	"stock info"
// @Header 200 {string} sign "签名信息"
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
// @Router /pub/stock/info [get]
func StockInfoHandler(c *gin.Context) {
	info := &services.ViewStock{}
	code:=c.Query("code")
	timestamp:=c.Query("timestamp")
	err := utils.Orm.Model(services.ViewStock{}).Where("code= ? and timestamp>= ? ", code,timestamp).Order("timestamp").First(info).Error
	if err == nil {
		if err == nil {
			bs,_:=json.Marshal(info)
			//md5str:=crypto.SHA256.New()
			hashbs:=sha256.Sum256(bs)
			log.Println(hashbs,len(hashbs))
			sign,signErr:= Privkey.Sign(rand.Reader,hashbs[0:32],crypto.SHA256)
			if signErr == nil {
				signStr:=base64.StdEncoding.EncodeToString(sign)
				c.Header("sign",signStr)
				log.Println(signStr)
			}else{
				log.Println(signErr)
			}
			c.JSON(200,info)
			return
		}
	}
	if err != nil {
		ErrJson(c,err.Error())
		return
	}
}

type VerObj struct {
	//stockInfo json: {"code":"AAPL","price":128.1,"name":"苹果","timestamp":1620292445,"UpdatedAt":"2021-05-06T17:14:05.878+08:00"}
	Data json.RawMessage `swaggertype:"object"`
	Sign []byte `swaggertype:"string" format:"base64" example:"UhRVNsT8B5Za6oO3APH0T9ebPMKHxDDhkscYuILl7lDepDMzyBaQsEu9vwTRIfoYBS8udfEanI/DUAhwnIdFJf9woIv7Oo+OS6q3sF3B5Vx9NN2ipXJ4wjTf2ct7FbS1vXAvTXSmA2svj+LF8P1PIEClITBqu/EWZXTpHvAlbGAAeF+hHO7/FquLHVDavLC+OENyb0CP+NvH+ytZ69tav0DqbGp+NGGil/ImZpPsetbOxwuhC/U1CV6Ap8qgRWe8s6IpOawXDAavLMHUmXVvORDf/XVzaQUJ5ob+vTsSTZwQsvj/4jmsODFt8eKFYL/7vyN/i3HkiDwhq0w85kqHgg=="`
}

// @Tags default
// @Summary　签名验证:
// @Description 签名验证
// @ID VerifyInfoHandler
// @Accept  json
// @Produce  json
// @Param     verObj   body    VerObj     true        "需要验证的对象" default(AAPL)
// @Success 200 {object} ApiOk	"ok info"
// @Failure 500 {object} ApiErr "失败时，有相应测试日志输出"
// @Router /pub/stock/sign_verify [post]
func VerifyInfoHandler(c *gin.Context) {
	vobj:=new(VerObj)
	err:=c.Bind(vobj)
	if err == nil {
		hashbs:=sha256.Sum256(vobj.Data)
		err=rsa.VerifyPKCS1v15(LocalCert.PublicKey.(*rsa.PublicKey),crypto.SHA256,hashbs[0:32],vobj.Sign,)
		if err == nil {
			c.JSON(200, ApiOk{"ok"})
			return
		}
	}
	if err != nil {
		ErrJson(c,err.Error())
		return
	}
}


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