package uni

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"log"
	"math"
	"math/big"
	"stock/utils"
	"strings"
	"time"
)

type TokenPairConf struct {
	//required
	PairAddre string
	//required
	TokenAddre string
	//required
	TokenDecimals int
	//eth or bsc
	ChainName string
	//token0  0; token1 1　由SubPairlog计算
	TokenIndex int
}

type TokenPrice struct {
	ID         uint
	PairAddre  string
	TokenAddre string `gorm:"index"`
	//token0  0; token1 1
	TokenIndex  int
	Reserve0    float64
	Reserve1    float64
	TokenPrice  float64
	BlockNumber uint64 `gorm:"index"`
	BlockTime   int64
	//eth or bsc
	ChainName string
	CreatedAt time.Time
}

//ethUsd 0xdac17f958d2ee523a2206206994597c13d831ec7
var chainUsdtDecimal = map[string]int{"bsc": 18, "eth": 6}

func getSyncLog(item types.Log, tpc *TokenPairConf) *TokenPrice {
	transferEvent := new(FanswapV2PairSync)
	err := pairAbi.UnpackIntoInterface(transferEvent, "Sync", item.Data)
	if err != nil {
		log.Fatal(err)
	}
	//transferEvent.From = common.HexToAddress(vLog.Topics[1].Hex())
	//transferEvent.To = common.HexToAddress(vLog.Topics[2].Hex())
	//log.Printf("res0: %v\n", BintTrunc2f(transferEvent.Reserve0,18,2))
	//log.Printf("res1: %v\n", BintTrunc2f(transferEvent.Reserve1,6,2))
	tp := new(TokenPrice)
	tp.PairAddre = tpc.PairAddre
	tp.TokenAddre = tpc.TokenAddre
	tp.TokenIndex = tpc.TokenIndex
	tp.ChainName = tpc.ChainName

	token0Decimal := chainUsdtDecimal[tpc.ChainName]
	token1Decimal := chainUsdtDecimal[tpc.ChainName] // usdtDicimal
	if tp.TokenIndex == 0 {
		token0Decimal = tpc.TokenDecimals
	} else {
		token1Decimal = tpc.TokenDecimals
	}
	//log.Println(token0Decimal,token1Decimal,transferEvent.Reserve0,transferEvent.Reserve1)
	tp.Reserve0 = BintTrunc2f(transferEvent.Reserve0, token0Decimal, 2)
	tp.Reserve1 = BintTrunc2f(transferEvent.Reserve1, token1Decimal, 2)
	if tp.TokenIndex == 0 {
		tp.TokenPrice = tp.Reserve1 / tp.Reserve0
	} else {
		tp.TokenPrice = tp.Reserve0 / tp.Reserve1
	}
	tp.TokenPrice = RoundPrice(tp.TokenPrice)
	tp.BlockNumber = item.BlockNumber
	//log.Printf("%v",tp)
	return tp
}

var pairAbi, _ = abi.JSON(strings.NewReader(string(FanswapV2PairABI)))

func initPairAbi() {
	//pairAbi, err := abi.JSON(strings.NewReader(string(FanswapV2PairABI)))
	//if err != nil {
	//	log.Fatal(err)
	//}
}

