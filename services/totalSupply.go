package services

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"stock/contracts"

	"github.com/robfig/cron/v3"
	"stock/utils"
	"time"
)



type TokenTotalSupply struct {
	ID          int `gorm:"primaryKey"`
	Token       string
	TotalSupply string
	DailyAmount string
	CreateAt    time.Time
}

var TokenAddressArr = []string{
	"0x011864d37035439e078d64630777ec518138af05",
}

var cronObj *cron.Cron = nil

func TokenTotalSupplyDailyData() {
	utils.Orm.AutoMigrate(TokenTotalSupply{})

	for _, address := range TokenAddressArr {
		contract, err := contracts.NewRei(common.HexToAddress(address), utils.EthConn)
		if err == nil {
			lastInfo := &TokenTotalSupply{}
			err = utils.Orm.Model(TokenTotalSupply{}).Where("token=?", address).Order("create_at DESC").Limit(1).First(lastInfo).Error
			if err != nil || lastInfo.ID == 0 {
				res, err := contract.TotalSupply(nil)
				if err == nil {
					info := &TokenTotalSupply{}
					info.Token = address
					info.DailyAmount = "0"
					info.TotalSupply = res.String()
					info.CreateAt = time.Now()
					utils.Orm.Save(info)
				}
			}
		}
	}

	if cronObj == nil {
		location, _ := time.LoadLocation("UTC")
		cronObj = cron.New(cron.WithLocation(location))
	}

	cronObj.AddFunc("0 0 * * *", func() {
		for _, address := range TokenAddressArr {
			contract, err := contracts.NewRei(common.HexToAddress(address), utils.EthConn)
			if err == nil {
				res, err := contract.TotalSupply(nil)
				if err == nil {
					currTotalSupply, _ := decimal.NewFromString(res.String())
					info := &TokenTotalSupply{}
					info.Token = address
					info.DailyAmount = "0"
					info.TotalSupply = res.String()
					info.CreateAt = time.Now()

					lastInfo := &TokenTotalSupply{}
					err = utils.Orm.Model(TokenTotalSupply{}).Where("token=?", info.Token).Order("create_at DESC").Limit(1).First(lastInfo).Error
					if err == nil && lastInfo.ID > 0 {
						lastTotalSupply, _ := decimal.NewFromString(lastInfo.TotalSupply)
						info.DailyAmount = currTotalSupply.Sub(lastTotalSupply).String()
					}
					utils.Orm.Save(info)
				}
			}
		}
	})
	cronObj.Start()
}

func TokenChartSupply(token string, amt int) (data []TokenTotalSupply, err error) {
	data = []TokenTotalSupply{}
	find := utils.Orm.Model(TokenTotalSupply{}).Where("token=?", token).Order("create_at DESC").Limit(amt).Find(&data)

	for find.Error == nil {
		return data, nil
	}
	return nil, find.Error
}

func GetTokenTotalSupply(token string) float64 {
	isExist := false
	for _, address := range TokenAddressArr {
		if address == token {
			isExist = true
			break
		}
	}
	if isExist {
		contract, err := contracts.NewRei(common.HexToAddress(token), utils.EthConn)
		if err == nil {
			res, err := contract.TotalSupply(nil)
			if err == nil {
				currTotalSupply, _ := decimal.NewFromString(res.String())
				temp, _ := decimal.NewFromString("1000000000000000000")
				currTotalSupply = currTotalSupply.Div(temp)
				totalSupply, _ := currTotalSupply.Float64()
				return totalSupply
			}
		}
	}
	return 0
}
