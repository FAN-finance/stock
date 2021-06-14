package services

import (
	"log"
	"math"
	"stock/utils"
	"strconv"
	"time"
)

func CacuBullPrice(lastAjustPriceBull, lastAjustPric, curPric float64) float64{
	return lastAjustPriceBull*((curPric-lastAjustPric)/lastAjustPric*3 +1)
}

/**/

var FirstBull ,LastBullAJ *CoinBull
func setFirstBull()  {
	firstBull := new(CoinBull)
	err := utils.Orm.First(firstBull).Error
	if err != nil {
		log.Fatal(err)
	}
	FirstBull=firstBull
	//return firstBull
}

func setLastBullAJ() {
	lastAj := new(CoinBull)
	err:= utils.Orm.Order("id desc").First(lastAj, "is_ajust_point=?", 1).Error
	if err != nil {
		log.Fatal(err)
	}
	log.Println("lastaj %v", lastAj)
	LastBullAJ=lastAj
	//return lastAj
}
func LastBullID() uint{
	lastAj := new(CoinBull)
	err:= utils.Orm.Order("id desc").First(lastAj).Error
	if err != nil {
		log.Fatal(err)
	}
	log.Println("lastja %v", lastAj)
	return lastAj.ID
}
func initCoinBull(){
	var err error
	utils.Orm.AutoMigrate(CoinBull{})
	bullCount := int64(0)
	utils.Orm.Model(CoinBull{}).Count(&bullCount)
	if bullCount == 0 {
		firstCoin := new(Coin)
		err = utils.Orm.Model(Coin{}).First(firstCoin).Error
		if err != nil {
			log.Fatal(err)
		}
		cb := new(CoinBull)
		cb.Btc = getCoinUSdPriceFromStr("1", firstCoin.Usd)
		cb.BtcBull = 10000
		cb.BaseChange = 0
		cb.BullChange = 0
		cb.IsAjustPoint = true
		cb.ID = 1
		err=utils.Orm.Save(cb).Error
		if err != nil {
			log.Fatal(err)
		}
	}
}
func SetAllBulls() {
	initCoinBull()
	setFirstBull()
	setLastBullAJ()
	lastBullId:=LastBullID()
	proc:=func()error{
		lastId,err:=SetBullsFromID(lastBullId)
		if err==nil{
			lastBullId=lastId
		}
		return err
	}
	utils.IntervalSync("SetAllBull",10,proc)
}
func SetBullsFromID(lastBullId uint) (uint,error) {
	//initCoinBull()
	//setFirstBull()
	//setLastBullAJ()

	var err error
	//firstBull :=FirstBull()
	//lastAj :=LastBullAJ()

	rows, err := utils.Orm.Model(Coin{}).Where("id>?",lastBullId).Select("id","usd").Rows()
	//rows,err:=utils.Orm.Raw("SELECT cast(usd as decimal(10,2))as `usd`,id FROM `coins` order by `usd` asc;").Rows()
	if err != nil {
		log.Println(err)
		return 0,err
	}
	defer rows.Close()
	counter:=0
	for rows.Next() {
		counter++
		if counter>10000{
			//return
		}
		coin := new(Coin)
		err=utils.Orm.ScanRows(rows, coin)
		if err != nil {
			log.Println(err)
			return 0,err
		}
		cb := new(CoinBull)
		cb.Btc = getCoinUSdPriceFromStr("1", coin.Usd)
		cb.BtcBull = CacuBullPrice(LastBullAJ.BtcBull, LastBullAJ.Btc, cb.Btc)
		cb.BaseChange = RoundPercentageChange(FirstBull.Btc,cb.Btc, 1)
		cb.BullChange = RoundPercentageChange(FirstBull.BtcBull,cb.BtcBull, 1)
		ajChange:= RoundPercentageChange(LastBullAJ.BtcBull,cb.BtcBull, 1)
		cb.Timestamp = time.Unix(coin.ID, 0).UTC()
		//|| cb.Timestamp.Sub(cb.Timestamp.Truncate(24*time.Hour).Add(2*time.Minute)).Seconds() < 25
		if  math.Abs(ajChange) > 10 {
			cb.IsAjustPoint = true
			LastBullAJ = cb
		}
		cb.ID=uint(coin.ID)
		err=utils.Orm.Create(cb).Error
		lastBullId=cb.ID
	}
	return lastBullId,err
}
func RoundPercentageChange(oldValue,newValue float64,deciaml int) float64{
	return float64(int(math.Trunc((newValue-oldValue)/oldValue* math.Pow10(deciaml+2))))/ math.Pow10(deciaml)
}
func getCoinUSdPriceFromStr(coin,usd string)float64{
	usdPrice,_:= strconv.ParseFloat(usd,64)
	coinPrice,_:= strconv.ParseFloat(coin,64)
	//log.Println(coinPrice,usdPrice)
	return usdPrice/coinPrice
}
type CoinBull struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	Timestamp time.Time
	//btc-usd价格
	Btc float64
	//bull价格
	BtcBull float64
	//bull相对于原点变化
	BullChange float64
	//源btc相对于原点变化
	BaseChange float64
	IsAjustPoint bool
}