func getStartBlockForSubTokenPrice(tokenAddre, chainName string) (startBlock int64) {
	err := utils.Orm.Raw(`	select t.block_number from token_prices t
		where t.token_addre=?  order by t.id desc limit 1;`, tokenAddre).Scan(&startBlock).Error
	if err != nil {
		log.Fatal(err)
	}
	if startBlock == 0 {
		err = utils.Orm.Raw(`select t.id from block_prices t
where t.block_time>unix_timestamp('2021-06-01') order by t.id limit 1`).Scan(&startBlock).Error
		if startBlock == 0 {
			log.Fatal(err)
		}
	}
	return
}
func SubPairlog(tpc *TokenPairConf) {
	fromBlock := int64(0)
	if tpc.ChainName == "bsc" {
		fromBlock = 8540473
	} else {
		fromBlock = getStartBlockForSubTokenPrice(tpc.TokenAddre, tpc.ChainName)
	}
	pairAddressHex, tokenAddreHex := tpc.PairAddre, tpc.TokenAddre
	pairAddre := common.HexToAddress(pairAddressHex)
	tokenAddre := common.HexToAddress(tokenAddreHex)
	tokenIndex := 100
	fw, err := NewFanswapV2Pair(common.HexToAddress(pairAddressHex), utils.EthConn)
	if err != nil {
		log.Fatal(err)
	}
	var token0, token1 common.Address
	token0, err = fw.Token0(&bind.CallOpts{Pending: true})
	if err != nil {
		log.Fatal(err)
	}
	token1, err = fw.Token1(&bind.CallOpts{Pending: true})
	if err != nil {
		log.Fatal(err)
	}
	if token1 == tokenAddre {
		tokenIndex = 1
	}
	if token0 == tokenAddre {
		tokenIndex = 0
	}
	if tokenIndex == 100 {
		log.Fatal("token配制错误")
	}
	tpc.TokenIndex = tokenIndex
	log.Println("get tokenIndex:", tokenIndex)

	//log.Println(tokenIndex)
	logTransferSig := []byte("Sync(uint112,uint112)")
	logTransferSigHash := crypto.Keccak256Hash(logTransferSig)

	fromBlockNum := new(big.Int)
	//toBlockNum:=new(big.Int)
	query := ethereum.FilterQuery{
		Addresses: []common.Address{pairAddre},
		FromBlock: fromBlockNum.SetInt64(fromBlock),
		//ToBlock: toBlockNum.SetInt64(12676762),
		Topics: [][]common.Hash{[]common.Hash{logTransferSigHash}},
	}
	log.Println("getlog", fromBlock, pairAddressHex)
	logs1, err := utils.EthConn.FilterLogs(context.Background(), query)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("getlog len(logs)", len(logs1))
	tps := []*TokenPrice{}
	for _, item := range logs1 {
		//log.Println(item) // pointer to event log
		if item.Topics[0].Hex() == logTransferSigHash.Hex() {
			//log.Printf("Log Name: Sync\n")
			tps = append(tps, getSyncLog(item, tpc))
			fromBlock = int64(item.BlockNumber)
		}
	}
	if len(tps) > 0 {
		utils.Orm.CreateInBatches(tps, 2000)
	}

	log.Println("begin sublog fromBlock", fromBlock)
	logs := make(chan types.Log)
	query.FromBlock = fromBlockNum.SetInt64(fromBlock)
RETRY:
	sub, err := utils.EthConn.SubscribeFilterLogs(context.Background(), query, logs)
	defer sub.Unsubscribe()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("sublog", fromBlock, pairAddressHex)
	//count:=0
	for {
		//count++;
		//if count>3{
		//	return
		//}
		select {
		case err := <-sub.Err():
			time.Sleep(1 * time.Second)
			log.Println("subLogERR", err)
			goto RETRY
		case vLog := <-logs:
			//log.Println(vLog) // pointer to event log
			tps := []*TokenPrice{}

			if vLog.Topics[0].Hex() == logTransferSigHash.Hex() {
				tps = append(tps, getSyncLog(vLog, tpc))
			}
			if len(tps) > 0 {
				utils.Orm.CreateInBatches(tps, 2000)
			}
			fromBlock = int64(vLog.BlockNumber)
		}
	}
}

func BintTrunc2f(bigInt *big.Int, decimal int, point int) float64 {
	bint := big.NewInt(0)
	bint.Set(bigInt)
	bint.Quo(bint, big.NewInt(int64(math.Pow10(decimal-point))))
	bf := new(big.Float)
	bf.SetInt(bint).Quo(bf, big.NewFloat(math.Pow10(point)))
	tmpfloat, _ := bf.Float64()
	return tmpfloat
}
func RoundPrice(price float64) float64 {
	res, _ := decimal.NewFromFloat(price).Round(18).Float64()
	return res
}

