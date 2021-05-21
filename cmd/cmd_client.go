package main

import (
	"flag"
	"fmt"
	"github.com/sonirico/visigoth/pkg/client"
	"log"
	"os"
	"runtime"
	"strings"
)

var (
	cmdBindTo string
)

func main() {
	var host, port string
	flag.StringVar(&host, "host", "localhost", "server address")
	flag.StringVar(&port, "port", "7373", "server port")
	flag.Parse()

	if strings.Compare(strings.TrimSpace(host), "") == 0 {
		log.Fatal("-host parameter is required")
	}

	if strings.Compare(strings.TrimSpace(port), "") == 0 {
		log.Fatal("-port parameter is required")
	}

	cmdBindTo = fmt.Sprintf("%s:%s", host, port)

	client.NewCmdClient(cmdBindTo).
		Repl(os.Stdin, os.Stdout)

	runtime.Goexit()
}
