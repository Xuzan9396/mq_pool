package mq_pool

import (
	logs "log"
)

var pubMqPool *RabbitPool

type OptionProduct func(*RabbitPool)

func WithProductMaxConnection(maxConnection int32) OptionProduct {
	return func(o *RabbitPool) {
		if maxConnection <= 0 {
			return
		}
		o.SetMaxConnection(maxConnection)
	}
}

// 初始化发布池, maxConnect 最大连接数 ,如果为0 则不设置
func InitRabbitmqProduct(rabAddr string, rabPort int, rabUser, rabPwd, vhost string, options ...OptionProduct) error {

	pubMqPool = NewProductPool()
	//设置最大连接是2个
	for _, option := range options {
		option(pubMqPool)
	}

	err := pubMqPool.ConnectVirtualHost(rabAddr, rabPort, rabUser, rabPwd, vhost)
	if err != nil {
		logs.Println("InitRabbitmq Err:", err)
		return err
	}
	return nil
}

type PublishInfo struct {
	ExChangeName string
	QueueName    string
	RouteKey     string
	Body         string
	ExchangeType string // "fanout" //  Fanout：广播，将消息交给所有绑定到交换机的队列 "direct" //Direct：定向，把消息交给符合指定routing key 的队列  "topic"  //Topic：通配符，把消息交给符合routing pattern（路由模式） 的队列
}

// 发布消息
func PublishPool(mq *PublishInfo) *RabbitMqError {

	if pubMqPool == nil {
		return &RabbitMqError{Message: "连接池为空"}
	}

	data := GetRabbitMqDataFormat(
		mq.ExChangeName, //交换机名
		mq.ExchangeType, //模式
		mq.QueueName,    //队列名
		mq.RouteKey,     //key
		mq.Body,         //内容
	)
	return pubMqPool.Push(data)
}
