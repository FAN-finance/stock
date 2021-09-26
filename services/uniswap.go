package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"log"
	"stock/services/uni"
	"stock/utils"
	"strconv"
	"time"
)

type TokenDayData struct {
	Typename         string `json:"__typename"`
	DailyVolumeETH   string `json:"dailyVolumeETH"`
	DailyVolumeToken string `json:"dailyVolumeToken"`
	//每天成交量
	DailyVolumeUSD string `json:"dailyVolumeUSD"`
	Date           int64  `json:"date"`
	ID             string `json:"id"`
	//每天最后一次价格
	PriceUSD            string `json:"priceUSD"`
	TotalLiquidityETH   string `json:"totalLiquidityETH"`
	TotalLiquidityToken string `json:"totalLiquidityToken"`
	TotalLiquidityUSD   string `json:"totalLiquidityUSD"`
}

var SwapBlockGraphApi = "https://api.thegraph.com/subgraphs/name/blocklytics/ethereum-blocks"
var blockGraph = `{"operationName":"blocks","variables":{},"query":"query blocks {\n  t1624159500: blocks(first: 1, orderBy: timestamp, orderDirection: desc, where: {timestamp_gt: 1624159500, timestamp_lt: 1624160100}) {\n    number\n    __typename\n  }\n  t1624073100: blocks(first: 1, orderBy: timestamp, orderDirection: desc, where: {timestamp_gt: 1624073100, timestamp_lt: 1624073700}) {\n    number\n    __typename\n  }\n  t1623641100: blocks(first: 1, orderBy: timestamp, orderDirection: desc, where: {timestamp_gt: 1623641100, timestamp_lt: 1623641700}) {\n    number\n    __typename\n  }\n}\n"}`


var SwapGraphApi = "https://api.thegraph.com/subgraphs/name/ianlapham/uniswapv2" //"https://api.thegraph.com/subgraphs/name/uniswap/uniswap-v2"
var tokenDayDataGraph = `{"operationName":"tokenDayDatas","variables":{"tokenAddr":"%s","skip":0},"query":"query tokenDayDatas($tokenAddr: String\u0021, $skip: Int\u0021) {\n  tokenDayDatas(first: %d, skip: $skip, orderBy: date, orderDirection: desc, where: {token: $tokenAddr}) {\n    id\n    date\n    priceUSD\n    totalLiquidityToken\n    totalLiquidityUSD\n    totalLiquidityETH\n    dailyVolumeETH\n    dailyVolumeToken\n    dailyVolumeUSD\n    __typename\n  }\n}\n"}`

func ReqSwapGraph(body string) ([]byte, error) {
	bs, err := utils.ReqResBody(SwapGraphApi, "", "POST", nil, []byte(body))
	if err != nil {
		return bs, err
	}
	if bytes.HasPrefix(bs, []byte(`{"errors"`)) {
		serr:=new(subGrahpErr)
		err=json.Unmarshal(bs,serr)
		if err == nil {
			errmsg:=""
			for _, item := range serr.Errors {
				errmsg+=item.Message+";"
			}
			err=errors.New(errmsg)
		}
	}
	return bs, err
}

type subGrahpErr struct {
	Errors []struct {
		Message string `json:"message" example:"Failed to decode block.number value: subgraph Qmc7K8dKoadu1VcHfAV45pN4sPnwZcU2okV6cuU4B7qQp1 has only indexed up to block number 13015747 and data for block number 123412341 is therefore not yet available"`
	} `json:"errors"`
}

func GetTokenDayData(tokenAddre string, days int) ([]byte, error) {
	bs, err := ReqSwapGraph(fmt.Sprintf(tokenDayDataGraph, tokenAddre, days))
	if err == nil {
		//使其直接返回字符串
		bs = bytes.TrimPrefix(bs, []byte(`{"data":{"tokenDayDatas":`))
		bs = bytes.TrimSuffix(bs, []byte(`}}`))
		//log.Println(string(bs))
	}
	return bs, err
}

