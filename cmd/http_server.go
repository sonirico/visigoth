package main

import (
	"fmt"
	"github.com/sonirico/visigoth/internal/repos"
	"github.com/sonirico/visigoth/internal/server"
	"log"
	"net/http"
)

func main() {
	done := make(chan bool)
	repo := repos.NewIndexRepo()
	httpServer := server.NewHttpServer(repo)
	go func() {
		err := http.ListenAndServe(":9000", httpServer)
		done <- true
		if err != nil {
			log.Fatal(err)
		}
	}()
	fmt.Println("server listening at localhost:9000")
	<-done
}
