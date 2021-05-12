
## stock info api
为美股预言机，提供数据源服务api

目前提供苹果和特斯拉每秒的股价
- 苹果代码 AAPL；特斯拉代码 TSLA


### ~~gen rsa key~~
目前无需使用了,改用 ethereum签名方法.
```shell script
openssl req -new -newkey rsa:2048 -days 1000 -nodes -x509 -keyout asset/key.pem -out asset/cert.pem -subj "/C=GB/ST=bj/L=bj/O=uprets/OU=ruprets/CN=*"
```

### ~~gen eth wallet~~
```shell script
#目前无需使用了,启动时,会自动创建 asset/pkey 私钥文件；应用接口 /pub/stock/node_wallets,会反返回所有节点的钱包地址
geth  account new --keystore asset
#asset 目录只保留一个wallet文件即可，只会使用用一个wallet文件
```

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
curl -X GET "http://localhost:8001/pub/stock/aggre_info?code=AAPL&timestamp=1620383144" -H "accept: application/json"
Response body:
{
  "Code": "AAPL",
  "Sign": "rqOfrJkrOmwA3WATCG5KHrjSRfK/HzjpL9ZX6LhP3nMy2tag5H+X5wE1AetWyeguMfngX3lZ3WUbWhCWzI4a8gE=",
  "Price": 129.74,
  "Timestamp": 1620383144,
  "Signs": [
    {
      "Code": "AAPL",
      "Node": "http://localhost:8001",
      "Timestamp": 1620383144,
      "Price": 129.74,
      "Sign": "rqOfrJkrOmwA3WATCG5KHrjSRfK/HzjpL9ZX6LhP3nMy2tag5H+X5wE1AetWyeguMfngX3lZ3WUbWhCWzI4a8gE="
    },
    {
      "Code": "AAPL",
      "Node": "http://localhost:8001",
      "Timestamp": 1620383144,
      "Price": 129.74,
      "Sign": "rqOfrJkrOmwA3WATCG5KHrjSRfK/HzjpL9ZX6LhP3nMy2tag5H+X5wE1AetWyeguMfngX3lZ3WUbWhCWzI4a8gE="
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
curl -X GET "http://localhost:8001/pub/stock/aggre_info?code=AAPL&timestamp=1620383144" -H "accept: application/json"
Response body:
{
  "Code": "AAPL",
  "Sign": "rqOfrJkrOmwA3WATCG5KHrjSRfK/HzjpL9ZX6LhP3nMy2tag5H+X5wE1AetWyeguMfngX3lZ3WUbWhCWzI4a8gE=",
  "Price": 129.74,
  "Timestamp": 1620383144,
  "Signs": [
    {
      "Code": "AAPL",
      "Node": "http://localhost:8001",
      "Timestamp": 1620383144,
      "Price": 129.74,
      "Sign": "rqOfrJkrOmwA3WATCG5KHrjSRfK/HzjpL9ZX6LhP3nMy2tag5H+X5wE1AetWyeguMfngX3lZ3WUbWhCWzI4a8gE="
    },
    {
      "Code": "AAPL",
      "Node": "http://localhost:8001",
      "Timestamp": 1620383144,
      "Price": 129.74,
      "Sign": "rqOfrJkrOmwA3WATCG5KHrjSRfK/HzjpL9ZX6LhP3nMy2tag5H+X5wE1AetWyeguMfngX3lZ3WUbWhCWzI4a8gE="
    }
  ]
}
```
#### 签名sign的计算方式
目前使用 ethereum签名方式:

- **message= Code+","+Timestamp+"," +Price**;
- sign=crypto.Sign(Keccak256(message),edcasaKey)

~~使用第一步生成的rsa 私钥，把Response body签名后生成．~~
~~代码大致为：~~
~~sign=Privkey.Sign(sha256.sum(ResponseBody),crypto.SHA256)~~

#### ~~验签~~

~~当客户端需要验证sign时，需要使用第一步生成的rsa证书文件asset/cert.pem~~
~~验证代码大致为：　ok=rsa.VerifyPKCS1v15(publicCert,sha256.sum(ResponseBody),sign）~~

~~签名和验签代码详见：~~
~~main.go中方法　StockInfoHandler　VerifyInfoHandler~~



