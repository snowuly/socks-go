// +build ignore

package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	"socks-go/socks"
)

var (
	errVer       = errors.New("socks version not supported")
	errMethod    = errors.New("socks only only supported no authentication method")
	errExtraData = errors.New("socks authentication get extra data")
	errCmd       = errors.New("socks cmd not supported")
	errAddrType  = errors.New("socks addr type not supported")
)

const (
	socksVer5       = 5
	methodNoAuth    = 0
	socksCmdConnect = 1
	debug           = false
)

var (
	readTimeout = 20 // second
	serverAddr  = "127.0.0.1:8080"
	password    = "test"
)

func main() {
	run()
}

func run() {

	ln, err := net.Listen("tcp", ":1081")
	if err != nil {
		log.Fatal(err)
	}

	go socks.StartProxy()

	key := md5.Sum([]byte(password))
	block, _ := aes.NewCipher(key[:])

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("accept:", err)
			continue
		}
		go handleClient(conn, block)
	}
}

func handleClient(conn net.Conn, block cipher.Block) {
	var closed = false
	defer func() {
		if !closed {
			conn.Close()
		}
	}()

	var err error
	if err = handShake(conn); err != nil {
		log.Println("socks handshake:", err)
		return
	}

	var rawAddr []byte
	if rawAddr, _, err = getRequest(conn); err != nil {
		log.Println("socks getRequest:", err)
		return
	}

	if _, err = conn.Write([]byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0}); err != nil {
		return
	}

	remote, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.Println("connect remote server:", err)
		return
	}
	sconn := socks.NewConn(remote, block)

	closed = true
	go func() {
		defer conn.Close()
		if err := sconn.InitRead(false); err != nil {
			// possiblely invalid domain name or ip address, just ignore the error
			return
		}
		io.Copy(conn, sconn)
	}()

	if err = sconn.InitWrite(true); err != nil {
		log.Println("init write:", err)
		return
	}
	sconn.Write(rawAddr)
	io.Copy(sconn, conn)
	sconn.Close()
}

func getRequest(conn net.Conn) (rawAddr []byte, host string, err error) {
	setReadTimeout(conn)
	// 1(ver) + 1(cmd) + 1(rsv) + atyp(1) + 255(domainname) + port(2)
	buf := make([]byte, 261)
	var n int

	// read until get the possible domainname length
	if n, err = io.ReadAtLeast(conn, buf, 5); err != nil {
		return
	}

	if buf[0] != socksVer5 {
		err = errVer
		return
	}
	if buf[1] != socksCmdConnect {
		err = errCmd
		return
	}

	reqLen := -1
	switch buf[3] {
	case 1: // ipv4
		reqLen = 10
	case 3: // domain name
		reqLen = 7 + int(buf[4])
	case 4: // ipv6
		reqLen = 22
	default:
		err = errAddrType
		return
	}
	if n < reqLen {
		if _, err = io.ReadFull(conn, buf[n:reqLen]); err != nil {
			return
		}
	} else if n > reqLen {
		err = errExtraData
		return
	}
	rawAddr = buf[3:reqLen]

	if debug {
		switch buf[3] {
		case 1:
			host = net.IP(buf[4:8]).String()
		case 3:
			host = string(buf[5 : buf[4]+5])
		case 4:
			host = net.IP(buf[4:20]).String()
		}
		port := binary.BigEndian.Uint16(buf[reqLen-2 : reqLen])
		host = net.JoinHostPort(host, strconv.Itoa(int(port)))
	}
	return

}

func handShake(conn net.Conn) (err error) {
	buf := make([]byte, 8)

	var n int
	setReadTimeout(conn)
	if n, err = io.ReadAtLeast(conn, buf, 2); err != nil {
		return
	}

	if buf[0] != socksVer5 {
		return errVer
	}
	nmethod := int(buf[1])
	msgLen := nmethod + 2

	if n < msgLen {
		if _, err = io.ReadFull(conn, buf[n:msgLen]); err != nil {
			return
		}
	} else if n > msgLen {
		return errExtraData
	}
	for _, method := range buf[2:msgLen] {
		if method == methodNoAuth {
			_, err = conn.Write([]byte{socksVer5, 0})
			return
		}
	}
	return errMethod
}

func setReadTimeout(conn net.Conn) {
	if readTimeout != 0 {
		conn.SetReadDeadline(time.Now().Add(time.Duration(readTimeout) * time.Second))
	}
}
