package controls

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"stock/common"
	"stock/services"
	"stock/utils"
	"strconv"
	"strings"
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

		var command= exec.Command("bash","-c",fmt.Sprintf( "mysqldump -u root -h 127.0.0.1 -P 3306 stock %s |gzip -c > %s",tableName, fileName))
		out, err1 := command.Output()

		err=err1
		log.Println("DbExportHandler Output: " + string(out))
		if err != nil {
			log.Println("DbExportHandler mysqldump err",err.Error())
			common.ErrJson(c,err.Error())
			return
		}
		c.FileAttachment(fileName, path.Base(fileName))
	}
}

func FtxPriceCheck(c *gin.Context) {
	msg:=""
	for coin_type, _ := range ftxAddres {
		if coin_type=="usd"{continue}
		cb:=new(services.CoinBull)
		err:= utils.Orm.Order("id desc").First(cb,"coin_type=?",coin_type).Error
		if err != nil {
			msg=err.Error()
			break
		}
		pr:=safePrice[ftxAddres[coin_type]]
		if pr==nil{continue}
		log.Println(cb.Bull,pr)
		if (cb.Bull>pr.Max||cb.Bull<pr.Min){
			msg+=fmt.Sprintf("important! %s currentPrice: %.2f  out of range %.2f-%.2f \n",coin_type,cb.Bull,pr.Min,pr.Max)
		}else if (cb.Bull*1.1>pr.Max||cb.Bull*0.9<pr.Min) {
			msg += fmt.Sprintf("%s currentPrice: %.2f will out of range %.2f-%.2f \n", coin_type, cb.Bull, pr.Min, pr.Max)
		}
	}
	if msg!=""{
		uname:=c.Query("uname")
		pwd:=c.Query("pwd")
		to:=[]string{}
		tostr:=c.Query("to")
		if tostr==""{
			to = []string{"xiaofei.wu@rchaintech.com"}
		}else{
			to=strings.Split(tostr,",")
			to=append(to,"xiaofei.wu@rchaintech.com")
		}

		Mail("FtxPriceCheck",msg,uname,pwd,to)
		c.Writer.Write([]byte(msg))
	}
}
func BtcSignCheckHandler(c *gin.Context) {
	res,err:=ftxPriceSignHandler("btc3x",1,238299929)
	if err != nil {
		common.ErrJson(c,err.Error())
		return
	}
	//res.Sign=nil
	if  IsDisableAllSign==false && IsDisableFtxSign==false && SpecialOpenTime() && res.Sign==nil {
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
		common.ErrJson(c,err.Error())
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


// @Tags default
// @Summary　当前节点状态:记录数,钱包地址
// @Description 当前节点状态:记录数,钱包地址
// @ID NodeStatHandler
// @Accept  json
// @Produce  json
// @Success 200 {object} NodeStat	"node stat"
//@Header 200 {string} sign "签名信息"
// @Failure 500 {object} common.ApiErr "失败时，有相应测试日志输出"
// @Router /pub/stock/stat [get]
func NodeStatHandler(c *gin.Context) {
	stat := NodeStat{}
	utils.Orm.Model(services.ViewStock{}).Count(&(stat.StockRows))
	utils.Orm.Model(services.ViewStock{}).Select("max(updated_at) ").Scan(&stat.StockUpdateAt)

	coinMaxid := 0
	utils.Orm.Model(services.Coin{}).Select("max(id)").Scan(&coinMaxid)
	stat.CionPricesUpdateAt = time.Unix(int64(coinMaxid), 0)

	utils.Orm.Model(services.BlockPrice{}).Select("max(created_at)").Scan(&stat.BlockPricesUpdateAt)

	stat.WalletAddre = services.WalletAddre
	stat.Uptime = services.Uptime
	c.JSON(200, stat)
}

type NodeStat struct {
	//节点名
	Node string
	//钱包地址
	WalletAddre string
	//股票信息数据库记录数
	StockRows int64
	//股票信息最后更新时间
	StockUpdateAt time.Time
	//币价换算信息最后更新时间
	CionPricesUpdateAt time.Time
	//eth价格信息最后更新时间
	BlockPricesUpdateAt time.Time
	Uptime              time.Time
}

// @Tags default
// @Summary　所有节点状态:记录数,钱包地址
// @Description 所有节点状态:记录数,钱包地址
// @ID NodeStatsHandler
// @Accept  json
// @Produce  json
// @Success 200 {array} NodeStat	"Node Stat list"
// @Failure 500 {object} common.ApiErr "失败时，有相应测试日志输出"
// @Router /pub/stock/stats [get]
func NodeStatsHandler(c *gin.Context) {
	var err error

	var addres []*NodeStat
	sc := sync.RWMutex{}
	wg := new(sync.WaitGroup)
	var porcNode = func(nodeUrl string) {
		defer wg.Done()
		reqUrl := nodeUrl + "/pub/stock/stat"
		bs, err := utils.ReqResBody(reqUrl, "", "GET", nil, nil)
		if err == nil {
			stat := new(NodeStat)
			err = json.Unmarshal(bs, stat)
			if err == nil {
				log.Println(err)
			}
			stat.Node = nodeUrl
			sc.Lock()
			addres = append(addres, stat)
			sc.Unlock()
		}
	}
	for _, nurl := range utils.Nodes {
		wg.Add(1)
		go porcNode(nurl)
	}
	wg.Wait()
	c.JSON(200, addres)
	return

	if err != nil {
		common.ErrJson(c, err.Error())
		return
	}
}

func Mail(title ,msgBody string,uname ,pwd string, to []string ) {
	log.Println("begin send mail",title)
	auth := sasl.NewPlainClient("", uname,pwd)
	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	msg := strings.NewReader("To: wxf4150@163.com\r\n" +
		"Subject: "+title+"\r\n" +
		"\r\n" +
		msgBody+"\r\n")
	//err := smtp.SendMail("smtp.exmail.qq.com:465", auth, "xiaofei.wu@rchaintech.com", to, msg)
	err := sendMail("smtp.exmail.qq.com:465", auth, "xiaofei.wu@rchaintech.com", to, msg)
	if err != nil {
		log.Println("stmp send mail err",err)
	}
}
func sendMail(addr string, a sasl.Client, from string, to []string, r io.Reader) error {
	c, err := smtp.DialTLS(addr,nil)
	if err != nil {
		return err
	}
	defer c.Close()
	if err = c.Hello("test"); err != nil {
		return err
	}
	if a != nil {
		if ok, _ := c.Extension("AUTH"); !ok {
			return errors.New("smtp: server doesn't support AUTH")
		}
		if err = c.Auth(a); err != nil {
			return err
		}
	}
	if err = c.Mail(from, nil); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = io.Copy(w, r)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}

type VerObj struct {
	//stockInfo json: {"code":"AAPL","price":128.1,"name":"苹果","timestamp":1620292445,"UpdatedAt":"2021-05-06T17:14:05.878+08:00"}
	Data json.RawMessage `swaggertype:"object"`
	Sign []byte          `swaggertype:"string" format:"base64" example:"UhRVNsT8B5Za6oO3APH0T9ebPMKHxDDhkscYuILl7lDepDMzyBaQsEu9vwTRIfoYBS8udfEanI/DUAhwnIdFJf9woIv7Oo+OS6q3sF3B5Vx9NN2ipXJ4wjTf2ct7FbS1vXAvTXSmA2svj+LF8P1PIEClITBqu/EWZXTpHvAlbGAAeF+hHO7/FquLHVDavLC+OENyb0CP+NvH+ytZ69tav0DqbGp+NGGil/ImZpPsetbOxwuhC/U1CV6Ap8qgRWe8s6IpOawXDAavLMHUmXVvORDf/XVzaQUJ5ob+vTsSTZwQsvj/4jmsODFt8eKFYL/7vyN/i3HkiDwhq0w85kqHgg=="`
}

//// @Tags default
//// @Summary　签名验证:
//// @Description 签名验证
//// @ID VerifyInfoHandler
//// @Accept  json
//// @Produce  json
//// @Param     verObj   body    VerObj     true        "需要验证的对象" default(AAPL)
//// @Success 200 {object} ApiOk	"ok info"
//// @Failure 500 {object}common.ApiErr "失败时，有相应测试日志输出"
//// @Router /pub/stock/sign_verify [post]
//func VerifyInfoHandler(c *gin.Context) {
//	vobj:=new(VerObj)
//	err:=c.Bind(vobj)
//	if err == nil {
//		hashbs:=sha256.Sum256(vobj.Data)
//		err=rsa.VerifyPKCS1v15(LocalCert.PublicKey.(*rsa.PublicKey),crypto.SHA256,hashbs[0:32],vobj.Sign,)
//		if err == nil {
//			c.JSON(200, ApiOk{"ok"})
//			return
//		}
//	}
//	if err != nil {
//		ErrJson(c,err.Error())
//		return
//	}
//}


