package socks

import (
	"net"
	"time"
)

func SetReadTimeout(conn net.Conn) {
	if readTimeout != 0 {
		conn.SetReadDeadline(time.Now().Add(readTimeout))
	}
}

func PipeThenClose(src, dst net.Conn, port uint16) {
	defer dst.Close()
	buf := leakbuf.Get()
	defer leakbuf.Put(buf)

	for {
		SetReadTimeout(src)
		n, err := src.Read(buf)

		if n > 0 {
			TrafficAdd(port, n)

			_, err := dst.Write(buf[:n])
			if err != nil {
				return
			}
		}
		if err != nil {
			return
		}
	}

}
