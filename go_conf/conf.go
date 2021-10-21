package go_conf

// 导入其它的包
import (
	"encoding/json"
	"github.com/go-redis/redis/v7"
	"log"
)

type node map[string]string

const ChannelAdd = "channel_add"
const ChannelDel = "channel_del"

type Root struct {
	addr string
	auth string
	data map[string]node
	sub_redis *redis.Client
	req_redis *redis.Client
}

type Msg struct {
	Name string `json:"_name"`
	Kid string `json:"_kid"`
	Seq string `json:"_seq"`
	Data string `json:"_data"`
}

func (this *Root)Start(addr string, auth string) {
	this.addr = addr
	this.auth = auth
	this.sub_redis = getRedis(addr, auth)
	this.req_redis = getRedis(addr, auth)
	this.data = make(map[string]node, 100)

	this.Subscribe(ChannelAdd, ChannelDel)
	select {}
}

func getRedis(addr string, auth string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr, // 使用默认数据库
		Password: auth,                // 没有密码则置空
		DB:       0,                    // 使用默认的数据库
	})
}

func (this *Root)onMsg(op string, msg Msg) {
	_, ok := this.data[msg.Kid]
	if !ok {
		this.data[msg.Kid] = make(node, 10)
	}

	switch op {
	case ChannelAdd:
		this.data[msg.Kid][msg.Name] = msg.Data
	case ChannelDel:
		delete(this.data[msg.Kid], msg.Name)
	}
	log.Println(this.data)
	this.req_redis.Incr(msg.Seq)
}

func (this *Root)Subscribe(channels ...string) {
	rdb := this.sub_redis
	pong, err := rdb.Ping().Result() // 检查是否连接
	if err != nil {
		log.Fatal(err)
	}

	// 连接成功啦
	log.Println(pong)

	// 订阅全部消息
	pubsub := rdb.Subscribe(channels...)
	// 等待消息返回，原因是上一个方法不是立即返回的，囧
	_, err = pubsub.Receive()
	if err != nil {
		log.Fatal(err)
	}

	// 用管道来接收消息
	ch := pubsub.Channel()

	// 处理消息
	for info := range ch {
		log.Println(info.Channel, ":", info.Payload)
		msg := Msg{}
		json.Unmarshal([]byte(info.Payload), &msg)
		log.Println(msg)
		this.onMsg(info.Channel, msg)
	}
}