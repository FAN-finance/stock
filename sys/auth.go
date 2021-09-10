package sys

import (
	jwtgin "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"stock/common"
	"stock/services"
	"stock/sys/mmlogin"
	"stock/sys/mmlogin/application/auth"
	"stock/utils"
	"time"
)


type loginModel struct {
	//Username string `form:"username" json:"username" binding:"required"`
	//Password string `form:"password" json:"password" binding:"required"`

	Address string `form:"address" json:"address" binding:"required"`
	Signature string `form:"signature" json:"signature" binding:"required"`

}
type loginRes struct {
	Address string `form:"address" json:"address" binding:"required"`
	IsAdmin bool `form:"isAdmin" json:"isAdmin" binding:"required"`
	Token string `form:"token" json:"token" binding:"false"`
}

func SetJwtSecretFile()[]byte{
	//secretKey
	var signKeyJwt []byte
	sfile:="asset/jwt_secret"
	fi, err := os.Stat(sfile)
	if err == nil && !fi.IsDir() {
		signKeyJwt,err=ioutil.ReadFile(sfile)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("loaded jwt_secret")
	}else if err!=nil && os.IsNotExist(err){
		signKeyJwt =[]byte(utils.RandStr(32))
		err1:=ioutil.WriteFile(sfile,signKeyJwt,0644)
		if err1 != nil {
			log.Println("WriteFile err",err1)
		}
	}else{
		log.Println("SetJwtSecretFile stat",err)
	}
	if signKeyJwt==nil{
		signKeyJwt =[]byte(utils.RandStr(32))
	}
	log.Println("signKeyJwt:",string(signKeyJwt))
	return signKeyJwt
}

var AuthMiddleware *jwtgin.GinJWTMiddleware
func InitJwt(routeGroup *gin.RouterGroup) {
	mmlogin.InitMMLogin()
	var identityKey = "id"
	var signKeyJwt =SetJwtSecretFile()
	authMiddleware1, err := jwtgin.New(&jwtgin.GinJWTMiddleware{
		Realm: "test zone",
		Key:   signKeyJwt,
		IdentityKey:identityKey,
		PayloadFunc: func(data interface{}) jwtgin.MapClaims {
			// Set custom claim, to be checked in Authorizator method
			if v, ok := data.(*loginRes); ok {
				return jwtgin.MapClaims{
					identityKey: v.Address,
					"isAdmin": v.IsAdmin,
				}
			}
			return jwtgin.MapClaims{}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginVals loginModel
			if binderr := c.ShouldBind(&loginVals); binderr != nil {
				return "", jwtgin.ErrMissingLoginValues
			}
			userID := loginVals.Address
			password := loginVals.Signature
			in := auth.NewAuthorizeInput(userID, password)
			err:= mmlogin.Apps.Auth.AuthorizeOnly(nil,in);
			if err!=nil{
				return nil,err
			}else{
				lres:=loginRes{userID,false,""}
				if services.IsAdmin(userID){
					lres.IsAdmin=true
				}
				c.Set("loginres",lres)
				return &lres,nil
			}
			//if userID == "admin" && password == "admin" {
			//	return userID, nil
			//}
			//return "", jwtgin.ErrFailedAuthentication
		},
		Authorizator: func(user interface{}, c *gin.Context) bool {
			//recover claims instance
			lres:=new(loginRes)
			claims:=jwtgin.ExtractClaims(c)
			lres.IsAdmin=claims["isAdmin"].(bool)
			lres.Address=claims["id"].(string)
			c.Set("loginres",lres)
			return true
		},
		LoginResponse: func(c *gin.Context, code int, token string, t time.Time) {
			res,_:=c.Get("loginres")
			lres:=res.(loginRes)
			lres.Token=token
			c.JSON(200,lres)
			//common.NewResBody(c,lres)
			//c.JSON(http.StatusOK, gin.H{
			//	"code":    http.StatusOK,
			//	"token":   token,
			//	"expire":  t.Format(time.RFC3339),
			//	"message": "login successfully",
			//	"cookie":  cookie,
			//})
		},
		SendCookie: false,
		//CookieName:   cookieName,
		//CookieDomain: cookieDomain,
		TimeFunc: func() time.Time { return time.Now().Add(time.Duration(5) * time.Minute) },
		Timeout:  time.Hour * 24,
		Unauthorized:func(c *gin.Context, code int, msg string){
			common.ErrJsonWithCode(c,msg+" jwt",http.StatusUnauthorized)
		},
	})
	if err != nil {
		log.Fatalln(err)
	}
	AuthMiddleware = authMiddleware1
	routeGroup.Use(AuthMiddleware.MiddlewareFunc())
}


// @Tags sys
// @Summary　 钱包登陆Challenge
// @Description 钱包登陆Challenge
// @ID ChallengeHandler
// @Accept  json
// @Produce  json
// @Param     address   query    string     true    "钱包地址" default(0x24C93Aaec52539a60240BCd2E972AB672D33eD79)
// @Success 200 {object} auth.ChallengeOutput	"Challenge　code"
// @Failure 500 {object} common.ResBody "失败时，有相应测试日志输出"
// @Router /pub/pre_login [get]
func ChallengeHandler(c *gin.Context)  {
	addressHex := c.Query("address")
	in := auth.NewChallengeInput(addressHex)

	out, err := mmlogin.Apps.Auth.Challenge(nil, in)
	if err != nil {
		common.ResErrMsg(c,err.Error())
		return
	}
	//common.ResErrMsg(c,"test 400 json")
	//common.ErrJson(c,"test 400 json")
	//c.JSON(400,"400 test")
	//return
	c.JSON(200,out)
	//common.NewResBody(c,out)
}

// @Tags sys
// @Summary　 钱包登陆
// @Description 钱包登陆
// @ID LoginHandler
// @Accept  json
// @Produce  json
// @Param     logreq   body    loginModel     true    "钱包登陆对象"
// @Success 200 {object} loginRes	"登陆结果"
// @Failure 500 {object} common.ResBody "失败时，有相应测试日志输出"
// @Router /pub/login [post]
func LoginHandler(){}


func HelloJwtHandler(c *gin.Context) {
	claims := jwtgin.ExtractClaims(c)
	user, _ := c.Get("id")
	c.JSON(200, gin.H{
		"userID": user,
		//"userName": user.(*User).UserName,
		"text":     "Hello World.",
		"claims":claims,
	})
}

func IsAdmin(c *gin.Context)bool{
	claims := jwtgin.ExtractClaims(c)
	isadmin,ok:=claims["isAdmin"]
	if ok && isadmin.(bool){
		return true
	}
	return false
}

// @Tags sys
// @Summary　 用户信息
// @Description 用户信息
// @ID UserInfoHandler
// @Accept  json
// @Produce  json
// @Success 200 {object} userInfo	"用户信息"
// @Failure 500 {object} common.ResBody "失败时，有相应测试日志输出"
// @Router /sys/user_info [get]
func UserInfoHandler(c *gin.Context){
	uinfo:=new(userInfo)
	uinfo.Name="wallet"
	uinfo.Avatar="https://price.btcfans.com/assets/price/coin-logo/bitcoin.png?big"
	uinfo.Address=AuthMiddleware.IdentityHandler(c).(string)
	uinfo.IsAdmin=services.IsAdmin(uinfo.Address)
	if uinfo.IsAdmin{
		uinfo.Access= "admin"
	}else{
		uinfo.Access= "guest"
	}
	c.JSON(200,uinfo)
}

type userInfo struct {
	loginRes
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Access string `json:"access"`
}