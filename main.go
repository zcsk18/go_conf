package main

import "github.com/zcsk18/go_conf/go_conf"

func main() {
	var root  = go_conf.Root{}
	root.Start("127.0.0.1:6379", "")
}

