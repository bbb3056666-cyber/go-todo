package todo

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

type todoStore interface {
	Create(todo Todo) (Todo, error)
}

type Service struct {
	store todoStore
}

func NewService(store todoStore) *Service {
	return &Service{store: store}
}

func (s *Service) Create(title string) (Todo, error) {
	title = strings.TrimSpace(title)
	if length := utf8.RuneCountInString(title); length == 0 || length > 100 {
		return Todo{}, fmt.Errorf("校验标题失败: %w", ErrInvalidTitle)
	}
	todo, err := s.store.Create(Todo{Title: title, Done: false, CreatedAt: time.Now().UTC()})
	if err != nil {
		return Todo{}, fmt.Errorf("创建待办失败: %w", err)
	}
	return todo, nil
}
