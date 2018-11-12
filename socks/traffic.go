package socks

import (
	"log"
	"sync"
	"time"
)

type Traffic struct {
	sync.Mutex
	stats map[uint16]int64
}

func NewTraffic() *Traffic {
	return &Traffic{stats: make(map[uint16]int64)}
}

func (t *Traffic) Add(port uint16, n int) {
	t.Lock()
	t.stats[port] += int64(n)
	t.Unlock()
}

func (t *Traffic) Run() {
	tick := time.Tick(10 * time.Second)

	for {
		<-tick
		log.Println(t.stats)
	}
}
