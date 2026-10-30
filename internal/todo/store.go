package todo

import (
	"strconv"
	"sync"
)

// MemoryStore holds tasks shared by all callers in this process.
type MemoryStore struct {
	mu     sync.Mutex
	nextID uint64
	todos  map[string]Todo
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{todos: make(map[string]Todo)}
}

func (s *MemoryStore) Create(todo Todo) (Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	todo.ID = strconv.FormatUint(s.nextID, 10)
	s.todos[todo.ID] = todo
	return todo, nil
}