func getTokenTimes(interval string, count int) []int64 {
	if count == 0 {
		count = 10
	}
	//now := time.Now().UTC().Add(-10 * time.Minute).Truncate(time.Minute)
	now := time.Now().UTC().Add(0 * time.Minute).Truncate(time.Minute)
	span := time.Hour
	switch interval {
	case "60s":
		span = time.Second * 60
	case "15minite":
		span = time.Minute * 15
	case "hour":
		span = time.Hour
	case "day":
		span = time.Hour * 24
	case "1w":
		span = time.Hour * 24 * 7
	case "1m":
		span = time.Hour * 24 * 30
	default:
		log.Println("err interval; valid interval: 15minite hour day 1w 1m ", interval)
	}
	timeItems := []int64{}
	ttime := now.Truncate(span)

	for i := count - 1; i >= -1; i-- {
		item := ttime.Add(-1 * span * time.Duration(i))
		timeItems = append(timeItems, item.Unix())
	}

	return timeItems
}

func getBlockPrices(timeItems []int64, maxBlockHeight int64) ([]*uni.BlockPrice, error) {
	bps := []*uni.BlockPrice{}

	sql := ""
	args := []interface{}{}
	for idx, item := range timeItems {
		sql += fmt.Sprintf("(select id,price, %d as block_time1 from block_prices where block_time< ? and id < %d order by block_time desc limit 1)", item, maxBlockHeight+1)
		if idx < len(timeItems)-1 {
			sql += " union all"
		}
		args = append(args, item)
	}
	sql = fmt.Sprintf("select  id,price, block_time1 as block_time from (%s) aa;", sql)
	err := utils.Orm.Raw(sql, args...).Scan(&bps).Error

	//for  _,item := range timeItems {
	//	bp:=new(BlockPrice)
	//	err:=utils.Orm.Raw("select id,price from block_prices where block_time< ? order by block_time desc limit 1;",item).First(bp).Error
	//	if err != nil {
	//		log.Println(err)
	//		continue
	//	}
	//	bp.BlockTime=uint64(item)
	//	bps=append(bps,bp)
	//}
	return bps, err
}

const getBlockHeight = `{"query":"{\n  _meta \n  \t{block\n      {number}\n  \t}\n}\n","variables":null}`

func GetTokenTimesPrice(tokenAddre string, interval string, count int) ([]*uni.BlockPrice, error) {
	//{"operationName":"blocks","variables":{},"query":"query blocks {
	//	\nt1620871200: token(id: \"0xbc396689893d065f41bc2c6ecbee5e0085233447\", block: {number: 12423239}) {\n    derivedETH\n    __typename\n  }\n
	//	\n}\n"}
	//
	//var times []int64
	//times=[]int64{12427306,12429525}
	times := getTokenTimes(interval, count)
	//log.Println(times)
	body, reqerr := ReqSwapGraph(getBlockHeight)
	if reqerr != nil {
		return nil, reqerr
	}
	result := gjson.Parse(string(body))
	blockHeight := result.Get("data").Get("_meta").Get("block").Get("number").Int()
	bps, err := getBlockPrices(times, blockHeight)
	if err == nil {
		gql := `{"operationName":"blocks","variables":{},"query":"query blocks {`
		for _, item := range bps {
			gql += fmt.Sprintf(`\nt%d: token(id: \"%s\", block: {number: %d}) {\n    derivedETH\n  }`, item.BlockTime, tokenAddre, item.ID)
		}
		gql += `\n}\n"}`
		bs, err1 := ReqSwapGraph(gql)
		err = err1
		if err == nil {
			//使其直接返回字符串
			bs = bytes.TrimPrefix(bs, []byte(`{"data":`))
			bs = bytes.TrimSuffix(bs, []byte(`}`))

			res := map[string]struct {
				DerivedETH string `json:"derivedETH"`
			}{}

			err = json.Unmarshal(bs, &res)
			if err == nil {
				//log.Println(res)
				if len(res) == 0 {
					return nil, errors.New("没找到任何数据，稍后重试")
				}
				preTime := uint64(0)
				for idx, item := range bps {
					key := fmt.Sprintf("t%d", item.BlockTime)
					resItem, ok := res[key]
					if ok {
						/*ethValue, _ := strconv.ParseFloat(resItem.DerivedETH, 64)
						log.Println(ethValue);*/
						ethPrice, _ := decimal.NewFromString(resItem.DerivedETH)
						if interval == "60s" && idx == (len(bps)-1) {
							item.Price = uni.BlockPrice{}.GetPrice()
						}
						price, _ := ethPrice.Mul(decimal.NewFromFloat(item.Price)).Float64()
						item.Price = RoundPrice(price)
						//log.Println("key", key, item.Price,ethValue)
					} else {
						if idx > 0 {
							item.Price = bps[idx-1].Price
						}
					}
					if idx > 0 {
					}
					if len(bps) > 1 {
						preTime, item.BlockTime = item.BlockTime, preTime //bps[idx-1].BlockTime
					}
				}
				if len(bps) > 1 {
					bps = bps[1:len(bps)]
				}
			}
		}
		if err != nil {
			log.Println(string(bs))
		}
		//log.Println(gql ,string(bs),err)
	}
	if err != nil {
		log.Println(err)
	}
	return bps, err
}

