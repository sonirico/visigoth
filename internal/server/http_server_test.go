package server

import (
	"bytes"
	"encoding/json"
	"github.com/sonirico/visigoth/internal"
	"github.com/sonirico/visigoth/internal/repos"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func encode(any interface{}) []byte {
	b, _ := json.Marshal(any)
	return b
}

func TestIndexHttpServer_ServeHTTP_Search(t *testing.T) {
	repo := repos.NewIndexRepo()
	server := NewHttpServer(repo)

	// Set up server state
	repo.Put("languages", internal.NewDocRequest("rust", "lenguaje con compilador grunon"))
	repo.Put("languages", internal.NewDocRequest("golang", "lenguaje con ratas azules"))
	repo.Put("languages", internal.NewDocRequest("nodejs", "lenguaje con node_modules/"))

	tests := []struct {
		name         string
		method       string
		uri          string
		expectedCode int
		expectedBody string
		resultSize   int
	}{
		{
			name:         "Search terms query param is mandatory",
			method:       http.MethodGet,
			uri:          "/api/search/", // "terms" query param is missing
			expectedCode: http.StatusUnprocessableEntity,
			expectedBody: "",
			resultSize:   -1,
		},
		{
			name:         "Search terms requires index name as path param",
			method:       http.MethodGet,
			uri:          "/api/search/?terms=ratas", // Missing path param here
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
			resultSize:   -1,
		},
		{
			name:         "Search terms yields not found for non existing indexes",
			method:       http.MethodGet,
			uri:          "/api/search/editors/?terms=vim",
			expectedCode: http.StatusNotFound,
			expectedBody: "{\"message\":\"Index not found\"}",
			resultSize:   -1,
		},
		{
			name:         "Search terms yields results if there are matches",
			method:       http.MethodGet,
			uri:          "/api/search/languages/?terms=lenguaje",
			expectedCode: http.StatusOK,
			resultSize:   3,
		},
		{
			name:         "Search should not yield results if there are not matches",
			method:       http.MethodGet,
			uri:          "/api/search/languages/?terms=php",
			expectedCode: http.StatusOK,
			resultSize:   0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.uri, nil)
			res := httptest.NewRecorder()
			server.ServeHTTP(res, req)

			if res.Code != test.expectedCode {
				t.Errorf("unexpected status code. want %d, have %d", test.expectedCode, res.Code)
			}

			if test.resultSize > -1 {
				content := make(map[string][]struct{})
				_ = json.Unmarshal([]byte(res.Body.String()), &content)
				results := content["results"]
				if len(results) != test.resultSize {
					t.Errorf("unexpected resultset size. want %d, have %d",
						test.resultSize, len(results))
				}
			} else if strings.TrimSpace(res.Body.String()) != test.expectedBody {
				t.Errorf("Want '%s', got '%s'", test.expectedBody, res.Body)
			}
		})
	}
}

func TestIndexHttpServer_ServeHTTP_Index(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		uri          string
		expectedCode int
		data         []byte
		init         func(repo repos.IndexRepo)
		assert       func(repo repos.IndexRepo) bool
	}{
		{
			name:         "Index should not raise not found if index does not exist, but create it",
			method:       http.MethodPut,
			uri:          "/api/index/colours/",
			expectedCode: http.StatusAccepted,
			data:         encode(&PutRequestPayload{Doc: "cyan", Terms: "es un color primario"}),
		},
		{
			name:         "Index should raise unprocessable entity on invalid payloads",
			method:       http.MethodPut,
			uri:          "/api/index/colours/",
			expectedCode: http.StatusUnprocessableEntity,
			data: encode(map[string]string{
				"invalid": "payload",
			}),
		},
		{
			name:         "Dropping an index that does not exist should raise not found",
			method:       http.MethodDelete,
			uri:          "/api/index/vegetables",
			data:         []byte{},
			expectedCode: http.StatusNotFound,
			init: func(repo repos.IndexRepo) {
				repo.Put("fruits", internal.NewDocRequest("香蕉", "西班牙的香蕉是地球上最天天及了"))
			},
			assert: func(repo repos.IndexRepo) bool {
				return repo.Has("fruits")
			},
		},
		{
			name:         "Dropping an index that exists should drop it and return no content",
			method:       http.MethodDelete,
			uri:          "/api/index/fruits",
			data:         []byte{},
			expectedCode: http.StatusNoContent,
			init: func(repo repos.IndexRepo) {
				repo.Put("fruits", internal.NewDocRequest("香蕉", "西班牙的香蕉是地球上最天天及了"))
			},
			assert: func(repo repos.IndexRepo) bool {
				return !repo.Has("fruits")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.uri, bytes.NewReader(test.data))
			res := httptest.NewRecorder()
			repo := repos.NewIndexRepo()
			if test.init != nil {
				test.init(repo)
			}
			server := NewHttpServer(repo)
			server.ServeHTTP(res, req)

			if res.Code != test.expectedCode {
				t.Errorf("unexpected status code. want %d, have %d", test.expectedCode, res.Code)
				return
			}

			if test.assert != nil && !test.assert(repo) {
				return
			}
		})
	}
}

func TestIndexHttpServer_ServeHTTP_Alias(t *testing.T) {
	// Set up server state
	repo := repos.NewIndexRepo()
	repo.Put("languages", internal.NewDocRequest("rust", "lenguaje con compilador grunon"))
	repo.Alias("rustaceans", "languages")

	tests := []struct {
		name         string
		method       string
		uri          string
		expectedCode int
		data         []byte
	}{
		{
			name:         "No other verbs that put and post should be accepted",
			method:       http.MethodGet,
			uri:          "/api/alias/languages/",
			expectedCode: http.StatusMethodNotAllowed,
			data:         []byte{},
		},
		{
			name:         "Index should raise not found if index does not exist",
			method:       http.MethodPut,
			uri:          "/api/alias/colours/",
			expectedCode: http.StatusNotFound,
			data:         []byte("{\"as\": \"colours:top:5\"}"),
		},
		{
			name:         "Index should raise unprocessable entity on invalid payloads",
			method:       http.MethodPut,
			uri:          "/api/alias/languages/",
			expectedCode: http.StatusUnprocessableEntity,
			data:         []byte("{\"invalid_key\": \"languages:top:5\"}"),
		},
		{
			name:         "Index should yield created entity response when created",
			method:       http.MethodPut,
			uri:          "/api/alias/languages/",
			expectedCode: http.StatusCreated,
			data:         []byte("{\"as\": \"languages:top:5\"}"),
		},
		{
			name:         "Unaliasing a non-existing alias should yield not found",
			method:       http.MethodDelete,
			uri:          "/api/alias/rustaceos", // This alias does not exist
			expectedCode: http.StatusNotFound,
			data:         []byte{},
		},
		{
			name:         "Unaliasing a existing alias should return no content (ok)",
			method:       http.MethodDelete,
			uri:          "/api/alias/rustaceans", // This alias does exist
			expectedCode: http.StatusNoContent,
			data:         []byte{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.uri, bytes.NewReader(test.data))
			res := httptest.NewRecorder()
			server := NewHttpServer(repo)
			server.ServeHTTP(res, req)

			if res.Code != test.expectedCode {
				t.Errorf("unexpected status code. want %d, have %d", test.expectedCode, res.Code)
				return
			}
		})
	}
}
