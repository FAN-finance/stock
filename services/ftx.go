package services

import (
	"fmt"
	"gorm.io/gorm"
	"log"
	"math"
	"regexp"
	"stock/utils"
	"strconv"
	"time"
)

func CacuBullPrice(lastAjustPriceBull, lastAjustPric, curPric float64, item_type string) float64 {
	//return lastAjustPriceBull * ((curPric-lastAjustPric)/lastAjustPric*3 + 1)
	return lastAjustPriceBull * ((curPric-lastAjustPric)/lastAjustPric*float64(ftxMultipleMap[item_type]) + 1)
}
func CacuBearPrice(lastAjustPriceBear, lastAjustPric, curPric float64, item_type string) float64 {
	//return lastAjustPriceBull * ((curPric-lastAjustPric)/lastAjustPric*3 + 1)
	return lastAjustPriceBear * (1 - (curPric-lastAjustPric)/lastAjustPric*float64(ftxMultipleMap[item_type]))
}
func getMultipleFromCoinType(coinType string) int {
	preg := regexp.MustCompile("(\\d+)x")
	items := preg.FindStringSubmatch(coinType)
	multi, _ := strconv.Atoi(items[1])
	return multi
}

var ftxList = []string{"mvi", "btc", "eth", "vix", "gold", "eur", "ndx", "govt"}

//var ftxList=[]string{"btc" , "eth" ,  "gold"}
var ftxMultipleMap = map[string]int{
	"mvi": 2,
	"btc": 3,
	"eth": 3,
	"vix": 3,
	//"ust":  20,
	"gold": 10,
	"eur":  20,
	"ndx":  10,
	"govt": 20,
}

//btc 3x：110054.79
//eth 3x：7900.56
//vix 3x：53.7
//govt 20x：5268
//gold 10x：19022.8
//eur 20x：244.66
//ndx 10x：136488
var ftxAJInitValueMap = map[string]float64{
	"mvi2x": 99,
	//"btc3x": 110054.79,
	//"eth3x": 7900.56,
	"btc3x": 102810.79,
	"eth3x": 6411,
	"vix3x": 53.7,
	//"ust":  20,
	"gold10x": 19022.8,
	"eur20x":  244.66,
	"ndx10x":  136488,
	"govt20x": 5268,
	"mvi2s":   112.0,
	"btc3s":   98265.0,
	"eth3s":   6205.8,
	//"mvi2s":1626825600,
	//"btc3s":1626825600,
	//"eth3s":1626825600,
	//"vix3s":1626825600,
	//"gold10s":1626825600,
	//"eur20s":1626825600,
	//"ndx10s":1626825600,
	//"govt20s":1626825600,
}

var ftxXMap = map[int]float64{
	2:  15,
	3:  10,
	10: 3,
	20: 1.5,
}

func getFtxBullType(item_type string) string {
	return fmt.Sprintf("%s%dx", item_type, ftxMultipleMap[item_type])
}
func getFtxBearType(item_type string) string {
	return fmt.Sprintf("%s%ds", item_type, ftxMultipleMap[item_type])
}

func getFtxXRate(itemType string) float64 {
	return ftxXMap[ftxMultipleMap[itemType]]
	//
	//itemType = strings.ToLower(itemType)
	//xArr := []string{"2x", "3x", "10x", "20x"}
	//for _, item := range xArr {
	//	if strings.Contains(itemType, item) {
	//		return ftxXMap[item]
	//	}
	//}
	//return 10
}

/*ndx10x vix3x*/
/**/

var FirstBull, LastBullAJ = map[string]*CoinBull{}, map[string]*CoinBull{}

func setFirstBull(coinType string) {
	if cointTypeInitTime[coinType] > int(time.Now().Unix()) {
		log.Println("setFirstBull 初始化禁用", coinType)
		return
	}
	firstBull := new(CoinBull)
	err := utils.Orm.Where("coin_type=?", coinType).First(firstBull).Error
	if err != nil {
		log.Fatal(err)
	}
	FirstBull[coinType] = firstBull
	//return firstBull
}

