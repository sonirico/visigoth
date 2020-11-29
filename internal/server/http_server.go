package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sonirico/visigoth/internal/repos"
	"github.com/sonirico/visigoth/internal/search"
	"github.com/sonirico/visigoth/pkg/entities"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type PutRequestPayload struct {
	Terms string `json:"content"`
	Doc   string `json:"doc"`
}

type AliasRequestPayload struct {
	As string `json:"as"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

const (
	comma = byte(44)
)

var (
	commaSer                      = []byte{comma}
	healthOKResponseSerialized, _ = json.Marshal(HealthResponse{
		Status: "ok",
	})
	indexDoesNotExistResponse, _ = json.Marshal(ErrorResponse{
		Message: "Index not found",
	})
	defaultEngine     = search.HitsSearchEngine
	defaultSerializer = search.JsonHitsSearchResultSerializer
)

type apiController struct {
	repo repos.IndexRepo
}

func (s *apiController) handleAlias(w http.ResponseWriter, r *http.Request) {
	switch verb := r.Method; {
	case verb == http.MethodDelete:
		w.Header().Set("content-type", "application/json")
		alias, err := parseIndex(r.URL.Path) // TODO: validate index name
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if s.repo.UnAlias(alias, "") {
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	case verb == http.MethodPost || verb == http.MethodPut:
		w.Header().Set("content-type", "application/json")
		iname, err := parseIndex(r.URL.Path) // TODO: validate index name
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		payload := new(AliasRequestPayload)
		body, _ := ioutil.ReadAll(r.Body)
		if err := json.Unmarshal(body, payload); err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		alias := strings.TrimSpace(payload.As)
		if len(alias) < 1 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		if !s.repo.Alias(alias, iname) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusCreated)
			w.Header().Add("location", fmt.Sprintf("/api/index/%s", alias))
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *apiController) handleHealth(w http.ResponseWriter, r *http.Request) {
	if http.MethodGet != r.Method {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(healthOKResponseSerialized)
	if err != nil {
		log.Fatal("could not send back status health")
	}
}

func (s *apiController) handleSearch(w http.ResponseWriter, r *http.Request) {
	if http.MethodGet != r.Method {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	terms, ok := r.URL.Query()["terms"]
	if !ok {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	w.Header().Set("content-type", "application/json")
	iname, err := parseIndex(r.URL.Path) // TODO: validate index name
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	result, err := s.repo.Search(iname, strings.Join(terms, " "), defaultEngine)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write(indexDoesNotExistResponse)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("{\"results\": ["))
	var pending []byte
	for {
		row, done := result.Next()
		if row != nil {
			if pending != nil {
				_, _ = w.Write(pending)
				_, _ = w.Write(commaSer)
			}
			pending = row.Ser(defaultSerializer)
		}
		if done {
			break
		}
	}
	_, _ = w.Write(pending)
	_, _ = w.Write([]byte("]}"))
}

func (s *apiController) handleIndex(w http.ResponseWriter, r *http.Request) {
	switch v := r.Method; {
	case v == http.MethodDelete:
		iname, err := parseIndex(r.URL.Path) // TODO: validate index name
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write(indexDoesNotExistResponse)
			return
		}

		if s.repo.Drop(iname) {
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	case v == http.MethodPost || v == http.MethodPut:
		w.Header().Set("content-type", "application/json")
		iname, err := parseIndex(r.URL.Path) // TODO: validate index name
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write(indexDoesNotExistResponse)
			return
		}
		payload := new(PutRequestPayload)
		body, _ := ioutil.ReadAll(r.Body) // TODO: Streaming
		if err := json.Unmarshal(body, payload); err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		if len(payload.Terms) < 1 || len(payload.Doc) < 1 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		docRequest := entities.NewDocRequest(payload.Doc, payload.Terms)
		s.repo.Put(iname, docRequest)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *apiController) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/search/", s.handleSearch)
	mux.HandleFunc("/api/index/", s.handleIndex)
	mux.HandleFunc("/api/alias/", s.handleAlias)
	mux.HandleFunc("/_health/", s.handleHealth)
	return mux
}

func NewApiController(repo repos.IndexRepo) *apiController {
	return &apiController{repo: repo}
}

type httpServer struct {
	addr string
	ctrl *apiController
}

func (s *httpServer) Serve(ctx context.Context) {
	handler := s.ctrl.Handler()
	server := &http.Server{Addr: s.addr, Handler: handler}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Println("apiController, listenAndServeError", err)
		}

	}()
	select {
	case <-ctx.Done():
		if err := server.Shutdown(ctx); err != nil {
			log.Println("htpServer, error on shutdown", err)
		}
	}
}

func NewHttpServer(addr string, repo repos.IndexRepo) Server {
	return &httpServer{addr: addr, ctrl: NewApiController(repo)}
}
