package todo

import (
	"errors"
	"strings"
	"testing"
	"time"
)

type recordingStore struct {
	calls int
	saved Todo
	err   error
}

func (s *recordingStore) Create(todo Todo) (Todo, error) {
	s.calls++
	s.saved = todo
	if s.err != nil {
		return Todo{}, s.err
	}
	todo.ID = "server-id"
	return todo, nil
}

func TestServiceCreate(t *testing.T) {
	for _, tt := range []struct {
		name, input, want string
		invalid           bool
	}{
		{name: "one", input: "a", want: "a"},
		{name: "trim Unicode whitespace", input: "\u3000\t 学习  Go\n语言 \r\n", want: "学习  Go\n语言"},
		{name: "100 ASCII", input: strings.Repeat("a", 100), want: strings.Repeat("a", 100)},
		{name: "100 Chinese", input: strings.Repeat("中", 100), want: strings.Repeat("中", 100)},
		{name: "100 emoji trimmed", input: " \t" + strings.Repeat("😀", 100) + "\n", want: strings.Repeat("😀", 100)},
		{name: "empty", invalid: true},
		{name: "whitespace", input: " \t\r\n\u3000", invalid: true},
		{name: "101 ASCII", input: strings.Repeat("a", 101), invalid: true},
		{name: "101 Chinese", input: strings.Repeat("中", 101), invalid: true},
		{name: "101 emoji", input: " " + strings.Repeat("😀", 101) + " ", invalid: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			store := &recordingStore{}
			service := NewService(store)
			before := time.Now()
			got, err := service.Create(tt.input)
			after := time.Now()
			if tt.invalid {
				if !errors.Is(err, ErrInvalidTitle) {
					t.Fatalf("error=%v, want invalid title", err)
				}
				if store.calls != 0 {
					t.Fatalf("invalid title wrote %d times", store.calls)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if store.calls != 1 || got.ID != "server-id" || got.Title != tt.want || got.Done {
				t.Fatalf("unexpected result: %+v, calls=%d", got, store.calls)
			}
			if got.CreatedAt.Before(before) || got.CreatedAt.After(after) || got.CreatedAt.Location() != time.UTC {
				t.Fatalf("invalid server time: %v", got.CreatedAt)
			}
			wantSaved := got
			wantSaved.ID = ""
			if store.saved != wantSaved {
				t.Fatalf("saved=%+v returned=%+v", store.saved, got)
			}
		})
	}
}

func TestServiceStoreError(t *testing.T) {
	for _, cause := range []error{errors.New("storage unavailable"), errors.New("write failed")} {
		t.Run(cause.Error(), func(t *testing.T) {
			store := &recordingStore{err: cause}
			_, err := NewService(store).Create("valid")
			if !errors.Is(err, cause) || err == cause {
				t.Fatalf("error must wrap cause: %v", err)
			}
			if store.calls != 1 {
				t.Fatalf("calls=%d", store.calls)
			}
		})
	}
}
