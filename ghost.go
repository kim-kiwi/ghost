package ghost

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type CollectionOfResource[T any] interface {
	Create(r T) (int64, error)
	Get(id int64) (T, error)
	Delete(id int64) error
	List() ([]T, error)
}

type CreateRes struct {
	Id int64 `json:"id"`
}

type Middleware = func(http.HandlerFunc) http.HandlerFunc
type Check func(w http.ResponseWriter, r *http.Request) bool

func NewMiddleware(check Check) Middleware {
	return func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !check(w, r) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			handler(w, r)
		}
	}
}

func defaultMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return handler
}

func Collection[T any](mux *http.ServeMux, path string, collection CollectionOfResource[T]) {
	CollectionWithMiddleware(mux, path, collection, defaultMiddleware)
}

func CollectionWithMiddleware[T any](mux *http.ServeMux, path string, collection CollectionOfResource[T], middleware Middleware) {
	mux.HandleFunc(fmt.Sprintf("POST /%s", path), middleware(func(w http.ResponseWriter, r *http.Request) {
		var req T
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		id, create_err := collection.Create(req)
		if create_err != nil {
			fmt.Println(create_err)
			http.Error(w, "Failed to Create", http.StatusBadRequest)
			return
		}
		res := CreateRes{id}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
	}))
	mux.HandleFunc(fmt.Sprintf("GET /%s/{id}", path), middleware(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
		f, read_err := collection.Get(id)
		if read_err != nil {
			fmt.Println(read_err)
			http.Error(w, "Failed to Read", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(f)
	}))
	mux.HandleFunc(fmt.Sprintf("DELETE /%s/{id}", path), middleware(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
		delete_err := collection.Delete(id)
		if delete_err != nil {
			fmt.Println(delete_err)
			http.Error(w, "Failed to Read", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	mux.HandleFunc(fmt.Sprintf("GET /%s", path), middleware(func(w http.ResponseWriter, r *http.Request) {
		f, list_err := collection.List()
		if list_err != nil {
			fmt.Println(list_err)
			http.Error(w, "Failed to List", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(f)
	}))
	// if updater, ok := resrc.(Updater[T]); ok {
	// }
	// if deleter, ok := resrc.(Deleter); ok {
	// }
}

// GET /users -> []User
// GET /users?id=10 -> []User
// GET /users?created_before=2024 -> []User

// GET /users/123 -> User
// GET /users/456 -> User

// GET /status -> Status
// GET /status/ping -> Ping


// // 둘이 유사한데
// GET /users/123 -> User
// GET /ping/status -> { uptime: 102013120 }

// GET /resource/:something -> ???
// :something -> 이부분을 뭔가 구조체나 인터페이스로 사용자가 정의할 수 있어야함.

// /status: Status
//  - /ping: Ping

// /users: []User
//  - /{id}: User

// 이렇게?
// 괜찮은데... Go로 어떻게 표현하지?
// 일단 이름은 sub-resource?

// ```
// ghost.ResourceHandler()
// ghost.SubResourceHandler()
// ```
// 이건 안되는게 User의 경우에는 하나의 User리소스가 메인과 서브 리소스를 모두 관리함;;;
// 말 그대로 서브라서 하나의 독립적인 리소스가 되면 안 될 것 같은데... 그러면 리소스 정의하는 쪽에서 서브 리소스를 만들어야하나?

// 아닌가? 이거 될 것 같기도..?

// ```
// status_resrc := NewStatusResource()
// ghost.ResourceHandler("status", status_resrc)
// ghost.SubResourceHandler("status/ping", status_resrc.PingResource())
// ```

// 근데 이건 좀 구린데 뭔가뭔가...

// ```
// ghost.ResourceHandler("status", NewStatusResource())
// ```

// 그냥 이렇게 하면 알아서 서브까지 처리해주면 좋겠는데...

// ```
// type StatusResource struct {}

// func (StatusResource) Sub() map[string]ghost.Resource {}
// ```

// 이렇게 하고 재귀적으로 Sub호출??? 나쁘지 않긴 한데??? 캐싱이나 once.Do 같은걸로 하면 되나?
// 흠...

// 더 간단한 방법을 찾아야해

// GET /resource/:something -> ???
// :something -> 이부분을 뭔가 구조체나 인터페이스로 사용자가 정의할 수 있어야함.

// /status: Status
//  - /ping: Ping

// /users: []User
//  - /{id}: User

// 여기서 ":something을 커스텀" 이라는 파트를 잘 봐야해.