func setLastBullAJ(coinType string) {
	if cointTypeInitTime[coinType] > int(time.Now().Unix()) {
		log.Println("setLastBullAJ 初始化禁用", coinType)
		return
	}

	lastAj := new(CoinBull)
	err := utils.Orm.Order("timestamp desc").First(lastAj, "coin_type=? and is_ajust_point=?", coinType, 1).Error
	if err != nil {
		log.Fatal(err)
	}

	if lastAj.Rebalance == 0 {
		lastAj.Rebalance = lastAj.Bull
	}
	log.Println("lastaj %v", lastAj)
	LastBullAJ[coinType] = lastAj
	//return lastAj
}

func LastBullTimeStamp(coinType string) int64 {
	cb := new(CoinBull)
	err := utils.Orm.Order("timestamp desc").Where("coin_type=?", coinType).First(cb).Error
	if err != nil {
		log.Fatal(err)
	}
	log.Println("lastbullTime %v", cb)
	return cb.Timestamp
}
func LastBullPriceID() int {
	cb := new(CoinBull)
	err := utils.Orm.Order("price_id desc").First(cb).Error
	if err != nil {
		log.Fatal(err)
	}
	log.Println("LastBullPriceID %v", cb)
	return cb.PriceID
}

//从coinGecko数据coins表初始化　bull
func initCoinBull(coinType string) {
	var err error
	utils.Orm.AutoMigrate(CoinBull{})
	bullCount := int64(0)
	utils.Orm.Model(CoinBull{}).Where("coin_type=?", coinType).Count(&bullCount)
	if bullCount == 0 {
		firstCoin := new(Coin)
		err = utils.Orm.Model(Coin{}).First(firstCoin).Error
		if err != nil {
			log.Fatal(err)
		}
		cb := new(CoinBull)
		cb.CoinType = coinType
		coin_name := coinType[:3]
		rawPrice := "1"
		if coin_name == "eth" {
			rawPrice = firstCoin.Eth
		}
		cb.RawPrice = getCoinUSdPriceFromStr(rawPrice, firstCoin.Usd)
		cb.Rebalance = 10000
		cb.Bull = cb.Rebalance
		cb.RawChange = 0
		cb.BullChange = 0
		cb.IsAjustPoint = true
		//cb.ID = 1
		err = utils.Orm.Save(cb).Error
		if err != nil {
			log.Fatal(err)
		}
	}
}
func getCoinType(itemType string, bullOrBear string) string {
	coinType := ""
	if bullOrBear == "bull" {
		coinType = getFtxBullType(itemType)
	} else if bullOrBear == "bear" {
		coinType = getFtxBearType(itemType)
	} else {
		log.Fatalln("err item_type", itemType)
	}
	return coinType
}

//1627005600 2021-07-23T02:00:00.000Z
//1626825600 2021-07-21
var cointTypeInitTime = map[string]int{
	"mvi2x":   0,
	"btc3x":   0,
	"eth3x":   0,
	"vix3x":   0,
	"govt20x": 0,
	"gold10x": 0,
	"eur20x":  0,
	"ndx10x":  0,
	"mvi2s":   1627005600,
	"btc3s":   1627005600,
	"eth3s":   1627005600,
	"vix3s":   3626825600,
	"gold10s": 2626825600,
	"eur20s":  2626825600,
	"ndx10s":  2626825600,
	"govt20s": 2626825600,
}

//从twelvedata数据market_pirces表初始化　第一个bull
func initCoinBullFromTw(itemType string, bullOrBear string) bool {
	coinType := getCoinType(itemType, bullOrBear)
	if cointTypeInitTime[coinType] > int(time.Now().Unix()) {
		log.Println("初始化禁用", coinType)
		return false
	}

	var err error
	//utils.Orm.AutoMigrate(CoinBull{})
	bullCount := int64(0)
	utils.Orm.Model(CoinBull{}).Where("coin_type=?", coinType).Count(&bullCount)
	if bullCount == 0 {
		firstPrice := new(MarketPrice)
		err = utils.Orm.Model(MarketPrice{}).Order("timestamp").Where("item_type=? and timestamp>=?", itemType, cointTypeInitTime[coinType]).First(firstPrice).Error
		if err != nil {
			log.Fatal(err)
		}
		cb := new(CoinBull)
		cb.CoinType = coinType
		cb.RawPrice = firstPrice.Price
		cb.TargetPriceOfCheckpoint = firstPrice.Price
		cb.PriceID = int(firstPrice.ID)
		cb.Rebalance = ftxAJInitValueMap[coinType]
		cb.Bull = cb.Rebalance
		cb.RawChange = 0
		cb.BullChange = 0
		cb.IsAjustPoint = true
		//cb.ID = 1
		err = utils.Orm.Save(cb).Error
		if err != nil {
			log.Fatal(err)
		}
		return true
	}
	return false
}

