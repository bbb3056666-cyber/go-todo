package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bbb3056666/go-todo/internal/todo"
)

func main() {
	if err := newServer().ListenAndServe(); err != nil {
		log.Fatal(fmt.Errorf("启动 HTTP 服务失败: %w", err))
	}
}

func newServer() *http.Server {
	store := todo.NewMemoryStore()
	service := todo.NewService(store)
	mux := http.NewServeMux()
	mux.Handle("POST /todos", todo.NewHandler(service))
	return &http.Server{
		Addr:    ":8080",
		Handler: mux,
		// Bound slow headers while allowing normal connection latency.
		ReadHeaderTimeout: 5 * time.Second,
		// Allow time to read the small request body without holding connections indefinitely.
		ReadTimeout: 10 * time.Second,
		// The in-memory operation and small JSON response should finish within this budget.
		WriteTimeout: 10 * time.Second,
		// Keep connections reusable for a minute, then reclaim idle connections.
		IdleTimeout: 60 * time.Second,
	}
}
