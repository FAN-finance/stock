


**系统架构图**
<img src="./assets/1422759379.png" alt="总体架构图" width="600" />


如图所示，整个系统分为3部分：
- [数据源](dataSource.md)：我们的数据源基于coingecko　uniswap  twelvedata coinmarketcap
- 预言机节点集群：节点集群负责从数据源获取数据，响应第三方的数据请求
- 第三方应用：主要通过从**oss存储**获取预言机价格数据．