//for coingecko
func SetAllBulls(coinType string) {
	initCoinBull(coinType)
	setFirstBull(coinType)
	setLastBullAJ(coinType)
	lastBullTime := LastBullTimeStamp(coinType)
	//lastBullTime,_=SetBullsFromID(lastBullTime,coinType)
	//return
	proc := func() error {
		lastId, err := SetBullsFromID(lastBullTime, coinType)
		if lastId > 0 {
			lastBullTime = lastId
		}
		return err
	}
	utils.IntervalSync("SetAllBull"+coinType, 10, proc)
}

//更新twelvedata数据源bull数据
//initAll 有新标杆币时,使用initAll=false初始化新标杆币数据
func SetAllBullsFromTw(initAll bool) {
	ftxItmes := []string{""}
	//ftxItmes = []string{"eth"}
	utils.Orm.Model(MarketPrice{}).Where("item_type in (?)", ftxList).Distinct().Pluck("item_type", &ftxItmes)
	if len(ftxItmes) == 0 {
		log.Println("none ftxItmes")
		return
	}
	for _, itemType := range ftxItmes {
		ok := initCoinBullFromTw(itemType, "bull")
		coin_type := getCoinType(itemType, "bull")
		setFirstBull(coin_type)
		setLastBullAJ(coin_type)
		if ok && !initAll { //非整表coinBull清空初始化时, 可初始化特定标杆币的历史数据
			log.Println("初始化历史数据:", coin_type)
			_, err := SetBullsForTw(cointTypeInitTime[coin_type], []string{itemType}, coin_type)
			if err != nil {
				log.Fatalln(err)
			}
		}
	}
	for _, itemType := range ftxItmes {
		ok := initCoinBullFromTw(itemType, "bear")
		coin_type := getCoinType(itemType, "bear")
		setFirstBull(coin_type)
		setLastBullAJ(coin_type)
		if ok && !initAll { //非整表coinBull清空初始化时, 可初始化特定标杆币的历史数据
			log.Println("初始化历史数据:", coin_type)
			_, err := SetBullsForTw(cointTypeInitTime[coin_type], []string{itemType}, coin_type)
			if err != nil {
				log.Fatalln(err)
			}
		}
	}

	log.Println("开始处理所有ftx")
	lastStat := LastBullPriceID()

	//lastStat:=1468544
	//lastStat,_=SetBullsForTw(lastStat,ftxItmes, "")
	//log.Println(lastStat)
	//return

	proc := func() error {
		lastId, err := SetBullsForTw(lastStat, ftxItmes, "")
		if err == nil {
			lastStat = lastId
		}
		return err
	}
	utils.IntervalSync("SetAllBullsFromTw", 20, proc)
}

