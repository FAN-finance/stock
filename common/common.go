package common

import (
	jwtgin "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
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
	Data interface{} `json:"data" swaggertype:"object"`
	Time time.Time
}
type ResPager struct {
	Page  int         `example:"1" json:"page"`
	PageSize int         `example:"10" json:"page_size"`
	Total    int         `example:"100" json:"total"`
	Result   interface{} `json:"result" swaggertype:"object"`
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

func ResErrMsg(c *gin.Context, err string ) {
	//msg:=GetCCErrorStr(err)
	resb:=new(ResBody)
	resb.Msg=err
	resb.Code=500
	resb.Time=time.Now()
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
	resb.Time=time.Now()
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
func IsAdmin(c *gin.Context)bool{
	claims := jwtgin.ExtractClaims(c)
	isadmin,ok:=claims["isAdmin"]
	if ok && isadmin.(bool){
		return true
	}
	return false
}


//对外配制编辑项
//区别于controls系统正在使用的map-config结构（ 如SafePrices　ftxAddres）; 在web-ui上编辑时,map项不能按添加顺序显示，操作上很不方便
//SafePrices　FtxTokenAddres换用数组方式的配制．
type RawDicConfig struct{
	//禁用所有签名
	IsDisableAllSign bool
	//禁用ftx签名
	IsDisableFtxSign bool
	SafePrices []sp
	FtxTokenAddres []fa
}
//ftxAddres
type fa struct {
	FtxName string
	TokenAddre string
}
//SafePrices
type sp struct {
	Min float64
	Max float64
	TokenAddre string
	//备注　可选项，可输入名字　
	Comment string
}

func (conig *RawDicConfig)BasicVerify(){
	//basic verify
	if len(conig.SafePrices)==0{
		log.Fatal("dic_config SafePrices err")
	}
	if len(conig.FtxTokenAddres)==0{
		log.Fatal("dic_config FtxTokenAddres err")
	}
}

