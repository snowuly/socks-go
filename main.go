package main

import (
	"flag"

	"github.com/snowuly/socks-go/cmd"
)

var cflag = flag.Bool("c", false, "client mode")

func main() {
	flag.Parse()
	if *cflag {
		cmd.RunClient()
	} else {
		cmd.RunServer()
	}
}
