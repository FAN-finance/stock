
## stock info api
- 美股报价　：

目前提供苹果和特斯拉每秒的股价；苹果代码 AAPL；特斯拉代码 TSLA

- uniswap

交易所的token pair报价，及图表相关信息

- 加密货币间的兑换价格查询

目前支持的62种加密货币符号如下：btc,aed,ars,aud,bch,bdt,bhd,bits,bmd,bnb,brl,byn,cad,chf,clp,cny,czk,dkk,dot,eos,eth,eur,gbp,hkd,huf,idr,ils,inr,jpy,krw,kwd,link,lkr,ltc,mmk,mxn,myr,ngn,nok,nzd,php,pkr,pln,rub,sar,sats,sek,sgd,thb,try,twd,uah,usd,vef,vnd,xag,xau,xdr,xlm,xrp,yfi,zar

- any-api 单字段查询

- ftx btc3x eth3x

### 节点eth钱包
启动时,会自动创建 asset/pkey 私钥文件；应用接口 /pub/stock/node_wallets,会反返回所有节点的钱包地址


### current nodes ip list 
- node1: 62.234.169.68
- node2: 62.234.188.160 
- node0: 49.232.234.250

国外节点
gnode 34.94.44.103   #google
mnode 52.250.67.202  #micrsoft
vnode 45.76.76.192  # vultr
anode 18.191.204.14  #aws

为方便识别结点, 自己hosts文件里加上 node0 node1这样的ip 解析.
### oss list
- https://snode0.oss-cn-beijing.aliyuncs.com/pub/stock/aggre_info/AAPL/1/1620383144
- https://snode1.oss-cn-beijing.aliyuncs.com/pub/stock/aggre_info/AAPL/1/1620383144
- https://snode2.oss-cn-beijing.aliyuncs.com/pub/stock/aggre_info/AAPL/1/1620383144

国外节点
- https://gnode.oss-us-west-1.aliyuncs.com/pub/stock/aggre_info/AAPL/1/1620383144
- https://mnode.oss-us-west-1.aliyuncs.com/pub/stock/aggre_info/AAPL/1/1620383144
- https://vnode1.oss-us-west-1.aliyuncs.com/pub/stock/aggre_info/AAPL/1/1620383144
- https://anode1.oss-us-west-1.aliyuncs.com/pub/stock/aggre_info/AAPL/1/1620383144





### startup 
启动参数详见min.go 中的"pflag.Parse()"代码段
```shell script

go build 

./stock --db --port 8001 'root:password@tcp(localhost:3306)/mydb?loc=Local&parseTime=true&multiStatements=true' --nodes=http://node0:8001,http://node1:8001,http://node2:8001
#nodes参数指定, 其它节点列表
#stock启动后，会另外启动一个线程，这个线程会在美股开盘时间，每隔１秒抓取苹果和特斯拉股价．

#stock启动后,会在8001端口，响应获取股价的http请求．
#获取苹果这个时间点1620383144的股价
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
### startup args
启动参数详见min.go 中的"pflag.Parse()"代码段

```shell script
./stock -h
Usage of /tmp/go-build868767577/b001/exe/main:
  -d, --db string             mysql database url (default "root:password@tcp(localhost:3306)/mydb?loc=Local&parseTime=true&multiStatements=true")
  -e, --env string            环境名字debug prod test (default "debug")
      --infura string         infura的项目id,需要自行去https://infura.io申请 (default "infura_proj_id")
  -j, --job                   是否抓取数据 (default true)
  -n, --nodes strings         所有节点列表,节点间用逗号分开 (default [http://localhost:8001,http://localhost:8001])
  -p, --port string           api　service port (default "8001")
      --swapGraphApi string   swap theGraphApi (default "https://api.thegraph.com/subgraphs/name/ianlapham/uniswapv2")
```

### swagger api doc
http://localhost:8001/docs/index.html


### 签名
获取股价的接口　pub/stock/aggre_info，．
签名主要可用于验证数据由相应节点提供．

example:
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
json中的　Sign字段为签名；　Sign值由　Timestamp／TextPrice／Code字段使用节点钱包签名得到．

#### 签名sign的计算方式
目前使用 go-ethereum签名方式:

```go
hash := crypto.Keccak256Hash(
		common.LeftPadBytes(big.NewInt(s.Timestamp).Bytes(), 32),
		[]byte(s.TextPrice),
		[]byte(s.Code),
	)
prefixedHash := crypto.Keccak256Hash(
        []byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%v", len(hash))),
        hash.Bytes(),
    )
sign=crypto.Sign(Keccak256(message),edcasaKey)
```

对应的**solidity**验签代码：
```js
 prefixedHash=keccak256(abi.encodePacked(Timestamp, TextPrice, Code)).toEthSignedMessageHash()
 prefixedHash.recover(sign)
```






