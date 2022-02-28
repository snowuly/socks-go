package main

import (
	"flag"
	"log"

	"github.com/snowuly/socks-go/cmd"
)

var cmode = flag.String("c", "", "client mode and it's value is server address")
var pwd = flag.String("s", "", "password")
var port = flag.Uint("p", 0, "port")

func main() {
	flag.Parse()

	if *pwd == "" {
		log.Fatal("password is required")
	}

	if *cmode == "" { // server mode
		port := uint16(*port)
		if port <= 1024 || port > 65535 {
			log.Fatal("port error")
		}
		cmd.RunServer(port, *pwd)
	} else { // comde is server address
		cmd.RunClient(*cmode, *pwd)
	}
}