//生成杠杆币数据　twelvedata
//specCoinType 用于指定初始化特定ftx;
func SetBullsForTw(lastStat int, ftxList []string, specCoinType string) (int, error) {
	//initCoinBull(coinType)
	//setFirstBull(coinType)
	//setLastBullAJ(coinType)
	coins := []MarketPrice{}
	idx_add := 0
	proc := func(tx *gorm.DB, batch int) error {
		for _, coin := range coins {
			idx_add++
			//log.Println(coin.ID,batch)
			cbs := []*CoinBull{}
			for _, bullOrBear := range []string{"bull", "bear"} {
				coinType := getCoinType(coin.ItemType, bullOrBear)
				if coin.Timestamp < cointTypeInitTime[coinType] { //只处理coinType指定初始化时间以后的数据
					continue
				}
				if len(specCoinType) > 0 && specCoinType != coinType { //如果指定了specCoinType, 则只处理specCoinType指定标杆币
					continue
				}
				cb := new(CoinBull)
				//coinType := coin.ItemType
				cb.CoinType = coinType
				cb.RawPrice = coin.Price

				ftxCal := GetFtxCalInstance()
				ftxCal.SetLastRawAndRebalance(LastBullAJ[coinType].TargetPriceOfCheckpoint, LastBullAJ[coinType].Rebalance)
				ftxCal.SetRebalanceThreshold(getFtxXRate(coin.ItemType) / 100.0)

				if bullOrBear == "bull" {
					ftxCal.SetLeverage(ftxMultipleMap[coin.ItemType])
					//cb.Bull = CacuBullPrice(LastBullAJ[coinType].Rebalance, LastBullAJ[coinType].RawPrice, cb.RawPrice, coin.ItemType)
				} else if bullOrBear == "bear" {
					ftxCal.SetLeverage(-ftxMultipleMap[coin.ItemType])
					//cb.Bull = CacuBearPrice(LastBullAJ[coinType].Rebalance, LastBullAJ[coinType].RawPrice, cb.RawPrice, coin.ItemType)
				}

				// 每天14点，检测是否在过去24小时之内触发过调仓
				//now := time.Now()
				now := time.Unix(cb.Timestamp, 0)
				if now.Hour() == 14 && now.Minute() < 20 {
					ltime := time.Unix(LastBullAJ[coinType].Timestamp, 0)
					lastRebalanceTime := ltime
					if now.Sub(lastRebalanceTime).Hours() >= 24 {
						cb.IsAjustPoint = true
						cb.Rebalance = ftxCal.GetLETFPriceOfCheckpoint()
						ftxCal.Rebalance(coin.Price)
					}
				}

				cb.Bull = ftxCal.FeedPrice(coin.Price)
				cb.TargetPriceOfCheckpoint = ftxCal.GetTargetPriceOfCheckpoint()
				cb.RawChange = RoundPercentageChange(LastBullAJ[coinType].RawPrice, cb.RawPrice, 1)
				cb.BullChange = RoundPercentageChange(FirstBull[coinType].Bull, cb.Bull, 1)
				cb.Timestamp = int64(coin.Timestamp)
				cb.CreatedAt = time.Now()
				cb.Rebalance = LastBullAJ[coinType].Rebalance
				cb.PriceID = int(coin.ID)
				//cb.ID = uint(coin.ID)
				//|| cb.Timestamp.Sub(cb.Timestamp.Truncate(24*time.Hour).Add(2*time.Minute)).Seconds() < 25

				cb.Rebalance = ftxCal.GetLETFPriceOfCheckpoint()
				if ftxCal.GetTargetPriceOfCheckpoint() != LastBullAJ[coinType].RawPrice {
					cb.IsAjustPoint = true
					LastBullAJ[coinType] = cb
				}

				cbs = append(cbs, cb)
			}
			err := utils.Orm.CreateInBatches(cbs, 10).Error
			if err != nil {
				break
			}
			lastStat = int(coin.ID)
		}
		return nil
	}
	var err error
	if len(specCoinType) > 0 { //
		err = utils.Orm.Model(MarketPrice{}).Order("id").Where(" timestamp>? and item_type in(?)", lastStat, ftxList).FindInBatches(&coins, 2, proc).Error
	} else {
		err = utils.Orm.Model(MarketPrice{}).Order("id").Where(" id>? and item_type in(?)", lastStat, ftxList).FindInBatches(&coins, 2, proc).Error
	}
	if err != nil {
		log.Println("SetBullsForTw db err", err)
	}
	return lastStat, err

	//
	//var rows *sql.Rows
	//if len(specCoinType) > 0 { //
	//	rows, err = utils.Orm.Model(MarketPrice{}).Order("id").Where(" timestamp>? and item_type in(?)", lastStat, ftxList).Rows()
	//} else {
	//	rows, err = utils.Orm.Model(MarketPrice{}).Order("id").Where(" id>? and item_type in(?)", lastStat, ftxList).Rows()
	//}
	//if err != nil {
	//	log.Println(err)
	//	return lastStat, err
	//}
	//defer rows.Close()
	//counter := 0
	//for rows.Next() {
	//	counter++
	//	if counter > 200 {
	//		//return lastStat,nil
	//	}
	//	coin := new(MarketPrice)
	//	err = utils.Orm.ScanRows(rows, coin)
	//	if err != nil {
	//		log.Println(err)
	//		return lastStat, err
	//	}
	//	cbs := []*CoinBull{}
	//	for _, bullOrBear := range []string{"bull", "bear"} {
	//		coinType := getCoinType(coin.ItemType, bullOrBear)
	//		if coin.Timestamp < cointTypeInitTime[coinType] { //只处理coinType指定初始化时间以后的数据
	//			continue
	//		}
	//		if len(specCoinType) > 0 && specCoinType != coinType { //如果指定了specCoinType, 则只处理specCoinType指定标杆币
	//			continue
	//		}
	//		cb := new(CoinBull)
	//		//coinType := coin.ItemType
	//		cb.CoinType = coinType
	//
	//		cb.RawPrice = coin.Price
	//		if bullOrBear == "bull" {
	//			cb.Bull = CacuBullPrice(LastBullAJ[coinType].Rebalance, LastBullAJ[coinType].RawPrice, cb.RawPrice, coin.ItemType)
	//		} else if bullOrBear == "bear" {
	//			cb.Bull = CacuBearPrice(LastBullAJ[coinType].Rebalance, LastBullAJ[coinType].RawPrice, cb.RawPrice, coin.ItemType)
	//		}
	//		cb.RawChange = RoundPercentageChange(LastBullAJ[coinType].RawPrice, cb.RawPrice, 1)
	//		cb.BullChange = RoundPercentageChange(FirstBull[coinType].Bull, cb.Bull, 1)
	//		cb.Timestamp = int64(coin.Timestamp)
	//		cb.CreatedAt = time.Now()
	//		cb.Rebalance = LastBullAJ[coinType].Rebalance
	//		cb.PriceID = int(coin.ID)
	//		//cb.ID = uint(coin.ID)
	//		//|| cb.Timestamp.Sub(cb.Timestamp.Truncate(24*time.Hour).Add(2*time.Minute)).Seconds() < 25
	//		ajChange := cb.RawChange
	//
	//		rate := getFtxXRate(coin.ItemType)
	//		if math.Abs(ajChange) >= rate {
	//			cb.IsAjustPoint = true
	//			if ajChange > 0 {
	//				if bullOrBear == "bull" {
	//					cb.Rebalance = LastBullAJ[coinType].Rebalance * (1 + rate/100.0)
	//				} else if bullOrBear == "bear" {
	//					cb.Rebalance = LastBullAJ[coinType].Rebalance * (1 - rate/100.0)
	//				}
	//			} else {
	//				if bullOrBear == "bull" {
	//					cb.Rebalance = LastBullAJ[coinType].Rebalance * (1 - rate/100.0)
	//				} else if bullOrBear == "bear" {
	//					cb.Rebalance = LastBullAJ[coinType].Rebalance * (1 + rate/100.0)
	//				}
	//			}
	//			LastBullAJ[coinType] = cb
	//		}
	//		// 每天14点，检测是否在过去24小时之内触发过调仓
	//		//now := time.Now()
	//		now := time.Unix(cb.Timestamp, 0)
	//		if now.Hour() == 14 && now.Minute() < 20 {
	//			ltime := time.Unix(LastBullAJ[coinType].Timestamp, 0)
	//			lastRebalanceTime := ltime
	//			if now.Sub(lastRebalanceTime).Hours() >= 24 {
	//				cb.IsAjustPoint = true
	//				cb.Rebalance = cb.Bull
	//				LastBullAJ[coinType] = cb
	//			}
	//		}
	//		cbs = append(cbs, cb)
	//	}
	//	err = utils.Orm.CreateInBatches(cbs, 10).Error
	//	lastStat = int(coin.ID)
	//}
	//return lastStat, err
}

