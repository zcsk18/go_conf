package go_conf

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
)

var _root *Root

func netSrv(root *Root) {
	_root = root
	var unixAddr *net.UnixAddr

	os.Remove("/tmp/go_conf.sock")
	unixAddr, _ = net.ResolveUnixAddr("unix", "/tmp/go_conf.sock")

	unixListener, _ := net.ListenUnix("unix", unixAddr)

	defer func () {
		unixListener.Close()
		os.Remove("/tmp/go_conf.sock")
	}()

	for {
		unixConn, err := unixListener.AcceptUnix()
		if err != nil {
			continue
		}

		fmt.Println("A client connected : " + unixConn.RemoteAddr().String())
		go unixPipe(unixConn)
	}
}

func sendMsg(conn *net.UnixConn, msg []byte) {
	l := Int2Byte(len(msg))
	conn.Write(l)
	conn.Write(msg)
}

func unixPipe(conn *net.UnixConn) {
	ipStr := conn.RemoteAddr().String()
	defer func() {
		fmt.Println("disconnected :" + ipStr)
		conn.Close()
	}()
	reader := bufio.NewReader(conn)
	request := Request{}
	responce := Responce{}

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		message = string(message[:len(message)-1])
		fmt.Println(message)
		json.Unmarshal([]byte(message), &request)

		switch request.Method {
		case "getConf":
			responce = getConf(request.Params["kid"], request.Params["name"])
		}

		buf,err := json.Marshal(responce)
		if err != nil {
			return
		}

		sendMsg(conn, buf)
	}
}

func getConf(kid string, name string) (res Responce) {
	res.Ret = "ok"
	if name == "" {
		ret,err := json.Marshal(_root.data[kid])
		if err != nil {
			res.Ret = "fail"
			res.Data = err.Error()
			return
		}
		res.Data = string(ret)
	} else {
		res.Data = _root.data[kid][name]
	}
	return
}