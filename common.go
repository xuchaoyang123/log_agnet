package main

import (
	"net"
)

//ColletcdEntry 要收集日志的结构体配置项
type ColletcdEntry struct {
	Path  string `json:"path"`
	Topic string `json:"topic"`
}

//获取本机ip的函数
func GetOutBoundIp() (ip string, err error) {
	conn, err := net.Dial("udp", "1.1.1.1:1")
	if err != nil {
		return
	}
	defer conn.Close()

	localaddr := conn.LocalAddr().(*net.UDPAddr)
	//fmt.Println(localaddr.String())  //ip:port
	ip = localaddr.IP.String() //ip
	return

}
