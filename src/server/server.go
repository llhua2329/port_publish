package main

import (
	"flag"
	"log"
	"port_publish/src/common"
	"time"
)

func main() {
	userPort := flag.Int("userPort", 30000, "user connect port")
	clientPort := flag.Int("clientPort", 30001, "client connect port")
	flag.Parse()

	var client, user common.Acceptor
	err := client.Run(*clientPort)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = user.Run(*userPort)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Printf("server startup, userPort:%d and clientPort:%d are listen", *userPort, *clientPort)
	ctrl := common.NewConnection()
	ctrl.Conn = <-client.Conn
	log.Printf("ctrl connection connected")
	go ctrl.Read()
	go ctrl.Write()
	ctrl.IsOpen = true

	recv := make([]byte, 1)
	for {
		select {
		case userConn := <-user.Conn: // 处理新连接
			if ctrl.IsOpen == false {
				log.Printf("ctrl connection is closed , not idle connection")
				userConn.Close()
			} else {
				go func() {
					log.Printf("%s send new connect notify to client", time.Now())
					b := []byte{0xff}
					ctrl.Send <- b // 通知client 有新的连接

					log.Printf("%s wait for clientConn", time.Now())
					clientConn := <-client.Conn // 等待client来连接
					log.Printf("%s accept client connected:%s", time.Now(), clientConn.RemoteAddr())
					go common.SwapConn(userConn, clientConn)
				}()
			}
		case recv = <-ctrl.Recv:
			if recv[0] == 0xFE {
				log.Printf("heart from ctrl client")
			}
		case <-ctrl.ReadClose: // 当ctrl connection断开的时候
			log.Printf("ctrl connction close")
			ctrl.IsOpen = false
			go func() {
				ctrl.Conn = <-client.Conn
				log.Printf("ctrl connection connected")
				go ctrl.Read()
				go ctrl.Write()
				ctrl.IsOpen = true
			}()
		}
	}
}