package services

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
	"stock/utils"
	"fmt"
	"net/http"
)

type mwRes struct {
	C string `json:"C" example:"d-8D0BBC5-B,0|Iz7u,1|Iz7v,1"`
	M []struct {
		A []string `json:"A"`
		H string        `json:"H" example:"MainHub"`
		M string        `json:"M" example:"resubscribe"`
	} `json:"M"`
}

func SaveMwmsg(msg []byte){
	//msg=[]byte(` {"C":"d-9C6E9159-B,0|LD_o,7|LD_p,1","M":[{"H":"MainHub","M":"newMessage","A":["{\"h\":{\"t\":\"/zigman2/quotes/210219788/delayed\",\"a\":\"broadcast\",\"s\":\"EC2-600fd49a\",\"n\":27160305909},\"b\":{\"z\":\"\\/Date(1625192869662)\\/\",\"t\":\"210219788\",\"l\":14526.75,\"v\":8879,\"c\":-21.75,\"y\":\"normal\",\"e\":-5,\"b\":null,\"a\":null,\"em\":0}}"]},{"H":"MainHub","M":"newMessage","A":["{\"h\":{\"t\":\"/zigman2/quotes/210369575/delayed\",\"a\":\"broadcast\",\"s\":\"EC2-600fd49a\",\"n\":27160310300},\"b\":{\"z\":\"\\/Date(1625192871107)\\/\",\"t\":\"210369575\",\"l\":132.328125,\"v\":25860,\"c\":0.0625,\"y\":\"normal\",\"e\":-5,\"b\":null,\"a\":null,\"em\":0}}"]}]}`)
	mres:=new(mwRes)
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
					mprice.ItemType= mwIdMap[idkey]
					mprice.Price,_=strconv.ParseFloat(value,64)
					mprice.TimeStamp,_=strconv.Atoi(z)
					err=utils.Orm.Save(mprice).Error
					if err != nil {
						log.Println(err)
					}
					log.Println("process mm", mwIdMap[idkey],value,z)
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
	//ftx类型　btc3x, eth3x, vix3x, ust20x, gold10x, eur20x,ndx10x,govt20x
	ItemType string `gorm:"uniqueIndex:type_timestamp,priority:1"`
	Price float64
	TimeStamp int `gorm:"uniqueIndex:type_timestamp,priority:2"`
	CreatedAt time.Time
}
var mwChanneMap =map[string]string{
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

var mwIdMap =map[string]string{
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
func SubMw(){
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
	for key, channel := range mwChanneMap {
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
				SaveMwmsg(message)
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