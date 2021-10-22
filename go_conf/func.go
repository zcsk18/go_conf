package go_conf

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v7"
	"log"
	"strconv"
	"time"
)

func (this *Root) Start(addr string, auth string) {
	fmt.Println("start")

	this.addr = addr
	this.auth = auth
	this.sub_redis = getRedis(addr, auth)
	this.req_redis = getRedis(addr, auth)
	go checkRds(this.sub_redis)
	go checkRds(this.req_redis)

	this.data = make(map[string]node, 100)

	this.initData()
	go netSrv(this)
	this.Subscribe(ChannelAdd, ChannelDel)
}

func (this *Root) AddData(kid string, name string, data string) {
	_, ok := this.data[kid]
	if !ok {
		this.data[kid] = make(node, 10)
	}

	this.data[kid][name] = data
}

func (this *Root) DelData(kid string, name string) {
	_, ok := this.data[kid]
	if !ok {
		return
	}

	delete(this.data[kid], name)
}

func (this *Root) initData() {
	this.readLock()
	rds := this.req_redis
	cmdAll := rds.HGetAll(activityConfNum)
	/*
	pipeline := rds.Pipeline()
	result := make([]*redis.StringStringMapCmd, 0)
	for kid, _ := range(cmdAll.Val()) {
		result = append(result, pipeline.HGetAll(fmt.Sprintf(activityConfKid, kid)))
	}
	_, err := pipeline.Exec()
	if err != nil {
		panic(err)
	}
	*/

	//i := 0
	for kid, _ := range(cmdAll.Val()) {
		cmdKid := rds.HGetAll(fmt.Sprintf(activityConfKid, kid))
		//cmdKid := result[i]
		//i += 1
		if len(cmdKid.Val()) == 0 {
			fmt.Println("del kid:" + kid)
			rds.HDel(activityConfNum, kid)
		}

		for name, str := range (cmdKid.Val()) {
			this.AddData(kid, name, str)
		}
	}

	fmt.Println(this.data)
}

func (this *Root)onMsg(op string, msg Msg) {
	_, ok := this.data[msg.Kid]
	if !ok {
		this.data[msg.Kid] = make(node, 10)
	}

	switch op {
	case ChannelAdd:
		this.AddData(msg.Kid, msg.Name, msg.Data)
	case ChannelDel:
		this.DelData(msg.Kid, msg.Name)
	}
	this.req_redis.Incr(msg.Seq)
	fmt.Println(this.data)
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
	this.unLockRead()

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

func (this *Root) readLock() {
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

func (this *Root) unLockRead() {
	rds := this.req_redis
	rds.ZRem(readLockKey, getHostName())
}

func (this *Root) lockWithRetry(str string, keys []string) bool {
	rds := this.req_redis
	start_time := now()
	hasSha := false

	for true {
		now := now()
		keys[1] = strconv.FormatInt(now - lockTime, 10)
		keys[2] = strconv.FormatInt(now, 10)
		keys[4] = strconv.FormatInt(now, 10)
		var cmd *redis.Cmd

		script := redis.NewScript(str)

		if hasSha == false {
			cmd = script.Eval(rds, keys, len(keys))
			hasSha = true
		} else {
			cmd = script.EvalSha(rds, keys, len(keys))
		}
		if cmd.Err() != nil {
			fmt.Println(keys)
			panic(cmd.Err())
		}

		ret, _ := cmd.Bool()
		if ret {
			break
		}

		fmt.Println("lock")
		fmt.Println(cmd)
		if now - start_time > lockTime {
			panic("lock failed")
			return false
		}

		time.Sleep(time.Duration(100) * time.Millisecond)
	}

	return true
}