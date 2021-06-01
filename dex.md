###Api Path:  

#####获取流动性价格信息　
/pub/dex/lp_price/{pair}/{timestamp}


#####获取token价格信息　
/pub/dex/token_price/{token}/{timestamp}

#####获取token不同时间区间的价格图表信息 
/pub/dex/token_chart_prices/{token}/{count}/{interval}/{timestamp}

#####获取token相应天数的统计图表信息
/pub/dex/token_day_datas/{token}/{days}/{timestamp}
含token各种统计信息，详见swag 文件档



详见swag 文件档：
http://62.234.169.68:8001/docs/index.html



对应的**solidity**hash字段拼接顺序为Timestamp BigPrice：
```js
 prefixedHash=keccak256(abi.encodePacked(Timestamp,BigPrice)).toEthSignedMessageHash()
 prefixedHash.recover(sign)
```
