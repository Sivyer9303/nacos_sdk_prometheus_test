package main

import (
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/prometheus/common/log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

func main() {
	//// 启动同名服务
	//var wg sync.WaitGroup
	//tick := time.Tick(time.Second * 15)
	//portStart := 7788
	//// 启动一个线程，一直获取service list
	//go func() {
	//	tick := time.Tick(time.Second * 5)
	//	client, err := getNamingClient()
	//	if err != nil {
	//		log.Fatal("err ... ", err)
	//	}
	//	for {
	//		select {
	//		case <-tick:
	//			_, err := client.GetService(vo.GetServiceParam{
	//				ServiceName: "prometheuse_service",
	//				GroupName:   "DEFAULT_GROUP",
	//			})
	//			if err != nil {
	//				log.Fatal("get service fail ...", err)
	//			}
	//			//bytes, err := json.Marshal(service)
	//			//if err != nil {
	//			//	log.Fatal("Marshal service fail ...",err)
	//			//}
	//			//fmt.Println(string(bytes))
	//		}
	//	}
	//}()
	//for i := 0; i < 1; i++ {
	//	wg.Add(1)
	//	select {
	//	case <-tick:
	//		// 每三十秒注册一个
	//		rand.Seed(time.Now().Unix())
	//		go registerSelf(portStart+rand.Intn(1000), wg)
	//	}
	//}
	//wg.Wait()

	// 注册一个服务，然后看文件名，然后再注册一个服务
	var wg sync.WaitGroup
	wg.Add(1)
	client, err := getNamingClient()
	if err != nil {
		return
	}
	go func() {
		http.ListenAndServe(":4567", nil)
		wg.Done()
	}()
	// 先注册自己
	_, err1 := client.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          ip,
		Port:        uint64(4567),
		GroupName:   "MALL_GROUP",
		Weight:      1,
		Enable:      true,
		Ephemeral:   true,
		Healthy:     true,
		ServiceName: "prometheuse_service",
		ClusterName: "ps",
	})
	if err1 != nil {
		log.Fatal("register fail")
	}
	// 获取服务
	client.GetService(vo.GetServiceParam{
		ServiceName: "prometheuse_service",
		GroupName:   "MALL_GROUP",
	})
	wg.Wait()
}

var ip = createRandomIp()

func registerSelf(port int, wg sync.WaitGroup) {
	log.Info("start register self")
	defer wg.Done()
	// 启动服务
	namingClient, client_err := getNamingClient()
	if client_err != nil {
		log.Fatal("create naming client err ,", client_err)
	}
	// 先注册自己
	_, err := namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          ip,
		Port:        uint64(port),
		GroupName:   "DEFAULT_GROUP",
		Weight:      1,
		Enable:      true,
		Ephemeral:   true,
		Healthy:     true,
		ServiceName: "prometheuse_service",
		ClusterName: "ps",
	})
	if err != nil {
		log.Fatal("register fail")
	}
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}

var serverConfigs = []constant.ServerConfig{
	//{
	//	IpAddr:      "127.0.0.1",
	//	ContextPath: "/nacos",
	//	Port:        8849,
	//	Scheme:      "http",
	//},
	{
		IpAddr:      "127.0.0.1",
		ContextPath: "/nacos",
		Port:        8848,
		Scheme:      "http",
	},
	//{
	//	IpAddr:      "127.0.0.1",
	//	ContextPath: "/nacos",
	//	Port:        8847,
	//	Scheme:      "http",
	//},
}
var nowCount = 0

func getNamingClient() (naming_client.INamingClient, error) {
	var clientConfig = *constant.NewClientConfig(
		//constant.WithNamespaceId("501689b2-129f-450c-8735-b04a5978b016"), //当namespace是public时，此处填空字符串。
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("/tmp/nacos/log"),
		//constant.WithCacheDir("/tmp/nacos/cache"),
		constant.WithLogLevel("debug"),
		constant.WithUsername("nacos"),
		constant.WithPassword("nacos"),
		constant.WithNotLoadCacheAtStart(false),
	)
	sc := []constant.ServerConfig{}
	sc = append(sc, serverConfigs...)
	nowCount++
	return clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
}

// create random ip addr
func createRandomIp() string {
	rand.Seed(time.Now().Unix())
	ip := fmt.Sprintf("%d.%d.%d.%d", rand.Intn(255), rand.Intn(255), rand.Intn(255), rand.Intn(255))
	return ip
}
