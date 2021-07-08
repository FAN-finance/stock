package services

import (
	"bytes"
	"encoding/json"
	"github.com/rickar/cal/v2"
	"github.com/rickar/cal/v2/aa"
	"github.com/rickar/cal/v2/us"
	"log"
	"stock/utils"
	"time"
)

type resModel struct {
	Data struct {
		Diff json.RawMessage
	}
}
type Stock struct {
	ID        int     `gorm:"column:id;primary_key" `
	Code      string  `gorm:"index:code_time,priority:1" json:"f12"`         //代码
	Price     float64 `json:"f2" gorm:"DEFAULT:null;type:decimal(10,2)"`     //最新价
	StockName string  `json:"f14"`                                           //名称
	Mk        int     `json:"f13"`                                           //市场 1 sh 01 0 sz 02
	Diff      float32 `json:"f3" gorm:"DEFAULT:0;"`                          //最新涨百分比
	Timestamp int64   `gorm:"index:code_time,priority:2;" json:"timestamp" ` //unix 秒数
	UpdatedAt *time.Time
}
type ViewStock struct {
	Code      string  `json:"Code" gorm:"column:code"`     //代码 苹果代码 AAPL ,特斯拉代码 TSLA
	Price     float64 `json:"Price" gorm:"DEFAULT:null;"`  //最新价
	StockName string  `json:"StockName" `                  //名称
	Timestamp int64   `json:"Timestamp" gorm:"DEFAULT:0;"` //unix 秒数
	//rfc3339 fortmat
	UpdatedAt *time.Time `json:"UpdatedAt" example:"2021-05-07T18:25:44.27+08:00"`
}
type StockData struct {
	//股票代码
	StockCode string
	//最高最低价１最高　２最低价
	DataType int
	//合约代码
	Code            string
	IsMarketOpening bool
	MarketOpenTime  int64
	//	计算平均价格的节点的签名　Sign_Hash值由 Timestamp　DataType BigPrice Code计算
	Sign  []byte
	Price float64 `json:"Price" gorm:"DEFAULT:null;"` //平均价
	// Multiply the Price by 1000000000000000000 to remove decimals
	BigPrice  string
	Timestamp int64 `json:"Timestamp" gorm:"DEFAULT:0;"` //unix 秒数
	//所有节点签名列表
	Signs []StockNode
	//所有节点平均价格签名列表
	AvgSigns []StockNode
}
type StockNode struct {
	//股票代码
	StockCode string
	//最高最低价１最高　２最低价
	DataType int
	//合约代码
	Code        string
	Node        string  //节点名字
	NodeAddress string  //节点地址
	Timestamp   int64   `json:"Timestamp" gorm:"DEFAULT:0;"` //unix 秒数
	Price       float64 //新价
	// Multiply the Price by 1000000000000000000 to remove decimals
	BigPrice string
	// Sign_Hash值由 Timestamp　DataType BigPrice Code计算
	Sign []byte
}

func (s *StockNode) SetSign() {
	//msg:=fmt.Sprintf("%s,%d,%f",s.Code,s.Timestamp, s.Price)
	hash := s.GetHash()
	s.Sign = SignMsg(hash)
}
func (s *StockData) SetSign() {
	//msg := fmt.Sprintf("%s,%d,%f", s.Code, s.Timestamp, s.Price)
	hash := s.GetHash()
	s.Sign = SignMsg(hash)
}

