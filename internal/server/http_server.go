package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/sonirico/visigoth/internal/repos"
	"github.com/sonirico/visigoth/internal/search"
	"github.com/sonirico/visigoth/pkg/entities"
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
	defaultSerializer = search.JSONHitsSearchResultSerializer
)

type APIController struct {
	repo repos.IndexRepo
}

func (s *APIController) handleAlias(w http.ResponseWriter, r *http.Request) {
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

func (s *APIController) handleHealth(w http.ResponseWriter, r *http.Request) {
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

func (s *APIController) handleSearch(w http.ResponseWriter, r *http.Request) {
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
			pending = row.Ser(&defaultSerializer)
		}
		if done {
			break
		}
	}
	_, _ = w.Write(pending)
	_, _ = w.Write([]byte("]}"))
}

func (s *APIController) handleIndex(w http.ResponseWriter, r *http.Request) {
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

func (s *APIController) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/search/", s.handleSearch)
	mux.HandleFunc("/api/index/", s.handleIndex)
	mux.HandleFunc("/api/alias/", s.handleAlias)
	mux.HandleFunc("/_health/", s.handleHealth)
	return mux
}

func NewAPIController(repo repos.IndexRepo) *APIController {
	return &APIController{repo: repo}
}

type httpServer struct {
	addr string
	ctrl *APIController
}

func (s *httpServer) Serve(ctx context.Context) {
	handler := s.ctrl.Handler()
	server := http.Server{Addr: s.addr, Handler: handler}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Println("APIController, listenAndServeError", err)
		}
	}()
	<-ctx.Done()
	if err := server.Shutdown(ctx); err != nil {
		log.Println("htpServer, error on shutdown", err)
	}
}

func NewHTTPServer(addr string, repo repos.IndexRepo) Server {
	return &httpServer{addr: addr, ctrl: NewAPIController(repo)}
}