func GetTokenTimesPriceFromPair(pairAddr, tokenAddr string, interval string, count int) ([]*uni.BlockPrice, error) {
	times := getTokenTimes(interval, count)
	//log.Println(times)
	body, reqErr := ReqSwapGraph(getBlockHeight)
	if reqErr != nil {
		return nil, reqErr
	}
	result := gjson.Parse(string(body))
	blockHeight := result.Get("data").Get("_meta").Get("block").Get("number").Int()
	bps, err := getBlockPrices(times, blockHeight)

	//GetTokenTimesPriceFromPair 不需要以太坊价格
	for _, bp := range bps {
		bp.Price = 0
	}
	if err == nil {
		gql := `{"operationName":"blocks","variables":{},"query":"query blocks {`
		for _, item := range bps {
			gql += fmt.Sprintf(`\nt%d: pair(id: \"%s\", block: {number: %d}) { \n    \ttoken0{id},\n     token1{id},\n     token0Price,\n     token1Price }`, item.BlockTime, pairAddr, item.ID)
		}
		gql += `\n}\n"}`

		bs, err1 := ReqSwapGraph(gql)
		//log.Println("GetTokenTimesPriceFromPair",gql)
		err = err1
		if err == nil {
			//使其直接返回字符串
			dataJson := gjson.ParseBytes(bs).Get("data")
			if !dataJson.Exists() {
				err = errors.New("data err")
				return nil, err
			}
			if dataJson.Exists() {
				//log.Println(res)
				key := fmt.Sprintf("t%d", bps[len(bps)-1].BlockTime)
				resItem := dataJson.Get(key)
				if resItem.Exists() == false {
					return nil, errors.New("no data from thegraph")
				}
				token0Address := resItem.Get("token0").Get("id").Str
				token1Address := resItem.Get("token1").Get("id").Str
				if tokenAddr != token0Address && tokenAddr != token1Address {
					return nil, errors.New("the token address not found in pair")
				}

				preTime := uint64(0)
				for idx, item := range bps {
					key = fmt.Sprintf("t%d", item.BlockTime)
					resItem = dataJson.Get(key)
					if resItem.Exists() {

						item.Price = RoundPrice(resItem.Get("token0Price").Float())
						if token0Address == tokenAddr {
							item.Price = RoundPrice(resItem.Get("token1Price").Float())
						}
						item.CreatedAt = time.Unix(int64(preTime), 0)
					} else {
						if idx > 0 {
							item.Price = bps[idx-1].Price
						}
					}
					if len(bps) > 1 {
						preTime, item.BlockTime = item.BlockTime, preTime //bps[idx-1].BlockTime
					}
				}
				if len(bps) > 1 {
					bps = bps[1:len(bps)]
				}
			}
		}
		if err != nil {
			log.Println(string(bs))
		}
	}
	if err != nil {
		log.Println(err)
	}
	return bps, err
}

func RoundPrice(price float64) float64 {
	res, _ := decimal.NewFromFloat(price).Round(18).Float64()
	return res
}

