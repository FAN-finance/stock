package services

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"stock/utils"
	"strings"
	"time"
)

func SubCoinsPrice (){
	utils.Orm.AutoMigrate(Coin{})
	//wss://cables.coingecko.com/cable
	//{command: "subscribe", identifier: "{"channel":"CEChannel"}"}
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	defer signal.Stop(interrupt)

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
			log.Println("interrupt coingecko")

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
