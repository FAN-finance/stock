package controls

import (
	"github.com/gin-gonic/gin"
)

type config struct{
	//禁用所有签名
	IsDisableAllSign bool
	//禁用ftx签名
	IsDisableFtxSign bool
	SafePrice map[string]*mm
	FtxTokenAddres map[string]string
}

var Config =new(config)
func InitConfig(){
	Config.IsDisableAllSign=IsDisableAllSign
	Config.IsDisableFtxSign=IsDisableFtxSign
	Config.SafePrice=safePrice
	Config.FtxTokenAddres=ftxAddres
}


// @Tags default
// @Summary　配制信息
// @Description 配制信息
// @ID ConfigHandler
// @Accept  json
// @Produce  json
// @Success 200 {object} config	"config json"
// @Failure 500 {object} controls.ApiErr "失败时，有相应测试日志输出"
// @Router /pub/system/config [get]
func ConfigHandler(c *gin.Context) {
	c.JSON(200, Config)
}