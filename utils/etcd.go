package utils
import(
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"strings"
	"time"
)

var EtcdCli *clientv3.Client
func InitEtcd(endpoins string) {
	cfg := clientv3.Config{
		Endpoints:strings.Split(endpoins,","),
		DialTimeout: 5* time.Second,
	}
	cli, err := clientv3.New(cfg)
	if err != nil {
		log.Fatal("etcd connect",err)
	}
	EtcdCli=cli
	log.Println("etcd ok")
}