//func getStartBlockSubPair(tokenAddre, chainName string) (startBlock int64) {
//	err := utils.Orm.Raw(`	select t.block_number from token_prices t
//		where t.token_addre=?  order by t.id desc limit 1;`, tokenAddre).Scan(&startBlock).Error
//	if err != nil {
//		log.Fatal(err)
//	}
//	if startBlock == 0 {
//		err = utils.Orm.Raw(`select t.id from block_prices t
//where t.block_time>unix_timestamp('2021-06-01') order by t.id limit 1`).Scan(&startBlock).Error
//		if startBlock == 0 {
//			log.Fatal(err)
//		}
//	}
//	return
//}

type TokenInfo struct {
	ID        int
	ChainName string
	Addre     string
	//decimal
	Point       uint8
	Name        string
	Symbol      string
	//is 1$ stableCoin: usdc or usdt
	IsUsd bool
	//是否需要将价格转换为usd价
	LoadUsd bool
	TotalSupply int
	UpdatedAt   time.Time
}

func (TokenInfo) CreateOrInit(chainName, addre string) *TokenInfo {
	ti := new(TokenInfo)
	err := utils.Orm.Order("id desc").First(ti, "addre=?", addre).Error
	if err != nil {
		ti = GetTokenInfoFromChain(chainName, addre)
		err = utils.Orm.Save(ti).Error
		if err != nil {
			log.Fatal(err)
		}
	}
	return ti
}
func GetTokenInfoFromChain(chainName, addre string) *TokenInfo {
	addre = strings.ToLower(addre)
	ethConn := utils.GetEthConn(chainName)
	defer ethConn.Close()
	token, err := NewFanswapV2ERC20(common.HexToAddress(addre), ethConn)
	if err != nil {
		log.Fatal("new erc20 token err:", err)
	}
	ti := new(TokenInfo)
	ti.ChainName = chainName
	ti.Addre = addre
	ti.Name, err = token.Name(nil)
	if err != nil {
		log.Fatal("get token Name err:", err)
	}
	ti.Symbol, _ = token.Symbol(nil)
	ti.Point, _ = token.Decimals(nil)
	ts, _ := token.TotalSupply(nil)
	ti.TotalSupply = int(bintTrunc(ts, int(ti.Point)))
	//log.Println(ti)
	return ti
}


type PairLog struct {
	Id uint
	//bsc eth polygon
	ChainName string `gorm:"type:varchar(256);index:idx_chain_name"`
	PairID uint
	Reserve0  string `gorm:"type:varchar(256);"`
	Reserve1  string `gorm:"type:varchar(256);"`
	Block uint64
	BlockTime uint64
	UpdatedAt time.Time
	TxHash string  `gorm:"type:varchar(256);"`
}
type UniPrice struct {
	Id uint
	Symbol    string `gorm:"type:varchar(50);index:idx_main,priority:1"`
	PairID uint `gorm:"index:idx_main,priority:2"`
	Price float64
	BlockTime uint64  `gorm:"index:idx_main,priority:3"`
	UpdatedAt time.Time
	//6 decimal
	Vol float64
}
type PairInfo struct {
	Id uint
	//bsc eth polygon
	ChainName string `gorm:"type:varchar(256);"`
	//uniswap pancake xxSwap
	ProjName  string `gorm:"type:varchar(256);"`
	Pair      string `gorm:"type:varchar(50);index:idx_main,priority:1"`
	Symbol    string `gorm:"type:varchar(50);"`
	Token0    string `gorm:"type:varchar(256);"`
	Token1    string `gorm:"type:varchar(256);"`
	Point0    uint8
	Point1    uint8
	Reserve0  string `gorm:"type:varchar(256);"`
	Reserve1  string `gorm:"type:varchar(256);"`
	Price0    float64
	Price1    float64
	Symbol0   string `gorm:"type:varchar(256);index:idx_symbol0"`
	Symbol1   string `gorm:"type:varchar(256);index:idx_symbol1"`
	IsUsd0    bool
	IsUsd1    bool
	LoadUsd0  bool
	LoadUsd1  bool
	Vol0      float64
	Vol1      float64
	//lp$
	VolUsd    float64
	UpdatedAt time.Time
	BlockNum  int64
	BlockTime uint32 `gorm:"index:idx_main,priority:2"`

	//IsInternal bool   `gorm:"index:idx_ti,priority:2"`
	//PathID     int    `gorm:"type:tinyint;uniqueIndex:idx_node_t,priority:3"`
	//Pathstr    string `gorm:"type:varchar(256);`
}

