package services

import (
	"github.com/stretchr/testify/assert"
	"stock/utils"
	"testing"
)

func Test_IsFtxDataOk(t *testing.T) {

	//ttime,err:=time.ParseInLocation("2006-01-02 15:04:05","2021-09-03 03:04:05",time.Local)
	//assert.Nil(t,err,"err should nil")
	//t.Log(int(time.Now().Unix()-ttime.Unix()),ttime)
	//assert.True(t, IsFtxDataNew("govt20x",int(time.Now().Unix()-ttime.Unix() )),"eth3s数据应该是新的")
	//return
	sec:=180
	assert.True(t, IsFtxDataNew("btc3x",sec),"btc3x数据应该是新的")
	assert.True(t, IsFtxDataNew("eth3s",sec),"eth3s数据应该是新的")
	assert.True(t, IsFtxDataNew("govt20x",sec),"govt20x数据应该是新的")
	assert.True(t, IsGraphEthPriceDataNew(sec),"GraphEthPrice数据应该是新的")
}

func init(){
	utils.InitDb("root:1M8x8G1S5J@tcp(62.234.169.68:3306)/stock?loc=Local&parseTime=true&multiStatements=true")
	//services.InitEConn("891eeaa3c7f945b880608e1cc9976284")
}