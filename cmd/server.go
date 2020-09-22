package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/sonirico/visigoth/pkg/entities"
	"log"
	"net/http"
	"strings"

	"github.com/sonirico/visigoth/internal/repos"
	"github.com/sonirico/visigoth/internal/server"
)

var (
	bindToTcp  string
	bindToHttp string
)

type healthIndex struct {
	Ok bool `json:"ok"`
}

func init() {
	var tcpPort, httpPort string
	flag.StringVar(&httpPort, "http", "7374", "HTTP port to bind to")
	flag.StringVar(&tcpPort, "tcp", "7373", "TCP port to bind to")
	flag.Parse()

	if strings.Compare(strings.TrimSpace(httpPort), "") == 0 {
		log.Fatal("-http parameter is required")
	}

	if strings.Compare(strings.TrimSpace(tcpPort), "") == 0 {
		log.Fatal("-tcp parameter is required")
	}

	bindToTcp = fmt.Sprintf("localhost:%s", tcpPort)
	bindToHttp = fmt.Sprintf(":%s", httpPort)
}

func main() {
	repo := repos.NewIndexRepo()
	node := server.NewNode(repo)
	transporter := server.NewVTPTransport()
	tcpServer := server.NewTcpServer(bindToTcp, node, transporter)
	httpServer := server.NewHttpServer(repo)
	done := make(chan bool)

	go func() {
		log.Println("indexing debug documents...")
		data, _ := json.Marshal(&healthIndex{Ok: true})
		repo.Put("__health__", entities.NewDocRequest("health", string(data)))
	}()

	go func() {
		log.Println("tcp server listening on ", bindToTcp)
		tcpServer.Serve()
		done <- true
	}()

	go func() {
		log.Println("http server listening on ", bindToHttp)
		err := http.ListenAndServe(":9000", httpServer)
		done <- true
		if err != nil {
			log.Fatal(err)
		}
	}()

	<-done
	log.Println("Bye")
}
