package services

import "stock/utils"

var StatPathMap = map[string]string{
	"other":"其它",
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
	StatPath2IDMap["other"]=100
	for key, value := range StatPath2IDMap {
		StatID2PathMap[value]=key
		StatID2ResMap[value]=StatPathMap[key]
	}
}

func ApiStat()(map[string]interface{},error) {
	res := map[string]interface{}{}
	allStat := map[string]interface{}{}
	err := utils.Orm.Raw(
		`select substr(from_unixtime(min(timestamp)),1,10) begin_date,
       sum(stats.counter)                                                   counter
from api_stats stats
where stats.is_internal = 0`).First(&allStat).Error
	res["allStat"] = allStat

	type apiRankStat struct{
		PathId int
		Counter int
		PathName string
	}
	apiRankStats := []*apiRankStat{}
	err = utils.Orm.Raw(
		`select stats.path_id,
       sum(stats.counter) counter
from api_stats stats
 where stats.is_internal=0 and stats.timestamp>truncate(unix_timestamp() / (3600 * 24), 0) * 3600 * 24- (3600 * 24)*30
group by stats.path_id  order by counter desc;`).
		Find(&apiRankStats).Error
	if err == nil {
		for _, item := range apiRankStats {
			item.PathName=StatID2ResMap[item.PathId]
		}
		res["apiRankStats"] = apiRankStats
	}

	day30Stat := []map[string]interface{}{}
	err = utils.Orm.Raw(
		`select substr(from_unixtime(aa.timespan),1,10) daystr,
       sum((case when is_internal = 1 then 0  else aa.counter end)) as oss,
       sum((case when is_internal = 1 then aa.counter else 0 end)) as internal
from (
         select stats.is_internal,
                (truncate(timestamp / (3600 * 24), 0) * 3600 * 24) as timespan,
                sum(stats.counter)                                                 counter
         from api_stats stats
#          where stats.is_internal = 0
         where stats.timestamp>truncate(unix_timestamp() / (3600 * 24), 0) * 3600 * 24- (3600 * 24)*30
         group by stats.is_internal, timespan
     ) aa
group by daystr`).
		Find(&day30Stat).Error
	if err == nil {
		res["day30Stat"] = day30Stat
	}

	day30HourRankStat := []map[string]interface{}{}
	err = utils.Orm.Raw(
		`select concat( hour(from_unixtime(truncate(timestamp / (3600), 0) * 3600)),'点') as hour_name,
#                 stats.is_internal,
#                 hour(date_add(
#                         FROM_UNIXTIME(0), interval
#                         (truncate(timestamp / (3600), 0) * 3600) +
#                         TIMESTAMPDIFF(SECOND, NOW(), UTC_TIMESTAMP()) SECOND
#                     ))             timespan,
       sum(stats.counter)                                             counter
from api_stats stats
where stats.is_internal = 0
  and stats.timestamp > truncate(unix_timestamp() / (3600 * 24), 0) * 3600 * 24 - (3600 * 24) * 30
group by hour_name
order by counter desc`).
		Find(&day30HourRankStat).Error
	if err == nil {
		res["day30HourRankStat"] = day30HourRankStat
	}

	last48HourStat:= []map[string]interface{}{}
	err = utils.Orm.Raw(
		`select from_unixtime(truncate(timestamp / (3600), 0) * 3600) as datestr,
sum(stats.counter)                                             counter
from api_stats stats
where stats.is_internal = 0
and stats.timestamp > truncate(unix_timestamp() / (3600), 0) * 3600 - 3600 * 48
group by datestr;`).
		Find(&last48HourStat).Error
	if err == nil {
		res["last48HourStat"] = last48HourStat
	}

	last48Hour10MinStat:= []map[string]interface{}{}
	err = utils.Orm.Raw(
		`select from_unixtime(truncate(timestamp / (600), 0) * 600) as datestr,
sum(stats.counter)                                             counter
from api_stats stats
where stats.is_internal = 0
and stats.timestamp > truncate(unix_timestamp() / (600), 0) * 600 - 3600 * 48
group by datestr`).
		Find(&last48Hour10MinStat).Error
	if err == nil {
		res["last48Hour10MinStat"] = last48Hour10MinStat
	}
	return res, err
}

/*

-- 总访问次数
select substr(from_unixtime(min(timestamp)),1,10) begin_date,
       sum(stats.counter)                                                   counter
from api_stats stats
where stats.is_internal = 0;


-- 最近30天访问最多的api
select stats.path_id,
       sum(stats.counter) counter
from api_stats stats
 where stats.is_internal=0 and stats.timestamp>truncate(unix_timestamp() / (3600 * 24), 0) * 3600 * 24- (3600 * 24)*30
group by stats.path_id  order by counter desc;



-- 最近30天每天访问次数统计
select substr(from_unixtime(aa.timespan),1,10) daystr,
       sum((case when is_internal = 1 then 0  else aa.counter end)) as oss,
       sum((case when is_internal = 1 then aa.counter else 0 end)) as internal
from (
         select stats.is_internal,
                (truncate(timestamp / (3600 * 24), 0) * 3600 * 24) as timespan,
                sum(stats.counter)                                                 counter
         from api_stats stats
#          where stats.is_internal = 0
         where stats.timestamp>truncate(unix_timestamp() / (3600 * 24), 0) * 3600 * 24- (3600 * 24)*30
         group by stats.is_internal, timespan
     ) aa
group by daystr;


-- 最近30天小时次数排名.
select concat( hour(from_unixtime(truncate(timestamp / (3600), 0) * 3600)),'点') as hour_name,
#                 stats.is_internal,
#                 hour(date_add(
#                         FROM_UNIXTIME(0), interval
#                         (truncate(timestamp / (3600), 0) * 3600) +
#                         TIMESTAMPDIFF(SECOND, NOW(), UTC_TIMESTAMP()) SECOND
#                     ))             timespan,
       sum(stats.counter)                                             counter
from api_stats stats
where stats.is_internal = 0
  and stats.timestamp > truncate(unix_timestamp() / (3600 * 24), 0) * 3600 * 24 - (3600 * 24) * 30
group by hour_name
order by counter desc


-- 最近48小时每小时访问量
select from_unixtime(truncate(timestamp / (3600), 0) * 3600) as datestr,
sum(stats.counter)                                             counter
from api_stats stats
where stats.is_internal = 0
and stats.timestamp > truncate(unix_timestamp() / (3600), 0) * 3600 - 3600 * 48
group by datestr;
# order by counter desc;


-- 最近48小时每10分钟访问量
select from_unixtime(truncate(timestamp / (600), 0) * 600) as datestr,
sum(stats.counter)                                             counter
from api_stats stats
where stats.is_internal = 0
and stats.timestamp > truncate(unix_timestamp() / (600), 0) * 600 - 3600 * 48
group by datestr
*/


