package controls

import (
	jwtgin "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"net/http"
	"stock/mmlogin"
	"stock/mmlogin/application/auth"
	"stock/services"
	"stock/utils"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ResBody struct {
	Code int    `example:"200" json:"code"`
	Msg  string `example:"ok" json:"msg"`
	//Data interface{} `json:"data"`
	Data interface{} `json:"result"`
	Time time.Time
}
type ResPager struct {
	Page  int         `example:"1" json:"page"`
	PageSize int         `example:"10" json:"page_size"`
	Total    int         `example:"100" json:"total"`
	Result   interface{} `json:"result"`
}
func NewResBody(c *gin.Context, data interface{}) {
	resb:=new(ResBody)
	resb.Msg="success"
	resb.Code=200
	resb.Data=data
	resb.Time=time.Now()
	c.JSON(200, resb)
}
func NewResListBody(c *gin.Context, pageNum,pageSize,total int, listData interface{}) {
	resb:=new(ResBody)
	resPager:=new(ResPager)
	resPager.Page=pageNum
	resPager.PageSize=pageSize
	resPager.Result=listData
	resPager.Total=total
	resb.Data=resPager
	resb.Time=time.Now()
	resb.Code=200
	resb.Msg="success"
	c.JSON(200, resb)
}

func ResErrMsgWithCode(c *gin.Context, err string ) {
	//msg:=GetCCErrorStr(err)
	resb:=new(ResBody)
	resb.Msg=err
	resb.Code=500
	c.JSON(200, resb)
}
func ResErrWithCode(c *gin.Context, err error, code int ) {
	//msg:=GetCCErrorStr(err)
	resb:=new(ResBody)
	resb.Msg=err.Error()
	if code==0{
		code=500
	}
	resb.Code=code
	c.JSON(200, resb)
}

type ApiErr struct {
	Error string `json:"Error"`
}
type ApiOk struct {
	Msg string `json:"Msg" example:"ok"`
}

func ErrJson(c *gin.Context, msg string) {
	if strings.HasPrefix(msg, "40") {
		c.JSON(400, ApiErr{msg})
		return
	}
	c.JSON(500, ApiErr{msg})
}
func OkJson(c *gin.Context, err error) {
	if err != nil {
		ErrJson(c, err.Error())
	} else {
		c.JSON(200, ApiOk{"ok"})
	}
}

//cache *******
func SetCacheRes(c *gin.Context, ckey string, setHeaderCache bool, process func() (interface{}, error), debug bool) {
	SetCacheResExpire(c, ckey, setHeaderCache, 100, process, debug)
}
func SetCacheResExpire(c *gin.Context, ckey string, setHeaderCache bool, expire int64, process func() (interface{}, error), debug bool) {
	var res interface{}
	var err error
	if debug {
		res, err = process()
	} else {
		log.Println("cache process", ckey)
		res, err = utils.CacheFromLru(1, ckey, int(expire), process)
	}
	if err == nil {
		headerTtl := utils.CalcExpiration(expire, ckey) - time.Now().Unix()
		log.Println("set cache header", ckey, headerTtl)
		if !debug && setHeaderCache {
			SetExireHeader(c, headerTtl)
		}
		c.JSON(200, res)
	} else {
		ErrJson(c, err.Error())
	}
}
func SetExireHeader(c *gin.Context, seconds int64) {
	c.Header("Cache-Control", "max-age="+strconv.Itoa(int(seconds)))
}

//api stat
var ReqStatMap = new(sync.Map)

func MiddleWareStat() gin.HandlerFunc {
	return func(c *gin.Context) {
		//key:=c.Request.URL.Path
		key := c.FullPath()
		_, ok := services.StatPath2IDMap[key]
		if !ok {
			key = "other"
		}
		counter, _ := ReqStatMap.LoadOrStore(key, 0)
		ReqStatMap.Store(key, counter.(int)+1)
		//log.Println("req ",key,counter)
	}
}

type ApiStat struct {
	ID         uint
	Timestamp  int64  `gorm:"index:idx_ti,priority:1;uniqueIndex:idx_node_t,priority:2"`
	IsInternal bool   `gorm:"index:idx_ti,priority:2"`
	PathID     int    `gorm:"type:tinyint;uniqueIndex:idx_node_t,priority:3"`
	Pathstr    string `gorm:"type:varchar(256);`
	Counter    int
	CreatedAt  time.Time
	NodeAddr   string `gorm:"type:varchar(50);uniqueIndex:idx_node_t,priority:1"`
}

func isUrlInternal(key string) bool {
	if strings.HasPrefix(key, "/pub/internal/dex/token_info") {
		return false
	} else if strings.HasPrefix(key, "/pub/internal/") {
		return true
	} else if strings.HasPrefix(key, "/pub/stock/info/") {
		return true
	}
	return false
}
func (rs *ApiStat) BeforeCreate(tx *gorm.DB) (err error) {
	rs.IsInternal = isUrlInternal(rs.Pathstr)
	rs.PathID = services.StatPath2IDMap[rs.Pathstr]
	return nil
}
func SaveStat() {
	utils.Orm.AutoMigrate(ApiStat{})
	proc := func() error {
		rss := []*ApiStat{}
		f := func(k, v interface{}) bool {
			key := k.(string)
			value := v.(int)
			//log.Println(key,value)
			rs := new(ApiStat)
			rs.Pathstr = key
			rs.Counter = value
			rs.Timestamp = time.Now().Unix()
			rs.NodeAddr = services.WalletAddre
			rss = append(rss, rs)
			ReqStatMap.Delete(key)
			//log.Println("rs ",key,value)
			return true
		}
		ReqStatMap.Range(f)
		if len(rss) > 0 {
			err := utils.Orm.CreateInBatches(rss, 100).Error
			return err
		}
		return nil
	}
	utils.IntervalSync("saveStat", 600, proc)
}

//jwt
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
				if userID==services.WalletAddre{
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
			NewResBody(c,gin.H{
				"code":    http.StatusOK,
				"token":   token,
				"expire":  t.Format(time.RFC3339),
				"message": "login successfully",
				"userInfo":lres,
			})
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
			ErrJson(c,msg+" jwt")
		},
	})
	if err != nil {
		log.Fatalln(err)
	}
	AuthMiddleware = authMiddleware1
	routeGroup.Use(AuthMiddleware.MiddlewareFunc())
}
 func ChallengeHandler(c *gin.Context)  {
	addressHex := c.Query("address")
	in := auth.NewChallengeInput(addressHex)

	out, err := mmlogin.Apps.Auth.Challenge(nil, in)
	if err != nil {
		OkJson(c,err)
		return
	}
	NewResBody(c,out)
}

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