package main

import (
	"fmt"
	"github.com/Xuzan9396/mq_pool"
	"github.com/spf13/viper"
	"strconv"
	"strings"
	"time"
)

func main() {
	viper.Set("mq.username", "xuzan")
	viper.Set("mq.password", "rd272623iuyt")
	viper.Set("mq.addr_url", "127.0.0.1:5672")
	viper.Set("mq.vhost", "/")
	rabAddr := strings.Split(viper.GetString("mq.addr_url"), ":")[0]
	rabPort, _ := strconv.Atoi(strings.Split(viper.GetString("mq.addr_url"), ":")[1])
	rabUser := viper.GetString("mq.username")
	rabPwd := viper.GetString("mq.password")
	vhost := viper.GetString("mq.vhost")
	maxConnet := int32(2)
	maxChannel := int32(2)
	err := mq_pool.InitRabbitmqConsume(rabAddr, rabPort, rabUser, rabPwd, vhost,
		mq_pool.WithConsumeMaxConnection(maxConnet),
		mq_pool.WithConsumeMaxConsumeChannel(maxChannel),
	)
	if err != nil {
		panic(err)
		return
	}
	model := &mq_pool.MqModel{
		Exchange:     "test.exchange",
		ExchangeType: mq_pool.EXCHANGE_TYPE_FANOUT,
		Queue:        "test.queue",
		Routeingkey:  "test.routeKey",
		Ctag:         "test.tag",
		IsTry:        true,
		IsAutoAck:    false,
		MaxReTry:     2,
		Callback: func(name string, data []byte) error {
			fmt.Printf("name:%s,data:%s", name, string(data))
			time.Sleep(time.Second)
			return nil
		},
	}

	model2 := &mq_pool.MqModel{
		Exchange:     "test.exchange",
		ExchangeType: mq_pool.EXCHANGE_TYPE_FANOUT,
		Queue:        "test_v2.queue",
		Routeingkey:  "test_v2.routeKey",
		Ctag:         "test_v2.tag",
		IsTry:        true,
		IsAutoAck:    false,
		MaxReTry:     2,
		Callback: func(name string, data []byte) error {
			fmt.Printf("name:%s,data:%s", name, string(data))
			time.Sleep(time.Second)
			return nil
		},
	}

	go func() {
		time.Sleep(time.Second * 5)
		mq_pool.Shutdown()
	}()
	mq_pool.RunConsume(model, model2)
	select {}
}
