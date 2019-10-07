package socks

import (
	"log"
	"time"
)

type Info struct {
	data int64
	last int64
	ch   chan int
}

var traffic = make(map[uint16]*Info)

func TrafficAdd(port uint16, n int) {
	traffic[port].ch <- n
}

func initPorts(ports []uint16) {
	for _, port := range ports {
		traffic[port] = &Info{ch: make(chan int, 100)}
	}
}

func TrafficRun(ports []uint16) {
	initPorts(ports)

	for _, info := range traffic {
		go func(info *Info) {
			for n := range info.ch {
				info.data += int64(n)
			}
		}(info)
	}

	printStat()

}

func printStat() {
	tick := time.Tick(5 * time.Minute)

	for {
		<-tick

		for port, info := range traffic {
			if info.last != info.data {
				log.Printf("%d:%d:%d\n", port, info.data, info.data-info.last)
				info.last = info.data
			}
		}
	}
}
