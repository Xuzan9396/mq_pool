package main

import (
	"fmt"
	"github.com/Xuzan9396/mq_pool"
	"github.com/spf13/viper"
	"log"
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
	err := mq_pool.InitRabbitmqProduct(rabAddr, rabPort, rabUser, rabPwd, vhost, mq_pool.WithProductMaxConnection(maxConnet))
	if err != nil {
		log.Fatal(err)
		return
	}

	i := 0
	for {
		mq_pool.PublishPool(&mq_pool.PublishInfo{
			ExChangeName: "test.exchange",
			QueueName:    "test.queue",
			RouteKey:     "test.routeKey",
			ExchangeType: mq_pool.EXCHANGE_TYPE_FANOUT,
			Body:         fmt.Sprintf(`{"msg_id":2,"user_id":1000558,"content":"xxxxx测试,%d"}`, i),
		})
		log.Println("写入消息:", i)
		i++
		time.Sleep(time.Millisecond * 5000)
	}

}