type LpSnapPairInfo struct {
	Block int64 `json:"block"`
	//lp数量
	LiquidityTokenTotalSupply string `json:"liquidityTokenTotalSupply"`
	Reserve0                  string `json:"reserve0"`
	Reserve1                  string `json:"reserve1"`
	//pair市值
	ReserveUSD     string `json:"reserveUSD"`
	Timestamp      int64  `json:"timestamp"`
	Token0PriceUSD string `json:"token0PriceUSD"`
	Token1PriceUSD string `json:"token1PriceUSD"`
}

var infoLpDataGraph = `{"query":"{\n  liquidityPositionSnapshots(first:1,orderBy:timestamp,orderDirection:desc where:{pair:\"%s\"}) {\n  # id\n  timestamp\n  block\n  token0PriceUSD\n  token1PriceUSD\n  reserve0\n  reserve1\n  reserveUSD\n  liquidityTokenTotalSupply\n  # liquidityTokenBalance\n  }\n}","variables":null}`

func GetLpPairInfo(pairAddre string) (pair *LpSnapPairInfo, err error) {
	bs, err := ReqSwapGraph(fmt.Sprintf(infoLpDataGraph, pairAddre))
	if err == nil {
		//使其直接返回字符串
		bs = bytes.TrimPrefix(bs, []byte(`{"data":{"liquidityPositionSnapshots":`))
		bs = bytes.TrimSuffix(bs, []byte(`}}`))
		//log.Println(string(bs))
		ps := []*LpSnapPairInfo{}
		err = json.Unmarshal(bs, &ps)
		if err == nil {
			if len(ps) == 0 {
				return nil, errors.New("pair not found")
			}
			return ps[0], err
		}
	}
	return nil, err
}

type TokenInfo struct {
	Decimals           string `json:"decimals"`
	DerivedETH         string `json:"derivedETH"`
	PriceUsd           float64
	Name               string  `json:"name"`
	Symbol             string  `json:"symbol"`
	ReserveUSD         float64 `json:"reserveUSD"`
	TotalLiquidity     string  `json:"totalLiquidity"`
	TotalSupply        string  `json:"totalSupply"`
	TradeVolume        string  `json:"tradeVolume"`
	TradeVolumeUSD     string  `json:"tradeVolumeUSD"`
	TxCount            string  `json:"txCount"`
	UntrackedVolumeUSD string  `json:"untrackedVolumeUSD"`
	BlockTime          uint64  `json:",omitempty"`
}
type OneDayStat struct {
	VolumeUsd float64 `json:"VolumeUsd"`
	//变化百分比
	VolumeChange float64 `json:"volumeChange"`

	LiquidityUsd float64 `json:"liquidityUsd"`
	//变化百分比
	LiquidityChange float64 `json:"liquidityChange"`

	TxCount float64 `json:"txCount"`
	//变化百分比
	TxCountChange float64 `json:"txCountChange"`

	PriceChange float64 `json:"priceChange"`
}