//生成杠杆币数据　only for coin from coingecko
func SetBullsFromID(lastBullTime int64, coinType string) (int64, error) {
	//initCoinBull(coinType)
	//setFirstBull(coinType)
	//setLastBullAJ(coinType)

	var err error

	coin_name := coinType[:3]
	rows, err := utils.Orm.Model(Coin{}).Where(" id>?", lastBullTime).Rows()
	//rows,err:=utils.Orm.Raw("SELECT cast(usd as decimal(10,2))as `usd`,id FROM `coins` order by `usd` asc;").Rows()
	if err != nil {
		log.Println(err)
		return 0, err
	}
	defer rows.Close()
	counter := 0
	for rows.Next() {
		counter++
		if counter > 10 {
			//return 0,nil
		}
		coin := new(Coin)
		err = utils.Orm.ScanRows(rows, coin)
		if err != nil {
			log.Println(err)
			return 0, err
		}
		cb := new(CoinBull)
		cb.CoinType = coinType
		rawPrice := "1"
		if coin_name == "eth" {
			rawPrice = coin.Eth
		}

		cb.RawPrice = getCoinUSdPriceFromStr(rawPrice, coin.Usd)
		cb.Bull = CacuBullPrice(LastBullAJ[coinType].Rebalance, LastBullAJ[coinType].RawPrice, cb.RawPrice, coinType)
		cb.RawChange = RoundPercentageChange(LastBullAJ[coinType].RawPrice, cb.RawPrice, 1)
		cb.BullChange = RoundPercentageChange(FirstBull[coinType].Bull, cb.Bull, 1)
		cb.Timestamp = coin.ID
		cb.CreatedAt = time.Now()
		cb.Rebalance = LastBullAJ[coinType].Rebalance
		//cb.ID = uint(coin.ID)
		//|| cb.Timestamp.Sub(cb.Timestamp.Truncate(24*time.Hour).Add(2*time.Minute)).Seconds() < 25
		ajChange := cb.RawChange
		rate := getFtxXRate(coinType)
		if math.Abs(ajChange) >= rate {
			cb.IsAjustPoint = true
			if ajChange > 0 {
				cb.Rebalance = LastBullAJ[coinType].Rebalance * (1 + rate/100.0)
			} else {
				cb.Rebalance = LastBullAJ[coinType].Rebalance * (1 - rate/100.0)
			}
			LastBullAJ[coinType] = cb
		}

		// 每天14点，检测是否在过去24小时之内触发过调仓
		//now := time.Now()
		now := time.Unix(cb.Timestamp, 0)
		if now.Hour() == 14 && now.Minute() < 20 {
			ltime := time.Unix(LastBullAJ[coinType].Timestamp, 0)
			lastRebalanceTime := ltime
			if now.Sub(lastRebalanceTime).Hours() >= 24 {
				cb.IsAjustPoint = true
				cb.Rebalance = cb.Bull
				LastBullAJ[coinType] = cb
			}
		}

		err = utils.Orm.Create(cb).Error
		if err == nil {
			lastBullTime = cb.Timestamp
		}
	}
	return lastBullTime, err
}
func RoundPercentageChange(oldValue, newValue float64, deciaml int) float64 {
	return float64(int(math.Trunc((newValue-oldValue)/oldValue*math.Pow10(deciaml+2)))) / math.Pow10(deciaml)
}
func getCoinUSdPriceFromStr(coin, usd string) float64 {
	usdPrice, _ := strconv.ParseFloat(usd, 64)
	coinPrice, _ := strconv.ParseFloat(coin, 64)
	//log.Println(coinPrice,usdPrice)
	return usdPrice / coinPrice
}

