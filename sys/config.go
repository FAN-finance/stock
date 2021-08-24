package sys

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"stock/common"
	"stock/controls"
)

// @Tags default
// @Summary　获取字典配制信息
// @Description 字典配制信息
// @ID ConfigHandler
// @Accept  json
// @Produce  json
// @Param     is_using   query    string     false   "配制json"
// @Success 200 {object} common.RawDicConfig	"config json"
// @Failure 500 {object} common.ResBody "失败时，有相应测试日志输出"
// @Router /pub/dic_config [get]
func ConfigHandler(c *gin.Context) {
	if c.Query("is_using")==""{
		common.NewResBody(c,RawDic)
	}else{
		common.NewResBody(c,controls.Config)
	}
}

// @Tags default
// @Summary　修改字典配制信息
// @Description 修改字典配制信息
// @ID ConfigHandler
// @Accept  json
// @Produce  json
// @Param     dic   body    common.RawDicConfig     true    "配制json"
// @Success 200 {object} common.RawDicConfig	"data"
// @Success 200 {object} common.RawDicConfig	"config json"
// @Failure 500 {object} common.ResBody "失败时，有相应测试日志输出"
// @Router /pub/dic_config [post]
func ConfigUpdateHandler(c *gin.Context) {
	if !IsAdmin(c){
		common.ResErrMsg(c,"you are not administrator")
		return
	}
	tDicConf:=new(common.RawDicConfig)
	err:=c.ShouldBindBodyWith(tDicConf,binding.JSON)
	c.GetRawData()
	if err != nil {
		common.ResErrMsg(c,err.Error())
		return
	}
	tDicConf.BasicVerify()
	rawBs,_:=c.Get(gin.BodyBytesKey)
	err=ioutil.WriteFile("./asset/dic_config1.json",rawBs.([]byte),0644)
	if err != nil {
		common.ResErrMsg(c,err.Error())
		return
	}

	//dic-config admin; only on admin-server
	shName:="./asset/sync_dic_config.sh"
	_,errSync:=os.Stat(shName)
	if errSync == nil {
		var command = exec.Command("bash", "-c", shName )
		out, err1 := command.Output()
		err = err1
		log.Println("sync dic config" + string(out))
		if err != nil {
			log.Println("sync dic config err",err.Error())
			common.ResErrMsg(c,err.Error())
			return
		}
	}else{
		log.Println("not exist  sync-dic sh")
	}
	RawDic=tDicConf
	controls.BeginUseRawDicConfig(RawDic)
	c.JSON(200, RawDic)
}



var RawDic =new(common.RawDicConfig)
func InitDicConfig(){
	bs,err:=ioutil.ReadFile("./asset/dic_config.json")
	if err != nil {
		log.Fatal(err,"dic_config err")
	}
	err=json.Unmarshal(bs, RawDic)
	if err != nil {
		log.Fatal("dic_config err")
	}
	RawDic.BasicVerify()
	controls.BeginUseRawDicConfig(RawDic)

	//RawDic.IsDisableAllSign=IsDisableAllSign
	//RawDic.IsDisableFtxSign=IsDisableFtxSign
	//for key, value := range safePrice {
	//	RawDic.SafePrices=append(RawDic.SafePrices,sp{mm{value.Min,value.Max},key ,""} )
	//}
	//for key, value := range ftxAddres {
	//	RawDic.FtxTokenAddres=append(RawDic.FtxTokenAddres,fa{key,value})
	//}
}

