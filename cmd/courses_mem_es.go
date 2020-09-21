package main

import (
	"github.com/sonirico/visigoth/internal"
	"github.com/sonirico/visigoth/internal/repos"
	"log"
	"net/http"

	"github.com/sonirico/visigoth/internal/server"
)

func main() {
	repo := repos.NewIndexRepo()
	server := server.NewHttpServer(repo)
	done := make(chan bool)

	go func() {
		log.Println("indexing some documents...")
		repo.Put("cursos", internal.NewDocRequest("/c/python/", "curso de programacion python mega guay"))
		repo.Put("cursos", internal.NewDocRequest("/c/java/", "curso de programacion java mega guay"))
	}()

	go func() {
		log.Println("listening on localhost:9000")
		err := http.ListenAndServe(":9000", server)
		if err != nil {
			log.Fatal(err)
		}
		done <- true
	}()

	<-done
}
