## 获取最新价格

最新价格：是指**最近１小时内**的最高价或最低价．

#### 价格类型列表
最新价格有以下３种价格类型列.
1.  [美股报价](stock.md)　：如苹果代码 AAPL；特斯拉代码 TSLA

可报价的股票代码，可从这里查询https://twelvedata.com/docs#symbol-search

1.  [uniswap交易所的token报价]
在uniswap-info: https://v2.info.uniswap.org/ 上建立的pair相关token都可以报价
1.  [ftx杠杆币模拟报价]
目前提交供以下７种x杠杆币：btc3x, eth3x, vix3x, govt20x, gold10x, eur20x,ndx10x,mvi2x

#### 价格实时性：
- 价格是有一定延时的，比如价格数据接口，在预言机端api有１分钟左右的缓存．

- uniswap交易所的token价格，数据源来自thegraph,有时thegraph的块处理延迟，大多数情况下在10-30秒左右．个别极端情况会达到10分钟．偶尔出现服务不可用．

- 美股价格，数据源于twelvedata端，未见延迟

- ftx杠杆币价格，每分钟去twelvedata抓取新数据，新数据每分钟生成ftx杠杆代币价格，会有１分钟左右的延迟

