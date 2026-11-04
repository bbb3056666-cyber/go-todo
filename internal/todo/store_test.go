package todo

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestMemoryStoreCreate(t *testing.T) {
	for _, title := range []string{"学习 Go", "重复标题"} {
		t.Run(title, func(t *testing.T) {
			s := NewMemoryStore()
			input := Todo{Title: title, CreatedAt: time.Now().UTC()}
			first, err := s.Create(input)
			if err != nil {
				t.Fatal(err)
			}
			second, err := s.Create(input)
			if err != nil {
				t.Fatal(err)
			}
			if first.ID == "" || second.ID == "" || first.ID == second.ID {
				t.Fatalf("invalid IDs: %q, %q", first.ID, second.ID)
			}
			if len(s.todos) != 2 {
				t.Fatalf("saved count = %d", len(s.todos))
			}
			for _, got := range []Todo{first, second} {
				if got.Title != input.Title || got.Done != input.Done || !got.CreatedAt.Equal(input.CreatedAt) {
					t.Fatalf("saved content changed: %+v", got)
				}
				if s.todos[got.ID] != got {
					t.Fatalf("returned and saved values differ: %+v", got)
				}
			}
			first.Title = "modified"
			if s.todos[first.ID].Title != title {
				t.Fatal("returned value aliases stored value")
			}
		})
	}
}

func TestMemoryStoreConcurrentCreate(t *testing.T) {
	for _, count := range []int{1, 100} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			s := NewMemoryStore()
			results := make(chan Todo, count)
			var wg sync.WaitGroup
			for i := 0; i < count; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					got, err := s.Create(Todo{Title: "shared"})
					if err != nil {
						t.Error(err)
						return
					}
					results <- got
				}()
			}
			wg.Wait()
			close(results)
			ids := map[string]bool{}
			for got := range results {
				if got.ID == "" || ids[got.ID] {
					t.Fatalf("invalid or duplicate ID: %q", got.ID)
				}
				ids[got.ID] = true
				if s.todos[got.ID] != got {
					t.Fatalf("missing saved todo: %+v", got)
				}
			}
			if len(ids) != count || len(s.todos) != count {
				t.Fatalf("returned=%d saved=%d want=%d", len(ids), len(s.todos), count)
			}
		})
	}
}
