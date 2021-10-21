package go_conf

// 导入其它的包
import (
	"encoding/json"
	"github.com/go-redis/redis/v7"
	"log"
	"os"
	"time"
)

type node map[string]string

const ChannelAdd = "channel_add"
const ChannelDel = "channel_del"

const activityConfKid = "h:activity_conf_kid:%s"
const activityConfNum = "h:activity_conf_num"

const readLockKey = "s:read_lock_key";
const writeLockKey = "s:write_lock_key";

const lockTime = 3;

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

func getRedis(addr string, auth string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr, // 使用默认数据库
		Password: auth,                // 没有密码则置空
		DB:       0,                    // 使用默认的数据库
	})
}

func getHostName() string {
	name, err := os.Hostname()
	if err != nil {
		return ""
	}
	return name
}

func now() int64 {
	return time.Now().Unix()
}

func (this *Root) Start(addr string, auth string) {
	this.addr = addr
	this.auth = auth
	this.sub_redis = getRedis(addr, auth)
	this.req_redis = getRedis(addr, auth)
	this.data = make(map[string]node, 100)

	this.initData()
	this.Subscribe(ChannelAdd, ChannelDel)
	select {}
}

func (this *Root) initData() {

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

func (this *Root) getScript() string {
	return `
		local ret = redis.call('zrangebyscore', KEYS[1], KEYS[2], KEYS[3])
		if #ret > 0 then
			return 0
		end

		redis.call('zadd', KEYS[4], KEYS[5], KEYS[6])
		return 1;
	`
}

func (this *Root) lock() {
	start_time := now()
	script := this.getScript()
	keys := []string{
		writeLockKey,
		string(start_time - lockTime),
		string(start_time),
		readLockKey,
		string(start_time),
		getHostName(),
	}
	this.lockWithRetry(script, keys)
}

func (this *Root) lockWithRetry(script string, keys []string) {
	rds := this.req_redis
	start_time := now()
	sha := ""

	for true {
		now := now()
		keys[1] = string(now - lockTime)
		keys[2] = string(now)
		keys[4] = string(now)
		if sha == "" {
			cmd := rds.Eval(script, keys, len(keys))

		} else {

		}


	}



	rds.Eval(script, keys)


}