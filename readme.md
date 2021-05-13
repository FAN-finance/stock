
## stock info api
为美股预言机，提供数据源服务api

目前提供苹果和特斯拉每秒的股价
- 苹果代码 AAPL；特斯拉代码 TSLA

### 节点eth钱包
启动时,会自动创建 asset/pkey 私钥文件；应用接口 /pub/stock/node_wallets,会反返回所有节点的钱包地址

### mysql table:
```mysql
-- auto-generated definition
create table stocks
(
    id         bigint auto_increment
        primary key,
    code       varchar(191)    null,
    price      float           null,
    stock_name longtext        null,
    mk         bigint          null,
    diff       float default 0 null,
    timestamp  bigint          null,
    updated_at datetime(3)     null
);

create index code_time
    on stocks (code, timestamp);
```

### current nodes ip list 
- node1: 62.234.169.68
- node2: 62.234.188.160 
- node0: 49.232.234.250
为方便识别结点, 自己hosts文件里加上 node0 node1这样的ip 解析.
### oss list
- https://snode0.oss-cn-beijing.aliyuncs.com/pub/stock/aggre_info/AAPL/1620383145
- https://snode1.oss-cn-beijing.aliyuncs.com/pub/stock/aggre_info/AAPL/1620383145
- https://snode2.oss-cn-beijing.aliyuncs.com/pub/stock/aggre_info/AAPL/1620383145 

### startup 
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
```shell script
./stock -h
Usage of /tmp/go-build868767577/b001/exe/main:
  -d, --db string       mysql database url (default "root:password@tcp(localhost:3306)/mydb?loc=Local&parseTime=true&multiStatements=true")
  -e, --env string      环境名字debug prod test (default "debug")
  -n, --nodes strings   所有节点列表,节点间用逗号分开 (default [http://localhost:8001,http://localhost:8001])
  -p, --port string     api　service port (default "8001")

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

- ~~**message= Code+","+Timestamp+"," +Price**; solidity 无法使用拼串方式生成message,再hash,也没有float字段，hash过程换成如下方法~~
- ~~sign=crypto.Sign(Keccak256(message),edcasaKey)~~

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

 



