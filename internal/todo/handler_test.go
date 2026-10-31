package todo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func testMux(store todoStore) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("POST /todos", NewHandler(NewService(store)))
	return mux
}

func assertEnvelope(t *testing.T, response *httptest.ResponseRecorder, status int) map[string]json.RawMessage {
	t.Helper()
	if response.Code != status {
		t.Fatalf("status=%d want=%d body=%s", response.Code, status, response.Body.String())
	}
	mediaType, _, err := mime.ParseMediaType(response.Header().Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		t.Fatalf("Content-Type=%q", response.Header().Get("Content-Type"))
	}
	var envelope map[string]json.RawMessage
	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(&envelope); err != nil {
		t.Fatal(err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		t.Fatalf("response must be one JSON object: %v", err)
	}
	if len(envelope) != 3 {
		t.Fatalf("envelope fields=%v", envelope)
	}
	var code int
	if err := json.Unmarshal(envelope["code"], &code); err != nil {
		t.Fatal(err)
	}
	wantCode := status
	if status == http.StatusCreated {
		wantCode = 0
	}
	if code != wantCode {
		t.Fatalf("code=%d want=%d", code, wantCode)
	}
	var msg string
	if err := json.Unmarshal(envelope["msg"], &msg); err != nil {
		t.Fatal(err)
	}
	if msg == "" || (status == http.StatusCreated && msg != "ok") {
		t.Fatalf("msg=%q", msg)
	}
	if status != http.StatusCreated && string(envelope["data"]) != "null" {
		t.Fatalf("error data=%s", envelope["data"])
	}
	return envelope
}

func TestHandlerCreateSuccess(t *testing.T) {
	for _, tt := range []struct{ name, title, want string }{
		{"one", "a", "a"},
		{"whitespace", "\t\n\u3000学习  Go\n语言 \r", "学习  Go\n语言"},
		{"100 ASCII", strings.Repeat("a", 100), strings.Repeat("a", 100)},
		{"100 Chinese", strings.Repeat("中", 100), strings.Repeat("中", 100)},
		{"100 emoji trimmed", " \t" + strings.Repeat("😀", 100) + "\n", strings.Repeat("😀", 100)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemoryStore()
			body, err := json.Marshal(map[string]string{"title": tt.title})
			if err != nil {
				t.Fatal(err)
			}
			response := httptest.NewRecorder()
			before := time.Now().Truncate(time.Second)
			testMux(store).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(string(body)+" \n\t")))
			envelope := assertEnvelope(t, response, http.StatusCreated)
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(envelope["data"], &fields); err != nil {
				t.Fatal(err)
			}
			if len(fields) != 4 {
				t.Fatalf("data fields=%v", fields)
			}
			var id, title, createdAt string
			var done bool
			for _, field := range []struct {
				key    string
				target any
			}{{"id", &id}, {"title", &title}, {"done", &done}, {"created_at", &createdAt}} {
				raw, ok := fields[field.key]
				if !ok || string(raw) == "null" {
					t.Fatalf("missing or null field %s", field.key)
				}
				if err := json.Unmarshal(raw, field.target); err != nil {
					t.Fatal(err)
				}
			}
			created, err := time.Parse(time.RFC3339, createdAt)
			if err != nil {
				t.Fatal(err)
			}
			if id == "" || title != tt.want || done || created.Before(before) || created.After(time.Now()) {
				t.Fatalf("invalid data: %s", envelope["data"])
			}
			saved, ok := store.todos[id]
			if !ok || len(store.todos) != 1 || saved.Title != title || saved.Done != done || !saved.CreatedAt.Equal(created) {
				t.Fatalf("saved=%+v response=%s", saved, envelope["data"])
			}
		})
	}
}

func TestHandlerInvalidRequest(t *testing.T) {
	cases := []struct{ name, body string }{
		{"empty body", ""}, {"whitespace body", " \t\n"}, {"malformed", `{"title":`},
		{"null object", `null`}, {"array", `[]`}, {"string", `"x"`}, {"number", `1`}, {"boolean", `true`},
		{"second object", `{"title":"valid"}{"title":"another"}`},
		{"second null", `{"title":"valid"} null`}, {"trailing garbage", `{"title":"valid"} garbage`},
		{"missing", `{}`}, {"empty title", `{"title":""}`}, {"blank title", `{"title":" \t\n\u3000"}`},
		{"number title", `{"title":1}`}, {"boolean title", `{"title":false}`},
		{"array title", `{"title":[]}`}, {"object title", `{"title":{}}`}, {"null title", `{"title":null}`},
		{"101 ASCII", `{"title":"` + strings.Repeat("a", 101) + `"}`},
		{"101 Chinese", `{"title":"` + strings.Repeat("中", 101) + `"}`},
		{"101 emoji trimmed", `{"title":" ` + strings.Repeat("😀", 101) + ` "}`},
	}
	for _, field := range []string{"done", "id", "created_at", "other", "Title", "TITLE"} {
		cases = append(cases, struct{ name, body string }{"unknown " + field, `{"title":"valid","` + field + `":"client"}`})
	}
	for _, field := range []string{"Title", "TITLE"} {
		cases = append(cases, struct{ name, body string }{"only " + field, `{"` + field + `":"client"}`})
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemoryStore()
			seed, err := NewService(store).Create("existing")
			if err != nil {
				t.Fatal(err)
			}
			response := httptest.NewRecorder()
			testMux(store).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(tt.body)))
			assertEnvelope(t, response, http.StatusBadRequest)
			if len(store.todos) != 1 || store.todos[seed.ID] != seed {
				t.Fatalf("invalid request changed storage: %+v", store.todos)
			}
		})
	}
}