//杠杆币对象
type CoinBull struct {
	ID uint `gorm:"primarykey"`
	//原币价格抓取时对应的时间秒数
	Timestamp int64
	//杠杆币的类型：btc3x eth3x vix3x ust20x gold10x eur20x ndx10x　govt20x
	CoinType string
	//bull价格
	Bull float64
	//调仓价格
	Rebalance float64
	//bull相对于原点变化
	BullChange float64
	//原币 usd价格
	RawPrice float64
	//last rebalance raw price
	TargetPriceOfCheckpoint float64
	//原币相对于原点变化
	RawChange float64
	//是否是调仓点
	IsAjustPoint bool
	CreatedAt    time.Time
	//market_price表的id
	PriceID int
}

func (CoinBull CoinBull) TableName() string {
	return "coin_bull"
}

type FtxChartDate struct {
	Timestamp uint
	//杠杆币价格
	Bull float64
	////杠杆区间最高
	//Hight float64
	////杠杆区间最低
	//Low float64
	////Btc价格
	//RawPrice float64
	////Btc区间最高
	//RawPriceHight float64
	////Btc区间最低
	//RawPriceLow float64
}

//杠杆图表
func GetFtxTimesPrice(coin_type string, interval, count int) ([]*FtxChartDate, error) {
	datas := []*FtxChartDate{}

	sql := `
select bulls.*,dates.secon1 as timestamp
from (select truncate((dates.id - 1) / @interval, 0) as id1,
             min(dates.date)                     datestr,
             min(dates.secon1)                   secon1,
             max(dates.secon2)                   secon2
      from stock.dates dates
      where dates.secon1 > truncate(unix_timestamp() / (15 * 60 * @interval), 0) * 15 * 60 * @interval  - 15 * 60 * @interval * @count
        and dates.secon1 < unix_timestamp()
      group by id1
     ) dates
         left join (select
                            truncate(coin_bull.timestamp / (15 * 60 * @interval), 0) * 15 * 60 * @interval as b_timestamp,
                           cast(avg(coin_bull.bull) as decimal(9, 3))                          bull
#                           cast(max(coin_bull.bull) as decimal(9, 3))                          hight,
#                           cast(avg(coin_bull.bull) as decimal(9, 3))                          low,
#                           cast(avg(coin_bull.raw_price) as decimal(9, 3))                     raw_price,
#                           cast(max(coin_bull.raw_price) as decimal(9, 3))                     raw_price_hight,
#                           cast(avg(coin_bull.raw_price) as decimal(9, 3))                     raw_price_low
                    from coin_bull
                    where coin_bull.timestamp > unix_timestamp() - 15 * 60 * @interval * @count
                      and coin_bull.timestamp < unix_timestamp()
                      and coin_bull.coin_type = @coin_type
                    group by b_timestamp
) bulls
                   on dates.secon1 <= bulls.b_timestamp and dates.secon2 > bulls.b_timestamp
order by dates.id1 `

	err := utils.Orm.Raw(sql, map[string]interface{}{"interval": interval, "count": count, "coin_type": coin_type}).Scan(&datas).Error
	if err == nil {
		for idx, data := range datas {
			//数据不连续时，取得的第一个时间点可能为０
			if idx == 0 {
				if data.Bull == 0 {
					sql = `
select cast(coin_bull.bull as decimal(9, 3))      bull,
       cast(coin_bull.raw_price as decimal(9, 3)) raw_price
from coin_bull
where coin_bull.timestamp < ?
  and coin_bull.coin_type = ?
order by timestamp desc
limit 1;
`
					cdata := new(FtxChartDate)
					err = utils.Orm.Raw(sql, data.Timestamp, coin_type).Scan(cdata).Error
					if err != nil {
						log.Println(err)
					} else {
						data.Bull = cdata.Bull
						//data.RawPrice = cdata.RawPrice
					}
				}
			}
			if idx > 0 {
				if data.Bull == 0 {
					data.Bull = datas[idx-1].Bull
					//data.RawPrice = datas[idx-1].RawPrice
				}
			}
		}
	}
	return datas, err
}

