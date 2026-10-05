package ghost

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type Repo[T any] interface {
	Create(r T) (int64, error)
	Get(id int64) (T, error)
	Delete(id int64) error
	List() ([]T, error)
	Find(predicate func(T) bool) (T, error)
	FindMany(predicate func(T) bool) ([]T, error)
}

type CreateRes struct {
	Id int64 `json:"id"`
}

type Api[T interface{ ToD() D }, D interface{ ToT() T }] struct {
	path       string
	repo       Repo[D]
	createFrom func(T, *http.Request) (D, error)
	listFrom   func(*http.Request) ([]T, error)
	deleteFrom func(int64, *http.Request) error
	middleware func(handler http.HandlerFunc) http.HandlerFunc
}

func NewApi[T interface{ ToD() D }, D interface{ ToT() T }](path string, repo Repo[D]) *Api[T, D] {
	return &Api[T, D]{
		path: path,
		repo: repo,
	}
}

func (a *Api[T, D]) CreateFrom(fn func(T, *http.Request) (D, error)) {
	a.createFrom = fn
}

func (a *Api[T, D]) ListFrom(fn func(*http.Request) ([]T, error)) {
	a.listFrom = fn
}

func (a *Api[T, D]) DeleteFrom(fn func(int64, *http.Request) error) {
	a.deleteFrom = fn
}

func (a *Api[T, D]) SetMiddleware(fn func(handler http.HandlerFunc) http.HandlerFunc) {
	a.middleware = fn
}

func AddApiToMux[T interface{ ToD() D }, D interface{ ToT() T }](mux *http.ServeMux, api *Api[T, D]) {
	api_post := func(w http.ResponseWriter, r *http.Request) {
		var req T
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		var data D
		if api.createFrom != nil {
			data, err = api.createFrom(req, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		} else {
			data = req.ToD()
		}

		id, err := api.repo.Create(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		res := CreateRes{id}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
	}
	api_get := func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
		f, err := api.repo.Get(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(f.ToT())
	}
	api_delete := func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
		if api.deleteFrom != nil {
			err = api.deleteFrom(id, r)
		} else {
			err = api.repo.Delete(id)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
	api_list := func(w http.ResponseWriter, r *http.Request) {
		var l []T
		var err error
		if api.listFrom != nil {
			l, err = api.listFrom(r)
		} else {
			dl, err := api.repo.List()
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			l = make([]T, len(dl))
			for i, v := range dl {
				l[i] = v.ToT()
			}
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(l)
	}
	if api.middleware != nil {
		api_post = api.middleware(api_post)
		api_get = api.middleware(api_get)
		api_delete = api.middleware(api_delete)
		api_list = api.middleware(api_list)
	}
	mux.HandleFunc(fmt.Sprintf("POST /%s", api.path), api_post)
	mux.HandleFunc(fmt.Sprintf("GET /%s/{id}", api.path), api_get)
	mux.HandleFunc(fmt.Sprintf("DELETE /%s/{id}", api.path), api_delete)
	mux.HandleFunc(fmt.Sprintf("GET /%s", api.path), api_list)
}
