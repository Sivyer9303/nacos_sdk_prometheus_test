package main

import (
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/prometheus/common/log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

func main() {
	// 启动同名服务两个
	var wg sync.WaitGroup
	tick := time.Tick(time.Second * 30)
	portStart := 7788
	for i := 0; i < 20; i++ {
		wg.Add(1)
		select {
		case <-tick:
			go registerSelf(portStart+i, wg)
		}
	}
	wg.Wait()
}

func registerSelf(port int, wg sync.WaitGroup) {
	defer wg.Done()
	// 启动服务
	namingClient, _ := getNamingClient()
	// 先注册自己
	_, err := namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          "localhost",
		Port:        uint64(port),
		Weight:      1,
		Enable:      true,
		Ephemeral:   true,
		Healthy:     true,
		ServiceName: "prometheuse_service",
	})
	if err != nil {
		log.Fatal("register fail")
	}
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}
