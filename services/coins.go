package services

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"log"
	"stock/utils"
	"strconv"
	"strings"
	"time"
	"github.com/gorilla/websocket"
	"os"
	"os/signal"
	"net/url"
	"net/http"
)
type mmRes struct {
	C string `json:"C" example:"d-8D0BBC5-B,0|Iz7u,1|Iz7v,1"`
	M []struct {
		A []string `json:"A"`
		H string        `json:"H" example:"MainHub"`
		M string        `json:"M" example:"resubscribe"`
	} `json:"M"`
}

func SaveMMmsg(msg []byte){
	//msg=[]byte(` {"C":"d-9C6E9159-B,0|LD_o,7|LD_p,1","M":[{"H":"MainHub","M":"newMessage","A":["{\"h\":{\"t\":\"/zigman2/quotes/210219788/delayed\",\"a\":\"broadcast\",\"s\":\"EC2-600fd49a\",\"n\":27160305909},\"b\":{\"z\":\"\\/Date(1625192869662)\\/\",\"t\":\"210219788\",\"l\":14526.75,\"v\":8879,\"c\":-21.75,\"y\":\"normal\",\"e\":-5,\"b\":null,\"a\":null,\"em\":0}}"]},{"H":"MainHub","M":"newMessage","A":["{\"h\":{\"t\":\"/zigman2/quotes/210369575/delayed\",\"a\":\"broadcast\",\"s\":\"EC2-600fd49a\",\"n\":27160310300},\"b\":{\"z\":\"\\/Date(1625192871107)\\/\",\"t\":\"210369575\",\"l\":132.328125,\"v\":25860,\"c\":0.0625,\"y\":\"normal\",\"e\":-5,\"b\":null,\"a\":null,\"em\":0}}"]}]}`)
	mres:=new(mmRes)
	err:=json.Unmarshal(msg,mres)
	if err == nil {
		for _, msg := range mres.M {
			if msg.M=="newMessage"{
				if len(msg.A)>0{
					msgstr:=msg.A[0]
					dataObj:=gjson.Get(msgstr, "b")
					idkey:=dataObj.Get("t").String()
					value:=dataObj.Get("l").String()
					z:=dataObj.Get("z").String()[6:16]

					mprice:=new(MarketPrice)
					mprice.ItemType=mmIdMap[idkey]
					mprice.Price,_=strconv.ParseFloat(value,64)
					mprice.TimeStamp,_=strconv.Atoi(z)
					err=utils.Orm.Save(mprice).Error
					if err != nil {
						log.Println(err)
					}
					log.Println("process mm",mmIdMap[idkey],value,z)
				}
			}
		}
	}
	if err != nil {
		log.Println(err)
	}
}
type MarketPrice struct{
	ID uint
	//ftx类型　btc3x, eth3x, vix3x, ust20x, gold10x, eur20x,ndx10x
	ItemType string `gorm:"index"`
	Price float64
	TimeStamp int
	CreatedAt time.Time
}
var mmChanneMap =map[string]string{
	//VX00
	"vix3x": "/zigman2/quotes/210387616/delayed",
	//TY00
	"ust20x": "/zigman2/quotes/210369575/delayed",
	//NQ00
	"ndx10x": "/zigman2/quotes/210219788/delayed",
	//EURUSD
	"eur20x": "/zigman2/quotes/210561242/realtime/sampled",
	//GC00
	"gold10x": "/zigman2/quotes/210034565/delayed",
}

var mmIdMap =map[string]string{
	//VX00
	"210387616":"vix3x",
	//TY00
	"210369575": "ust20x",
	//NQ00
	"210219788": "ndx10x",
	"210034565":"gold10x",
	"210561242":"eur20x",
}


