package sys

import (
	"encoding/json"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"stock/common"
	"stock/controls"
)

// @Tags sys
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
		//c.JSON(200,RawDic)
		c.JSON(200,controls.GetRawDicConfig())
	}else{
		c.JSON(200,controls.GetControlConfig())
	}
}

// @Tags sys
// @Summary　修改字典配制信息
// @Description 修改字典配制信息
// @ID ConfigUpdateHandler
// @Accept  json
// @Produce  json
// @Param     Authorization   header    string     true    "token"
// @Param     dic   body    common.RawDicConfig     true    "配制json"
// @Success 200 {object} common.RawDicConfig	"data"
// @Failure 500 {object} common.ResBody "失败时，有相应测试日志输出"
// @Router /sys/dic_config [post]
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
	//RawDic=tDicConf
	//controls.BeginUseRawDicConfig(RawDic)
	c.JSON(200, tDicConf)
}




var dicFile="./asset/dic_config.json"
func InitDicConfig(){
	var rawDic =new(common.RawDicConfig)
	bs,err:=ioutil.ReadFile(dicFile)
	if err != nil {
		log.Fatal(err,"dic_config err")
	}
	err=json.Unmarshal(bs, rawDic)
	if err != nil {
		log.Fatal("dic_config err")
	}
	rawDic.BasicVerify()
	controls.BeginUseRawDicConfig(rawDic)

}
func InitDicConfigWatch(){
	go WatchDicFile(dicFile)
}
func WatchDicFile(fpath string){
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	done := make(chan bool)
	go func() {
		defer close(done)
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					log.Println("WatchDicFile exit",event)
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event)
					InitDicConfig()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					log.Println("error exit:", err)
					return
				}
				log.Println("error:", err)
			}
		}
	}()
	log.Println("add watch file ",fpath)

	err = watcher.Add(fpath)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

