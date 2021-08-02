### api 列表
|api-path|api名字|类别|
|--|----------------|----------------|
|"/pub/dex/ftx_chart_prices/:coin_type/:count/:interval/:timestamp"|          "获取杠杆btc代币不同时间区间的价格图表信息" |ftx杠杆代币|
|"/pub/dex/ftx_price/:coin_type/:data_type/:timestamp"|                       "获取ftx token价格信息"         |ftx杠杆代币|

价格签名对应的**solidity**hash验签recover：
```js
    struct Object {
        address token;
        uint256 price;
        uint256 timestamp;
    }
prefixedHash=keccak256(abi.encode(Object)).toEthSignedMessageHash()
prefixedHash.recover(sign)
```

##  概念
ftx杠杆币是，随某种主流币种价格，以相应的倍数涨跌的币.

bull类币种是看多币，会以相应的倍数上涨，张的越多盈利越多，我们使用的主流币名字加3x 10x；这样的后缀表示bull币名字.

bear类币种是看空币，会以相应的倍数上下跌，跌的越多盈利越多，我们使用的主注流币名字加3s 10s; 这样的后缀表示 bear币名字.

杠杆币存在优势：
比如３倍的bull,它会定期调仓，在牛市时，比它应的原始币种，会取得比３倍以上的收益；当遇上熊市时，会损失会小于３倍.　bear与之相反.

我们模拟了以下几种杠杆币价格
- bull类: mvi2x,usd, btc3x, eth3x, vix3x, gold10x, eur20x,ndx10x
- bear类: govt20x,mvi2s, btc3s, eth3s

ftx相关文档:
[doc](https://help.ftx.com/hc/zh-cn/articles/360032973651-%E8%AE%BE%E8%AE%A1%E5%8E%9F%E7%90%86%E7%AC%AC%E5%9B%9B%E8%AE%B2-%E6%9D%A0%E6%9D%86%E4%BB%A3%E5%B8%81%E7%9A%84%E8%B0%83%E4%BB%93%E6%9C%BA%E5%88%B6)




## 杠杆币的计算公式如下
- 上一个杠杆价格：lastBullPrice
- 当前原始币价格：currRawPrice
- 上一个原始价格：lastRawPrice
- 倍数：type

#### 买多的报价
 bullPrice = lastBullPrice*(1+(currRawPrice-lastRawPrice)/lastRawPrice*type )
#### 买空的报价
 bearPrice = lastBullPrice*(1-(currRawPrice-lastRawPrice)/lastRawPrice*type)

#### 初始价格
模似杠杆币时,我们给定了一个初始价格, 对应上面讲的lastBullPrice
```shell script
	"mvi2x": 99,
	"btc3x": 110054.79,
	"eth3x": 7900.56,
	"vix3x": 53.7,
	"gold10x": 19022.8,
	"eur20x":  244.66,
	"ndx10x":  136488,
	"govt20x": 5268,
	"mvi2s": 112.0,
	"btc3s": 98265.0,
	"eth3s": 6205.8,
```
#### 模似杠杆币数据的开始时间
 bull数据初始化时间,使用相应原始币从2021-06-01开始的数据
 
 bear数据初始化时间,使用相应原始币从2021-07-23 10:00日开始的数据
 
### 调仓
```shell script
 倍数和波动触发值
 "2x":  15%,
 "3x":  10%,
 "10x": 3%,
 "20x": 1.5%
```
调仓的触发条件
1. 日内价格达到波动触发值，触发调仓rebalance
1. 24小时之内未触发，在北京时间下午两点触发一次调仓
1. 调仓价格,是我们设定的杠杆币上的另一个价格,应该是直接对应相应ftx的盈利计算,暂时不对外显示.
 
### 调仓价格
 它的初始值同杠杆币的初始价格一样.但后续计算方式有区另,如下:
 
 公式字段说明
 - currentReBalance: 当前调仓价格 
 - rate: 30"除以"杠杆倍数 
 - preRebalance: 上一个调仓价格 
 - bullPrice: 当前杠杆币价格
  
#### 调仓方式-: 当原始价格变化rawChange的绝对值大于rate bull调仓公式如下:
rawChange>0时:
  currentReBalance = preRebalance * (1 + rate/100.0)

rawChange<0时:
currentReBalance = preRebalance * (1 - rate/100.0)

bear类同上, rate取反(rate=-1*rate),再使用上面的公式

#### 调仓方式二: 24小时之内未触发，在北京时间下午两点触发一次调仓
currentReBalance=bullPrice




