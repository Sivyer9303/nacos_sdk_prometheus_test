package main

import (
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/prometheus/common/log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

func main() {
	// 启动同名服务
	var wg sync.WaitGroup
	tick := time.Tick(time.Second * 15)
	portStart := 7788
	// 启动一个线程，一直获取service list
	go func() {
		tick := time.Tick(time.Second * 5)
		client, err := getNamingClient()
		if err != nil {
			log.Fatal("err ... ", err)
		}
		for {
			select {
			case <-tick:
				_, err := client.GetService(vo.GetServiceParam{
					ServiceName: "prometheuse_service",
					GroupName:   "DEFAULT_GROUP",
				})
				if err != nil {
					log.Fatal("get service fail ...", err)
				}
				//bytes, err := json.Marshal(service)
				//if err != nil {
				//	log.Fatal("Marshal service fail ...",err)
				//}
				//fmt.Println(string(bytes))
			}
		}
	}()
	for i := 0; i < 5; i++ {
		wg.Add(1)
		select {
		case <-tick:
			// 每三十秒注册一个
			go registerSelf(portStart+i, wg)
		}
	}
	wg.Wait()
}

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
		Ip:          "localhost",
		Port:        uint64(port),
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

func getNamingClient() (naming_client.INamingClient, error) {
	var clientConfig = *constant.NewClientConfig(
		//constant.WithNamespaceId("501689b2-129f-450c-8735-b04a5978b016"), //当namespace是public时，此处填空字符串。
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("/tmp/nacos/log"),
		constant.WithCacheDir("/tmp/nacos/cache"),
		constant.WithLogLevel("debug"),
		constant.WithUsername("nacos"),
		constant.WithPassword("nacos"),
	)
	var serverConfigs = []constant.ServerConfig{
		{
			IpAddr:      "127.0.0.1",
			ContextPath: "/nacos",
			Port:        8848,
			Scheme:      "http",
		},
	}
	return clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
}
