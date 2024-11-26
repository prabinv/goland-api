package todo

import (
	"context"
	"fmt"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/prabinv/goland-api/internal/db"
	"iter"
	"slices"
	"strings"
)

type Todo struct {
	Id   int    `json:"id"`
	Name string `json:"item"`
	Done bool   `json:"done"`
}

type Service struct {
	db db.Storage
}

func (t Todo) GetDBTodoItem() db.TodoItem {
	return db.TodoItem{
		Id:   t.Id,
		Task: t.Name,
		Done: t.Done,
	}
}

func NewTodoService(db db.Storage) *Service {
	return &Service{
		db: db,
	}
}

func (s *Service) AddTodo(ctx context.Context, t Todo) (*Todo, error) {
	allTodos, err := s.GetTodos(ctx)
	if err != nil {
		return nil, err
	}

	indexOfTodo := slices.IndexFunc(allTodos, func(item Todo) bool {
		return item.Name == t.Name && !item.Done
	})

	if indexOfTodo >= 0 {
		return nil, fmt.Errorf("todo already exists %s", t.Name)
	}

	addedTodoItem, err := s.db.InsertItem(ctx, t.GetDBTodoItem())
	if err != nil {
		return nil, err
	}
	addedTodo := &Todo{
		Id:   addedTodoItem.Id,
		Name: addedTodoItem.Task,
		Done: addedTodoItem.Done,
	}

	return addedTodo, nil
}

func filter[V any](i iter.Seq[V], f func(V) bool) iter.Seq[V] {
	return func(yield func(V) bool) {
		for e := range i {
			if f(e) && !yield(e) {
				return
			}
		}
	}
}

func (s *Service) GetTodos(ctx context.Context) ([]Todo, error) {
	items, err := s.db.GetAllItems(ctx)
	if err != nil {
		return nil, err
	}
	var allTodos = make([]Todo, 0)

	for _, item := range items {
		allTodos = append(allTodos, Todo{
			Id:   item.Id,
			Name: item.Task,
			Done: item.Done,
		})
	}
	return allTodos, nil
}

func (s *Service) FilterTodos(ctx context.Context, query string) ([]Todo, error) {
	var filteredTodos []Todo
	allTodos, err := s.GetTodos(ctx)

	if err != nil {
		return nil, err
	}

	if query != "" {
		filterTodos := func(item Todo) bool {
			return fuzzy.Match(strings.ToLower(query), strings.ToLower(item.Name))
		}
		it := slices.Values(allTodos)

		filteredTodos = slices.Collect(filter(it, filterTodos))
		if filteredTodos == nil {
			filteredTodos = make([]Todo, 0)
		}
	} else {
		filteredTodos = allTodos
	}

	return filteredTodos, nil
}

func (s *Service) DeleteTodo(ctx context.Context, id int) error {
	return s.db.DeleteItem(ctx, id)
}

func (s *Service) UpdateTodo(ctx context.Context, id int, item Todo) (*Todo, error) {
	updatedTodoItem, err := s.db.UpdateItem(ctx, id, item.GetDBTodoItem())
	if err != nil {
		return nil, err
	}

	return &Todo{
		Id:   updatedTodoItem.Id,
		Name: updatedTodoItem.Task,
		Done: updatedTodoItem.Done,
	}, nil
}

func NewTodo(item string) Todo {
	return Todo{
		Name: item,
		Done: false,
	}
}
