package transport

import (
	"encoding/json"
	"fmt"
	"github.com/prabinv/goland-api/internal/todo"
	"net/http"
	"strconv"
)

type TodoItem struct {
	Item string `json:"item"`
}

type Server struct {
	mux *http.ServeMux
}

func NewServer(s *todo.Service) *Server {

	server := &Server{
		mux: http.NewServeMux(),
	}
	server.mux.HandleFunc("GET /todo", server.HandleGetTodos(s))

	server.mux.HandleFunc("GET /todo/search", server.HandleSearchTodos(s))

	server.mux.HandleFunc("POST /todo", server.HandlePostTodo(s))

	server.mux.HandleFunc("PUT /todo/{id}", server.HandlePutTodo(s))

	server.mux.HandleFunc("DELETE /todo/{id}", server.HandleDeleteTodo(s))

	return server
}

func (s *Server) Serve() error {
	return http.ListenAndServe(":8080", s.mux)
}

func (s *Server) HandleDeleteTodo(svc *todo.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := r.PathValue("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		if err := svc.DeleteTodo(r.Context(), id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) HandleGetTodos(svc *todo.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		todos, err := svc.GetTodos(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		resp, err := json.Marshal(todos)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		if _, err = w.Write(resp); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (s *Server) HandlePostTodo(svc *todo.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var todoItem TodoItem
		ctx := r.Context()

		if err := json.NewDecoder(r.Body).Decode(&todoItem); err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		fmt.Printf("todo item %+v", todoItem)

		newTodo := todo.NewTodo(todoItem.Item)
		addedTodo, err := svc.AddTodo(ctx, newTodo)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			if _, err = w.Write([]byte(err.Error())); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}

		resp, err := json.Marshal(addedTodo)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		if _, err = w.Write(resp); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (s *Server) HandlePutTodo(svc *todo.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		idString := r.PathValue("id")
		id, err := strconv.Atoi(idString)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		var todoItem todo.Todo
		if err := json.NewDecoder(r.Body).Decode(&todoItem); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if id != todoItem.Id {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		updatedTodo, err := svc.UpdateTodo(ctx, id, todoItem)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		resp, err := json.Marshal(updatedTodo)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		if _, err = w.Write(resp); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (s *Server) HandleSearchTodos(svc *todo.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		queryValues := r.URL.Query()
		query := queryValues.Get("q")
		filteredTodos, err := svc.FilterTodos(r.Context(), query)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		resp, err := json.Marshal(filteredTodos)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		if _, err = w.Write(resp); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
