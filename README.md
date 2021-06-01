# 开始

## 概述

FAN Oracle将给智能合约链外部可信的数据源。


**系统整体架构图**
<img src="assets/1422759379.png" alt="总体架构图" width="600" />
如图所示，整个系统分为四个部分：
- 数据源：我们的数据源基于coingecko　uniswap 东方财富网
- 预言机节点集群：节点集群负责从数据源获取数据，第三方应用有请求时，节点集体签名后，发送给第三方应用
- 第三方应用：主要通过从**oss存储**获取预言机价格数据，（oss数据源于预言机节点）．然后把数据提交给合约，合约验证白名单中的钱包地址和数据签名是否一致．



## 目前提供服务类别

- [美股报价](stock.md)　：目前提供苹果和特斯拉每秒的股价；苹果代码 AAPL；特斯拉代码 TSLA

- [uniswap交易所的token pair报价](dex.md)，及图表相关信息

- [加密货币间的兑换价格查询](coins.md)　;目前支持的62种加密货币符号如下：btc,aed,ars,aud,bch,bdt,bhd,bits,bmd,bnb,brl,byn,cad,chf,clp,cny,czk,dkk,dot,eos,eth,eur,gbp,hkd,huf,idr,ils,inr,jpy,krw,kwd,link,lkr,ltc,mmk,mxn,myr,ngn,nok,nzd,php,pkr,pln,rub,sar,sats,sek,sgd,thb,try,twd,uah,usd,vef,vnd,xag,xau,xdr,xlm,xrp,yfi,zar

- [any-api 单字段查询](anyapi.md)
