package main

import (
	"context"
	"encoding/json"
	"flag"
	vindex "github.com/sonirico/visigoth/internal/index"
	"github.com/sonirico/visigoth/internal/repos"
	"github.com/sonirico/visigoth/internal/server"
	"github.com/sonirico/visigoth/pkg/analyze"
	"github.com/sonirico/visigoth/pkg/entities"
	"log"
	"os"
	"os/signal"
	"strings"
)

var (
	bindToTcp  string
	bindToHttp string
)

type healthIndex struct {
	Ok bool `json:"ok"`
}

func init() {
	flag.StringVar(&bindToHttp, "http", "localhost:7374", "HTTP port to bind to")
	flag.StringVar(&bindToTcp, "tcp", "localhost:7373", "TCP port to bind to")
	flag.Parse()

	if strings.Compare(strings.TrimSpace(bindToHttp), "") == 0 {
		log.Fatal("-http parameter is required")
	}

	if strings.Compare(strings.TrimSpace(bindToTcp), "") == 0 {
		log.Fatal("-tcp parameter is required")
	}
}

func main() {
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Kill, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	tokenizer := analyze.NewKeepAlphanumericTokenizer()
	analyzer := analyze.NewTokenizationPipeline(&tokenizer,
		analyze.NewLowerCaseTokenizer(),
		analyze.NewStopWordsFilter(analyze.SpanishStopWords),
		analyze.NewSpanishStemmer(true))
	repo := repos.NewIndexRepo(vindex.NewMemoryIndexBuilder(&analyzer))
	node := server.NewNode(repo)
	transporter := server.NewVTPTransport()
	tcpServer := server.NewTcpServer(bindToTcp, node, transporter)
	httpServer := server.NewHttpServer(bindToHttp, repo)
	done := make(chan struct{})

	go func() {
		log.Println("indexing debug documents...")
		data, _ := json.Marshal(&healthIndex{Ok: true})
		repo.Put("__health__", entities.NewDocRequest("health", string(data)))
	}()

	go func() {
		log.Println("tcp server listening on ", bindToTcp)
		tcpServer.Serve(ctx)
		log.Println("tcp server shutdown")
	}()

	go func() {
		log.Println("http server listening on ", bindToHttp)
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
