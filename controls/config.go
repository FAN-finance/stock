package controls

import (
	"stock/common"
)

var IsKovan bool
type config struct{
	//开启测试环境地址转换
	IsKovan bool
	//禁用所有签名
	IsDisableAllSign bool
	//禁用ftx签名
	IsDisableFtxSign bool
	SafePrice map[string]*mm
	FtxTokenAddres map[string]string
	//kovan地址转换
	KovanAddreMap map[string]string
	//是否禁用两时段签名功能
	IsDisableSpecialOpenTime bool
	//是否禁用新数据检查　
	IsDisableCheckFtxDataNewCheck bool
}

func BeginUseRawDicConfig(dicCfg *common.RawDicConfig){
	IsKovan=dicCfg.IsKovan
	IsDisableAllSign=dicCfg.IsDisableAllSign
	IsDisableFtxSign=dicCfg.IsDisableFtxSign
	IsDisableSpecialOpenTime=dicCfg.IsDisableSpecialOpenTime
	IsDisableCheckFtxDataNewCheck=dicCfg.IsDisableCheckFtxDataNewCheck
	for _, item := range dicCfg.KovanAddreMap {
		KovanAddreMap[item.Main]=item.Kovan
	}
	for _, item := range dicCfg.SafePrices {
		safePrice[item.TokenAddre].Max=item.Max
		safePrice[item.TokenAddre].Min=item.Min
	}
	for _, item := range dicCfg.FtxTokenAddres {
		ftxAddres[item.FtxName]=item.TokenAddre
		addresFtx[item.TokenAddre]=item.FtxName
	}
}

func GetRawDicConfig()(dicCfg *common.RawDicConfig){
	dicCfg=new(common.RawDicConfig)
	dicCfg.IsKovan=IsKovan
	dicCfg.IsDisableAllSign=IsDisableAllSign
	dicCfg.IsDisableFtxSign=IsDisableFtxSign
	//dicCfg.KovanAddreMap=KovanAddreMap
	for key, value := range KovanAddreMap {
		dicCfg.KovanAddreMap=append(dicCfg.KovanAddreMap,common.AM{key, value})
	}
	dicCfg.IsDisableSpecialOpenTime=IsDisableSpecialOpenTime
	dicCfg.IsDisableCheckFtxDataNewCheck=IsDisableCheckFtxDataNewCheck
	dicCfg.SafePrices=[]common.SP{}
	for key, item := range safePrice {
		sp:=common.SP{item.Min,item.Max,key,""}
		dicCfg.SafePrices=append(dicCfg.SafePrices,sp)
	}
	dicCfg.FtxTokenAddres=[]common.FA{}
	for key, addr := range ftxAddres {
		fa:=common.FA{key,addr}
		dicCfg.FtxTokenAddres=append(dicCfg.FtxTokenAddres,fa)
	}
	return dicCfg
}

func GetControlConfig() *config{
	var Config =new(config)
	Config.IsDisableAllSign=IsDisableAllSign
	Config.IsDisableFtxSign=IsDisableFtxSign
	Config.SafePrice=safePrice
	Config.FtxTokenAddres=ftxAddres
	Config.IsKovan=IsKovan
	Config.KovanAddreMap=KovanAddreMap
	Config.IsDisableSpecialOpenTime=IsDisableSpecialOpenTime
	Config.IsDisableCheckFtxDataNewCheck=IsDisableCheckFtxDataNewCheck
	return Config
}