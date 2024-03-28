## rabbitmq 支持连接池connect,channel复用，自动重连，多连接池, 自动重试机制,direct,fanout 失败重试机制
例如开启了: 重试，最大重试次数为2次，2次都失败后支持失败回调
```yaml
IsTry:        true,
IsAutoAck:    false,
MaxReTry:     2,
```
fanout 模式下，如果多个订阅消费者,如果只有一个消费者失败了，会单独重试发这个消费者,MaxReTry 重试次数，不会影响其他消费者




开发语言 golang
依赖库



### 功能说明
1. 自定义连接池大小及最大处理channel数
2. 消费者底层断线自动重连
3. 生产者底层断线自动重连
4. 底层使用轮循方式复用tcp
5. 生产者每个tcp对应一个channel,防止channel写入阻塞造成内存使用过量
6. 支持rabbitmq exchangeType
7. 默认交换机、队列、消息都会持久化磁盘
8. 默认值

| 名称 | 说明 |
| --- | --- |
| tcp最大连接数 | 5 |
| 生产者消费发送失败最大重试次数 | 5 |
| 消费者最大channel信道数(每个连接自动平分) | 100(每个tcp10个) |



### 使用
1. 消费者
```
package mq_model

import (
	"github.com/Xuzan9396/mq_pool"
	"github.com/spf13/viper"
	"log"
	"strconv"
	"strings"
)

func InitMq() {
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
		log.Fatal("mq_pool.InitRabbitmqConsume err:", err)
		return
	}

    // 多个消费者事件
	mq_pool.RunConsume(NewMqSendModel())
	//mq_pool.RunConsume(NewMqSendModel(), NewMqSendModelGuild())

}

func ShutDown() {
	mq_pool.Shutdown()
}

func NewMqSendModel() *mq_pool.MqModel {
	res := &mq_pool.MqModel{
		Exchange:     "send_gift_fanout.exchange",
		ExchangeType: mq_pool.EXCHANGE_TYPE_FANOUT,
		Queue:        "send_gift_fanout_rank.queue",
		Routeingkey:  "send_gift_fanout_rank.routeKey",
		Ctag:         "send_gift_fanout_rank.tag",
		IsTry:        true,
		IsAutoAck:    false,
		MaxReTry:     2,
		Callback:     handler_mq.NewModel().SendGiftCallback,
	}
	return res
}

func NewMqSendModelGuild() *mq_pool.MqModel {
	res := &mq_pool.MqModel{
		Exchange:     "send_gift_fanout.exchange",
		ExchangeType: mq_pool.EXCHANGE_TYPE_FANOUT,
		Queue:        "send_gift_fanout_guild.queue",
		Routeingkey:  "send_gift_fanout_guild.routeKey",
		Ctag:         "send_gift_fanout_guild.tag",
		IsTry:        true,
		IsAutoAck:    false,
		MaxReTry:     2,
		Callback:     handler_mq.NewModel().SendGiftGuild,
		EventFail: func(code int, e error, data []byte) {
			if code == mq_pool.RCODE_RETRY_MAX_ERROR {
				//handler_mq.NewModel().SendGiftGuildFail(data)
			}
		},
	}

	return res
}

```


2.  生产者
```
	maxConnet := int32(2)
	err := mq_pool.InitRabbitmqProduct(rabAddr, rabPort, rabUser, rabPwd, vhost, mq_pool.WithProductMaxConnection(maxConnet))
	if err != nil {
        log.Fatal("mq_pool.InitRabbitmqProduct err:", err)
        return
    }
   	err := mq_pool.PublishPool(&mq_pool.PublishInfo{
		ExChangeName: "send_gift_fanout.exchange",
		QueueName:    "",
		RouteKey:     "",
		ExchangeType: mq_pool.EXCHANGE_TYPE_FANOUT,
		Body:         string(resByte),
	}) 
```


> * 参数说明

| 名称 | 类型 | 说明 |
| --- | --- | --- |
| ExchangeName|  string | 交换机名称 |
| ExchangeType | string | 交换机类型: <br>EXCHANGE_TYPE_FANOUT<br>EXCHANGE_TYPE_DIRECT<br>EXCHANGE_TYPE_TOPIC |
| Route| string | 路由键 |
| QueueName | string | 队列名称 |
| IsTry | bool | 是否重试<br>如果开启重试后， 在成功回调用返回true会对消息进行重试, 重试时间为 5000~15000 MS|
| IsAutoAck | bool | 是否自动确认消息, true: 组件底层会自动对消息进行确认<br> false: 手动进行消息确认，在成功会调中需进行手动确认` _ = retryClient.Ack()` |
| MaxReTry | int | 重试最大次数s, 需isTry为true |
| EventFail | func | 失败回调 |
| EventSuccess | func | 成功回调 |


4. 错误码说明
> 错误码为
>
> 1. 生产者push时返回的  *RabbitMqError
> 2. 消费者事件监听回返的 code
>
| 错误码 | 说明 |
| --- | --- |
|501|生产者发送超过最大重试次数|
|502|获取信道失败, 一般为认道队列数用尽|
|503|交换机/队列/绑定失败|
|504|连接失败|
|506|信道创建失败|
|507|超过最大重试次数|
