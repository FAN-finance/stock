
###Api Path:  
/pub/stock/info 
详见swag 文件档：
http://62.234.169.68:8001/docs/index.html


###example:
```shell script
curl -X GET "http://localhost:8001/pub/stock/aggre_info/AAPL/1620383145" -H "accept: application/json"
Response body:
{
  "Code": "AAPL",
  "Sign": "3KGU6Rd0hMbPDkjWwVhX7qCbW8RE5WEWZSDQ0vupqNMLj/aTsyRNH6c0/yKfbiEMa8f98cGkUK1vyrR6AQrlNQE=",
  "Price": 129.74,
  "TextPrice": "129740000000000000000",
  "Timestamp": 1620383144,
  "Signs": [
    {
      "Code": "AAPL",
      "Node": "http://localhost:8001",
      "Timestamp": 1620383144,
      "Price": 129.74,
      "TextPrice": "129740000000000000000",
      "Sign": "3KGU6Rd0hMbPDkjWwVhX7qCbW8RE5WEWZSDQ0vupqNMLj/aTsyRNH6c0/yKfbiEMa8f98cGkUK1vyrR6AQrlNQE="
    },
    {
      "Code": "AAPL",
      "Node": "http://localhost:8001",
      "Timestamp": 1620383144,
      "Price": 129.74,
      "TextPrice": "129740000000000000000",
      "Sign": "3KGU6Rd0hMbPDkjWwVhX7qCbW8RE5WEWZSDQ0vupqNMLj/aTsyRNH6c0/yKfbiEMa8f98cGkUK1vyrR6AQrlNQE="
    }
  ]
}
```

对应的**solidity**hash字段拼接顺序为　Timestamp, TextPrice, Code：
```js
 prefixedHash=keccak256(abi.encodePacked(Timestamp, TextPrice, Code)).toEthSignedMessageHash()
 prefixedHash.recover(sign)
```
