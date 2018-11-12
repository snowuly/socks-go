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
	"strconv"

	"github.com/snowuly/socks/secureconn"
)

var (
	errExtraData = errors.New("socks get extra data")
)

func main() {
	run()
}

func run() {
	config := map[string]string{
		"8387": "chenermao",
		"8389": "chenyinuo",
	}

	for port, pwd := range config {
		md5Sum := md5.Sum([]byte(pwd))
		block, _ := aes.NewCipher(md5Sum[:])
		go createServer(port, block)
	}

	ch := make(chan struct{})
	<-ch
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
	closed := false
	defer func() {
		if !closed {
			conn.Close()
		}
	}()
	sconn := secureconn.New(conn, block)

	if err := sconn.InitRead(); err != nil {
		log.Println("init read:", err)
		return
	}

	buf := make([]byte, 32*1024)

	var n int
	var err error

	if n, err = io.ReadAtLeast(sconn, buf, 2); err != nil {
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
		log.Println("unsupport addr type: %d\n", buf[0])
		return
	}
	if n < reqLen {
		if _, err = io.ReadFull(sconn, buf[:reqLen]); err != nil {
			log.Println("read remote addr:", err)
			return
		}
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
		log.Println("connect remote:", err)
		return
	}
	if !closed {
		remote.Close()
	}
	go func() {
		if n > reqLen {
			remote.Write(buf[reqLen:n])
		}
		io.Copy(remote, sconn)
		conn.Close()
		remote.Close()
		closed = true
	}()

	if err = sconn.InitWrite(); err != nil {
		log.Println("init write:", err)
		return
	}
	io.Copy(sconn, remote)

}