func (Stock) TableName() string {
	return "stocks"
}
func (ViewStock) TableName() string {
	return "stocks"
}
func UsdStockTime() bool {
	now := time.Now().UTC()
	week := now.Weekday()
	if week == 0 || week == 6 {
		//log.Println("周未休息两小时")
		return false
	}

	y, m, d := now.Date()
	stime := time.Date(y, m, d, 13, 30, 0, 0, time.UTC)
	etime := time.Date(y, m, d, 20, 00, 0, 0, time.UTC)
	if now.Unix() < stime.Unix() || now.Unix() > etime.Unix() {
		return false
	}
	return true
}
func GetUsdStockCacheTime() int64 {
	now := time.Now().UTC()
	y, m, d := now.Date()
	week := now.Weekday()
	//周未
	if week == 0 || week == 6 {
		nextMondyDays := 2
		if week == 0 {
			nextMondyDays = 1
		}
		nextOpen := time.Date(y, m, d, 13, 30, 0, 0, time.UTC).Add(24 * time.Hour * time.Duration(nextMondyDays))
		return int64(nextOpen.Sub(now).Seconds())
	}
	stime := time.Date(y, m, d, 13, 30, 0, 0, time.UTC)
	etime := time.Date(y, m, d, 20, 00, 0, 0, time.UTC)
	if now.Unix() < stime.Unix() || now.Unix() > etime.Unix() {
		if now.Unix() < stime.Unix() { //当天未开盘
			return int64(stime.Sub(now).Seconds())
		}
		//当天已经收盘
		return int64(stime.Add(24 * time.Hour).Sub(now).Seconds())
	}
	return 100
}

//苹果代码  AAPL  ,苹果代码 TSLA
func GetStocks() {
	utils.Orm.AutoMigrate(Stock{})
	//科技股列表
	//techUrl := `https://push2.eastmoney.com/api/qt/clist/get?np=1&fltt=2&invt=2&fields=f1,f2,f3,f4,f12,f13,f14&pn=1&pz=60&fid=f3&po=1&fs=b:MK0216&ut=f057cbcbce2a86e2866ab8877db1d059&forcect=1&cb=cbCallback&&callback=jQuery34105542308523132689_1620291099859&_=1620291099869`
	//汽车能源类股列表
	//carUrl := `https://push2.eastmoney.com/api/qt/clist/get?np=1&fltt=2&invt=2&fields=f1,f2,f3,f4,f12,f13,f14&pn=1&pz=30&fid=f3&po=1&fs=b:MK0219&ut=f057cbcbce2a86e2866ab8877db1d059&forcect=1&cb=cbCallback&&callback=jQuery34105741003303298395_1620284654258&_=1620284654273`

	specUrl := "https://push2.eastmoney.com/api/qt/ulist/get?np=1&fltt=2&invt=2&fields=f2,f3,f4,f12,f13,f14,f128&pn=1&pz=30&fid=&po=1&secids=105.AAPL,105.TSLA&ut=f057cbcbce2a86e2866ab8877db1d059&cb=cbCallback&_=1620350834301"
	sleep := 60
	for {
		now := time.Now().UTC()
		//周未休息两小时
		week := now.Weekday()
		if week == 0 || week == 6 {
			log.Println("周未休息两小时")
			time.Sleep(2 * time.Hour)
			continue
		}

		y, m, d := now.Date()
		stime := time.Date(y, m, d, 13, 30, 0, 0, time.UTC)
		etime := time.Date(y, m, d, 20, 00, 0, 0, time.UTC)
		if now.Unix() >= stime.Unix() && now.Unix() <= etime.Unix() {
			GetCarStock(specUrl)
			sleep = 3
			log.Println("休息", sleep, "s")
		} else {
			sleep = int(now.Add(time.Minute).Truncate(time.Minute).Add(time.Second * 2).Sub(now).Seconds())
			if sleep <= 0 {
				sleep = 60
			}
			log.Println("非开盘时段休息", sleep, "s")
		}
		time.Sleep(time.Second * time.Duration(sleep))
	}
}

func GetCarStock(dataUrl string) {
	bs, err := utils.ReqResBody(dataUrl, "https://wap.eastmoney.com/", "GET", nil, nil)
	if err == nil {
		bs = bytes.TrimPrefix(bs, []byte("cbCallback("))
		bs = bytes.TrimSuffix(bs, []byte(");"))
		//log.Println(string(bs))
		res := new(resModel)
		err = json.Unmarshal(bs, res)
		if err == nil {
			//itemsStr := fixFloatNull(fields, string(res.Data.Diff))
			sts := make([]*Stock, 60)
			err = json.Unmarshal([]byte(res.Data.Diff), &sts)
			if err == nil {
				for _, st := range sts {
					st.Timestamp = time.Now().Unix()
				}
				err = utils.Orm.CreateInBatches(sts, 60).Error
				if err == nil {
					log.Println(err)
				}
			}
		}
	}
	if err != nil {
		log.Println(err)
	}
}

