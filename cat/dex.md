###  api 列表:
|api-path|api名字|类别|                                                                                                      
|--|----------------|----------------|                                                                                   
|"/pub/dex/lp_price/:pair/:timestamp"|                                        "获取lp价格信息"               |uniswap交易所 |     
|"/pub/dex/token_chart_prices/:token/:count/:interval/:timestamp"|            "获取token不同时间区间的价格图表信息"    |uniswap交易所 |    
|"/pub/dex/token_day_datas/:token/:days/:timestamp"|                          "获取token相应天数的统计图表信息"      |uniswap交易所 |    
|"/pub/dex/token_price/:token/:data_type/:timestamp"|                         "获取token价格信息"             |uniswap交易所 |    
|"/pub/dex/pair/token_chart_prices/:pair/:token/:count/:interval/:timestamp"| "从Pair获取token不同时间区间的价格图表信息"|uniswap交易所 |   
|"/pub/dex/pair/token_price/:pair/:token/:data_type/:timestamp"|              "从Pair获取token价格信息"       |uniswap交易所|      
|"/pub/dex/token/token_chart_supply/:token/:amount/:timestamp"|               "获取某个token的totalSupply的变化量"|uniswap交易所|    


详见swag 文件档：
http://62.234.169.68:8001/docs/index.html


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