func msgIdx()int {
	msgIndex++
	return msgIndex
}
var msgIndex =0
func SubMM (){
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	defer signal.Stop(interrupt)

	//u := url.URL{Scheme: "wss", Host: "cables.coingecko.com", Path: "/cable"}
	urlstr:=fmt.Sprintf( "wss://mwstream.wsj.net/BG2/signalr/connect?transport=webSockets&clientProtocol=2.1&connectionToken=%s&connectionData=%%5B%%7B%%22name%%22%%3A%%22mainhub%%22%%7D%%5D&tid=7","98770c33-3bf0-4af2-96b1-3749a71e964c%3A" )
	log.Printf("connecting to %s",urlstr )

BEGIN:
	wcHeader:=http.Header{}
	wcHeader.Set( "Origin","https://www.marketwatch.com")
	c, _, err := websocket.DefaultDialer.Dial(urlstr, wcHeader)
	if err != nil {
		log.Println("dial:", err)
		time.Sleep(5*time.Second)
		c.Close()
		goto BEGIN
		//log.Fatal("dial:", err)
	}
	defer c.Close()

	log.Println("sub")
	time.Sleep(5*time.Second)
	//err =c.WriteMessage(websocket.TextMessage,[]byte(`{"H":"mainhub","M":"subscribe","A":["/zigman2/quotes/210387616/delayed","","0"],"I":1}`))
	//err =c.WriteMessage(websocket.TextMessage,[]byte(`{"H":"mainhub","M":"subscribe","A":["/news/metasearch/mktw/wsj/watchlist/company/US/NQ00","","0"],"I":1}`))
	//err =c.WriteMessage(websocket.TextMessage,[]byte(`{"H":"mainhub","M":"subscribe","A":["/news/metasearch/mktw/wsj/watchlist/company/US/TY00","","0"],"I":2}`))
	//err =c.WriteMessage(websocket.TextMessage,[]byte(`{"H":"mainhub","M":"subscribe","A":["/news/metasearch/mktw/wsj/watchlist/company/US/VX00","","0"],"I":3}`))
	for key, channel := range mmChanneMap {
		err =c.WriteMessage(websocket.TextMessage,[]byte(fmt.Sprintf( `{"H":"mainhub","M":"subscribe","A":["%s","","0"],"I":%d}`,channel,msgIdx())))
		if err != nil {
			log.Println(err)
		}else{
			log.Println("sub ",key,channel)
		}
	}

	done := make(chan struct{})
	go func() {
		defer func() {
			log.Println("SubCoinsPrice done")
			close(done)
		}()
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("SubCoinsPrice read err", err)
				return
			}
			msgStr:=string(message)
			//if msgStr==`{"type":"welcome"}`{
			//	c.WriteMessage(websocket.TextMessage,[]byte(`{"command":"subscribe","identifier":"{\"channel\":\"CEChannel\"}"}`))
			//}
			if strings.HasPrefix(msgStr, `{"C":"`) {
				SaveMMmsg(message)
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(30*time.Second)
	defer ticker.Stop()
	connectTimer := time.NewTimer(14*time.Minute)
	defer connectTimer.Stop()
	//time.Sleep(100*time.Second)
	for {
		select {
		case <-done:
			c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			c.Close()
			ticker.Stop()
			connectTimer.Stop()
			goto BEGIN
			//return
		case  <-connectTimer.C:
			c.Close()
			ticker.Stop()
			connectTimer.Stop()
			goto BEGIN
		case  <-ticker.C:
			pingMsg:=fmt.Sprintf(`{"H":"mainhub","M":"ping","A":[],"I":%d}`, msgIdx())
			err := c.WriteMessage(websocket.TextMessage, []byte(pingMsg))
			if err != nil {
				log.Println("hwrite:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}

}
func SubCoinsPrice (){
	utils.Orm.AutoMigrate(Coin{})
	//wss://cables.coingecko.com/cable
	//{command: "subscribe", identifier: "{"channel":"CEChannel"}"}
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: "cables.coingecko.com", Path: "/cable"}
	log.Printf("connecting to %s", u.String())

	BEGIN:
	wcHeader:=http.Header{}
	wcHeader.Set( "Origin","https://www.coingecko.com")
	c, _, err := websocket.DefaultDialer.Dial(u.String(), wcHeader)
	if err != nil {
		log.Println("dial:", err)
		time.Sleep(5*time.Second)
		goto BEGIN
		//log.Fatal("dial:", err)
	}
	defer c.Close()

	log.Println("sub")
	err =c.WriteMessage(websocket.TextMessage,[]byte(`{"command":"subscribe","identifier":"{\"channel\":\"CEChannel\"}"}`))

	done := make(chan struct{})
	go func() {
		defer func() {
			log.Println("SubCoinsPrice done")
			close(done)
		}()
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("SubCoinsPrice read err", err)
				return
			}
			msgStr:=string(message)
			if msgStr==`{"type":"welcome"}`{
				//if err != nil {
				//	log.Println("sub err",err)
				//	return
				//}
				c.WriteMessage(websocket.TextMessage,[]byte(`{"command":"subscribe","identifier":"{\"channel\":\"CEChannel\"}"}`))
			}
			if strings.HasPrefix(msgStr, `{"identifier":"{\"channel\":\"CEChannel\"}"`){
				res := gjson.GetBytes(message, "message.r")
				//log.Println("coins:",res.String())
				coin := new(Coin)
				err=json.Unmarshal([]byte(res.Raw),coin)
				if err != nil {
					log.Println("json err",err)
					continue
				}
				coin.ID=time.Now().Unix()
				err=utils.Orm.Create(coin).Error
				if err != nil {
					log.Println("db err",err)
				}
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(3*time.Second)
	defer ticker.Stop()


	//time.Sleep(100*time.Second)
	for {
		select {
		case <-done:
			c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			goto BEGIN
			//return
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Println("hwrite:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}

}

type Coin struct {
	ID int64  `json:"timestamp"`
	Aed  string `json:"aed" example:"131677.286"`
	Ars  string `json:"ars" example:"3390090.137"`
	Aud  string `json:"aud" example:"46356.355"`
	Bch  string `json:"bch" example:"53.1870804755832"`
	Bdt  string `json:"bdt" example:"3034612.765"`
	Bhd  string `json:"bhd" example:"13514.255"`
	Bits string `json:"bits" example:"1000000.0"`
	Bmd  string `json:"bmd" example:"35850.064"`
	Bnb  string `json:"bnb" example:"108.829611986654"`
	Brl  string `json:"brl" example:"187359.606"`
	Byn  string `json:"byn" example:"90667.9317124824"`
	Cad  string `json:"cad" example:"43281.603"`
	Chf  string `json:"chf" example:"32294.526"`
	Clp  string `json:"clp" example:"25973371.626"`
	Cny  string `json:"cny" example:"228178.489"`
	Czk  string `json:"czk" example:"749097.849"`
	Dkk  string `json:"dkk" example:"218677.182"`
	Dot  string `json:"dot" example:"1743.76558546874"`
	Eos  string `json:"eos" example:"5837.95191057802"`
	Eth  string `json:"eth" example:"14.634477965256"`
	Eur  string `json:"eur" example:"29404.975"`
	Gbp  string `json:"gbp" example:"25283.257"`
	Hkd  string `json:"hkd" example:"278216.216"`
	Huf  string `json:"huf" example:"10237685.096"`
	Idr  string `json:"idr" example:"511981939.084"`
	Ils  string `json:"ils" example:"116413.942"`
	Inr  string `json:"inr" example:"2601201.871"`
	Jpy  string `json:"jpy" example:"3933352.548"`
	Krw  string `json:"krw" example:"39750876.697"`
	Kwd  string `json:"kwd" example:"10783.161"`
	Link string `json:"link" example:"1293.98768705102"`
	Lkr  string `json:"lkr" example:"7108153.621"`
	Ltc  string `json:"ltc" example:"206.358015451212"`
	Mmk  string `json:"mmk" example:"58941940.669"`
	Mxn  string `json:"mxn" example:"714337.627"`
	Myr  string `json:"myr" example:"148240.016"`
	Ngn  string `json:"ngn" example:"14699654.659"`
	Nok  string `json:"nok" example:"300040.23"`
	Nzd  string `json:"nzd" example:"49404.758"`
	Php  string `json:"php" example:"1711124.288"`
	Pkr  string `json:"pkr" example:"5557587.824"`
	Pln  string `json:"pln" example:"131773.83"`
	Rub  string `json:"rub" example:"2627003.09"`
	Sar  string `json:"sar" example:"134460.613"`
	Sats string `json:"sats" example:"100000000.0"`
	Sek  string `json:"sek" example:"297708.398"`
	Sgd  string `json:"sgd" example:"47399.592"`
	Thb  string `json:"thb" example:"1120673.011"`
	Try  string `json:"try" example:"305001.592"`
	Twd  string `json:"twd" example:"988565.56"`
	Uah  string `json:"uah" example:"984617.894"`
	Usd  string `json:"usd" example:"35850.064"`
	Vef  string `json:"vef" example:"3589.666"`
	Vnd  string `json:"vnd" example:"825255819.416"`
	Xag  string `json:"xag" example:"1282.60452596673"`
	Xau  string `json:"xau" example:"18.8237932915211"`
	Xdr  string `json:"xdr" example:"24800.536"`
	Xlm  string `json:"xlm" example:"91330.3831976851"`
	Xrp  string `json:"xrp" example:"35896.5126466017"`
	Yfi  string `json:"yfi" example:"0.838268596014042"`
	Zar  string `json:"zar" example:"493451.04"`
}
