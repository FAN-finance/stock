package sys

import (
	"testing"
	"time"
)

func TestWatchDicFile(t *testing.T) {
	WatchDicFile("/home/wxf/go/src/ethproj/stock/asset/dic_config.json")
	time.Sleep(30*time.Second)
}