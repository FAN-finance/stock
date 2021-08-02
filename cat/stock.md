
### api 列表
|api-path|api名字|类别|
|--|----------------|----------------|
|"/pub/stock/aggre_info/:code/:data_type/:timestamp"|                         "获取共识美股价格"               |美股报价 |
|"/pub/dex/stock_chart_prices/:coin_type/:count/:interval/:timestamp"|        "获取股票不同时间区间的价格图表信息"      |美股报价|

### Api Path:  
/pub/stock/info 
详见swag 文件档：
http://62.234.169.68:8001/docs/index.html


### example:
```shell script
curl -X GET "http://62.234.169.68:8001/pub/stock/aggre_info/AAPL/1/1620383145" | jq .
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100  2012  100  2012    0     0   163k      0 --:--:-- --:--:-- --:--:--  163k
{
  "StockCode": "AAPL",
  "DataType": 1,
  "Code": "0xD87f6eCC45ABAD69446DA79a19D1E5Cf3B779098",
  "IsMarketOpening": true,
  "MarketOpenTime": 1620397800,
  "Sign": "2iODBBWb0zZQZza7op1Iuj5EjdhiXLe0PxlVaqiDlUdofDU/PMLl50t8Udz1xgxWJdBoLu8xN0SeHLDsIE80ths=",
  "Price": 146.325,
  "BigPrice": "146325000000000000000",
  "Timestamp": 1620383145,
  "Signs": [
    {
      "StockCode": "AAPL",
      "DataType": 1,
      "Code": "",
      "Node": "http://node1:8001",
      "NodeAddress": "0x0Bd6B9402BE15C9D80Fc0465584E5cb1e48c5C9e",
      "Timestamp": 1620383145,
      "Price": 146.325,
      "BigPrice": "146325000000000000000",
      "Sign": null
    },
    {
      "StockCode": "AAPL",
      "DataType": 1,
      "Code": "",
      "Node": "http://node0:8001",
      "NodeAddress": "0x0143d0AA3C8a405b1Dd936b0925572c19dDA6B3a",
      "Timestamp": 1620383145,
      "Price": 146.325,
      "BigPrice": "146325000000000000000",
      "Sign": null
    },
    {
      "StockCode": "AAPL",
      "DataType": 1,
      "Code": "",
      "Node": "http://node2:8001",
      "NodeAddress": "0x964998E87E8b9D1CD8b68D185f5F79b9e7c50f69",
      "Timestamp": 1620383145,
      "Price": 146.325,
      "BigPrice": "146325000000000000000",
      "Sign": null
    }
  ],
  "AvgSigns": [
    {
      "StockCode": "AAPL",
      "DataType": 1,
      "Code": "0xD87f6eCC45ABAD69446DA79a19D1E5Cf3B779098",
      "Node": "http://node1:8001",
      "NodeAddress": "0x0Bd6B9402BE15C9D80Fc0465584E5cb1e48c5C9e",
      "Timestamp": 1620383145,
      "Price": 146.325,
      "BigPrice": "146325000000000000000",
      "Sign": "2iODBBWb0zZQZza7op1Iuj5EjdhiXLe0PxlVaqiDlUdofDU/PMLl50t8Udz1xgxWJdBoLu8xN0SeHLDsIE80ths="
    },
    {
      "StockCode": "AAPL",
      "DataType": 1,
      "Code": "0xD87f6eCC45ABAD69446DA79a19D1E5Cf3B779098",
      "Node": "http://node0:8001",
      "NodeAddress": "0x0143d0AA3C8a405b1Dd936b0925572c19dDA6B3a",
      "Timestamp": 1620383145,
      "Price": 146.325,
      "BigPrice": "146325000000000000000",
      "Sign": "tGDwc511bjhDeYm3MW/4LZ9+U0Gyt1hS32LVDFqnEas+SxUkc185N6eCSa7GuQ1YBUszlvsQPGtoZ5C8uOPaBxs="
    },
    {
      "StockCode": "AAPL",
      "DataType": 1,
      "Code": "0xD87f6eCC45ABAD69446DA79a19D1E5Cf3B779098",
      "Node": "http://node2:8001",
      "NodeAddress": "0x964998E87E8b9D1CD8b68D185f5F79b9e7c50f69",
      "Timestamp": 1620383145,
      "Price": 146.325,
      "BigPrice": "146325000000000000000",
      "Sign": "L2eGCDNz+mr7FOsAF3a8I4DMCK54azztiNfArWn6BJUYdvGN4UT7MK4lEQh2kTcgDZ52QDo8MMkUbK8D5UevtBs="
    }
  ]
}
```

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
