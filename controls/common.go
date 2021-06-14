package controls

import (
	"github.com/gin-gonic/gin"
	"strings"
	"stock/utils"
	"log"
	"strconv"
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
