package socks

import (
	"bufio"
	"bytes"
	"log"
	"net"
	"strconv"
	"strings"
)

// StartProxy starts auto proxy agent at :1082/proxy.pac
func StartProxy() {
	ln, err := net.Listen("tcp", ":1082")

	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := ln.Accept()

		if err != nil {
			log.Print(err)
			continue
		}
		go handleProxy(conn)
	}
}

func handleProxy(c net.Conn) {
	defer c.Close()

	rd := bufio.NewReaderSize(c, 2048)

	line, err := rd.ReadSlice('\n')
	if err != nil {
		return
	}
	if !bytes.Equal(line[:14], []byte{71, 69, 84, 32, 47, 112, 114, 111, 120, 121, 46, 112, 97, 99}) { // GET /proxy.pac
		return
	}
	for i := 0; i < 20; i++ {
		line, err = rd.ReadSlice('\n')
		if err != nil {
			log.Print(err)
		}
		if len(line) <= 2 {
			writeResponse(c)
			return
		}
	}
}

func writeResponse(c net.Conn) {
	addr := c.LocalAddr().String()
	ip := addr[:strings.LastIndexByte(addr, ':')]

	c.Write([]byte("HTTP/1.1 200 OK\r\nCache-Control: max-age=600\r\nContent-Type: text/plain\r\nContent-Length: " + strconv.Itoa(len(ip)+61) + "\r\n\r\nfunction FindProxyForURL(url, host) { return \"SOCKS " + ip + ":1081\"; }"))
}
