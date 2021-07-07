### 签名算法
目前使用同ethereum－client－sdk发送交易时事件时相同的钱包椭圆曲线签名方式；solidity内置支持这种验签． 这种验签规则不同于常规的私钥签名公钥验签：
它的规则是**大致**逻辑是这样的： 
```js
returnAddre=verifySign(dataHash,signature);
return returnAddre==whiteList[returnAddre]
```
验证方法verifySign返回的地址在白名单地址列表里，就表示有效的一签名．


对应的正式的**solidity**验签代码如下：
```js
 prefixedHash=keccak256(abi.encodePacked(Timestamp, TextPrice, Code)).toEthSignedMessageHash()
 prefixedHash.recover(sign)
```

附上别人的一段验签sol demo
```js
 function sync(bytes calldata data) internal returns (OP memory) {
        (bytes memory o, bytes[] memory s) = abi.decode(data, (bytes, bytes[]));
        OP memory op = abi.decode(o, (OP));
        if (op.timestamp <= timestamps[op.token]) {
            op.price = prices[op.token];
            op.timestamp = timestamps[op.token];
        } else {
            prices[op.token] = op.price;
            timestamps[op.token] = op.timestamp;
            require(s.length == SIGNATURENUM);
            bytes32 hash = keccak256(o).toEthSignedMessageHash();
            address auth = address(0);
            for (uint256 i = 0; i < s.length; i++) {
                address addr = hash.recover(s[i]);
                require(addr > auth);
                require(authorization[addr]);
                auth = addr;
            }
        }

        require(op.timestamp + EXPIRY > block.timestamp);
        return op;
    }
```

更多使用规则参考：　@openzeppelin/contracts-upgradeable/utils/cryptography/ECDSAUpgradeable.sol"

有签名的接口，一般为单字段数据接口如：
- 美股价格股票接口
- uniswap token价格接口
- coin兑换价格接口
- any-api数据接口

其它见swag文档

列表数据接口暂不使用签名，主要由于列表json太大，字段太多，无法在太合约里拼接hash，并验签．


### 共识
目前我们使用简单共识，当返回价格数据时，会包含所有节点数据的签名signs
大致结构为：
```js
{"price":123,"signs":[
{node:"node1",price:123,sign:node1_sign},
{node:"node1",price:123,sign:node2_sign}
]}
#signs为各节点的签名数据列表，最外层的price为＂签名列表signs＂里价格的平均值．
```
当返回数据时signs字段列表的长度小于集群节点数量的　　(n/2)+1 时，返回失败．

signs字段内容第三方应用来验证和使用．

### 节点钱包列表
http://62.234.169.68:8001/pub/stock/stats