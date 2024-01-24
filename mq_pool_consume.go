package mq_pool

import (
	"fmt"
)

type MqModel struct {
	Exchange     string
	ExchangeType string
	Ctag         string
	Queue        string
	Routeingkey  string
	IsTry        bool  //是否重试
	IsAutoAck    bool  //自动消息确认 true 自动确认消息
	MaxReTry     int32 //最大重试次数
	Callback     func(quequeName string, data []byte) error
	EventFail    func(code int, e error, data []byte)
}

var consumeMqPool *RabbitPool

func RunConsume(list ...*MqModel) {
	//初始各个消费者队列
	if len(list) == 0 {
		return
	}
	for _, mq := range list {
		mq.RangeConsumer()
	}

	go func() {
		err := consumeMqPool.RunConsume()
		if err != nil {
			//zlog.F().Error(err)

			panic(err)
			return
		}
		return
	}()
}

type OptionConsume func(*RabbitPool)

func WithConsumeMaxConnection(maxConnection int32) OptionConsume {
	return func(o *RabbitPool) {
		if maxConnection <= 0 {
			return
		}
		o.SetMaxConnection(maxConnection)
	}
}
func WithConsumeMaxConsumeChannel(maxConsumeChannel int32) OptionConsume {
	return func(o *RabbitPool) {
		if maxConsumeChannel <= 0 {
			return
		}
		o.SetMaxConsumeChannel(maxConsumeChannel)
	}
}

// 初始化发布池
func InitRabbitmqConsume(rabAddr string, rabPort int, rabUser, rabPwd, vhost string, options ...OptionConsume) error {
	consumeMqPool = NewConsumePool()
	//consumeMqPool.SetMaxConnection(1)
	//consumeMqPool.SetMaxConsumeChannel(2)
	for _, option := range options {
		option(consumeMqPool)
	}
	err := consumeMqPool.ConnectVirtualHost(rabAddr, rabPort, rabUser, rabPwd, vhost)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if err != nil {
		return err
	}
	return nil

}

// 注册接收者
func eventFail(code int, e error, data []byte) { //消费失败的调用
	fmt.Printf("code:%d,%v,%s\n", code, e, string(data))
}

func (mq *MqModel) RangeConsumer() {
	//初始各个消费者队列

	var eventFailInfo func(code int, e error, data []byte)
	if mq.EventFail == nil {
		eventFailInfo = eventFail
	} else {
		eventFailInfo = mq.EventFail
	}
	consumeReceive := &ConsumeReceive{
		ExchangeName: mq.Exchange,     //交换机名
		ExchangeType: mq.ExchangeType, //交换机类型
		Route:        mq.Routeingkey,  //key
		Ctag:         mq.Ctag,
		QueueName:    mq.Queue,     //队列名
		IsTry:        mq.IsTry,     //是否重试
		IsAutoAck:    mq.IsAutoAck, //自动消息确认
		MaxReTry:     mq.MaxReTry,  //最大重试次数
		EventFail:    eventFailInfo,
		//EventFail: func(code int, e error, data []byte) { //消费失败的调用
		//	fmt.Printf("code:%d,%v,%s\n", code, e, string(data))
		//},
	}
	consumeReceive.EventSuccess = mq.successHandler(consumeReceive) //消费成功的调用

	consumeMqPool.RegisterConsumeReceive(consumeReceive)

}

// 根据不同队列的不同分发处理
func (mq *MqModel) successHandler(q *ConsumeReceive) func(data []byte, header map[string]interface{}, retryClient RetryClientInterface) bool {
	return func(data []byte, header map[string]interface{}, retryClient RetryClientInterface) bool { //如果返回true 则无需重试
		if !mq.IsAutoAck {
			defer retryClient.Ack()
		}
		if mq.Callback != nil {
			err := mq.Callback(q.QueueName, data)
			if err != nil {
				return false
			}
		}
		return true
	}
}

func Shutdown() error {
	if consumeMqPool == nil {
		return nil
	}
	return consumeMqPool.Shutdown()
}

// 废弃
/*func (c *MqModel) initChannel() {
	rabbitHost := fmt.Sprintf("amqp://%s:%s@%s/", viper.GetString("mq.username"), viper.GetString("mq.password"), viper.GetString("mq.addr_url"))
	var channel *amqp.Channel

	conn, err := amqp.Dial(rabbitHost)
	if err != nil {
		zlog.F().Error(err)
		return
	}

	channel, err = conn.Channel()
	if err != nil {
		zlog.F().Error(err)
		return
	}

	if err := channel.ExchangeDeclare(
		c.Exchange, // name of the exchange
		//"direct",   // type
		EXCHANGE_TYPE_FANOUT, // type
		true,                 // durable 持久化 交换机
		false,                // delete when complete
		false,                // internal
		false,                // noWait
		nil,                  // arguments
	); err != nil {
		zlog.F("transfer_micro").Error("ExchangeDeclare错误", c.Queue, err)
		return
	}
	// 声明一个queue
	queue, err := channel.QueueDeclare(
		c.Queue, // name of the queue
		true,    // durable 持久化 队列
		false,   // delete when unused
		false,   // exclusive 为true时，只能被当前连接使用，连接断开后自动删除
		false,   // noWait
		nil,     // arguments
	)

	if err != nil {
		zlog.F("transfer_micro").Error("QueueDeclare错误", c.Queue, err)
		return
	}

	if err = channel.QueueBind(
		c.Queue, // name of the queue
		"",      //gxz fanout模式下不需要key
		//c.BindingKey, // bindingKey
		c.Exchange, // sourceExchange
		false,      // noWait
		nil,        // arguments
	); err != nil {
		zlog.F("transfer_micro").Error("bind错误", c.Queue, err)

		return
	}
	zlog.F("transfer_micro").Infof("declared Queue (%q %d messages, %d consumers), binding to Exchange (key %q)",
		queue.Name, queue.Messages, queue.Consumers, c.Queue)

}*/
