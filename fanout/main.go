package main

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func producer() {
	viper.Set("mq.username", "xuzan")
	viper.Set("mq.password", "rd272623iuyt")
	viper.Set("mq.addr_url", "127.0.0.1:5672")
	viper.Set("mq.vhost", "/")

	conn, err := amqp.Dial("amqp://xuzan:rd272623iuyt@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	exchangeName := "logs" // 使用一个 fanout 类型的交换机
	err = ch.ExchangeDeclare(
		exchangeName, // 交换机名称
		"fanout",     // 交换机类型
		true,         // 持久性
		false,        // 自动删除
		false,        // 内部交换机
		false,        // 不等待
		nil,          // 参数
	)
	failOnError(err, "Failed to declare an exchange")

	for i := 0; i < 5; i++ {
		body := fmt.Sprintf("Log Message %d", i)
		err = ch.Publish(
			exchangeName, // 交换机名称
			"",           // 路由键为空，因为 fanout 交换机会广播给所有队列
			false,        // 强制
			false,        // 立即
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			})
		failOnError(err, "Failed to publish a message")
		fmt.Printf(" [x] Sent %s\n", body)
		time.Sleep(time.Second)
	}
}

func consumer(id int) {
	conn, err := amqp.Dial("amqp://xuzan:rd272623iuyt@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	exchangeName := "logs" // 使用相同的 fanout 交换机名称
	if err := ch.ExchangeDeclare(
		exchangeName, // name of the exchange
		"fanout",     // type
		true,         // durable
		false,        // delete when complete
		false,        // internal
		false,        // noWait
		nil,          // arguments
	); err != nil {
		log.Fatalf("Failed to declare an exchange: %s", err)
		return
	}
	_, err = ch.QueueDeclare(
		strconv.Itoa(id), // 随机生成队列名称
		true,             // 持久性
		false,            // 自动删除
		false,            // 独占性
		false,            // 不等待
		nil,              // 参数
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		strconv.Itoa(id), // 随机生成的队列名称
		"",               // 路由键为空，因为 fanout 交换机会广播给所有队列
		exchangeName,     // 交换机名称
		false,            // 不等待
		nil,              // 参数
	)
	failOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		//"",    // 随机生成的队列名称
		//"",    // 消费者名称
		//true,  // 自动确认
		//false, // 独占性
		//false, // 不等待
		//false, // 参数

		strconv.Itoa(id), // queue
		strconv.Itoa(id), // 消费者标签
		false,            // auto-ack
		false,            // exclusive 独占模式
		false,            // no-local
		false,            // no-wait
		nil,              // args
	)
	failOnError(err, "Failed to register a consumer")

	for msg := range msgs {
		fmt.Printf("Worker %d received a message: %s\n", id, msg.Body)
		if id == 1 {
			msg.Ack(true)
		}

	}
}

func main() {
	mode := "producer" // 设置运行模式，可以是 "producer" 或 "consumer"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}

	if mode == "product" {
		producer()
	} else if mode == "consume" {
		for i := 1; i <= 2; i++ {
			go consumer(i)
		}

		select {}
	} else {
		fmt.Println("Usage: go run main.go [producer|consumer]")
	}
}
