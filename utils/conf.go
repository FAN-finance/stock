package utils

import "sync"

//节点列表
var Nodes []string
//钱包地址白名单列表
var wList=new(sync.Map)
func SetWList(list []string) {
 for _, value := range list {
  wList.Store(value, true)
 }
}
func IsInWL(addre string)bool{
 _,ok:=wList.Load(addre)
 return ok
}

//twelvedata数据源的api-key,需要自行去https://twelvedata.com/申请
var TwKey string
type PubConf struct {
 //管理节点地址
 AdminAddr string
}
var InfuraID string
var NodeTwKeyMap=map[string]string{
"node0":"4e8a6b8b4afe47be815d9e3b4d8cf163",
"node1":"21cad25580b74ba3a0a2ba9be29057bb",
"node2":"bbc77d57030d48268f764e6a4c2c5bed",
}
