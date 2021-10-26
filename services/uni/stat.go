package uni

import (
	"fmt"
	"gorm.io/gorm"
	"log"
	"math"
	"stock/utils"
	"time"
)
type UniPriceIntervalStat struct {
	ID        int    `gorm:"primaryKey"`
	PairID uint `gorm:"uniqueIndex:idx_cat_interval_time,priority:1"`
	Symbol    string `gorm:"type:varchar(50);uniqueIndex:idx_cat_interval_time,priority:2"`
	Interval  int    `gorm:"uniqueIndex:idx_cat_interval_time,priority:3"`
	Timestamp int    `gorm:"uniqueIndex:idx_cat_interval_time,priority:4"`
	LastID    int
	LastState float64
	LastLp float64
	Vol float64
	UpdatedAt time.Time
}

func (istat UniPriceIntervalStat) GetLastID() int {
	err := utils.Orm.Order("id desc").Limit(1).Find(&istat).Error
	if err != nil {
		log.Fatalln(err)
	}
	return istat.LastID
}

type mapStat map[string]*UniPriceIntervalStat

var gIntervals mapStat
var intervals = []int{15 * 60, 60 * 60, 24 * 3600}
//var stocks = []string{"REI", "zUSD"}

func InitMapStat() {
	mstat := mapStat{}
	ps:=[]PairInfo{}
	utils.Orm.Find(&ps)
	for _, pinfo := range ps {
		for _, interval := range intervals {
			key := fmt.Sprintf("%d-%s-%d", pinfo.Id,pinfo.Symbol, interval)
			istat1 := new(UniPriceIntervalStat)
			istat1.PairID = pinfo.Id
			istat1.Symbol = pinfo.Symbol
			istat1.Interval = interval
			utils.Orm.Order("id desc").First(istat1,istat1)
			mstat[key] = istat1
		}
	}
	gIntervals = mstat
	log.Println("init stat for all pair",len(gIntervals))
	//return mstat
}
func (mstat mapStat) SetItem(price *UniPrice ,dbtx *gorm.DB) (err error) {
	for _, interval := range intervals {
		key := fmt.Sprintf("%d-%s-%d", price.PairID,price.Symbol, interval)
		istat := mstat[key]
		if istat == nil {
			//log.Println("mapInterval not init")
			continue
			return
		}

		preLp:=istat.LastLp
		if preLp==0{
			preLp=float64(price.Vol)
		}
		currLp:=float64(price.Vol)
		currVol:=math.Abs(currLp-preLp)

		tt := (int(price.BlockTime) / istat.Interval) * istat.Interval
		if istat.Timestamp != tt && istat.Timestamp != 0 {
			if istat.ID==0{
				err=dbtx.Create(istat).Error
			}else{
				err=dbtx.Save(istat).Error
			}
			if err != nil {
				return err
			}
			//reset
			istat.ID = 0
			istat.Vol=0
			istat.UpdatedAt = time.Time{}
			istat.LastState=0
		}

		istat.Timestamp = tt
		istat.LastState =float64( price.Price)
		istat.LastLp =float64( price.Vol)
		istat.Vol+=currVol
		istat.LastID = int(price.Id)
		//if istat.ID==0{
		//	err=dbtx.Create(istat).Error
		//}else{
		//
		//}
		err=dbtx.Save(istat).Error
		if err != nil {
			return err
		}
	}
	return nil
}


func SetUniStatFromLastId(lastId int) int {
	dbtx:=utils.Orm.Begin()
	mps:=[]*UniPrice{}
	counter:=0
	proc:=func(tx *gorm.DB, batch int) error{
		for _, mp := range mps {
			err:=gIntervals.SetItem(mp,dbtx)
			if err != nil {
				return err
			}
			lastId = int(mp.Id)
		}
		counter+=len(mps)
		return nil
	}
	err:=utils.Orm.Where("id>?", lastId).FindInBatches(&mps,100,proc ).Error
	if err != nil {
		log.Println(err)
		dbtx.Rollback()
	}else{
		dbtx.Commit()
	}
	return lastId
}

//生成uni_price 时分统计数据
func CronUniStat() {
	utils.Orm.AutoMigrate(UniPriceIntervalStat{})
	InitMapStat();
	lastId := UniPriceIntervalStat{}.GetLastID()
	//SetUniStatFromLastId(1190);return;
	proc := func() error {
		lastId = SetUniStatFromLastId(lastId)
		return nil
	}
	utils.IntervalSync("CronUniStat", 60, proc)
}