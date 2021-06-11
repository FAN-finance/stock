package utils

import (
	"fmt"
	"github.com/hashicorp/golang-lru"
	"hash/fnv"
	"log"
	"time"
)

var lcache *lru.Cache
func init(){
	var err error
	lcache,err=lru.New(80000)
	if err != nil {
		log.Println("lru cache error",err)
	}else{
		log.Println("lru cache init")
	}
}

func CacheFromLruWithFixKey(key string, myfunc func()(interface{},error))(interface{},error){
	res,ok:=lcache.Get(key)
	if ok{
		log.Println("lru hit key",key)
		return res,nil
	}
	log.Println("lru miss key",key)

	obj,err:=myfunc()
	if err == nil {
		lcache.Add(key,obj)
	}
	return obj,err
}
func CacheFromLru(version int ,key string,ttl int,myfunc func()(interface{},error))(interface{},error){
	if ttl>0{
		key=fmt.Sprintf("%s-%d-%d",key,version,CalcExpiration(int64(ttl),key))
	}
	return CacheFromLruWithFixKey(key, myfunc)

}
func CalcExpiration(ttl int64, key string) int64 {
	if ttl<60{
		log.Println("错误的ttl 应该大于60")
	}
	//we calculate the non discrete expiration, relative to current time
	now:=time.Now().Unix()
	expires := now
	var padding int64 = 0
	h := fnv.New32a()
	h.Write([]byte(key))
	padding = int64(h.Sum32()) % 60
	//ran := rand.New(rand.NewSource(padding))
	//rvalue := ran.Int63n(60)
	ttl -= padding

	expires += (ttl - (expires % ttl))
	//log.Println("padding:=", padding, "ttl",ttl,"expires",expires,"now",now,"e-n",expires-now)
	return expires
}
