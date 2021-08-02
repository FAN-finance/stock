 ### 获取币价换算　
/pub/coin_price/{coin}/{vs_coin}/{timestamp}

详见swag 文件档：
http://62.234.169.68:8001/docs/index.html

目前支持的62种加密货币符号如下：btc,aed,ars,aud,bch,bdt,bhd,bits,bmd,bnb,brl,byn,cad,chf,clp,cny,czk,dkk,dot,eos,eth,eur,gbp,hkd,huf,idr,ils,inr,jpy,krw,kwd,link,lkr,ltc,mmk,mxn,myr,ngn,nok,nzd,php,pkr,pln,rub,sar,sats,sek,sgd,thb,try,twd,uah,usd,vef,vnd,xag,xau,xdr,xlm,xrp,yfi,zar

### demo
```bash
curl -X GET "http://localhost:8001/pub/coin_price/eth/usd?timestamp=1620383144" -H "accept: application/json"
response body:
{
  "Price": 2613.47,
  "BigPrice": "2613469900000000000000",
  "Timestamp": 1620383144,
  "Coin": "eth",
  "VsCoin": "usd",
  "Sign": "MsnYH3X7MKtiXcUPaI/juPZW9URGJ3mdA5Tq1Nx4UD9uairxmtjIdx7aeTBs8vOmbJNG7CGM/BZOLW5ekqw6KAE=",
  "Signs": [
    {
      "Price": 2613.47,
      "BigPrice": "2613469900000000000000",
      "Node": "http://localhost:8001",
      "Timestamp": 1620383144,
      "Coin": "eth",
      "VsCoin": "usd",
      "Sign": "MsnYH3X7MKtiXcUPaI/juPZW9URGJ3mdA5Tq1Nx4UD9uairxmtjIdx7aeTBs8vOmbJNG7CGM/BZOLW5ekqw6KAE="
    },
    {
      "Price": 2613.47,
      "BigPrice": "2613469900000000000000",
      "Node": "http://localhost:8001",
      "Timestamp": 1620383144,
      "Coin": "eth",
      "VsCoin": "usd",
      "Sign": "MsnYH3X7MKtiXcUPaI/juPZW9URGJ3mdA5Tq1Nx4UD9uairxmtjIdx7aeTBs8vOmbJNG7CGM/BZOLW5ekqw6KAE="
    }
  ]
}
```

价格签名对应的**solidity**hash字段拼接顺序为 Timestamp Coin VsCoin BigPrice ：
```js
 prefixedHash=keccak256(abi.encodePacked(Timestamp,Coin,VsCoin,BigPrice)).toEthSignedMessageHash()
 prefixedHash.recover(sign)
```
