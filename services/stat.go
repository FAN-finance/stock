package services

var StatPathMap = map[string]string{
	"/pub/coin_price/:coin/:vs_coin/:timestamp":                                 "获取币价换算，多节点签名版",
	"/pub/dex/ftx_chart_prices/:coin_type/:count/:interval/:timestamp":          "获取杠杆btc代币不同时间区间的价格图表信息",
	"/pub/dex/ftx_price/:coin_type/:data_type/:timestamp":                       "获取ftx token价格信息",
	"/pub/dex/lp_price/:pair/:timestamp":                                        "获取lp价格信息",
	"/pub/dex/stock_chart_prices/:coin_type/:count/:interval/:timestamp":        "获取股票不同时间区间的价格图表信息",
	"/pub/dex/token_chain_price/:token/:data_type/:timestamp":                   "获取token链上价格信息",
	"/pub/dex/token_chart_prices/:token/:count/:interval/:timestamp":            "获取token不同时间区间的价格图表信息",
	"/pub/dex/token_day_datas/:token/:days/:timestamp":                          "获取token相应天数的统计图表信息",
	"/pub/dex/token_price/:token/:data_type/:timestamp":                         "获取token价格信息",
	"/pub/internal/coin_price/:coin/:vs_coin":                                   "获取币价换算，内部单节点",
	"/pub/internal/dex/ftx_price/:coin_type/:timestamp":                         "获取ftx coin最近一小时最高最低价格信息,内部单节点",
	"/pub/internal/dex/lp_price/:pair/:timestamp":                               "获取lp价格信息,内部单节点:",
	"/pub/internal/dex/token_chain_price/:token/:timestamp":                     "获取token最近一小时最高最低价格信息,内部单节点模式",
	"/pub/internal/dex/token_info/:token/:timestamp":                            "获取token信息,内部单节点",
	"/pub/internal/dex/token_price/:token/:timestamp":                           "获取token最近一小时最高最低价格信息,内部单节点",
	"/pub/internal/stock_avgprice":                                              "获取股票平均价格共识:",
	"/pub/internal/token_avgprice":                                              "获取token平均价格共识:",
	"/pub/stock/aggre_info/:code/:data_type/:timestamp":                         "获取共识美股价格:",
	"/pub/stock/any_api":                                                        "当前节点any-api",
	"/pub/stock/any_apis":                                                       "所有节点any-api",
	"/pub/stock/info/:code/:data_type/:timestamp":                               "获取美股价格",
	"/pub/stock/market_status/:timestamp":                                       "获取美股市场开盘状态",
	"/pub/stock/stat":                                                           "当前节点状态:记录数,钱包地址",
	"/pub/stock/stats":                                                          "所有节点状态:记录数,钱包地址",
	"/pub/dex/pair/token_chart_prices/:pair/:token/:count/:interval/:timestamp": "从Pair获取token不同时间区间的价格图表信息",
	"/pub/dex/pair/token_price/:pair/:token/:data_type/:timestamp":              "从Pair获取token价格信息",
	"/pub/internal/dex/pair/token_info/:pair/:token/:timestamp":                 "从pair获取token信息,内部单节点",
	"/pub/internal/dex/pair/token_price/:pair/:token/:timestamp":                "从Pair获取token最近一小时最高最低价格信息,内部单节点",
	"/pub/dex/token/token_chart_supply/:token/:amount/:timestamp":               "获取某个token的totalSupply的变化量",
}
var StatPath2IDMap = map[string]int{
	"/pub/coin_price/:coin/:vs_coin/:timestamp":                                 1,
	"/pub/dex/ftx_chart_prices/:coin_type/:count/:interval/:timestamp":          2,
	"/pub/dex/ftx_price/:coin_type/:data_type/:timestamp":                       3,
	"/pub/dex/lp_price/:pair/:timestamp":                                        4,
	"/pub/dex/stock_chart_prices/:coin_type/:count/:interval/:timestamp":        5,
	"/pub/dex/token_chain_price/:token/:data_type/:timestamp":                   6,
	"/pub/dex/token_chart_prices/:token/:count/:interval/:timestamp":            7,
	"/pub/dex/token_day_datas/:token/:days/:timestamp":                          8,
	"/pub/dex/token_price/:token/:data_type/:timestamp":                         9,
	"/pub/internal/coin_price/:coin/:vs_coin":                                   10,
	"/pub/internal/dex/ftx_price/:coin_type/:timestamp":                         11,
	"/pub/internal/dex/lp_price/:pair/:timestamp":                               12,
	"/pub/internal/dex/token_chain_price/:token/:timestamp":                     13,
	"/pub/internal/dex/token_info/:token/:timestamp":                            14,
	"/pub/internal/dex/token_price/:token/:timestamp":                           15,
	"/pub/internal/stock_avgprice":                                              16,
	"/pub/internal/token_avgprice":                                              17,
	"/pub/stock/aggre_info/:code/:data_type/:timestamp":                         18,
	"/pub/stock/any_api":                                                        19,
	"/pub/stock/any_apis":                                                       20,
	"/pub/stock/info/:code/:data_type/:timestamp":                               21,
	"/pub/stock/market_status/:timestamp":                                       22,
	"/pub/stock/stat":                                                           23,
	"/pub/stock/stats":                                                          24,
	"/pub/dex/pair/token_chart_prices/:pair/:token/:count/:interval/:timestamp": 25,
	"/pub/dex/pair/token_price/:pair/:token/:data_type/:timestamp":              26,
	"/pub/internal/dex/pair/token_info/:pair/:token/:timestamp":                 27,
	"/pub/internal/dex/pair/token_price/:pair/:token/:timestamp":                28,
	"/pub/dex/token/token_chart_supply/:token/:amount/:timestamp":               29,
}
var StatID2PathMap = map[int]string{ }
var StatID2ResMap = map[int]string{ }

func init(){
	for key, value := range StatPath2IDMap {
		StatID2PathMap[value]=key
		StatID2ResMap[value]=StatPathMap[key]
	}
}
