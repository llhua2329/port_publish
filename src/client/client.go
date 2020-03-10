package main

import (
	"flag"
	"log"
	"net"
	"os"
	"port_publish/src/common"
	"strconv"
	"time"
)

func main() {
	remoteIp := flag.String("remoteIp", "192.168.186.52", "remote ip")
	remotePort := flag.Int("remotePort", 30001, "remote port")
	localPort := flag.Int("localPort", 8080, "local port")
	flag.Parse()

	serverAddr := *remoteIp + ":" + strconv.Itoa(*remotePort)
	localAddr := "localhost" + ":" + strconv.Itoa(*localPort)
	controlConn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.Println("connect server failed")
		log.Fatal(err)
		return
	}
	defer controlConn.Close()

	ctrl := common.NewConnection()
	ctrl.Conn = controlConn
	go ctrl.Read()
	go ctrl.Write()

	recv := make([]byte, 1)
	tick := time.Tick(5 * time.Second)
	for {
		select {
		case <-tick: // 定时发送心跳
			heartbeat := []byte{0xFE}
			ctrl.Send <- heartbeat
		case recv = <-ctrl.Recv:
			if recv[0] == 0xFF { // 发起新的连接
				log.Printf("%s accept now\n", time.Now())
				go NewChannel(serverAddr, localAddr)
			}
		case <-ctrl.ReadClose:
			log.Printf("ctrl connect close\n")
			os.Exit(0)
		}
	}
}

func NewChannel(serverInfo string, localAddr string) {
	log.Printf("%s send server to create new connect\n", time.Now())
	remote, err := net.Dial("tcp", serverInfo)
	if err != nil {
		log.Println("connect to server failed")
		log.Fatal(err)
		return
	}
	defer remote.Close()
	log.Printf("%s send local connect\n", time.Now())
	local, err := net.Dial("tcp", localAddr)
	if err != nil {
		log.Println("connect to client failed:")
		log.Fatal(err)
		return
	}
	log.Printf("%s swap data\n", time.Now())
	defer local.Close()
	common.SwapConn(local, remote)
}