package sys

import (
	"github.com/gin-gonic/gin"
	"stock/common"
	"stock/services"
	"stock/utils"
	"strings"
	"time"
)

// @Tags data
// @Summary　ftx数据查看
// @Description ftx数据查看
// @ID FtxListHandler
// @Accept  json
// @Produce  json
// @Param   CoinType     query    string false  "ftx名字" default()
// @Param   IsAjustPoint     query    string false  "调仓点 3所有 2是调仓点 1非调仓点" default()
// @Param     current   query    int     false        "页码" default(1)
// @Param     pageSize   query    int     false    "每页行数" default(10)
// @Param     sort   query    string     false    "排序" default(id)
// @Success 200 {array} services.CoinBull	"res"
// @Failure 500 {object} common.ResBody "失败时，有相应测试日志输出"
// @Router /pub/ftxs [get]
func FtxListHandler(c *gin.Context) {
	qm := new(qMListFtx)
	err := c.BindQuery(qm)
	if err != nil {
		common.ResErrWithCode(c, err, 500)
		return
	}
	if qm.PageSize==0{
		qm.PageSize=10
	}
	if qm.Page==0{
		qm.Page=1
	}

	items := []*services.CoinBull{}
	dbquery := utils.Orm.Model(services.CoinBull{})
	if qm.CoinType!="" {
		dbquery = dbquery.Where("coin_type=?", qm.CoinType)
	}
	if qm.IsAjustPoint > 0 && qm.IsAjustPoint < 3  {
		dbquery = dbquery.Where("is_ajust_point=?", qm.IsAjustPoint)
	}
	if !qm.StartTime.IsZero() && !qm.EndTime.IsZero(){
		dbquery=dbquery.Where("timestamp> ? and timestamp<?", qm.StartTime.Unix(), qm.EndTime.Unix())
	}
	total := int64(0)
	err = dbquery.Count(&total).Error
	if err == nil {
		if qm.Sort!=""{
			dbquery=dbquery.Order(getSqlSortStr(qm.Sort))
		}else{
			dbquery=dbquery.Order("id desc")
		}
		err = dbquery.Limit(qm.PageSize).Offset(qm.Offset()).Find(&items).Error
		if err == nil {
			common.NewResListBody(c, qm.Page, qm.PageSize, total, items)
			return
		}
	}
	if err != nil {
		common.ResErrWithCode(c, err, 500)
		return
	}
}
func getSqlSortStr(str string) string{
	sqlSortStr:=""
	if !strings.HasPrefix(str,"+") && !strings.HasPrefix(str,"-"){
		sqlSortStr=str+" asc"
		return sqlSortStr
	}
	sqlSortStr=str[1:]
	if str[0]=='+'{
		sqlSortStr+=" asc"
	}
	if str[0]=='-'{
		sqlSortStr+=" desc"
	}
	return sqlSortStr
}
type qMListFtx struct {
	Page     int `example:"1" json:"current"  form:"current"`
	PageSize int `example:"10" json:"pageSize" form:"pageSize"`
	CoinType string `example:"btc3x" json:"CoinType" form:"CoinType"`
	//CreatedAt    time.Time  `example:"btc3x" json:"CreatedAt" form:"CreatedAt"`
	StartTime    time.Time  `example:"2021-09-07T00:00:00+08:00" json:"startTime" form:"startTime"`
	EndTime    time.Time  `example:"2021-09-07T00:00:00+08:00" json:"endTime" form:"endTime"`
	IsAjustPoint int `example:"1" json:"IsAjustPoint" form:"IsAjustPoint"`
	Sort string `example:"+id" json:"sort" form:"sort"`
}
func (qmp *qMListFtx) Offset() int{
	return (qmp.Page-1)*qmp.PageSize
}