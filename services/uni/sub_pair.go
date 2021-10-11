package uni

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
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
	TotalSupply int
	UpdatedAt   time.Time
}

func (TokenInfo) CreateOrInit(chainName, infuraID, addre string) *TokenInfo {
	ti := new(TokenInfo)
	err := utils.Orm.Order("id desc").First(ti, "addre=?", addre).Error
	if err != nil {
		ti = GetTokenInfoFromChain(chainName, infuraID, addre)
		err = utils.Orm.Save(ti).Error
		if err != nil {
			log.Fatal(err)
		}
	}
	return ti
}
func GetTokenInfoFromChain(chainName, infuraID, addre string) *TokenInfo {
	addre = strings.ToLower(addre)
	ethConn := utils.GetEthConn(chainName, infuraID)
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

type PairInfo struct {
	PairEvent
}
type PairEvent struct {
	Id uint
	//bsc eth polygon
	ChainName string `gorm:"type:varchar(256);"`
	//uniswap pancake xxSwap
	ProjName  string `gorm:"type:varchar(256);"`
	Pair      string `gorm:"type:varchar(50);index:idx_main,priority:1"`
	Token0    string `gorm:"type:varchar(256);"`
	Token1    string `gorm:"type:varchar(256);"`
	Point0    uint8  `gorm:"-"`
	Point1    uint8  `gorm:"-"`
	Reserve0  string `gorm:"type:varchar(256);"`
	Reserve1  string `gorm:"type:varchar(256);"`
	Price0    float64
	Price1    float64
	Symbol0   string `gorm:"type:varchar(256);index:idx_symbol0"`
	Symbol1   string `gorm:"type:varchar(256);index:idx_symbol1"`
	Vol0      float64
	Vol1      float64
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
}

//sub chainName's all uni-pair
//param  init true加载最近的历史事件；从上次中断的块号处，开始加载事件
func SubPair(chainName string, pairCfgs []SubPairConfig, init bool, infuraID string) {
	msgId := fmt.Sprintf("%s %s", "SubPair", chainName)
	log.Println("begin ", msgId, pairCfgs)

	utils.Orm.AutoMigrate(PairEvent{})
	utils.Orm.AutoMigrate(PairInfo{})
	utils.Orm.AutoMigrate(TokenInfo{})
	pinfos := map[string]*PairEvent{}
	pairAddres := []common.Address{}
	ethConn := utils.GetEthConn(chainName, infuraID)
	for _, paircf := range pairCfgs {
		pairAddressStr := strings.ToLower(paircf.Pair)
		pairAddre := common.HexToAddress(pairAddressStr)
		fw, err := NewFanswapV2Pair(pairAddre, ethConn)
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
		pinfo := new(PairEvent)
		pinfo.ChainName = chainName
		pinfo.ProjName = paircf.ProjName
		pinfo.Pair = pairAddressStr
		pinfo.Token0 = hexAddres(token0)
		pinfo.Token1 = hexAddres(token1)

		pinfos[pairAddressStr] = pinfo
		pairAddres = append(pairAddres, common.HexToAddress(pairAddressStr))

		//init token info
		ti := TokenInfo{}.CreateOrInit(chainName, infuraID, pinfo.Token0)
		pinfo.Point0 = ti.Point
		pinfo.Symbol0 = ti.Symbol
		ti = TokenInfo{}.CreateOrInit(chainName, infuraID, pinfo.Token1)
		pinfo.Point1 = ti.Point
		pinfo.Symbol1 = ti.Symbol

		//当前价
		reservs, _ := fw.GetReserves(nil)
		//log.Println(reservs)
		pinfo.caculateReservePrice(reservs.Reserve0, reservs.Reserve1)
		pinfo.BlockTime = reservs.BlockTimestampLast

		pairi := new(PairInfo)
		pairi.PairEvent = *pinfo

		err = utils.Orm.FirstOrCreate(pairi, "pair=?", pairi.Pair).Assign(pairi).Error
		if err != nil {
			log.Println("msgId  save pairInfo err:", err)
		}
	}

	logTransferSig := []byte("Sync(uint112,uint112)")
	logTransferSigHash := crypto.Keccak256Hash(logTransferSig)

	fromBlock := int64(0)
	if !init {
		pinfo := new(PairEvent)
		err := utils.Orm.Order("id desc").Where("chain_name=?", chainName).First(pinfo).Error
		if err == nil {
			fromBlock = pinfo.BlockNum
		}
	}
	lastBlockNum := uint64(0)
	if fromBlock == 0 {
		//bsc maximum block range: 5000
		lastBlock, err := ethConn.BlockNumber(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		lastBlockNum = lastBlock
		fromBlock = int64(lastBlock - 4000)
	}
	fromBlockNum := new(big.Int)
	//toBlockNum:=new(big.Int)
	query := ethereum.FilterQuery{
		Addresses: pairAddres,
		FromBlock: fromBlockNum.SetInt64(fromBlock),
		//ToBlock: toBlockNum.SetInt64(12676762),
		Topics: [][]common.Hash{[]common.Hash{logTransferSigHash}},
	}

	log.Println("getlog", fromBlock, pairAddres)
	logs1, err := ethConn.FilterLogs(context.Background(), query)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("getlog len(logs)", len(logs1))
	tps := []*PairEvent{}
	for _, item := range logs1 {
		//log.Println(item) // pointer to event log
		if item.Topics[0].Hex() == logTransferSigHash.Hex() {
			//log.Printf("Log Name: Sync\n")
			tps = append(tps, getPairSyncLog(item, pinfos))
			fromBlock = int64(item.BlockNumber)
		}
	}
	//save histrory event
	historyStat := map[string]int{}
	if len(tps) > 0 {
		//历史数据获取blockTime
		for _, tp := range tps {
			historyStat[tp.Pair] += 1
			header, _ := ethConn.HeaderByNumber(context.Background(), big.NewInt(tp.BlockNum))
			tp.BlockTime = uint32(header.Time)
		}
		utils.Orm.CreateInBatches(tps, 2000)
	}
	//没有新日志数据时，存储当前价格
	for _, pi := range pinfos {
		if historyStat[pi.Pair] == 0 {
			pinfo := *pi
			pinfo.BlockNum = int64(lastBlockNum)
			pinfo.BlockTime = uint32(time.Now().Unix())
			utils.Orm.Save(&pinfo)
		}
	}
	log.Println("finish", msgId, "init logs")
RETRY:
	log.Println("begin sublog fromBlock", fromBlock)
	logs := make(chan types.Log)
	query.FromBlock = fromBlockNum.SetInt64(fromBlock)

	sub, err := ethConn.SubscribeFilterLogs(context.Background(), query, logs)
	defer sub.Unsubscribe()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("sublog", fromBlock)
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
			tps := []*PairEvent{}

			if vLog.Topics[0].Hex() == logTransferSigHash.Hex() {
				tps = append(tps, getPairSyncLog(vLog, pinfos))
			}
			if len(tps) > 0 {
				utils.Orm.CreateInBatches(tps, 2000)
			}
			fromBlock = int64(vLog.BlockNumber) + 1
		}
	}
}
func getPairSyncLog(item types.Log, pInfos map[string]*PairEvent) *PairEvent {
	transferEvent := new(FanswapV2PairSync)
	err := pairAbi.UnpackIntoInterface(transferEvent, "Sync", item.Data)
	if err != nil {
		log.Fatal(err)
	}
	//log.Println(hexAddres(item.Address), transferEvent.Raw.BlockNumber)
	chainPinfo := *(pInfos[hexAddres(item.Address)])
	chainPinfo.caculateReservePrice(transferEvent.Reserve0, transferEvent.Reserve1)
	//chainPinfo.Price0,chainPinfo.Price1=caculateReservePrice(transferEvent.Reserve0,transferEvent.Reserve1,chainPinfo.Point0,chainPinfo.Point1)
	//resv0 := bintTrunc(transferEvent.Reserve0, int(chainPinfo.Point0))
	//resv1 := bintTrunc(transferEvent.Reserve1, int(chainPinfo.Point1))
	//chainPinfo.Price0 = float64(resv1) / float64(resv0)
	//chainPinfo.Price1 = float64(resv0) / float64(resv1)
	chainPinfo.BlockNum = int64(item.BlockNumber)
	chainPinfo.BlockTime = uint32(time.Now().Unix())
	//log.Printf("%v",tp)
	return &chainPinfo
}
func (pi *PairEvent) caculateReservePrice(bint0, bint1 *big.Int) {
	pi.Reserve0, pi.Reserve1 = bint0.String(), bint1.String()
	point0, point1 := pi.Point0, pi.Point1
	reserv0 := bint0.Mul(bint0, big.NewInt(int64(math.Pow10(18-int(point0)))))
	reserv1 := bint1.Mul(bint1, big.NewInt(int64(math.Pow10(18-int(point1)))))
	//reserv1:=pi.Reserve0* int64(math.Pow10(18-int(point1)))

	price0 := big.NewFloat(0).SetInt(reserv1)
	price0 = price0.Quo(price0, big.NewFloat(0).SetInt(reserv0))
	price1 := big.NewFloat(1)
	price1 = price1.Quo(price1, price0)

	pi.Price0, _ = price0.SetPrec(10).Float64()
	pi.Price1, _ = price1.SetPrec(10).Float64()

	pi.Vol0, pi.Vol1 = BintTrunc2f(reserv1, 18, 6), BintTrunc2f(reserv0, 18, 6)
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
