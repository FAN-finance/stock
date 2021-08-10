package utils

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)
/*
 mv stock.log stock.log.0
$ kill -USR1 `cat stock.pid`
$ sleep 1
$ gzip stock.log.0
*/
//接收进程SIGUSR1信号重新打开日志
func RotationLog(logFileName string)() {
	//logFileName="stock.log"
	if logFileName != "" {
		flog, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalln("open logfile Err", err)
		}
		log.SetOutput(flog)

		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, syscall.SIGUSR1)
		defer signal.Stop(interrupt)
		for {
			sigNum := os.Interrupt
			select {
			case sigNum = <-interrupt:
				log.Println("reOpen log sigNum", sigNum)
				cerr := flog.Close()
				if cerr != nil {
					log.Println("colse log err", err)
				}

				flog, err = os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
				if err != nil {
					log.Println("re open logfile Err", err)
				} else {
					log.Println("re open logfile", logFileName)
					log.SetOutput(flog)
					log.Println("re open logfile", logFileName)
				}
			}
		}
	}
}
