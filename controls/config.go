package controls

import (
	"log"
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
	//合约安全范围
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

	tmp_KovanAddreMap:=map[string]string{}
	for _, item := range dicCfg.KovanAddreMap {
		tmp_KovanAddreMap[item.Main]=item.Kovan
	}
	KovanAddreMap=tmp_KovanAddreMap

	tmp_sp := map[string]*mm{}
	for _, item := range dicCfg.SafePrices {
		tmpMm:=&mm{item.Min,item.Max}
		tmp_sp[item.TokenAddre]=tmpMm

		kAddre:=GetKovanAddreMap(item.TokenAddre)
		if item.TokenAddre!=kAddre{
			tmp_sp[kAddre]=tmpMm
		}
	}
	safePrice=tmp_sp

	tmp_ftxAddres := map[string]string{}
	tmp_addresFtx := map[string]string{}
	for _, item := range dicCfg.FtxTokenAddres {
		tmp_ftxAddres[item.FtxName]=item.TokenAddre
		tmp_addresFtx[item.TokenAddre]=item.FtxName

		kAddre:=GetKovanAddreMap(item.TokenAddre)
		if item.TokenAddre!=kAddre{
			tmp_addresFtx[kAddre]=item.FtxName
		}
	}
	ftxAddres=tmp_ftxAddres
	addresFtx=tmp_addresFtx
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
	log.Println(KovanAddreMap)
	Config.IsDisableSpecialOpenTime=IsDisableSpecialOpenTime
	Config.IsDisableCheckFtxDataNewCheck=IsDisableCheckFtxDataNewCheck
	return Config
}