type SubPairConfig struct {
	ProjName string
	Pair     string
	Symbol	 string
	ChainName	 string
}
func (pinfo *PairInfo)GetPrice() float64{
	if pinfo.Symbol==pinfo.Symbol0{
		return  pinfo.Price0
	}
	if pinfo.Symbol==pinfo.Symbol1{
		return  pinfo.Price1
	}
	return 0
}
func (pinfo *PairInfo)getPairInfo(pcfg SubPairConfig, ethConn *ethclient.Client){
	log.Println("getPairInfo",pcfg)
	pAddre:=strings.ToLower(pcfg.Pair)
	pinfo.Pair = pAddre
	fw:=pinfo.getConstractPair(ethConn)
	var token0, token1 common.Address
	var err error
	token0, err = fw.Token0(&bind.CallOpts{Pending: true})
	if err != nil {
		log.Fatal(err)
	}
	token1, err = fw.Token1(&bind.CallOpts{Pending: true})
	if err != nil {
		log.Fatal(err)
	}

	pinfo.ChainName = pcfg.ChainName
	pinfo.ProjName = pcfg.ProjName
	pinfo.Pair = pAddre
	pinfo.Symbol = pcfg.Symbol
	pinfo.Token0 = hexAddres(token0)
	pinfo.Token1 = hexAddres(token1)

	//当前价
	reservs, _ := fw.GetReserves(nil)
	//log.Println(reservs)
	//pinfo.caculateReservePrice(reservs.Reserve0, reservs.Reserve1)
	pinfo.BlockTime = reservs.BlockTimestampLast
}
type ps  map[string]*PairInfo

func (pinfos ps)initHistoryLog( ethConn *ethclient.Client,counts int)(lcount int){

	logTransferSig := []byte("Sync(uint112,uint112)")
	logTransferSigHash := crypto.Keccak256Hash(logTransferSig)

	lastBlock,err:= utils.EthLastBlock(ethConn)
	if err != nil {
		log.Fatal("lastBlock err",err)
	}
	pinfosMap := map[string]*PairInfo{}
	pairAddres := []common.Address{}
	chainName:=""
	for _, pinfo := range pinfos {
		pinfosMap[pinfo.Pair] = pinfo
		pairAddres = append(pairAddres, common.HexToAddress(pinfo.Pair))

		chainName=pinfo.ChainName
	}
	msgid:=fmt.Sprintf("initHistoryLog %s",chainName)

	fromBlock:=lastBlock-counts
	ftime:=time.Unix( int64(utils.EthBlockTime(uint64(fromBlock),ethConn)),0)
	log.Println(msgid,"from block",fromBlock,"blockTime:",ftime)

	for i:=lastBlock-counts; i<lastBlock;i+=2000{
		query := ethereum.FilterQuery{
			Addresses:pairAddres,
			FromBlock: big.NewInt(int64(i)),
			ToBlock: big.NewInt(int64(i+2000)),
			Topics: [][]common.Hash{[]common.Hash{logTransferSigHash}},
		}

		log.Println(msgid ," getlog fromBlock",i)
		logs1, err := ethConn.FilterLogs(context.Background(), query)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println(msgid ," getlog len(logs)", len(logs1))
		if len(logs1)==0{
			continue
		}
		plogs := []*PairLog{}
		ups := []*UniPrice{}
		for _, item := range logs1 {
			//log.Println(item) // pointer to event log
			if item.Topics[0].Hex() == logTransferSigHash.Hex() {
				event:=parseSyncEvent(item)
				plog,up:=syncEventHanlder(event,pinfos,true,ethConn)
				plogs=append(plogs,plog)
				ups=append(ups,up...)
			}
		}
		utils.Orm.CreateInBatches(plogs,2000)
		utils.Orm.CreateInBatches(ups,2000)
		lcount+=len(plogs)
	}

	for _, pi := range pinfos {
		utils.Orm.Save(pi)
	}
	log.Println("finish init history logs", msgid, "saved:",lcount)
	return lcount
}
func (pinfo *PairInfo)getConstractPair( ethConn *ethclient.Client)(conPair *FanswapV2Pair){
	fw, err := NewFanswapV2Pair(common.HexToAddress(pinfo.Pair), ethConn)
	if err != nil {
		log.Fatal(err)
	}
	return fw
}

