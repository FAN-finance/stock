package controls

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
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
var rawDic =new(rawDicConfig)
func InitDicConfig(){
	bs,err:=ioutil.ReadFile("./asset/dic_config.json")
	if err != nil {
		log.Fatal(err,"dic_config err")
	}
	err=json.Unmarshal(bs,rawDic)
	if err != nil {
		log.Fatal("dic_config err")
	}
	rawDic.BasicVerify()
	rawDic.BeginUse()
	//rawDic.IsDisableAllSign=IsDisableAllSign
	//rawDic.IsDisableFtxSign=IsDisableFtxSign
	//for key, value := range safePrice {
	//	rawDic.SafePrices=append(rawDic.SafePrices,sp{mm{value.Min,value.Max},key ,""} )
	//}
	//for key, value := range ftxAddres {
	//	rawDic.FtxTokenAddres=append(rawDic.FtxTokenAddres,fa{key,value})
	//}
}

func (dicCfg *rawDicConfig)BeginUse(){
	IsDisableAllSign=dicCfg.IsDisableAllSign
	IsDisableFtxSign=dicCfg.IsDisableFtxSign
	for _, item := range dicCfg.SafePrices {
		safePrice[item.TokenAddre].Max=item.Max
		safePrice[item.TokenAddre].Min=item.Min
	}
	for _, item := range dicCfg.FtxTokenAddres {
		ftxAddres[item.FtxName]=item.TokenAddre
		addresFtx[item.TokenAddre]=item.FtxName
	}

	Config.IsDisableAllSign=IsDisableAllSign
	Config.IsDisableFtxSign=IsDisableFtxSign
	Config.SafePrice=safePrice
	Config.FtxTokenAddres=ftxAddres
}
func (conig *rawDicConfig)BasicVerify(){
	//basic verify
	if len(rawDic.SafePrices)==0{
		log.Fatal("dic_config SafePrices err")
	}
	if len(rawDic.FtxTokenAddres)==0{
		log.Fatal("dic_config FtxTokenAddres err")
	}
}
type rawDicConfig struct{
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
	mm
	TokenAddre string
	//备注　可选项，可输入名字　
	Comment string
}

// @Tags default
// @Summary　获取字典配制信息
// @Description 字典配制信息
// @ID ConfigHandler
// @Accept  json
// @Produce  json
// @Param     is_using   query    string     false   "配制json"
// @Success 200 {object} controls.rawDicConfig	"config json"
// @Failure 500 {object} controls.ResBody "失败时，有相应测试日志输出"
// @Router /pub/dic_config [get]
func ConfigHandler(c *gin.Context) {
	if c.Query("is_using")==""{
		NewResBody(c,rawDic)
	}else{
		NewResBody(c,Config)
	}
}

// @Tags default
// @Summary　修改字典配制信息
// @Description 修改字典配制信息
// @ID ConfigHandler
// @Accept  json
// @Produce  json
// @Param     dic   body    controls.rawDicConfig     true    "配制json"
// @Success 200 {object} AnyApiRes	"data"
// @Success 200 {object} controls.rawDicConfig	"config json"
// @Failure 500 {object} controls.ResBody "失败时，有相应测试日志输出"
// @Router /pub/dic_config [post]
func ConfigUpdateHandler(c *gin.Context) {
	if !IsAdmin(c){
		ResErrMsg(c,"you are not administrator")
		return
	}
	err:=c.ShouldBindBodyWith(rawDic,binding.JSON)
	c.GetRawData()
	if err != nil {
		ResErrMsg(c,err.Error())
		return
	}
	rawDic.BasicVerify()
	rawBs,_:=c.Get(gin.BodyBytesKey)
	err=ioutil.WriteFile("./asset/dic_config1.json",rawBs.([]byte),0644)
	if err != nil {
		ResErrMsg(c,err.Error())
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
			ResErrMsg(c,err.Error())
			return
		}
	}else{
		log.Println("not exist  sync-dic sh")
	}

	rawDic.BeginUse()
	c.JSON(200, rawDic)
}