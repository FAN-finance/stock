package utils

import (
	"gorm.io/gorm"
	"gorm.io/driver/mysql"
	"log"
)
func InitDb(connstr string)  {
	//var connstr="root:emdata2015@tcp(192.168.9.100:3306)/dataExport"
	log.Println("connect:",connstr)
	db, err :=gorm.Open(mysql.Open(connstr), &gorm.Config{})
	if err!=nil{
		log.Println(err);
		panic(err)
	}
	log.Println("InitGormBD,dbType",connstr);
	Orm=db.Debug()
}
var Orm *gorm.DB