//TotalLiquidity TradeVolumeUSD  TxCount PriceUsd
func GetTokenInfosForStat(tokenAddre string, ethPrice float64) (OneDayStat, error) {
	//{"operationName":"blocks","variables":{},"query":"query blocks {
	//	\nt1620871200: token(id: \"0xbc396689893d065f41bc2c6ecbee5e0085233447\", block: {number: 12423239}) {\n    derivedETH\n    __typename\n  }\n
	//	\n}\n"}
	//
	var times []int64
	now := time.Now().UTC().Truncate(-3 * time.Minute)
	oneDay := now.Add(time.Hour * -24)
	twoDay := now.Add(time.Hour * -48)
	times = []int64{twoDay.Unix(), oneDay.Unix(), now.Unix()}
	//times=[]int64{12427306,12429525}
	//times = getTokenTimes(interval, count)
	body, err := ReqSwapGraph(getBlockHeight)
	if err == nil {
		result := gjson.Parse(string(body))
		blockHeight := result.Get("data").Get("_meta").Get("block").Get("number").Int()
		bps, err1 := getBlockPrices(times, blockHeight)
		err = err1
		log.Println(times)
		if err == nil {
			gql := `{"operationName":"blocks","variables":{},"query":"query blocks {`
			for _, item := range bps {
				gql += fmt.Sprintf(`\nt%d: token(id: \"%s\", block: {number: %d}) {\nsymbol\nname\ndecimals\ntotalSupply\ntradeVolume\ntradeVolumeUSD\nuntrackedVolumeUSD\ntxCount\ntotalLiquidity\nderivedETH\n\n}`, item.BlockTime, tokenAddre, item.ID)
			}
			gql += `\n}\n"}`

			bs, err1 := ReqSwapGraph(gql)
			err = err1
			if err == nil {
				//使其直接返回字符串
				bs = bytes.TrimPrefix(bs, []byte(`{"data":`))
				bs = bytes.TrimSuffix(bs, []byte(`}`))
				res := map[string]*TokenInfo{}
				err = json.Unmarshal(bs, &res)
				if err == nil {
					//log.Println(res)
					tokens := []*TokenInfo{}
					for _, item := range bps {
						key := fmt.Sprintf("t%d", item.BlockTime)
						log.Println("key", key, item.ID)
						resItem, ok := res[key]
						if ok {
							tokens = append(tokens, resItem)
							resItem.BlockTime = item.BlockTime
						}
					}
					log.Println("GetTokenInfosForStat got tokens", len(tokens))
					if len(tokens) < 3 {
						for i := 0; i < 3-len(tokens); i++ {
							tokens = append(tokens, tokens[len(tokens)-1])
						}
					}
					ost := new(OneDayStat)
					ost.LiquidityUsd, ost.LiquidityChange = get2DayPercentChangeFloat(parseFloat(tokens[2].TotalLiquidity)*bps[2].Price, parseFloat(tokens[1].TotalLiquidity)*bps[1].Price, parseFloat(tokens[0].TotalLiquidity)*bps[0].Price)
					//ost.LiquidityUsd=RoundPrice(ost.LiquidityUsd*ethPrice)

					ost.VolumeUsd, ost.VolumeChange = get2DayPercentChange(tokens[2].TradeVolumeUSD, tokens[1].TradeVolumeUSD, tokens[0].TradeVolumeUSD)

					ost.TxCount, ost.TxCountChange = get2DayPercentChange(tokens[2].TxCount, tokens[1].TxCount, tokens[0].TxCount)

					ost.PriceChange = getPercentChange(parseFloat(tokens[2].DerivedETH)*bps[2].Price, parseFloat(tokens[1].DerivedETH)*bps[1].Price)

					log.Println(*ost)
					return *ost, nil
					//ethValue, _ := strconv.ParseFloat(resItem.DerivedETH, 64)
					//item.Price = RoundPrice(ethValue * item.Price)
					//TotalLiquidity TradeVolumeUSD  TxCount PriceChange
				}
			}
			//log.Println(gql ,string(bs),err)
		}
	}
	if err != nil {
		log.Println(err)
	}
	return OneDayStat{}, err
}
func parseFloat(fstr string) float64 {
	fvalue, _ := strconv.ParseFloat(fstr, 64)
	return fvalue
}
func get2DayPercentChangeFloat(valueNow, value24HoursAgo, value48HoursAgo float64) (float64, float64) {
	log.Println(valueNow, value24HoursAgo, value48HoursAgo)
	// get volume info for both 24 hour periods
	var currentChange = valueNow - (value24HoursAgo)
	var previousChange = (value24HoursAgo) - (value48HoursAgo)
	if previousChange == 0 {
		return currentChange, 0
	}
	var adjustedPercentChange = (currentChange - previousChange) / previousChange * 100
	//	if (isNaN(adjustedPercentChange) || !isFinite(adjustedPercentChange)) {
	//		return [currentChange, 0]
	//}
	return currentChange, adjustedPercentChange
}
func get2DayPercentChange(valueNow, value24HoursAgo, value48HoursAgo string) (float64, float64) {
	log.Println(valueNow, value24HoursAgo, value48HoursAgo)
	// get volume info for both 24 hour periods
	var currentChange = parseFloat(valueNow) - parseFloat(value24HoursAgo)
	var previousChange = parseFloat(value24HoursAgo) - parseFloat(value48HoursAgo)
	if previousChange == 0 {
		return currentChange, 0
	}
	var adjustedPercentChange = (currentChange - previousChange) / previousChange * 100
	//	if (isNaN(adjustedPercentChange) || !isFinite(adjustedPercentChange)) {
	//		return [currentChange, 0]
	//}
	return currentChange, adjustedPercentChange
}
func getPercentChange(valueNow, value24HoursAgo float64) float64 {
	log.Println("getPercentChange", valueNow, value24HoursAgo)
	// get volume info for both 24 hour periods
	//var currentChange = parseFloat(valueNow)
	//var previousChange = parseFloat(value24HoursAgo)
	var currentChange = valueNow
	var previousChange = value24HoursAgo
	if previousChange == 0 {
		return 0
	}
	var adjustedPercentChange = (currentChange - previousChange) / previousChange * 100
	//	if (isNaN(adjustedPercentChange) || !isFinite(adjustedPercentChange)) {
	//		return [currentChange, 0]
	//}
	return adjustedPercentChange
}