/*
	日历支持以下节日排除
	新年元旦，
　　马丁·路德·金纪念日，
　　华盛顿诞辰，
　　耶稣受难日，
　　美国阵亡战士纪念日，
　　美国独立日，
　　美国劳工日，
　　感恩节，
　　圣诞节，
https://www.nyse.com/markets/hours-calendars
https://www.tradinghours.com/markets/nyse/holidays
//夏令时（3月11日至11月7日），冬令时（11月8日至次年3月11日）
//https://zhidao.baidu.com/question/374792000.html
//https://zhidao.baidu.com/question/76638117.html

*/
var stockCalendar *cal.BusinessCalendar

func InitCalendar() {
	if stockCalendar != nil {
		return
	}
	c := cal.NewBusinessCalendar()
	c.Name = "Bigco, Inc."
	c.Description = "Default company calendar"
	c.AddHoliday(
		us.NewYear,
		us.MlkDay,
		us.PresidentsDay,
		aa.GoodFriday,
		us.MemorialDay,
		us.IndependenceDay,
		us.LaborDay,
		//us.ColumbusDay,
		//us.VeteransDay,
		us.ThanksgivingDay,
		us.ChristmasDay,
	)
	c.SetWorkHours(10*time.Hour+30*time.Minute, 16*time.Hour)
	stockCalendar = c

}

func IsMarketTime(timestamp int64) bool {
	tmpCa := *stockCalendar
	if IsSummerTime(timestamp) {
		tmpCa.SetWorkHours(9*time.Hour+30*time.Minute, 16*time.Hour)
	} else {
		tmpCa.SetWorkHours(10*time.Hour+30*time.Minute, 17*time.Hour)
	}
	tmpDate := time.Unix(timestamp, 0)
	tmpDate = tmpDate.In(locUsaStock)
	return tmpCa.IsWorkTime(tmpDate)
}
func IsWorkTime(timestamp int64) (bool, int64) {
	if timestamp == 0 {
		timestamp = time.Now().Unix()
	}
	tmpCa := *stockCalendar
	if IsSummerTime(timestamp) {
		tmpCa.SetWorkHours(10*time.Hour+30*time.Minute, 16*time.Hour)
	} else {
		tmpCa.SetWorkHours(11*time.Hour+30*time.Minute, 17*time.Hour)
	}
	tmpDate := time.Unix(timestamp, 0)
	tmpDate = tmpDate.In(locUsaStock)
	return tmpCa.IsWorkTime(tmpDate), tmpCa.WorkdayStart(tmpDate).Unix()
}
func IsSignTime(timestamp int64) bool {
	if timestamp == 0 {
		timestamp = time.Now().Unix()
	}
	tmpCa := *stockCalendar
	if IsSummerTime(timestamp) {
		tmpCa.SetWorkHours(10*time.Hour+30*time.Minute, 16*time.Hour)
	} else {
		tmpCa.SetWorkHours(11*time.Hour+30*time.Minute, 17*time.Hour)
	}
	tmpDate := time.Unix(timestamp, 0)
	tmpDate = tmpDate.In(locUsaStock)
	return tmpCa.IsWorkTime(tmpDate)
}

var locUsaStock = time.FixedZone("usa-stock", -4*60*60)

//夏令时判断:（3月11日至11月7日），冬令时（11月8日至次年3月11日）
func IsSummerTime(timeStamp int64) bool {
	tmpDate := time.Unix(timeStamp, 0)
	tmpDate = tmpDate.In(locUsaStock)
	y := tmpDate.Year()
	stime := time.Date(y, 3, 11, 0, 0, 0, 0, locUsaStock)
	etime := time.Date(y, 11, 8, 0, 0, 0, 0, locUsaStock)
	return tmpDate.After(stime) && tmpDate.Before(etime)
}
