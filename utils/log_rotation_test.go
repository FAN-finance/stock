package utils

import (
	"log"
	"os"
	"syscall"
	"testing"
	"time"
)

func Test_Log(t *testing.T) {
	var timer = time.NewTicker(time.Duration(10) * time.Second)
	var print_timer = time.NewTicker(time.Duration(200) * time.Millisecond)
	var timerLog = time.NewTicker(time.Duration(5) * time.Second)
	defer timer.Stop()
	lfName := "stock.log"
	go RotationLog(lfName)
	for {
		select {
		case <-timer.C:
			return
		case <-print_timer.C:
			log.Println("test print")
		case <-timerLog.C:
			os.Rename(lfName, lfName+".archive."+time.Now().Format("01-02T15:04")+".log")
			err := syscall.Kill(os.Getpid(), syscall.SIGUSR1)
			if err != nil {
				log.Println("syscall.Kill", err)
			}
			//willl remove old file
		}
	}
}