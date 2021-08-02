# 获取数据
我们使用restful-api获取数据。

比如获取**苹果股票**价格数据访问如下地址：
https://vnode1.oss-us-west-1.aliyuncs.com/pub/stock/aggre_info/AAPL/1/1620383144

前面”https://vnode1.oss-us-west-1.aliyuncs.com/“ 是阿里ossEndpoint. 后面"/pub/stock/aggre_info/AAPL/1/1620383144"为restful-api的路径或是参数。 客户端可以直接使用返回的数据，并对数据包含的[签名](sign.md)做相应的判断。


endPoint节点列表
- https://gnode.oss-us-west-1.aliyuncs.com/pub/stock/aggre_info/AAPL/1/1620383144
- https://mnode.oss-us-west-1.aliyuncs.com/pub/stock/aggre_info/AAPL/1/1620383144
- https://vnode1.oss-us-west-1.aliyuncs.com/pub/stock/aggre_info/AAPL/1/1620383144
- https://anode1.oss-us-west-1.aliyuncs.com/pub/stock/aggre_info/AAPL/1/1620383144


apifull-api功能列表


|状态|api-path|api名字|类别|
|---|--|----------------|----------------|
|可用|"/pub/dex/lp_price/:pair/:timestamp"|                                        "获取lp价格信息"               |uniswap交易所 |
|可用|"/pub/dex/token_chart_prices/:token/:count/:interval/:timestamp"|            "获取token不同时间区间的价格图表信息"    |uniswap交易所 |
|可用|"/pub/dex/token_day_datas/:token/:days/:timestamp"|                          "获取token相应天数的统计图表信息"      |uniswap交易所 |
|可用|"/pub/dex/token_price/:token/:data_type/:timestamp"|                         "获取token价格信息"             |uniswap交易所 |
|可用|"/pub/dex/pair/token_chart_prices/:pair/:token/:count/:interval/:timestamp"| "从Pair获取token不同时间区间的价格图表信息"|uniswap交易所 |
|可用|"/pub/dex/pair/token_price/:pair/:token/:data_type/:timestamp"|              "从Pair获取token价格信息"       |uniswap交易所|
|可用|"/pub/dex/token/token_chart_supply/:token/:amount/:timestamp"|               "获取某个token的totalSupply的变化量"|uniswap交易所|
|禁用|"/pub/dex/token_chain_price/:token/:data_type/:timestamp"|                   "获取token链上价格信息,仅预定几个token"          |eth链上 |
|可用|"/pub/dex/ftx_chart_prices/:coin_type/:count/:interval/:timestamp"|          "获取杠杆btc代币不同时间区间的价格图表信息" |ftx杠杆代币|
|可用|"/pub/dex/ftx_price/:coin_type/:data_type/:timestamp"|                       "获取ftx token价格信息"         |ftx杠杆代币|
|可用|"/pub/stock/aggre_info/:code/:data_type/:timestamp"|                         "获取共识美股价格"               |美股报价 |
|可用|"/pub/dex/stock_chart_prices/:coin_type/:count/:interval/:timestamp"|        "获取股票不同时间区间的价格图表信息"      |美股报价|
|可用|"/pub/coin_price/:coin/:vs_coin/:timestamp"|获取币价换算|加密货币兑换价格查询|
|可用|"/pub/stock/any_apis"|                                                       "所有节点any-api"             |any-api|
|可用|"/pub/stock/market_status/:timestamp"|                                       "获取美股市场开盘状态"             |dic|
|可用|"/pub/stock/stat"|                                                           "当前节点状态:记录数,钱包地址"         |stat|
|可用|"/pub/stock/stats"|                                                          "所有节点状态:记录数,钱包地址"         |stat|



参数说明见swagger 文档：http://62.234.169.68:8001/docs/index.html
