package services

import (
	"log"
	"math"
	"regexp"
	"stock/utils"
	"strconv"
	"time"
)

func CacuBullPrice(lastAjustPriceBull, lastAjustPric, curPric float64) float64 {
	return lastAjustPriceBull * ((curPric-lastAjustPric)/lastAjustPric*3 + 1)
}
func getMultipleFromCoinType(coinType string) int {
	preg := regexp.MustCompile("(\\d+)x")
	items := preg.FindStringSubmatch(coinType)
	multi, _ := strconv.Atoi(items[1])
	return multi
}

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
	err := utils.Orm.Order("id desc").First(lastAj, "coin_type=? and is_ajust_point=?", coinType, 1).Error
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
	err := utils.Orm.Order("id desc").Where("coin_type=?", coinType).First(cb).Error
	if err != nil {
		log.Fatal(err)
	}
	log.Println("lastbullTime %v", cb)
	return cb.Timestamp
}
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
func SetAllBulls(coinType string) {
	initCoinBull(coinType)
	setFirstBull(coinType)
	setLastBullAJ(coinType)
	lastBullTime := LastBullTimeStamp(coinType)
	//lastBullTime,_=SetBullsFromID(lastBullTime,coinType)
	//return
	proc := func() error {
		lastId, err := SetBullsFromID(lastBullTime, coinType)
		if err == nil {
			lastBullTime = lastId
		}
		return err
	}
	utils.IntervalSync("SetAllBull"+coinType, 10, proc)
}

//only for coin from congecko
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
		cb.Bull = CacuBullPrice(LastBullAJ[coinType].Rebalance, LastBullAJ[coinType].RawPrice, cb.RawPrice)
		cb.RawChange = RoundPercentageChange(LastBullAJ[coinType].RawPrice, cb.RawPrice, 1)
		cb.BullChange = RoundPercentageChange(FirstBull[coinType].Bull, cb.Bull, 1)
		cb.Timestamp = coin.ID
		cb.CreatedAt = time.Now()
		cb.Rebalance = LastBullAJ[coinType].Rebalance
		//cb.ID = uint(coin.ID)
		//|| cb.Timestamp.Sub(cb.Timestamp.Truncate(24*time.Hour).Add(2*time.Minute)).Seconds() < 25
		ajChange := RoundPercentageChange(LastBullAJ[coinType].RawPrice, cb.RawPrice, 1)
		if math.Abs(ajChange) > 10 {
			cb.IsAjustPoint = true
			if ajChange > 0 {
				cb.Rebalance = LastBullAJ[coinType].Rebalance * 1.1
			} else {
				cb.Rebalance = LastBullAJ[coinType].Rebalance * 0.9
			}
			LastBullAJ[coinType] = cb
		}

		// 每天14点，检测是否在过去24小时之内触发过调仓
		now := time.Now()
		if now.Hour() == 14 && now.Minute() < 3 {
			lastRebalanceTime := LastBullAJ[coinType].CreatedAt
			if now.Sub(lastRebalanceTime).Hours() >= 24 {
				cb.IsAjustPoint = true
				cb.Rebalance = cb.Bull
				LastBullAJ[coinType] = cb
			}
		}

		err = utils.Orm.Create(cb).Error
		lastBullTime = cb.Timestamp
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
	//杠杆币的类型：btc3x eth3x vix3x ust20x gold10x eur20x ndx10x
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
}

func (CoinBull CoinBull) TableName() string {
	return "coin_bull"
}
