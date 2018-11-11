package secureconn

import (
	"crypto/cipher"
	"crypto/rand"
	"io"
	"net"
)

const (
	blockSize = 16
	bufSize   = 32 * 1024
)

type Conn struct {
	conn  net.Conn
	block cipher.Block
	enc   cipher.Stream
	dec   cipher.Stream
	wbuf  []byte
}

func (c *Conn) Read(b []byte) (n int, err error) {
	n, err = c.conn.Read(b)
	if n > 0 {
		c.dec.XORKeyStream(b[:n], b[:n])
	}
	return
}
func (c *Conn) Write(b []byte) (int, error) {
	n := len(b)
	c.enc.XORKeyStream(c.wbuf[:n], b)
	return c.conn.Write(c.wbuf[:n])
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) InitRead() (err error) {
	iv := make([]byte, blockSize)
	if _, err = io.ReadFull(c.conn, iv); err != nil {
		return
	}
	c.dec = cipher.NewCFBDecrypter(c.block, iv)
	return
}
func (c *Conn) InitWrite() (err error) {
	iv := make([]byte, blockSize)
	rand.Read(iv)
	c.enc = cipher.NewCFBEncrypter(c.block, iv)
	c.wbuf = make([]byte, bufSize)
	_, err = c.conn.Write(iv)
	return
}

func New(conn net.Conn, block cipher.Block) *Conn {
	return &Conn{conn, block, nil, nil, nil}
}
