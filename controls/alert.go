package controls

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"stock/utils"
	"strconv"
	"time"
)

func BtcSignCheckHandler(c *gin.Context) {
	res,err:=ftxPriceSignHandler("btc3x",1,238299929)
	if err != nil {
		ErrJson(c,err.Error())
		return
	}
	//res.Sign=nil
	if res.Sign==nil{
		time.Sleep(3*time.Second)
		c.JSON(200,"BtcSignCheck err")
		return
	}
	c.JSON(200,"BtcSignCheck ok")
}
func CoinDataCheckHandler(c *gin.Context) {
	d,_:=strconv.Atoi(c.Query("d"))
	if d==0{
		d=300
	}

	counter:=0
	err:=utils.Orm.Raw(`
select count(1) cc from  market_prices t where
t.item_type in('btc','eth') and timestamp >unix_timestamp()-?;
`,d).Scan(&counter).Error
	if err != nil {
		ErrJson(c,err.Error())
		return
	}
	if counter==0{
		time.Sleep(3*time.Second)
		c.JSON(200,"CoinData err")
		return
	}
	c.JSON(200,fmt.Sprintf("CoinData ok counter%d",counter))
}
func OkHandler(c *gin.Context) {
	c.JSON(200,"ok")
}

