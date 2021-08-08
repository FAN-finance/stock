package utils

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"
	"io/ioutil"
	//"math/rand"
	//"net/url"
	//"strings"
)

var DebugReq = true

func GetUa()string{
	return "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/80.0.3987.87 Chrome/80.0.3987.87 Safari/537.36"
}
func ReqResBody(url,ref, method string, header http.Header, bodyBs []byte) (bs []byte, err error) {
	resp, err1 := ReqRes(url,ref, method, header, bodyBs)
	err = err1
	if err == nil && resp!=nil {
		bs, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			err = errors.New(err.Error() + " " + string(bs))
			return
		}
	}
	return
}
var hclient = http.DefaultClient
func ReqRes(url, ref, method string, header http.Header, bodybs []byte) (resp *http.Response, err error) {
	reqTime:=time.Now()
	if DebugReq {
		log.Println(method, ":", url)
	}
	retryMaxDefault := 1
	retry := 0
	retryMax := retryMaxDefault
	breader := bytes.NewReader(bodybs)
	request, err1 := http.NewRequest(method, url, breader)
	err = err1
	if err == nil {
		request.Header.Set("User-Agent", GetUa())
		request.Header.Set("Referer", ref)
		for k, v := range header {
			request.Header.Set(k, v[0])
		}
	RETRY:
		//hclient := HClient()
		hclient.Timeout = 2 * time.Second
		resp, err = hclient.Do(request)
		err = err1
		if err == nil {
			if resp == nil {
				log.Println("retry resp nil")
				retry += 1
				if retry < 2 {
					time.Sleep(time.Millisecond * 100)
					goto RETRY
				}
				err = errors.New("err resp nil")
				log.Println("ReqRes", err.Error(),url,reqTime.Unix())
			} else if resp.StatusCode != 200 && resp.StatusCode != 206 {
				err = errors.New("err status:" + resp.Status)
			}
		} else {
			//net error
			//log.Printf("net err set retryMax %d to %d", retryMaxDefault, retryMax)
			//retryMax = 1
			//time.Sleep(5 * time.Second)
		}
		if err != nil {
			retry += 1
			if retry < retryMax && method != "HEAD" {
				time.Sleep(time.Millisecond * 100)
				goto RETRY
			}
		}
	}
	if DebugReq {
		if retry > 1 {
			log.Println("retry", retry, url,time.Now().Sub(reqTime).Seconds())
		}
		if err == nil {
			clen, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
			if clen > 1000000 {
				log.Println("datasize gt 1M:", url, float32(clen)/1024/1024, "M")
			}
		}
	}
	return
}
