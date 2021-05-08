package services

import (
	"bytes"
	"encoding/json"
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
	Code      string  `gorm:"index:code_time,priority:1" json:"f12"` //代码
	Price     float32 `json:"f2" gorm:"DEFAULT:null;"`            //最新价
	StockName string  `json:"f14"`                                //名称
	Mk        int     `json:"f13"`                                //市场 1 sh 01 0 sz 02
	Diff      float32 `json:"f3" gorm:"DEFAULT:0;"`               //最新涨百分比
	Timestamp  int64 `gorm:"index:code_time,priority:2;" json:"timestamp" ` //unix 秒数
	UpdatedAt *time.Time
}
type ViewStock struct {
	Code      string  `json:"Code" gorm:"column:code"` //代码 苹果代码 AAPL ,特斯拉代码 TSLA
	Price     float32 `json:"Price" gorm:"DEFAULT:null;"`        //最新价
	StockName string  `json:"StockName" `                            //名称
	Timestamp  int64 `json:"Timestamp" gorm:"DEFAULT:0;"` //unix 秒数
	//rfc3339 fortmat
	UpdatedAt *time.Time  `json:"UpdatedAt" example:"2021-05-07T18:25:44.27+08:00"`
}

func (Stock) TableName() string {
	return "stocks"
}
func (ViewStock) TableName() string {
	return "stocks"
}


//苹果代码  AAPL  ,苹果代码 TSLA
func GetStocks() {
	utils.Orm.AutoMigrate(Stock{})
	//科技股列表
	//techUrl := `https://push2.eastmoney.com/api/qt/clist/get?np=1&fltt=2&invt=2&fields=f1,f2,f3,f4,f12,f13,f14&pn=1&pz=60&fid=f3&po=1&fs=b:MK0216&ut=f057cbcbce2a86e2866ab8877db1d059&forcect=1&cb=cbCallback&&callback=jQuery34105542308523132689_1620291099859&_=1620291099869`
	//汽车能源类股列表
	//carUrl := `https://push2.eastmoney.com/api/qt/clist/get?np=1&fltt=2&invt=2&fields=f1,f2,f3,f4,f12,f13,f14&pn=1&pz=30&fid=f3&po=1&fs=b:MK0219&ut=f057cbcbce2a86e2866ab8877db1d059&forcect=1&cb=cbCallback&&callback=jQuery34105741003303298395_1620284654258&_=1620284654273`

	specUrl:="https://push2.eastmoney.com/api/qt/ulist/get?np=1&fltt=2&invt=2&fields=f2,f3,f4,f12,f13,f14,f128&pn=1&pz=30&fid=&po=1&secids=105.AAPL,105.TSLA&ut=f057cbcbce2a86e2866ab8877db1d059&cb=cbCallback&_=1620350834301"
	sleep := 60
	for {
		y, m, d := time.Now().UTC().Date()
		stime := time.Date(y, m, d, 13, 30, 0, 0, time.UTC)
		etime := time.Date(y, m, d, 20, 00, 0, 0, time.UTC)
		if time.Now().Unix() >= stime.Unix() && time.Now().Unix() <= etime.Unix() {
			GetCarStock(specUrl)
			sleep = 1
			log.Println("休息", sleep, "s")
		} else {
			sleep = int( time.Now().Add(time.Minute).Truncate(time.Minute).Add(time.Second*2).Sub(time.Now()).Seconds())
			if sleep <=0 {
				sleep = 60
			}
			log.Println("非开盘时段休息", sleep, "s")
		}
		time.Sleep(time.Second * time.Duration(sleep))
	}
}

func GetCarStock(dataUrl string) {
	bs, err := utils.ReqResBody(dataUrl, "https://wap.eastmoney.com/", "GET",nil,nil)
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
					st.Timestamp=time.Now().Unix()
				}
				err=utils.Orm.CreateInBatches(sts,60).Error
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
