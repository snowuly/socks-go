package socks

type Leakbuf struct {
	bufSize  int
	freeList chan []byte
}

var leakbuf = NewLeakbuf(8*1024, 2*1024)

func NewLeakbuf(bufSize, length int) *Leakbuf {
	return &Leakbuf{bufSize, make(chan []byte, length)}
}

func (lb *Leakbuf) Get() (b []byte) {
	select {
	case b = <-lb.freeList:
	default:
		b = make([]byte, lb.bufSize)
	}
	return
}

func (lb *Leakbuf) Put(b []byte) {
	select {
	case lb.freeList <- b:
	default:
	}
}
