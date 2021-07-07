### 节点eth钱包
启动时,会自动创建 asset/pkey 私钥文件；应用接口 /pub/stock/node_wallets,会反返回所有节点的钱包地址

### current nodes ip list 
- node1: 62.234.169.68
- node2: 62.234.188.160 
- node0: 49.232.234.250
为方便识别结点, 自己hosts文件里加上 node0 node1这样的ip 解析.

### oss list
- https://snode0.oss-cn-beijing.aliyuncs.com/pub/stock/aggre_info/AAPL/1620383145
- https://snode1.oss-cn-beijing.aliyuncs.com/pub/stock/aggre_info/AAPL/1620383145
- https://snode2.oss-cn-beijing.aliyuncs.com/pub/stock/aggre_info/AAPL/1620383145 

### os system
linux 系列即可．　在window系统上，下面的启动命令中的参数中的特殊字会有不兼容的处理方式，导致无法启动

### startup 
```shell script

git clone ...
go build 

./stock --db --port 8001 'root:password@tcp(localhost:3306)/mydb?loc=Local&parseTime=true&multiStatements=true' --nodes=http://node0:8001,http://node1:8001,http://node2:8001 --infura 891eeaa3c7f945b880608e1cc9976284
#infura 最后换成自己infura_proj_id；　infura的项目id,需要自行去https://infura.io申请
#nodes参数指定, 其它节点列表
#stock启动后，会另外启动一个线程，这个线程会在美股开盘时间，每隔１秒抓取苹果和特斯拉股价．

#stock启动后,会在8001端口，响应获取股价的http请求．
#例如获取苹果这个时间点1620383144的股价
curl -X GET "http://62.234.169.68:8001/pub/stock/aggre_info/AAPL/1620383145" -H "accept: application/json"
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

http://62.234.169.68:8001/docs/index.html


### oss EndPoint
通过oss 访问时，接口主机地址替换成oss 节点地址即可．
http://62.234.169.68:8001/　换成　https://snode1.oss-cn-beijing.aliyuncs.com/　即可

通过oss访问的目前限于价格签名类接口；其它如图表数据／any-api接口暂不使用oss访问，通过oss访问可能会有问题．