func (pinfo *PairInfo)CreateOrInit(pcfg SubPairConfig,ethConn *ethclient.Client){
	err:=utils.Orm.First(pinfo,"pair=?", strings.ToLower(pcfg.Pair)).Error
	if err != nil {
		pinfo.getPairInfo(pcfg,ethConn)
		err=utils.Orm.Save(pinfo).Error
		if err != nil {
			log.Fatal("pairInfo save err",err,pcfg)
		}
	}
	//init token info
	ti := TokenInfo{}.CreateOrInit(pcfg.ChainName, pinfo.Token0)
	pinfo.Point0 = ti.Point
	pinfo.Symbol0 = ti.Symbol
	pinfo.IsUsd0=ti.IsUsd
	pinfo.LoadUsd0=ti.LoadUsd
	ti = TokenInfo{}.CreateOrInit(pcfg.ChainName, pinfo.Token1)
	pinfo.Point1 = ti.Point
	pinfo.Symbol1 = ti.Symbol
	pinfo.IsUsd1=ti.IsUsd
	pinfo.LoadUsd1=ti.LoadUsd

	utils.Orm.Save(pinfo)
}

//sub chainName's all uni-pair
//param  init true加载最近的历史事件；从上次中断的块号处，开始加载事件
func SubPair(chainName string, pairCfgs []SubPairConfig, init bool, infuraID string) {

	msgId := fmt.Sprintf("%s %s", "SubPair", chainName)
	log.Println("begin ", msgId,"init",init, pairCfgs)
	pinfos := map[string]*PairInfo{}
	pairAddres := []common.Address{}
	ethConn := utils.GetEthConn(chainName)
	for _, paircf := range pairCfgs {
		pinfo := new(PairInfo)
		pinfo.CreateOrInit(paircf,ethConn)
		pinfos[pinfo.Pair] = pinfo
		pairAddres = append(pairAddres, common.HexToAddress(pinfo.Pair))
	}

	if init {
		ps(pinfos).initHistoryLog(ethConn,5000)
	}


	fromBlock := int64(0)
	pinfo := new(PairLog)
	err := utils.Orm.Order("id desc").Where("chain_name=?", chainName).First(pinfo).Error
	if err == nil {
		fromBlock = int64(pinfo.Block)
	}else if  !errors.Is(err,gorm.ErrRecordNotFound) {
		log.Fatal(msgId, "get chain last block err", err)
	}

	if fromBlock==0{
		lastBlock,err:=utils.EthLastBlock(ethConn)
		if err != nil {
			log.Fatal(err)
		}
		fromBlock=int64(lastBlock)
	}
	logTransferSig := []byte("Sync(uint112,uint112)")
	logTransferSigHash := crypto.Keccak256Hash(logTransferSig)
	query := ethereum.FilterQuery{
		Addresses:pairAddres,
		FromBlock: big.NewInt(fromBlock),
		Topics: [][]common.Hash{[]common.Hash{logTransferSigHash}},
	}
RETRY:
	query.FromBlock = big.NewInt(fromBlock)
	logs := make(chan types.Log)

	log.Println("begin sublog ",msgId, "fromBlock", fromBlock)
	sub, err := ethConn.SubscribeFilterLogs(context.Background(), query, logs)
	defer sub.Unsubscribe()
	if err != nil {
		time.Sleep(1 * time.Second)
		log.Println(msgId,err)
		goto RETRY
	}
	log.Println("watch pair log ",msgId, "fromBlock", fromBlock)
	for {
		select {
		case err := <-sub.Err():
			time.Sleep(1 * time.Second)
			log.Println(msgId,"subLogERR", err)
			goto RETRY
		case vLog := <-logs:
			//log.Println(vLog) // pointer to event log

			if vLog.Topics[0].Hex() == logTransferSigHash.Hex() {
				event:=parseSyncEvent(vLog)
				plog,up:=syncEventHanlder(event,pinfos,false,ethConn)

				dbtx:=utils.Orm.Begin()
				err=dbtx.Save(plog).Error
				if err == nil {
					err=dbtx.Save(up).Error
					if err == nil {
						for _, pi := range pinfos {
							//save pair_info which update in 5 secs
							if pi.UpdatedAt.After(time.Now().Add(-5*time.Second)){
								err=dbtx.Save(pi).Error
							}
							if err != nil {
								continue
							}
						}
						dbtx.Commit()
					}
				}
				if err != nil {
					dbtx.Rollback()
				}
			}
			fromBlock = int64(vLog.BlockNumber) + 1
		}
	}
}


