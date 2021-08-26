package controls

import (
	"stock/common"
)

type config struct{
	//禁用所有签名
	IsDisableAllSign bool
	//禁用ftx签名
	IsDisableFtxSign bool
	SafePrice map[string]*mm
	FtxTokenAddres map[string]string
}

func BeginUseRawDicConfig(dicCfg *common.RawDicConfig){
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

}
func GetControlConfig() *config{
	var Config =new(config)
	Config.IsDisableAllSign=IsDisableAllSign
	Config.IsDisableFtxSign=IsDisableFtxSign
	Config.SafePrice=safePrice
	Config.FtxTokenAddres=ftxAddres
	return Config
}