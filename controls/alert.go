package controls

import "github.com/gin-gonic/gin"

func OkHandler(c *gin.Context) {
	c.JSON(200,"ok")
}

//select count(1) from  market_prices t where
//t.item_type in('btc','eth') and timestamp >unix_timestamp()-300;
