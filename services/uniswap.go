package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"stock/utils"
)

var SwapGraphApi ="https://api.thegraph.com/subgraphs/name/ianlapham/uniswapv2"
var tokenDayDataGraph=`{"operationName":"tokenDayDatas","variables":{"tokenAddr":"%s","skip":0},"query":"query tokenDayDatas($tokenAddr: String\u0021, $skip: Int\u0021) {\n  tokenDayDatas(first: %d, skip: $skip, orderBy: date, orderDirection: desc, where: {token: $tokenAddr}) {\n    id\n    date\n    priceUSD\n    totalLiquidityToken\n    totalLiquidityUSD\n    totalLiquidityETH\n    dailyVolumeETH\n    dailyVolumeToken\n    dailyVolumeUSD\n    __typename\n  }\n}\n"}`
func GetTokenDayData(tokenAddre string, days int ){
	bs, err := utils.ReqResBody(SwapGraphApi, "", "POST", nil, []byte(fmt.Sprintf(tokenDayDataGraph,tokenAddre,days)))
	if err == nil {
		//使其直接返回字符串
		bs= bytes.TrimPrefix(bs,[]byte(`{"data":{"tokenDayDatas":`))
		bs= bytes.TrimSuffix(bs,[]byte(`}}`))
		log.Println(string(bs))
	}
}



type LpSnapPairInfo struct {
	Block                     int64  `json:"block"`
	//lp数量
	LiquidityTokenTotalSupply string `json:"liquidityTokenTotalSupply"`
	Reserve0                  string `json:"reserve0"`
	Reserve1                  string `json:"reserve1"`
	//pair市值
	ReserveUSD                string `json:"reserveUSD"`
	Timestamp                 int64  `json:"timestamp"`
	Token0PriceUSD            string `json:"token0PriceUSD"`
	Token1PriceUSD            string `json:"token1PriceUSD"`
}

var infoLpDataGraph=`{"query":"{\n  liquidityPositionSnapshots(first:1,orderBy:timestamp,orderDirection:desc where:{pair:\"%s\"}) {\n  # id\n  timestamp\n  block\n  token0PriceUSD\n  token1PriceUSD\n  reserve0\n  reserve1\n  reserveUSD\n  liquidityTokenTotalSupply\n  # liquidityTokenBalance\n  }\n}","variables":null}`
func GetLpPairInfo(pairAddre string ) (pair *LpSnapPairInfo,err error ){
	bs, err := utils.ReqResBody(SwapGraphApi, "", "POST", nil, []byte(fmt.Sprintf(infoLpDataGraph,pairAddre)))
	if err == nil {
		//使其直接返回字符串
		bs= bytes.TrimPrefix(bs,[]byte(`{"data":{"liquidityPositionSnapshots":`))
		bs= bytes.TrimSuffix(bs,[]byte(`}}`))
		//log.Println(string(bs))
		ps:=[]*LpSnapPairInfo{}
		err=json.Unmarshal(bs,&ps)
		if err == nil {
			if len(ps)==0{
				return nil,errors.New("pair not found")
			}
			return ps[0],err
		}
	}
	return nil,err
}

type TokenInfo struct {
	Decimals           string `json:"decimals"`
	DerivedETH         string `json:"derivedETH"`
	PriceUsd float64
	Name               string `json:"name"`
	Symbol             string `json:"symbol"`
	TotalLiquidity     string `json:"totalLiquidity"`
	TotalSupply        string `json:"totalSupply"`
	TradeVolume        string `json:"tradeVolume"`
	TradeVolumeUSD     string `json:"tradeVolumeUSD"`
	TxCount            string `json:"txCount"`
	UntrackedVolumeUSD string `json:"untrackedVolumeUSD"`
}
type PriceView struct {
	PriceUsd float64
	BigPrice string
	Node string
	Timestamp int64
	//Sign值由 Timestamp+BigPrice
	Sign []byte
	//Signs []*PriceView `json:",omitempty"`
}
type DataPriceView struct {
	PriceUsd float64
	BigPrice string
	//Node string
	Timestamp int64
	//Sign值由 Timestamp+BigPrice
	Sign []byte
	Signs []*PriceView `json:",omitempty"`
}

var infoTokenGraph=`{"query":"{\n  token(id:\"%s\") {\nsymbol\nname\ndecimals\ntotalSupply\ntradeVolume\ntradeVolumeUSD\nuntrackedVolumeUSD\ntxCount\ntotalLiquidity\nderivedETH\n\n  }\n}","variables":null}`
func GetTokenInfo(pairAddre string ) (token *TokenInfo,err error ){
	bs, err := utils.ReqResBody(SwapGraphApi, "", "POST", nil, []byte(fmt.Sprintf(infoTokenGraph,pairAddre)))
	if err == nil {
		//使其直接返回字符串
		bs= bytes.TrimPrefix(bs,[]byte(`{"data":{"token":`))
		bs= bytes.TrimSuffix(bs,[]byte(`}}`))
		log.Println("GetTokenInfo",string(bs))
		token=&TokenInfo{}
		err=json.Unmarshal(bs,&token)
		if err == nil {
			return token,err
		}
	}
	return nil,err
}


type PairInfo  struct {
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

var infoPairGraph=`{"query":"{\n  pair(id:\"%s\") {\ntoken0 {\n  id\n} \ntoken1 {\n  id\n} \nreserve0 \nreserve1 \ntotalSupply \nreserveETH \nreserveUSD \ntrackedReserveETH \ntoken0Price \ntoken1Price \nvolumeToken0 \nvolumeToken1 \nvolumeUSD \nuntrackedVolumeUSD \ntxCount \ncreatedAtTimestamp \ncreatedAtBlockNumber \nliquidityProviderCount\n  }\n}\n","variables":null}`
func GetPairInfo(pairAddre string ) (pair *PairInfo,err error ){
	bs, err := utils.ReqResBody(SwapGraphApi, "", "POST", nil, []byte(fmt.Sprintf(infoPairGraph,pairAddre)))
	if err == nil {
		//使其直接返回字符串
		bs= bytes.TrimPrefix(bs,[]byte(`{"data":{"pair":`))
		bs= bytes.TrimSuffix(bs,[]byte(`}}`))
		log.Println("GetPairInfo",string(bs))
		pair=&PairInfo{}
		err=json.Unmarshal(bs,&pair)
		if err == nil {
			if pair==nil{
				err=errors.New("pair is null ")
			}
			return pair,err
		}
	}
	return nil,err
}