//股票图表
func GetStockTimesPrice(coin_type string, interval, count int) ([]*FtxChartDate, error) {
	datas := []*FtxChartDate{}

	sql := `
select bulls.*, dates.secon1 as timestamp,dates.datestr
from (select truncate((dates.id - 1) / @interval, 0) as id1,
             min(dates.date)                    datestr,
             min(dates.secon1)                  secon1,
             max(dates.secon2)                  secon2
      from stock.dates dates
      where dates.secon1 > truncate(unix_timestamp()  / (15 * 60 * @interval), 0) * 15 * 60 * @interval - 15 * 60 * @interval * ( @count)
        and dates.secon1 < unix_timestamp()
      group by id1
     ) dates
         left join (select coin_bull.timestamp as b_timestamp,
                           coin_bull.last_state   bull
                    from interval_stats as coin_bull
                    where coin_bull.timestamp > unix_timestamp() - 15 * 60 * @interval * @count
                      and coin_bull.timestamp < unix_timestamp()
                      and coin_bull.cat = @coin_type and coin_bull.` + "`interval`" + `=15 * 60 * @interval
) bulls
                   on dates.secon1 = bulls.b_timestamp #and dates.secon2 > bulls.b_timestamp
order by dates.id1
limit ` + strconv.Itoa(count) + ";"

	err := utils.Orm.Raw(sql, map[string]interface{}{"interval": interval, "count": count, "coin_type": coin_type}).Scan(&datas).Error
	if err == nil {
		for idx, data := range datas {
			//数据不连续时，取得的第一个时间点可能为０
			if idx == 0 {
				if data.Bull == 0 {
					sql = `
select cast(coin_bull.last_state as decimal(9, 3))      bull
from interval_stats coin_bull
where coin_bull.timestamp < ?
 and coin_bull.cat = ?
order by timestamp desc
limit 1;
`
					cdata := new(FtxChartDate)
					err = utils.Orm.Raw(sql, data.Timestamp, coin_type).Scan(cdata).Error
					if err != nil {
						log.Println(err)
					} else {
						data.Bull = cdata.Bull
						//data.RawPrice = cdata.RawPrice
					}
				}
			}
			if idx > 0 {
				if data.Bull == 0 {
					data.Bull = datas[idx-1].Bull
					//data.RawPrice = datas[idx-1].RawPrice
				}
			}
		}
	}
	return datas, err
}

