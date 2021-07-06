package services

import (
	"log"
	"math"
	"regexp"
	"stock/utils"
	"strconv"
	"strings"
	"time"
)

func CacuBullPrice(lastAjustPriceBull, lastAjustPric, curPric float64, coin_type string) float64 {
	//return lastAjustPriceBull * ((curPric-lastAjustPric)/lastAjustPric*3 + 1)
	return lastAjustPriceBull * ((curPric-lastAjustPric)/lastAjustPric*float64(ftxMultipleMap[coin_type]) + 1)
}
func getMultipleFromCoinType(coinType string) int {
	preg := regexp.MustCompile("(\\d+)x")
	items := preg.FindStringSubmatch(coinType)
	multi, _ := strconv.Atoi(items[1])
	return multi
}

var ftxMultipleMap = map[string]int{
	"btc3x":   3,
	"eth3x":   3,
	"vix3x":   3,
	"ust20x":  20,
	"gold10x": 10,
	"eur20x":  20,
	"ndx10x":  10,
	"govt20x": 20,
}

var ftxXMap = map[string]float64{
	"3x":  10,
	"10x": 3,
	"20x": 1.5,
}

func getFtxXRate(coinType string) float64 {
	coinType = strings.ToLower(coinType)
	xArr := []string{"3x", "10x", "20x"}
	for _, item := range xArr {
		if strings.Contains(coinType, item) {
			return ftxXMap[item]
		}
	}
	return 10
}

/*ndx10x vix3x*/
/**/

var FirstBull, LastBullAJ = map[string]*CoinBull{}, map[string]*CoinBull{}

func setFirstBull(coinType string) {
	firstBull := new(CoinBull)
	err := utils.Orm.Where("coin_type=?", coinType).First(firstBull).Error
	if err != nil {
		log.Fatal(err)
	}
	FirstBull[coinType] = firstBull
	//return firstBull
}

func setLastBullAJ(coinType string) {
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
	err := utils.Orm.Order("id desc").First(cb).Error
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

//从twelvedata数据market_pirces表初始化　bull
func initCoinBullFromTw(coinType string) {
	var err error
	utils.Orm.AutoMigrate(CoinBull{})
	bullCount := int64(0)
	utils.Orm.Model(CoinBull{}).Where("coin_type=?", coinType).Count(&bullCount)
	if bullCount == 0 {
		firstPrice := new(MarketPrice)
		err = utils.Orm.Model(MarketPrice{}).Order("timestamp").Where("item_type=?", coinType).First(firstPrice).Error
		if err != nil {
			log.Fatal(err)
		}
		cb := new(CoinBull)
		cb.CoinType = coinType
		cb.RawPrice = firstPrice.Price
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
func SetAllBullsFromTw() {
	for _, coinType := range twSymbolMap {
		initCoinBullFromTw(coinType)
		setFirstBull(coinType)
		setLastBullAJ(coinType)
	}
	lastStat := LastBullPriceID()
	//lastStat,_=SetBullsForTw(lastStat)
	//log.Println(lastStat)
	//return
	proc := func() error {
		lastId, err := SetBullsForTw(lastStat)
		if err == nil {
			lastStat = lastId
		}
		return err
	}
	utils.IntervalSync("SetAllBullsFromTw", 60, proc)
}
func SetBullsForTw(lastStat int) (int, error) {
	//initCoinBull(coinType)
	//setFirstBull(coinType)
	//setLastBullAJ(coinType)
	var err error
	rows, err := utils.Orm.Model(MarketPrice{}).Order("id").Where(" id>?", lastStat).Rows() //	,[]string{"vix3x"}

	//rows,err:=utils.Orm.Raw("SELECT cast(usd as decimal(10,2))as `usd`,id FROM `coins` order by `usd` asc;").Rows()
	if err != nil {
		log.Println(err)
		return lastStat, err
	}
	defer rows.Close()
	counter := 0
	for rows.Next() {
		counter++
		if counter > 200 {
			//return lastStat,nil
		}
		coin := new(MarketPrice)
		err = utils.Orm.ScanRows(rows, coin)
		if err != nil {
			log.Println(err)
			return lastStat, err
		}
		cb := new(CoinBull)
		coinType := coin.ItemType
		cb.CoinType = coin.ItemType

		cb.RawPrice = coin.Price
		cb.Bull = CacuBullPrice(LastBullAJ[coinType].Rebalance, LastBullAJ[coinType].RawPrice, cb.RawPrice, coinType)
		cb.RawChange = RoundPercentageChange(LastBullAJ[coinType].RawPrice, cb.RawPrice, 1)
		cb.BullChange = RoundPercentageChange(FirstBull[coinType].Bull, cb.Bull, 1)
		cb.Timestamp = int64(coin.Timestamp)
		cb.CreatedAt = time.Now()
		cb.Rebalance = LastBullAJ[coinType].Rebalance
		cb.PriceID = int(coin.ID)
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
		lastStat = int(coin.ID)
	}
	return lastStat, err
}

//only for coin from coingecko
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
