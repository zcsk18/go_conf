package go_conf

import "github.com/go-redis/redis/v7"

type node map[string]string

const ChannelAdd = "channel_add"
const ChannelDel = "channel_del"

const activityConfKid = "h:activity_conf_kid:%s"
const activityConfNum = "h:activity_conf_num"

const readLockKey = "s:read_lock_key";
const writeLockKey = "s:write_lock_key";

const lockTime = 30;

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

type Request struct {
	Method string `json:"method"`
	Params map[string]string `json:"params"`
}

type Responce struct {
	Ret string `json:"ret"`
	Data string `json:"data"`
}