type CoinPriceView struct {
	Price     float64
	BigPrice  string
	Node      string
	Timestamp int64
	//目标币价
	Coin string
	//vs币价
	VsCoin string
	//Sign_Hash值由 Timestamp+Coin+VsCoin+BigPrice计算
	Sign []byte
	//Signs []*PriceView `json:",omitempty"`
}
type DataCoinPriceView struct {
	Price     float64
	BigPrice  string
	Timestamp int64
	//目标币价
	Coin string
	//vs币价
	VsCoin string
	//Sign_Hash值由 Timestamp+Coin+VsCoin+BigPricey计算
	Sign []byte
	//所有节点签名数据
	Signs []*CoinPriceView `json:",omitempty"`
}

type PriceView struct {
	//合约代码
	Code        string
	PriceUsd    float64
	BigPrice    string
	Node        string
	NodeAddress string
	Timestamp   int64
	//Sign_Hash值由 Timestamp,Code,BigPrice
	Sign []byte
	//Signs []*PriceView `json:",omitempty"`

	//debug msg
	Msg string `json:"msg,omitempty"`
}
type DataPriceView struct {
	//合约代码
	Code     string
	PriceUsd float64
	BigPrice string
	//Node string
	Timestamp int64
	//Sign_Hash值由 Timestamp,Code,BigPrice
	Sign  []byte
	Signs []*PriceView `json:",omitempty"`
}
type HLPriceView struct {
	PriceView
	//最高最低价１最高　２最低价
	DataType int
	//Sign_Hash值由 Timestamp，DataType,Code,BigPrice
	Sign []byte

	//debug msg
	Msg string `json:"msg,omitempty"`
}

//同HLPriceView ,但使用不一样的签名字段顺序
type HLPriceViewRaw struct {
	HLPriceView
}
type HLDataPriceView struct {
	DataPriceView
	//最高最低价１最高　２最低价
	DataType int
	//Sign_Hash值由 Timestamp，DataType,Code,BigPrice
	Sign     []byte         `json:""`
	Signs    []*HLPriceView `json:",omitempty"`
	AvgSigns []*HLPriceView `json:",omitempty"`

	//股票杠杆币需要这个字段
	IsMarketOpening bool
	//股票杠杆币需要这个字段
	MarketOpenTime int64
}

func (hldv *HLDataPriceView) Clean() {
	for _, p := range hldv.Signs {
		p.Sign = nil
	}
}
func (sd *StockData) Clean() {
	for _, p := range sd.Signs {
		p.Sign = nil
	}
}

var infoTokenGraph = `{"query":"{\n  token(id:\"%s\") {\nsymbol\nname\ndecimals\ntotalSupply\ntradeVolume\ntradeVolumeUSD\nuntrackedVolumeUSD\ntxCount\ntotalLiquidity\nderivedETH\n\n  }\n}","variables":null}`

func GetTokenInfo(pairAddre string) (token *TokenInfo, err error) {
	bs, err := ReqSwapGraph(fmt.Sprintf(infoTokenGraph, pairAddre))
	if err == nil {
		//使其直接返回字符串
		bs = bytes.TrimPrefix(bs, []byte(`{"data":{"token":`))
		bs = bytes.TrimSuffix(bs, []byte(`}}`))
		log.Println("GetTokenInfo", string(bs))
		token = &TokenInfo{}
		err = json.Unmarshal(bs, &token)
		if err == nil {
			return token, err
		}
	}
	return nil, err
}

