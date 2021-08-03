package controls

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"os/exec"
	"path"
	"stock/utils"
	"strconv"
	"sync"
	"time"
)

var backupLock =sync.Mutex{}
func DbExportHandler(c *gin.Context) {
	tableName:=c.Query("t")
	fileName:=""
	if tableName==""{
		fileName=fmt.Sprintf("asset/db_all_%s.sql.gz",time.Now().Truncate(time.Minute*5).Format("2006010215:04"))
	}else{
		fileName=fmt.Sprintf("asset/db_%s_%s.sql.gz",tableName,time.Now().Truncate(time.Minute*5).Format("2006010215:04"))
	}

	isExists:=true
	_,err:=os.Stat(fileName)
	if err != nil && os.IsNotExist(err) {
		isExists=false
	}
	if isExists{
		log.Println("DbExportHandler download exists file",fileName)
		c.FileAttachment(fileName, path.Base(fileName))
		return
	}else{
		backupLock.Lock()
		defer backupLock.Unlock()
		_,err:=os.Stat(fileName)
		if err == nil ||( err!=nil && !os.IsNotExist(err) ){
			c.FileAttachment(fileName, path.Base(fileName))
			log.Println("DbExportHandler download exists file",fileName)
			return
		}

		var command= exec.Command("bash","-c",fmt.Sprintf( "mysqldump --column-statistics=0 -u root -h 127.0.0.1 -P 3306 stock %s |gzip -c > %s",tableName, fileName))
		out, err1 := command.Output()

		err=err1
		log.Println("DbExportHandler Output: " + string(out))
		if err != nil {
			log.Println("DbExportHandler mysqldump err",err.Error())
			ErrJson(c,err.Error())
			return
		}
		c.FileAttachment(fileName, path.Base(fileName))
	}
}

func BtcSignCheckHandler(c *gin.Context) {
	res,err:=ftxPriceSignHandler("btc3x",1,238299929)
	if err != nil {
		ErrJson(c,err.Error())
		return
	}
	//res.Sign=nil
	if res.Sign==nil{
		time.Sleep(3*time.Second)
		c.JSON(200,"BtcSignCheck err")
		return
	}
	c.JSON(200,"BtcSignCheck ok")
}
func CoinDataCheckHandler(c *gin.Context) {
	d,_:=strconv.Atoi(c.Query("d"))
	if d==0{
		d=300
	}

	counter:=0
	err:=utils.Orm.Raw(`
select count(1) cc from  market_prices t where
t.item_type in('btc','eth') and timestamp >unix_timestamp()-?;
`,d).Scan(&counter).Error
	if err != nil {
		ErrJson(c,err.Error())
		return
	}
	if counter==0{
		time.Sleep(3*time.Second)
		c.JSON(200,"CoinData err")
		return
	}
	c.JSON(200,fmt.Sprintf("CoinData ok counter%d",counter))
}
func OkHandler(c *gin.Context) {
	c.JSON(200,"ok")
}

