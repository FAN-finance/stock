package services

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"stock/utils"
	"strconv"
	"strings"
	"time"
)

type cmRes struct {
	D struct {
		Cr struct {
			As       float64   `json:"as" example:"46171815477"`
			D        float64 `json:"d" example:"2.061100"`
			Fmc      float64 `json:"fmc" example:"61449390494.777206"`
			Fmc24hpc float64 `json:"fmc24hpc" example:"-1.434785"`
			ID       int   `json:"id" example:"52"`
			Mc       float64 `json:"mc" example:"28372299190.989712"`
			Mc24hpc  float64 `json:"mc24hpc" example:"-1.434785"`
			P        float64 `json:"p" example:"0.614494"`
			P1h      float64 `json:"p1h" example:"-0.059023"`
			P24h     float64 `json:"p24h" example:"-1.434785"`
			P30d     float64 `json:"p30d" example:"-28.712166"`
			P7d      float64 `json:"p7d" example:"-4.272356"`
			Ts       float64   `json:"ts" example:"99990349851"`
			V        float64 `json:"v" example:"2260174210.655913"`
			Vol24hpc float64 `json:"vol24hpc" example:"-2.796830"`
		} `json:"cr"`
		T int64 `json:"t" example:"1625823360223"`
	} `json:"d"`
	ID string `json:"id" example:"price"`
	S  string `json:"s" example:"0"`
}
var cmPrePrice=0.0
func SaveCmmsg(msg []byte) {
	//msg=[]byte(` {"C":"d-9C6E9159-B,0|LD_o,7|LD_p,1","M":[{"H":"MainHub","M":"newMessage","A":["{\"h\":{\"t\":\"/zigman2/quotes/210219788/delayed\",\"a\":\"broadcast\",\"s\":\"EC2-600fd49a\",\"n\":27160305909},\"b\":{\"z\":\"\\/Date(1625192869662)\\/\",\"t\":\"210219788\",\"l\":14526.75,\"v\":8879,\"c\":-21.75,\"y\":\"normal\",\"e\":-5,\"b\":null,\"a\":null,\"em\":0}}"]},{"H":"MainHub","M":"newMessage","A":["{\"h\":{\"t\":\"/zigman2/quotes/210369575/delayed\",\"a\":\"broadcast\",\"s\":\"EC2-600fd49a\",\"n\":27160310300},\"b\":{\"z\":\"\\/Date(1625192871107)\\/\",\"t\":\"210369575\",\"l\":132.328125,\"v\":25860,\"c\":0.0625,\"y\":\"normal\",\"e\":-5,\"b\":null,\"a\":null,\"em\":0}}"]}]}`)
	res := new(cmRes)
	err := json.Unmarshal(msg, res)
	if err == nil {
		mprice := new(MarketPrice)
		idkey:=strconv.Itoa(res.D.Cr.ID)
		mprice.ItemType =cmIdMap[idkey]
		mprice.Price= res.D.Cr.P
		mprice.Timestamp =int(res.D.T/1000)
		if  mprice.Price!=cmPrePrice {//价格有变时才新增
			cmPrePrice=mprice.Price
			err = utils.Orm.Save(mprice).Error
			if err != nil {
				log.Println(err)
			}
		}else{
			log.Println("skip price",cmPrePrice)
		}
		log.Println("process mm", cmIdMap[idkey], mprice.Price,mprice)
	}
	if err != nil {
		log.Println(err)
	}
}
//{"method":"subscribe","id":"price","data":{"cryptoIds":[9207,1,1027,2010,1839,6636,52,1975,2,512,1831,7083,74,9023,9022],"index":"detail"}}
var cmCoinMap =map[string]string{
	//Metaverse Index
	"mvi": "9207",
}

var cmIdMap =map[string]string{
	//Metaverse Index
	"9207": "mvi",
}


func SubCM(){
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	defer signal.Stop(interrupt)

	//u := url.URL{Scheme: "wss", Host: "cables.coingecko.com", Path: "/cable"}
	urlstr:=fmt.Sprintf( "wss://stream.coinmarketcap.com/price/latest")
	log.Printf("connecting to %s",urlstr )

BEGIN:
	wcHeader:=http.Header{}
	wcHeader.Set( "Origin","https://coinmarketcap.com/")
	c, _, err := websocket.DefaultDialer.Dial(urlstr, wcHeader)
	if err != nil {
		log.Println("dial:", err)
		time.Sleep(5*time.Second)
		if c != nil {
			c.Close()
		}
		goto BEGIN
		//log.Fatal("dial:", err)
	}
	defer c.Close()

	//time.Sleep(5*time.Second)
	ids:=[]string{}
	for _, channel := range cmCoinMap {
		ids=append(ids,channel)

	}
	subMsg:=fmt.Sprintf(`{"method":"subscribe","id":"price","data":{"cryptoIds":[%s],"index":"detail"}}`,strings.Join(ids,","))
	err =c.WriteMessage(websocket.TextMessage,[]byte(subMsg))
	if err != nil {
		log.Println(err)
	}
	log.Println("submsg",subMsg)
	done := make(chan struct{})
	go func() {
		defer func() {
			log.Println("SubCMPrice done")
			close(done)
		}()
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("SubCMPrice read err", err)
				return
			}
			msgStr:=string(message)
			//if msgStr==`{"type":"welcome"}`{
			//	c.WriteMessage(websocket.TextMessage,[]byte(`{"command":"subscribe","identifier":"{\"channel\":\"CEChannel\"}"}`))
			//}
			if strings.HasPrefix(msgStr, `{`) {
				SaveCmmsg(message)
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
			continue
			pingMsg:=fmt.Sprintf(`{"H":"mainhub","M":"ping","A":[],"I":%d}`, msgIdx())
			err := c.WriteMessage(websocket.TextMessage, []byte(pingMsg))
			if err != nil {
				log.Println("hwrite:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt coinMarket")

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

func MarketCapCharData(){
	stime:=int(time.Date(2021,6,1,0,0,0,0,time.UTC).Unix())
	bs, err := utils.ReqResBody("https://api.coinmarketcap.com/data-api/v3/cryptocurrency/detail/chart?id=9207&range=3M", "https://coinmarketcap.com/", "GET", nil, nil)
	if err == nil {
		data:=map[string]struct{V []float64}{}
		err=json.Unmarshal([]byte(gjson.GetBytes(bs,"data.points").Raw),&data)
		if err == nil {
			keys:=[]string{}
			for tsStr := range data {
				keys=append(keys,tsStr)
			}
			sort.Strings(keys)
			for _, key := range keys {
				ts,_:=strconv.Atoi(key)
				if ts>stime{
					log.Println( ts,time.Unix(int64(ts),0),data[key].V[0])
					mprice := new(MarketPrice)
					mprice.ItemType =cmIdMap["9207"]
					mprice.Price= data[key].V[0]
					mprice.Timestamp =ts
					err=utils.Orm.Save(mprice).Error
					if err != nil {
						log.Println(err)
					}

				}

			}
		}

	}
}