func getPairSyncLog(item types.Log, pInfos map[string]*PairInfo) *PairLog {
	plog:=new(PairLog)
	pinfo := *(pInfos[hexAddres(item.Address)])
	
	plog.PairID=pinfo.Id
	plog.Block = item.BlockNumber
	plog.BlockTime = uint64(time.Now().Unix())
	plog.TxHash=item.TxHash.String()
	transferEvent := new(FanswapV2PairSync)
	err := pairAbi.UnpackIntoInterface(transferEvent, "Sync", item.Data)
	if err != nil {
		log.Fatal(err)
	}
	plog.Reserve0,plog.Reserve1=transferEvent.Reserve0.String(), transferEvent.Reserve1.String()
	return plog
}

type syncEvent struct {
	Address common.Address
	BlockNumber uint64
	// hash of the transaction
	TxHash common.Hash
	Reserve0 *big.Int
	Reserve1 *big.Int
}
func parseSyncEvent(item types.Log) *syncEvent {
	event := new(syncEvent)
	transferEvent := new(FanswapV2PairSync)
	err := pairAbi.UnpackIntoInterface(transferEvent, "Sync", item.Data)
	if err != nil {
		log.Fatal(err)
	}
	event.Address = item.Address
	event.BlockNumber = item.BlockNumber
	event.TxHash = item.TxHash
	event.Reserve0 = transferEvent.Reserve0
	event.Reserve1 = transferEvent.Reserve1
	return event
}
func getUsdtPrice(projName,symbol string, pInfos map[string]*PairInfo)(price float64 ){
	for _, info := range pInfos {
		if info.ProjName==projName && (info.IsUsd0 || info.IsUsd1) && (info.Symbol1==symbol || info.Symbol0==symbol ){
			if info.Symbol0==symbol{
				return info.Price0
			}else{
				return info.Price1
			}
			//up:=new(UniPrice)
			//err:=utils.Orm.Order("id desc").Find(up,UniPrice{PairID: info.Id,Symbol: symbol}).Error
			//if err != nil {
			//	log.Fatal("getUsdtPrice",err)
			//}
			//return up.Price
		}
	}
	return 0
}
func syncEventHanlder(event *syncEvent, pInfos map[string]*PairInfo,useChainTime bool,ethConn *ethclient.Client) (*PairLog,[]*UniPrice) {
	pinfo := pInfos[hexAddres(event.Address)]
	if pinfo == nil {
		log.Fatalf("pair %x not int", event.Address)
	}
	plog := new(PairLog)
	plog.PairID = pinfo.Id
	plog.ChainName = pinfo.ChainName
	plog.Block = event.BlockNumber
	if useChainTime {
		plog.BlockTime = utils.EthBlockTime(event.BlockNumber, ethConn)
	} else {
		plog.BlockTime = uint64(time.Now().Unix())
	}

	plog.TxHash = event.TxHash.String()
	plog.Reserve0, plog.Reserve1 = event.Reserve0.String(), event.Reserve1.String()

	reserv0, reserv1 := event.Reserve0, event.Reserve1
	point0, point1 := pinfo.Point0, pinfo.Point1
	reserv0 = reserv0.Mul(reserv0, big.NewInt(int64(math.Pow10(18-int(point0)))))
	reserv1 = reserv1.Mul(reserv1, big.NewInt(int64(math.Pow10(18-int(point1)))))

	price0 := big.NewFloat(0).SetInt(reserv1)
	price0 = price0.Quo(price0, big.NewFloat(0).SetInt(reserv0))
	price1 := big.NewFloat(1)
	price1 = price1.Quo(price1, price0)
	p0, _ := price0.SetPrec(10).Float64()
	p1, _ := price1.SetPrec(10).Float64()
	vol0, vol1 := BintTrunc2f(reserv0, 18, 6), BintTrunc2f(reserv1, 18, 6)
	volUsd:=0.0
	if  pinfo.IsUsd0{
		volUsd=vol0
	}
	if  pinfo.IsUsd1{
		volUsd=vol1
	}
	if !pinfo.IsUsd0 && !pinfo.IsUsd1 {
		if  pinfo.LoadUsd1{
			tprice:= getUsdtPrice(pinfo.ProjName,pinfo.Symbol1,pInfos)
			log.Println("loadUsd pirce for pair",pinfo.Id ,pinfo.Symbol1,tprice)
			p0=p0*tprice
			volUsd=tprice*vol1
		}
		if pinfo.LoadUsd0{
			tprice:= getUsdtPrice(pinfo.ProjName,pinfo.Symbol0,pInfos)
			log.Println("loadUsd pirce for pair",pinfo.Id ,pinfo.Symbol0,tprice)
			p1=p1*tprice
			volUsd=tprice*vol0
		}
	}

	up := new(UniPrice)
	up.Symbol = pinfo.Symbol0
	up.PairID = pinfo.Id
	up.Price = p0
	up.BlockTime = plog.BlockTime
	up.Vol = vol0

	up1 := *up
	up1.Symbol = pinfo.Symbol1
	up1.Price = p1
	up.Vol = vol1

	pinfo.Vol0=vol0
	pinfo.Vol1=vol1
	pinfo.VolUsd=volUsd
	pinfo.Reserve0=plog.Reserve0
	pinfo.Reserve1=plog.Reserve0
	pinfo.Price0=p0
	pinfo.Price1=p1
	pinfo.UpdatedAt=time.Now()
	return plog, []*UniPrice{up, &up1}
}

//func caculateReservePrice(resv0,resv1 *big.Int,point0,point1 uint8)(price0,pirce1 float64){
//	reserv0 := bintTrunc(resv0, int(point0))
//	reserv1 := bintTrunc(resv1, int(point1))
//	return float64(reserv1) / float64(reserv0),float64(reserv0) / float64(reserv1)
//}

//Trunc 6bit for store db
//func bintTrunc6(bigInt *big.Int) int64 {
//	return bintTrunc(bigInt, 6)
//}
func bintTrunc(bigInt *big.Int, decimal int) int64 {
	bint := big.NewInt(0)
	bint.Set(bigInt)
	return bint.Quo(bint, big.NewInt(int64(math.Pow10(decimal)))).Int64()
}
func hexAddres(address common.Address) string {
	return fmt.Sprintf("0x%x", address)
}
