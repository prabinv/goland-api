package todo_test

import (
	"context"
	"github.com/prabinv/goland-api/internal/db"
	"github.com/prabinv/goland-api/internal/todo"
	"slices"
	"testing"
)

type MockDB struct {
	counter int
	items   []db.TodoItem
}

func (m *MockDB) Close() {
}

func (m *MockDB) GetAllItems(_ context.Context) ([]db.TodoItem, error) {
	return m.items, nil
}

func (m *MockDB) InsertItem(_ context.Context, item db.TodoItem) (*db.TodoItem, error) {
	item.Id = m.counter + 1
	m.items = append(m.items, item)
	m.counter += 1
	return &item, nil
}

func (m *MockDB) DeleteItem(_ context.Context, i int) error {
	slices.DeleteFunc(m.items, func(item db.TodoItem) bool {
		return item.Id == i
	})
	return nil
}

func (m *MockDB) UpdateItem(_ context.Context, i int, item db.TodoItem) (*db.TodoItem, error) {
	idx := slices.IndexFunc(m.items, func(todoItem db.TodoItem) bool {
		return todoItem.Id == i
	})
	m.items[idx] = item
	return &item, nil
}

func TestService_FilterTodos(t *testing.T) {
	tests := []struct {
		name       string
		todosToAdd []string
		query      string
		want       []todo.Todo
	}{
		{
			name:       "returns empty slice if there are no todos",
			todosToAdd: make([]string, 0),
			query:      "foo",
			want:       []todo.Todo{},
		}, {
			name:       "returns a slice of todos when query matches todo",
			todosToAdd: []string{"go shopping", "laundry", "wash dishes"},
			query:      "shop",
			want: []todo.Todo{
				todo.NewTodo("go shopping"),
			},
		}, {
			name:       "returns a slice of todos when query matches todo but has different casing",
			todosToAdd: []string{"go shopping", "laundry", "wash dishes"},
			query:      "ShOp",
			want: []todo.Todo{
				todo.NewTodo("go shopping"),
			},
		}, {
			name:       "returns a slice of todos when query matches todo fuzzily",
			todosToAdd: []string{"go shopping", "laundry", "wash dishes"},
			query:      "dsh",
			want: []todo.Todo{
				todo.NewTodo("wash dishes"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := todo.NewTodoService(&MockDB{})
			for _, todoItem := range tt.todosToAdd {
				addedTodo, err := s.AddTodo(ctx, todo.NewTodo(todoItem))
				if err != nil {
					t.Fatal(err)
				}
				if todoItem != addedTodo.Name {
					t.Errorf("AddTodos() = got %v, want %v", addedTodo.Name, todoItem)
				}
			}
			if got, _ := s.FilterTodos(ctx, tt.query); !areEqual(got, tt.want) {
				t.Errorf("FilterTodos() = %v, want %v", got, tt.want)
			}
		})
	}
}

func areEqual(got []todo.Todo, expected []todo.Todo) bool {
	if len(got) != len(expected) {
		return false
	}
	for i, t := range got {
		if t.Name != expected[i].Name {
			return false
		}
	}
	return true
}
