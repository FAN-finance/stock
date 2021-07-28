package controls

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"stock/services"
	"strings"
	"stock/utils"
	"log"
	"strconv"
	"sync"
	"time"
)

type ApiErr struct{
	Error string `json:"Error"`
}
type ApiOk struct{
	Msg string `json:"Msg" example:"ok"`
}

func ErrJson(c *gin.Context,msg string){
	if strings.HasPrefix( msg,"40"){
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

func SetCacheRes(c *gin.Context, ckey string,setHeaderCache bool,process func() (interface{}, error) , debug bool){
	SetCacheResExpire(c,ckey,setHeaderCache,100, process,debug)
}
func SetCacheResExpire(c *gin.Context, ckey string,setHeaderCache bool,expire int64,process func() (interface{}, error) , debug bool){
	var res interface{}
	var err error
	if debug {
		res, err = process()
	} else {
		log.Println("cache process",ckey)
		res, err = utils.CacheFromLru(1, ckey, int(expire), process)
	}
	if err == nil {
		headerTtl:=utils.CalcExpiration(expire,ckey)-time.Now().Unix()
		log.Println("set cache header",ckey,headerTtl)
		if !debug && setHeaderCache {
			SetExireHeader(c, headerTtl)
		}
		c.JSON(200,res)
	} else {
		ErrJson(c,err.Error())
	}
}
func SetExireHeader(c *gin.Context,seconds int64){
	c.Header("Cache-Control", "max-age="+strconv.Itoa(int(seconds)))
}

var ReqStatMap=new(sync.Map)
func  Stat() gin.HandlerFunc {
	return func(c *gin.Context) {
		//key:=c.Request.URL.Path
		key:=c.FullPath()
		_,ok:=services.StatPath2IDMap[key]
		if !ok{
			key="other"
		}
		counter,_:=ReqStatMap.LoadOrStore(key,0)
		ReqStatMap.Store(key,counter.(int)+1)
		log.Println("req ",key,counter)
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
func isUrlInternal(key string) bool{
	if strings.HasPrefix(key,"/pub/internal/dex/token_info"){
		return false
	}else if strings.HasPrefix(key,"/pub/internal/"){
		return true
	}else if strings.HasPrefix(key,"/pub/stock/info/"){
		return true
	}
	return false
}
func (rs *ApiStat) BeforeCreate(tx *gorm.DB) (err error) {
	rs.IsInternal=isUrlInternal(rs.Pathstr)
	rs.PathID=services.StatPath2IDMap[rs.Pathstr]
	return nil
}
func SaveStat(){
	utils.Orm.AutoMigrate(ApiStat{})
	proc:=func()(error) {
		rss:=[]*ApiStat{}
		f := func(k, v interface{}) bool {
			key:=k.(string)
			value:=v.(int)
			//log.Println(key,value)
			rs:=new(ApiStat)
			rs.Pathstr =key
			rs.Counter=value
			rs.Timestamp=time.Now().Unix()
			rs.NodeAddr=services.WalletAddre
			rss=append(rss,rs)
			ReqStatMap.Delete(key)
			//log.Println("rs ",key,value)
			return true
		}
		ReqStatMap.Range(f)
		if len(rss)>0 {
			err := utils.Orm.CreateInBatches(rss, 100).Error
			return err
		}
		return nil
	}
	utils.IntervalSync("saveStat",600,proc)
}