//生成股票图表数据
func SetStockStat() {
	utils.Orm.AutoMigrate(IntervalStat{})
	//istat1:=new(IntervalStat)
	//istat1.Cat="AAPL"
	//istat1.Interval=15*60
	initMapInterval()
	lastId := IntervalStat{}.GetLastID()
	//SetStockStatFromlastId(lastId)

	proc := func() error {
		lastId = SetStockStatFromlastId(lastId)
		return nil
	}
	utils.IntervalSync("SetStockStat", 60, proc)
}

func SetStockStatFromlastId(lastId int) int {
	rows, err := utils.Orm.Model(MarketPrice{}).Where(" id>? and item_type in(?)", lastId, stocks).Rows()
	if err == nil {
		idx := 0
		for rows.Next() {
			idx++
			//if idx>100{log.Println("break",idx); break;}
			mprice := new(MarketPrice)
			err = utils.Orm.ScanRows(rows, mprice)
			if err != nil {
				break
			}
			gIntervals.SetItem(mprice)
			lastId = int(mprice.ID)
		}
		if idx > 0 {
			gIntervals.Save()
		}
	}
	if err != nil {
		log.Println(err)
	}
	return lastId
}

type IntervalStat struct {
	ID        int    `gorm:"primaryKey"`
	Cat       string `gorm:"type:varchar(20);uniqueIndex:idx_cat_interval_time,priority:1"`
	Interval  int    `gorm:"uniqueIndex:idx_cat_interval_time,priority:2"`
	Timestamp int    `gorm:"uniqueIndex:idx_cat_interval_time,priority:2"`
	LastID    int
	LastState float64
	CreatedAt time.Time
	//UpdatedAt time.Time
	Saved bool `gorm:"-"`
}

func (istat IntervalStat) GetLastID() int {
	err := utils.Orm.Order("timestamp desc").Limit(1).Find(&istat).Error
	if err != nil {
		log.Fatalln(err)
	}
	return istat.LastID
}
func (istat *IntervalStat) SetTimestamp(ts int) {
	if istat.Timestamp != ts && istat.Timestamp != 0 {
		utils.Orm.Create(istat)
	}
	istat.Timestamp = ts
}

type mapInterval map[string]*IntervalStat

var gIntervals mapInterval
var intervals = []int{15 * 60, 60 * 60, 24 * 3600}
var stocks = []string{"AAPL", "TSLA"}

func initMapInterval() {
	mstat := mapInterval{}
	for _, cat := range stocks {
		for _, interval := range intervals {
			key := fmt.Sprintf("%s-%d", cat, interval)
			istat1 := new(IntervalStat)
			istat1.Cat = cat
			istat1.Interval = interval
			mstat[key] = istat1
		}
	}
	gIntervals = mstat
	//return mstat
}
func (mstat mapInterval) SetItem(price *MarketPrice) {
	for _, interval := range intervals {
		key := fmt.Sprintf("%s-%d", price.ItemType, interval)
		istat := mstat[key]
		if istat == nil {

		}
		tt := (price.Timestamp / istat.Interval) * istat.Interval
		if istat.Timestamp != tt && istat.Timestamp != 0 {
			if istat.Saved {
				utils.Orm.Create(istat)
			} else { //没有被保存过时使用FirstOrCreate方法，可以同数据库中的记录合并．
				utils.Orm.Assign(IntervalStat{LastState: istat.LastState, LastID: istat.LastID}).FirstOrCreate(istat, IntervalStat{Cat: price.ItemType, Interval: istat.Interval, Timestamp: istat.Timestamp})
				//utils.Orm.Save(istat)
			}
			istat.Saved = true
			istat.ID = 0
			istat.CreatedAt = time.Time{}
		}
		istat.Timestamp = tt
		istat.LastState = price.Price
		istat.LastID = int(price.ID)
	}
}
func (mstat mapInterval) Save() {
	for _, stat := range mstat {
		if stat.Timestamp > 0 {
			utils.Orm.Save(stat)
		}
	}
}