const pairInfoGraph = `{"query":"{\n  pair(id:\"%s\") {\n    id,\n reserveUSD, \n    token0Price,\n    token1Price,\n    token0{id,symbol,name,decimals,totalSupply,tradeVolume,tradeVolumeUSD,untrackedVolumeUSD,txCount,totalLiquidity,derivedETH},\n    token1{id,symbol,name,decimals,totalSupply,tradeVolume,tradeVolumeUSD,untrackedVolumeUSD,txCount,totalLiquidity,derivedETH}\n  \n  }\n}\n","variables":null}`

func GetTokenInfoFromPair(pairAddr, tokenAddr string) (token *TokenInfo, err error) {
	bs, err := ReqSwapGraph(fmt.Sprintf(pairInfoGraph, pairAddr))
	if err == nil {
		//使其直接返回字符串
		pairInfo := gjson.ParseBytes(bs).Get("data").Get("pair")
		token = &TokenInfo{}
		token0Address := pairInfo.Get("token0").Get("id").Str
		token1Address := pairInfo.Get("token1").Get("id").Str
		if token0Address == tokenAddr || tokenAddr == token1Address {
			token.PriceUsd = pairInfo.Get("token0Price").Float()
			tokenJson := pairInfo.Get("token1")
			if tokenAddr == token0Address {
				token.PriceUsd = pairInfo.Get("token1Price").Float()
				tokenJson = pairInfo.Get("token0")
			}
			err = json.Unmarshal([]byte(tokenJson.String()), &token)
			token.ReserveUSD = pairInfo.Get("reserveUSD").Float()

			if err == nil {
				return token, err
			}
		} else {
			err = errors.New("the token address not found in pair")
		}

	}
	return nil, err
}

type PairInfo struct {
	CreatedAtBlockNumber   string `json:"createdAtBlockNumber"`
	CreatedAtTimestamp     string `json:"createdAtTimestamp"`
	LiquidityProviderCount string `json:"liquidityProviderCount"`
	Reserve0               string `json:"reserve0"`
	Reserve1               string `json:"reserve1"`
	ReserveETH             string `json:"reserveETH"`
	ReserveUSD             string `json:"reserveUSD"`
	Token0                 struct {
		ID string `json:"id"`
	} `json:"token0"`
	Token0Price string `json:"token0Price"`
	Token1      struct {
		ID string `json:"id"`
	} `json:"token1"`
	Token1Price        string `json:"token1Price"`
	TotalSupply        string `json:"totalSupply"`
	TrackedReserveETH  string `json:"trackedReserveETH"`
	TxCount            string `json:"txCount"`
	UntrackedVolumeUSD string `json:"untrackedVolumeUSD"`
	VolumeToken0       string `json:"volumeToken0"`
	VolumeToken1       string `json:"volumeToken1"`
	VolumeUSD          string `json:"volumeUSD"`
}

var infoPairGraph = `{"query":"{\n  pair(id:\"%s\") {\ntoken0 {\n  id\n} \ntoken1 {\n  id\n} \nreserve0 \nreserve1 \ntotalSupply \nreserveETH \nreserveUSD \ntrackedReserveETH \ntoken0Price \ntoken1Price \nvolumeToken0 \nvolumeToken1 \nvolumeUSD \nuntrackedVolumeUSD \ntxCount \ncreatedAtTimestamp \ncreatedAtBlockNumber \nliquidityProviderCount\n  }\n}\n","variables":null}`

func GetPairInfo(pairAddre string) (pair *PairInfo, err error) {
	bs, err := ReqSwapGraph(fmt.Sprintf(infoPairGraph, pairAddre))
	if err == nil {
		//使其直接返回字符串
		bs = bytes.TrimPrefix(bs, []byte(`{"data":{"pair":`))
		bs = bytes.TrimSuffix(bs, []byte(`}}`))
		log.Println("GetPairInfo", string(bs))
		pair = &PairInfo{}
		err = json.Unmarshal(bs, &pair)
		if err == nil {
			if pair == nil {
				err = errors.New("pair is null ")
			}
			return pair, err
		}
	}
	return nil, err
}
