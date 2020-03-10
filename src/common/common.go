package common

import (
	"fmt"
	"net"
	"strconv"
)

type connection struct {
	Conn       net.Conn
	Recv       chan []byte
	Send       chan []byte
	ReadClose  chan struct{}
	closeWrite chan struct{}
	IsOpen     bool
}

func NewConnection() *connection {
	return &connection{
		Recv:       make(chan []byte),
		Send:       make(chan []byte),
		ReadClose:  make(chan struct{}),
		closeWrite: make(chan struct{}),
	}
}

func (c connection) Read() {
	defer c.Conn.Close()

	buf := make([]byte, 1024)
	for {
		len, err := c.Conn.Read(buf)
		if err != nil {
			fmt.Println("read error ", err)
			c.closeWrite <- struct{}{}
			break
		}
		if len == 0 {
			fmt.Println("read 0")
			c.closeWrite <- struct{}{}
			break
		}
		c.Recv <- buf[0:len]
	}
	c.ReadClose <- struct{}{}
}

func (c connection) Write() {
	for {
		select {
		case buf := <-c.Send:
			_, err := c.Conn.Write(buf)
			if err != nil {
				fmt.Println("write err ", err)
				return
			}
		case <-c.closeWrite:
			fmt.Println("close connection write")
			return
		}
	}
}

func CopyConnection(dst net.Conn, src net.Conn) {
	defer dst.Close()
	defer src.Close()
	for {
		recvBuff := make([]byte, 1024)
		len, err := src.Read(recvBuff)
		if err != nil {
			fmt.Println("read info:", err)
			return
		}
		len, err = dst.Write(recvBuff[:len])
		if err != nil {
			fmt.Println("write info", err)
		}
	}
}

func SwapConn(conn1 net.Conn, conn2 net.Conn) {
	go CopyConnection(conn1, conn2)
	CopyConnection(conn2, conn1)
}

type Acceptor struct {
	lister net.Listener
	Conn   chan net.Conn
}

func (l *Acceptor) Run(port int) error {
	lister, err := net.Listen("tcp", "0.0.0.0:" + strconv.Itoa(port))
	if err != nil {
		fmt.Println(err)
		return err
	}
	l.Conn = make(chan net.Conn)
	l.lister = lister
	go l.accept()
	return nil
}

func (l *Acceptor) accept() {
	for {
		conn, err := l.lister.Accept()
		if err != nil {
			fmt.Println("accept err", err)
			break
		}
		l.Conn <- conn
		fmt.Println("accept new connect ", "RemoteAddr:", conn.RemoteAddr())
	}
	l.lister.Close()
}