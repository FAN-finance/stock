package sys

import (
	jwtgin "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"log"
	"stock/common"
	"stock/mmlogin"
	"stock/mmlogin/application/auth"
	"stock/services"
	"stock/utils"
	"time"
	"errors"
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
}
var AuthMiddleware *jwtgin.GinJWTMiddleware
func InitJwt(routeGroup *gin.RouterGroup) {
	mmlogin.InitMMLogin()
	var identityKey = "id"
	var signKeyJwt = []byte(utils.RandStr(32))
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
				lres:=loginRes{userID,false}
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
			lres,_:=c.Get("loginres")
			common.NewResBody(c,lres)
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
			common.ResErrWithCode(c,errors.New(msg+" jwt"),code,)
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
		common.OkJson(c,err)
		return
	}
	common.NewResBody(c,out)
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