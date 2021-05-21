package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	vindex "github.com/sonirico/visigoth/internal/index"
	"github.com/sonirico/visigoth/internal/repos"
	"github.com/sonirico/visigoth/internal/server"
	"github.com/sonirico/visigoth/pkg/analyze"
	"github.com/sonirico/visigoth/pkg/entities"
)

var (
	bindToTCP  string
	bindToHTTP string
)

type healthIndex struct {
	Ok bool `json:"ok"`
}

func main() {
	flag.StringVar(&bindToHTTP, "http", "localhost:7374", "HTTP port to bind to")
	flag.StringVar(&bindToTCP, "tcp", "localhost:7373", "TCP port to bind to")
	flag.Parse()

	if strings.Compare(strings.TrimSpace(bindToHTTP), "") == 0 {
		log.Fatal("-http parameter is required")
	}

	if strings.Compare(strings.TrimSpace(bindToTCP), "") == 0 {
		log.Fatal("-tcp parameter is required")
	}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	tokenizer := analyze.NewKeepAlphanumericTokenizer()
	analyzer := analyze.NewTokenizationPipeline(&tokenizer,
		analyze.NewLowerCaseTokenizer(),
		analyze.NewStopWordsFilter(analyze.SpanishStopWords),
		analyze.NewSpanishStemmer(true))
	repo := repos.NewIndexRepo(vindex.NewMemoryIndexBuilder(&analyzer))
	node := server.NewNode(repo)
	transporter := server.NewVTPTransport()
	tcpServer := server.NewTCPServer(bindToTCP, node, transporter)
	httpServer := server.NewHTTPServer(bindToHTTP, repo)
	done := make(chan struct{})

	go func() {
		log.Println("indexing debug documents...")
		data, _ := json.Marshal(healthIndex{Ok: true})
		repo.Put("__health__", entities.NewDocRequest("health", string(data)))
	}()

	go func() {
		log.Println("tcp server listening on ", bindToTCP)
		tcpServer.Serve(ctx)
		log.Println("tcp server shutdown")
	}()

	go func() {
		log.Println("http server listening on ", bindToHTTP)
		httpServer.Serve(ctx)
		log.Println("http server shutdown")
	}()

	go func() {
		sig := <-signals
		log.Println("server received signal", sig)
		cancel()
		close(done)
	}()

	<-done
	log.Println("Bye")
}
