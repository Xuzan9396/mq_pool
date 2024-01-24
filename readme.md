## rabbitmq 连接池channel复用

开发语言 golang
依赖库


> 已在线上生产环镜运行， 5200W请求 qbs 3000 时， 连接池显示无压力<br>
> rabbitmq部署为线上集群

### 接下来的功能，`预计在1.0.15版本`
1增加批次消息处理，用以提高生产及消费的吞吐量

### 功能说明
1. 自定义连接池大小及最大处理channel数
2. 消费者底层断线自动重连
3. 生产者底层断线自动重连 `v1.0.12`
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
1. 初始化
```
var oncePool sync.Once
var instanceRPool *kelleyRabbimqPool.RabbitPool
func initrabbitmq() *kelleyRabbimqPool.RabbitPool {
	oncePool.Do(func() {
        //初始化生产者
		instanceRPool = kelleyRabbimqPool.NewProductPool()
        //初始化消费者
	    instanceConsumePool = kelleyRabbimqPool.NewConsumePool()
        //使用默认虚拟host "/"
		err := instanceRPool.Connect("192.168.1.202", 5672, "guest", "guest")
        //使用自定义虚
        //err:=instanceConsumePool.ConnectVirtualHost("192.168.1.202", 5672, "guest", "guest", "/testHost")
		if err != nil {
			fmt.Println(err)
		}
	})
	return instanceRPool
}
```


2.  生产者
```
var wg sync.WaitGroup
	for i:=0;i<100000; i++ {
		wg.Add(1)
		go func(num int) {
			defer wg.Done()
			data:=kelleyRabbimqPool.GetRabbitMqDataFormat("testChange5", kelleyRabbimqPool.EXCHANGE_TYPE_TOPIC, "textQueue5", "/", fmt.Sprintf("这里是数据%d", num))
			_=instanceRPool.Push(data)
		}(i)
	}
	wg.Wait()
```

3. 消费者
> 可定义多个消息者事件, 不通交换机, 队列, 路由
>
> 每个事件独立
>

```
nomrl := &rabbitmq.ConsumeReceive{
        #定义消费者事件
        ExchangeName: "testChange31",//队列名称
        ExchangeType: kelleyRabbimqPool.EXCHANGE_TYPE_DIRECT,
        Route:        "",
        QueueName:    "testQueue31",
        IsTry:true,//是否重试
        IsAutoAck: false, //是否自动确认消息
        MaxReTry: 5,//最大重试次数
        EventFail: func(code int, e error, data []byte) {
        	fmt.Printf("error:%s", e)
        },
        /***
         * 参数说明
         * @param data []byte 接收的rabbitmq数据
         * @param header map[string]interface{} 原rabbitmq header
         * @param retryClient RabbitmqPool.RetryClientInterface 自定义重试数据接口，重试需return true 防止数据重复提交
         ***/
        EventSuccess: func(data []byte, header map[string]interface{},retryClient kelleyRabbimqPool.RetryClientInterface)bool {//如果返回true 则无需重试
            _ = retryClient.Ack()//确认消息    	
            fmt.Printf("data:%s\n", string(data))
        	return true
        },
	}
	instanceConsumePool.RegisterConsumeReceive(nomrl)

	err := instanceConsumePool.RunConsume()
	if err != nil {
		fmt.Println(err)
	}
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
