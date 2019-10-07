package socks

import (
	"crypto/cipher"
	"crypto/rand"
	"io"
	"net"
)

const (
	blockSize = 16
)

type Conn struct {
	net.Conn
	block cipher.Block
	wbuf  []byte
	enc   cipher.Stream
	dec   cipher.Stream
}

func (c *Conn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	if n > 0 {
		c.dec.XORKeyStream(b[:n], b[:n])
	}
	return
}
func (c *Conn) Write(b []byte) (int, error) {
	n := len(b)
	c.enc.XORKeyStream(c.wbuf[:n], b)
	return c.Conn.Write(c.wbuf[:n])
}

func (c *Conn) Close() error {
	leakbuf.Put(c.wbuf)
	return c.Conn.Close()
}

func (c *Conn) InitRead(bypass bool) (err error) {
	if bypass {
		b := make([]byte, prefaceLen)
		if _, err = io.ReadFull(c.Conn, b); err != nil {
			return
		}
	}
	iv := make([]byte, blockSize)
	if _, err = io.ReadFull(c.Conn, iv); err != nil {
		return
	}
	c.dec = cipher.NewCFBDecrypter(c.block, iv)
	return
}
func (c *Conn) InitWrite(bypass bool) (err error) {
	iv := make([]byte, blockSize)
	rand.Read(iv)
	c.enc = cipher.NewCFBEncrypter(c.block, iv)
	if bypass {
		_, err = c.Conn.Write(preface)
		if err != nil {
			return
		}
	}
	_, err = c.Conn.Write(iv)
	return
}

func NewConn(conn net.Conn, block cipher.Block) *Conn {
	return &Conn{conn, block, leakbuf.Get(), nil, nil}
}
