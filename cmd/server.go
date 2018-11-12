package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"socks-go/socks"
	"strconv"
)

var (
	errExtraData = errors.New("socks get extra data")
)

func main() {
	run()
}

var traffic = socks.NewTraffic()

func run() {
	config := map[string]string{
		"8080": "chenermao",
		"8081": "duanmingming",
	}

	for port, pwd := range config {
		md5Sum := md5.Sum([]byte(pwd))
		block, _ := aes.NewCipher(md5Sum[:])
		go createServer(port, block)
	}
	traffic.Run()

}

func createServer(port string, block cipher.Block) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Println("create server:", err)
		return
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("accept:", err)
			continue
		}
		go handle(conn, block)
	}
}

func handle(conn net.Conn, block cipher.Block) {
	sconn := socks.NewConn(conn, block)

	closed := false
	defer func() {
		if !closed {
			sconn.Close()
		}
	}()

	if err := sconn.InitRead(); err != nil {
		return
	}

	buf := make([]byte, 266)

	var n int
	var err error

	socks.SetReadTimeout(sconn)
	if n, err = io.ReadFull(sconn, buf[:2]); err != nil {
		log.Println("read remote addr:", err)
		return
	}
	reqLen := -1
	switch buf[0] {
	case 1:
		reqLen = 7
	case 3:
		reqLen = 4 + int(buf[1])
	case 4:
		reqLen = 19
	default:
		// highly possible wrong password
		return
	}
	if _, err = io.ReadFull(sconn, buf[n:reqLen]); err != nil {
		return
	}
	port := binary.BigEndian.Uint16(buf[reqLen-2 : reqLen])

	var host string
	switch buf[0] {
	case 1:
		fallthrough
	case 4:
		host = net.IP(buf[1 : reqLen-2]).String()
	case 3:
		host = string(buf[2 : reqLen-2])
	}
	addr := net.JoinHostPort(host, strconv.Itoa(int(port)))

	var remote net.Conn
	if remote, err = net.Dial("tcp", addr); err != nil {
		return
	}
	log.Println("connected to:", addr)
	closed = true
	go func() {
		socks.PipeThenClose(sconn, remote, func(n int) {
			traffic.Add(port, n)
		})
	}()

	if err = sconn.InitWrite(); err != nil {
		sconn.Close()
		return
	}
	socks.PipeThenClose(remote, sconn, func(n int) {
		traffic.Add(port, n)
	})

}
