package services

import (
	"fmt"
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
var ftxList=[]string{"mvi2x" , "btc3x" , "eth3x" , "vix3x" , "gold10x", "eur20x", "ndx10x", "govt20x"}
var ftxMultipleMap = map[string]int{
	"mvi2x": 2,
	"btc3x": 3,
	"eth3x": 3,
	"vix3x": 3,
	//"ust20x":  20,
	"gold10x": 10,
	"eur20x":  20,
	"ndx10x":  10,
	"govt20x": 20,
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
	"btc3x": 110054.79,
	"eth3x": 7900.56,
	"vix3x": 53.7,
	//"ust20x":  20,
	"gold10x": 19022.8,
	"eur20x":  244.66,
	"ndx10x":  136488,
	"govt20x": 5268,
}

var ftxXMap = map[string]float64{
	"2x":  15,
	"3x":  10,
	"10x": 3,
	"20x": 1.5,
}

func getFtxXRate(coinType string) float64 {
	coinType = strings.ToLower(coinType)
	xArr := []string{"2x", "3x", "10x", "20x"}
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
		cb.PriceID=int(firstPrice.ID)
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
	for _, coinType := range ftxList {
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
//生成杠杆币数据　twelvedata
func SetBullsForTw(lastStat int) (int, error) {
	//initCoinBull(coinType)
	//setFirstBull(coinType)
	//setLastBullAJ(coinType)
	var err error
	rows, err := utils.Orm.Model(MarketPrice{}).Order("id").Where(" id>? and item_type in(?)", lastStat, ftxList).Rows() //	,[]string{"vix3x"}

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
      where dates.secon1 >= truncate((unix_timestamp() - 15 * 60 * @interval * @count) / (15 * 60 * @interval), 0) * 15 * 60 * @interval
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
order by dates.id1
limit ` + strconv.Itoa(count) + ";"

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
      where dates.secon1 >= truncate((unix_timestamp() - 15 * 60 * @interval * ( @count)) / (15 * 60 * @interval), 0) * 15 * 60 * @interval
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
