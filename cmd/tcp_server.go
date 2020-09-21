package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/sonirico/visigoth/internal"
	"github.com/sonirico/visigoth/internal/repos"
	"github.com/sonirico/visigoth/internal/server"
)

var (
	tcpBindTo string
)

func init() {
	var host, port string
	flag.StringVar(&host, "host", "localhost", "Host address to bind to")
	flag.StringVar(&port, "port", "7373", "Port to listen")
	flag.Parse()

	if strings.Compare(strings.TrimSpace(host), "") == 0 {
		log.Fatal("-host parameter is required")
	}

	if strings.Compare(strings.TrimSpace(port), "") == 0 {
		log.Fatal("-port parameter is required")
	}

	tcpBindTo = fmt.Sprintf("%s:%s", host, port)
}

func main() {
	repo := repos.NewIndexRepo()
	node := server.NewNode(repo)
	transporter := server.NewVTPTransport()
	server := server.NewTcpServer(tcpBindTo, node, transporter)
	done := make(chan bool)

	go func() {
		log.Println("indexing some documents...")
		repo.Put("cursos", internal.NewDocRequest("/c/马桶/", "curso de programacion python mega guay 马"))
		repo.Put("cursos", internal.NewDocRequest("/c/java/", "curso de programacion java mega guay"))
		repo.Put("comics", internal.NewDocRequest("mortadelo & filemon", "hola super intendente"))
	}()

	go func() {
		log.Println("listening on ", tcpBindTo)
		server.Serve()
		done <- true
	}()

	<-done
}
