package main

import (
	"errors"
	"fmt"
	"github.com/Xuzan9396/mq_pool"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
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
			fmt.Printf("11111111:name:%s,data:%s\n", name, string(data))

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
			fmt.Printf("22222222:name:%s,data:%s\n", name, string(data))
			return errors.New("错误了")
		},
		EventFail: func(code int, e error, data []byte) {
			if code == mq_pool.RCODE_RETRY_MAX_ERROR {
				fmt.Printf("入库:%s\n", string(data))
			}
		},
	}

	//mq_pool.RunConsume(model, model2)
	_ = model2
	mq_pool.RunConsume(model)
	setupSignalHandling()
	select {}
}
func setupSignalHandling() {
	// 创建一个信号通道
	sigChan := make(chan os.Signal, 1)
	// 监听SIGINT和SIGTERM信号
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 使用goroutine等待信号
	go func() {
		// 阻塞，直到收到信号
		<-sigChan

		// 收到信号后执行清理操作
		mq_pool.Shutdown()
		time.Sleep(time.Second * 3)
		os.Exit(0)

	}()
}
