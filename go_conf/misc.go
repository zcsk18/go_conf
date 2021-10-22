package go_conf

import (
	"github.com/go-redis/redis/v7"
	"os"
	"time"
	"unsafe"
)

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

func checkRds(rdb *redis.Client) {
	for  {
		_, err := rdb.Ping().Result() // 检查是否连接
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Duration(1) * time.Second)
	}
}

func Int2Byte(data int)(ret []byte){
	len := unsafe.Sizeof(data)
	ret = make([]byte, len)
	var tmp int = 0xff
	var index uint = 0
	for index=0; index<uint(len); index++{
		ret[index] = byte((tmp<<(index*8) & data)>>(index*8))
	}
	return ret
}