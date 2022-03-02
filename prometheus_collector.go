package main

import (
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"net/http"
	"strconv"
	"time"
)

func main2() {

	http.Handle("/metrics", promhttp.Handler())
	// config 测试 每十秒获取一次config
	go func() {
		configClient, err := GetConfigClient()
		if err != nil {
			log.Fatal("create config client err")
			return
		}
		// 每十秒往nacos中的config中写一个配置文件
		tick := time.Tick(10 * time.Second)
		start := 0
		// 监听文件
		configClient.ListenConfig(vo.ConfigParam{
			DataId:  "test-data0",
			Group:   "DEFAULT_GROUP",
			Content: "hello world!",
		})
		for {
			select {
			case <-tick:
				log.Info("start publish config to nacos")
				configClient.PublishConfig(vo.ConfigParam{
					DataId:  "test-data" + strconv.Itoa(start),
					Group:   "DEFAULT_GROUP",
					Content: "hello world!",
				})
				configClient.GetConfig(vo.ConfigParam{
					DataId:  "test-data" + strconv.Itoa(start),
					Group:   "DEFAULT_GROUP",
					Content: "hello world!",
				})
				start++
			}
		}
	}()
	// 每秒获取一次naming
	namingClient, _ := getNamingClient()
	// 先注册自己
	_, err := namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          "localhost",
		Port:        8883,
		Weight:      1,
		Enable:      true,
		Ephemeral:   true,
		Healthy:     true,
		ServiceName: "prometheuse_service",
	})
	if err != nil {
		log.Fatal("register fail ....", err)
		return
	} else {
		fmt.Println("register success")
	}
	go func() {
		tick := time.Tick(3 * time.Second)
		for {
			select {
			case <-tick:
				// 每秒刷新一次
				service, err := namingClient.GetService(vo.GetServiceParam{
					//GroupName: "DEFAULT_GROUP",
					ServiceName: "prometheuse_service",
					GroupName:   "DEFAULT_GROUP",
				})
				if err != nil {
					log.Fatal("oops. there is some err .... ", err)
				} else {
					hosts := service.Hosts
					log.Info(fmt.Sprintf("service info,hosts:=%v", hosts))
				}
			}
		}
	}()
	http.ListenAndServe(":8883", nil)
}

func GetConfigClient() (config_client.IConfigClient, error) {
	//create ServerConfig
	sc := []constant.ServerConfig{
		*constant.NewServerConfig("127.0.0.1", 8848, constant.WithContextPath("/nacos")),
	}

	//create ClientConfig
	cc := *constant.NewClientConfig(
		constant.WithNamespaceId(""),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("/tmp/nacos/log"),
		constant.WithCacheDir("/tmp/nacos/cache"),
		constant.WithLogLevel("debug"),
	)

	// create config client
	return clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
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
