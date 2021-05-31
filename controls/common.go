package controls

import (
	"github.com/gin-gonic/gin"
	"strings"
)

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
