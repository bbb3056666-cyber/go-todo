package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bbb3056666/go-todo/internal/todo"
)

func main() {
	store := todo.NewMemoryStore()
	service := todo.NewService(store)
	mux := http.NewServeMux()
	mux.Handle("POST /todos", todo.NewHandler(service))
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(fmt.Errorf("启动 HTTP 服务失败: %w", err))
	}
}