func TestHandlerRequestBodyLimit(t *testing.T) {
	const limit = 64 * 1024
	const validJSON = `{"title":"valid"}`
	for _, tt := range []struct {
		name   string
		body   string
		status int
	}{
		{"exactly 64KB", strings.Repeat(" ", limit-len(validJSON)) + validJSON, http.StatusCreated},
		{"over 64KB leading whitespace", strings.Repeat(" ", limit+1-len(validJSON)) + validJSON, http.StatusBadRequest},
		{"over 64KB trailing whitespace", validJSON + strings.Repeat(" ", limit+1-len(validJSON)), http.StatusBadRequest},
	} {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemoryStore()
			seed, err := NewService(store).Create("existing")
			if err != nil {
				t.Fatal(err)
			}
			response := httptest.NewRecorder()
			testMux(store).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(tt.body)))
			assertEnvelope(t, response, tt.status)
			if tt.status == http.StatusBadRequest && (len(store.todos) != 1 || store.todos[seed.ID] != seed) {
				t.Fatalf("oversized request changed storage: %+v", store.todos)
			}
		})
	}
}

func TestHandlerInternalError(t *testing.T) {
	for _, detail := range []string{"secret storage failure", "private backend address"} {
		t.Run(detail, func(t *testing.T) {
			store := &recordingStore{err: errors.New(detail)}
			response := httptest.NewRecorder()
			testMux(store).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(`{"title":"valid"}`)))
			body := response.Body.String()
			assertEnvelope(t, response, http.StatusInternalServerError)
			if strings.Contains(body, detail) {
				t.Fatalf("leaked internal error: %s", body)
			}
			if store.calls != 1 {
				t.Fatalf("calls=%d", store.calls)
			}
		})
	}
}

func TestHandlerRouting(t *testing.T) {
	for _, tt := range []struct {
		method, path string
		status       int
	}{
		{http.MethodGet, "/todos", http.StatusMethodNotAllowed},
		{http.MethodPost, "/missing", http.StatusNotFound},
	} {
		t.Run(tt.method+tt.path, func(t *testing.T) {
			store := NewMemoryStore()
			response := httptest.NewRecorder()
			testMux(store).ServeHTTP(response, httptest.NewRequest(tt.method, tt.path, nil))
			if response.Code != tt.status || len(store.todos) != 0 {
				t.Fatalf("status=%d saved=%d", response.Code, len(store.todos))
			}
		})
	}
}

func TestHandlerSharedConcurrentCreation(t *testing.T) {
	for _, tt := range []struct {
		name       string
		count      int
		concurrent bool
	}{
		{"duplicate submissions from different clients", 2, false},
		{"concurrent callers", 50, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemoryStore()
			server := httptest.NewServer(testMux(store))
			defer server.Close()
			results := make(chan Todo, tt.count)
			var wg sync.WaitGroup
			create := func() {
				defer wg.Done()
				transport := &http.Transport{}
				defer transport.CloseIdleConnections()
				client := &http.Client{Transport: transport, Timeout: 10 * time.Second}
				response, err := client.Post(server.URL+"/todos", "application/json", strings.NewReader(`{"title":"same"}`))
				if err != nil {
					t.Error(err)
					return
				}
				defer response.Body.Close()
				var envelope struct {
					Code int    `json:"code"`
					Msg  string `json:"msg"`
					Data Todo   `json:"data"`
				}
				if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
					t.Error(err)
					return
				}
				if response.StatusCode != http.StatusCreated || envelope.Code != 0 || envelope.Msg != "ok" || envelope.Data.Title != "same" || envelope.Data.Done || envelope.Data.CreatedAt.IsZero() {
					t.Errorf("unexpected response: %d %+v", response.StatusCode, envelope)
					return
				}
				results <- envelope.Data
			}
			for i := 0; i < tt.count; i++ {
				wg.Add(1)
				if tt.concurrent {
					go create()
				} else {
					create()
				}
			}
			wg.Wait()
			close(results)
			ids := map[string]bool{}
			for got := range results {
				if got.ID == "" || ids[got.ID] {
					t.Fatalf("invalid or duplicate ID: %q", got.ID)
				}
				ids[got.ID] = true
				if saved, ok := store.todos[got.ID]; !ok || saved.Title != got.Title || saved.Done != got.Done || !saved.CreatedAt.Equal(got.CreatedAt) {
					t.Fatalf("saved and returned differ: %s", fmt.Sprint(got))
				}
			}
			if len(ids) != tt.count || len(store.todos) != tt.count {
				t.Fatalf("returned=%d saved=%d want=%d", len(ids), len(store.todos), tt.count)
			}
		})
	}
}
