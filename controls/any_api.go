package controls

import (
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"net/url"
	"stock/common"
	"stock/utils"
	"log"
	"fmt"
	"sync"
	"encoding/json"
	"stock/services"
)

type AnyApiRes struct {
	//节点名
	Node string
	//数据源
	Req string
	//节点取回的数据：取回数据的类型可能为　bool／字符串／数字等；但为了统一签名处理，将取回的数据转换为string，
	Data string
	//签名：　hash= keccak256(abi.encodePacked(Req, Data).toEthSignedMessageHash()
	Sign []byte
}
// @Tags default
// @Summary　当前节点any-api
// @Description 当前节any-api
// @ID NodeAnyApiHandler
// @Accept  json
// @Produce  json
// @Param     req   query    string     true        "数据url" default(https://min-api.cryptocompare.com/data/pricemultifull?fsyms=ETH&tsyms=USD)
// @Param     path   query    string     false    "指向数据字段的json path" default(RAW.ETH.USD.VOLUME24HOUR)
// @Success 200 {object} AnyApiRes	"data"
// @Failure 500 {object} common.ApiErr "失败时，有相应测试日志输出"
// @Router /pub/stock/any_api [get]
func NodeAnyApiHandler(c *gin.Context) {
	reqUrl:=c.Query("req")
	jsonPath:=c.Query("path")
	bs, err := utils.ReqResBody(reqUrl, "", "GET", nil, nil)
	if err == nil {
		res := gjson.GetBytes(bs, jsonPath)
		ares := new(AnyApiRes)
		ares.Data = res.String()
		ares.Sign=services.GetStringsHash([][]byte{[]byte(reqUrl),[]byte(ares.Data)})
		ares.Req=reqUrl
		c.JSON(200,ares)
		return
	}
	common.ErrJson(c,err.Error())
}
// @Tags default
// @Summary　所有节点any-api
// @Description 所有节点any-api
// @ID NodeAnyApisHandler
// @Accept  json
// @Produce  json
// @Param     req   query    string     true        "数据url" default(https://min-api.cryptocompare.com/data/pricemultifull?fsyms=ETH&tsyms=USD)
// @Param     path   query    string     false    "指向数据字段的json path" default(RAW.ETH.USD.VOLUME24HOUR)
// @Success 200 {array} AnyApiRes	"any api data"
// @Failure 500 {object} common.ApiErr "失败时，有相应测试日志输出"
// @Router /pub/stock/any_apis [get]
func NodeAnyApisHandler(c *gin.Context) {
	reqUrl:=c.Query("req")
	jsonPath:=c.Query("path")

	var err error
	var addres []*AnyApiRes
	sc:=sync.RWMutex{}
	wg:= new(sync.WaitGroup)
	var porcNode=func(nodeUrl string) {
		defer wg.Done()
		reqUrl :=fmt.Sprintf( nodeUrl+"/pub/stock/any_api?req=%s&path=%s",url.QueryEscape(reqUrl),jsonPath)
		bs, err := utils.ReqResBody(reqUrl, "", "GET", nil, nil)
		if err == nil {
			stat:=new(AnyApiRes)
			err=json.Unmarshal(bs,stat)
			if err == nil {
				log.Println(err)
			}
			stat.Node=nodeUrl
			sc.Lock()
			addres=append(addres,stat)
			sc.Unlock()
		}
	}
	for _, nurl := range utils.Nodes {
		wg.Add(1)
		go porcNode(nurl)
	}
	wg.Wait()
	c.JSON(200,addres)
	return

	if err != nil {
		common.ErrJson(c,err.Error())
		return
	}
}

