
## stock info api
为美股预言机，提供数据源服务api

目前提供苹果和特斯拉每秒的股价
- 苹果代码 AAPL；特斯拉代码 TSLA


### gen rsa key
```shell script
openssl req -new -newkey rsa:2048 -days 1000 -nodes -x509 -keyout asset/key.pem -out asset/cert.pem -subj "/C=GB/ST=bj/L=bj/O=uprets/OU=ruprets/CN=*"
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

### startup 
```shell script


go build 

./stock --db --port 8001 'root:password@tcp(localhost:3306)/mydb?loc=Local&parseTime=true&multiStatements=true'
#stock启动后，会另外启动一个线程，这个线程会在美股开盘时间，每隔１秒抓取苹果和特斯拉股价．

#stock启动后,会在8001端口，响应获取股价的http请求．
#获取苹果这个时间点1620383144的股价
curl -X GET "http://localhost:8001/pub/stock/info?code=AAPL&timestamp=1620383144" -H "accept: application/json"
{
  "Code": "AAPL",
  "Price": 129.74,
  "StockName": "苹果",
  "Timestamp": 1620383144,
  "UpdatedAt": "2021-05-07T18:25:44.27+08:00"
}


```
### startup args
```shell script
./stock -h
Usage of /tmp/go-build868767577/b001/exe/main:
  -c, --cert string   pem encoded x509 cert (default "./asset/cert.pem")
  -d, --db string     mysql database url (default "root:password@tcp(localhost:3306)/mydb?loc=Local&parseTime=true&multiStatements=true")
  -e, --env string    环境名字debug prod test (default "debug")
  -k, --key string    pem encoded private key (default "./asset/key.pem")
  -p, --port string   api　service port (default "8001")

```

### swagger api doc
http://localhost:8001/docs/index.html

###签名
获取股价的接口　pub/stock/info，在http响应header里添加一个名字为sign的header，值为对响应body的添名．
签名主要可用于验证数据由我们服务器提供．验证为可选项

example:
```shell script
curl -X GET "http://localhost:8001/pub/stock/info?code=AAPL&timestamp=1620383144" -H "accept: application/json"
Response body:
{
  "Code": "AAPL",
  "Price": 129.74,
  "StockName": "苹果",
  "Timestamp": 1620383144,
  "UpdatedAt": "2021-05-07T18:25:44.27+08:00"
}
Response headers:
content-length: 118 
 content-type: application/json; charset=utf-8 
 date: Sat, 08 May 2021 01:35:04 GMT 
 sign: m6XI2hZ2AD0BcgGqIOG4YSDaBQMVTuTiN7dVuDvLPdfY5+IUa24gi8aIKE4mw5Z43kue5PDltworBpK597QbUPXOIZi+hPpebcXjwgkGfcvwdHbOqVhb6NlAQIdoAeMOzA/05En4wjubaqX4Mr1sL5Yiq3lKHjIX5nlbLf33lErPuBim7TlZpQu6FNkm7aro1igH+doIOVYZPVxpBl8eu+Vzu8iBiQiAgx0tlLFEEs+J8Kx5Lnrrd1lHUyWQdoKR52tYtilF1Owt4QGzbCEAHaVzfrRS40DYi2g4gCshZGpn3f8PXzz9b/rLn2YZTeKlMBuLVRMN01hnzwzhr+te9Q== 
 vary: Origin 
```
####签名sign的计算方式：
使用第一步生成的rsa 私钥，把Response body签名后生成．
代码大致为：
sign=Privkey.Sign(sha256.sum(ResponseBody),crypto.SHA256)

####验签
当客户端需要验证sign时，需要使用第一步生成的rsa证书文件asset/cert.pem
验证代码大致为：　ok=rsa.VerifyPKCS1v15(pubiicCert,sha256.sum(ResponseBody),sign）

签名和验签代码详见：
main.go中方法　StockInfoHandler　VerifyInfoHandler



