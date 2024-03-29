package main

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/valyala/fasthttp"
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type UserRepository struct {
	users sync.Map
}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) Create(user User) {
	r.users.Store(user.ID, user)
}

func (r *UserRepository) Read(id string) (User, bool) {
	user, ok := r.users.Load(id)
	if !ok {
		return User{}, false
	}
	return user.(User), true
}

func (r *UserRepository) Update(id string, newUser User) {
	r.users.Store(id, newUser)
}

func (r *UserRepository) Delete(id string) {
	r.users.Delete(id)
}

func (r *UserRepository) List() []User {
	var userList []User
	r.users.Range(func(key, value interface{}) bool {
		userList = append(userList, value.(User))
		return true
	})
	return userList
}

func CreateUser(ctx *fasthttp.RequestCtx, userRepository *UserRepository) {
	var user User
	if err := json.Unmarshal(ctx.PostBody(), &user); err != nil {
		ctx.Error("Invalid request body", fasthttp.StatusBadRequest)
		return
	}
	userRepository.Create(user)
	ctx.SetStatusCode(fasthttp.StatusCreated)
	ctx.SetBody([]byte("User created successfully"))
}

func GetUser(ctx *fasthttp.RequestCtx, userRepository *UserRepository) {
	id := string(ctx.QueryArgs().Peek("id"))
	if id == "" {
		ctx.Error("User ID not provided", fasthttp.StatusBadRequest)
		return
	}
	user, found := userRepository.Read(id)
	if !found {
		ctx.Error("User not found", fasthttp.StatusNotFound)
		return
	}
	jsonResponse, _ := json.Marshal(user)
	ctx.SetContentType("application/json")
	ctx.SetBody(jsonResponse)
}

func UpdateUser(ctx *fasthttp.RequestCtx, userRepository *UserRepository) {
	id := string(ctx.QueryArgs().Peek("id"))
	var newUser User
	if err := json.Unmarshal(ctx.PostBody(), &newUser); err != nil {
		ctx.Error("Invalid request body", fasthttp.StatusBadRequest)
		return
	}
	userRepository.Update(id, newUser)
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody([]byte("User updated successfully"))
}

func DeleteUser(ctx *fasthttp.RequestCtx, userRepository *UserRepository) {
	id := string(ctx.QueryArgs().Peek("id"))
	userRepository.Delete(id)
	ctx.SetStatusCode(fasthttp.StatusNoContent)
	ctx.SetBody([]byte("User deleted successfully"))
}

func ListUsers(ctx *fasthttp.RequestCtx, userRepository *UserRepository) {
	users := userRepository.List()
	jsonResponse, _ := json.Marshal(users)
	ctx.SetContentType("application/json")
	ctx.SetBody(jsonResponse)
}

func main() {
	userRepository := NewUserRepository()

	requestHandler := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/create":
			switch string(ctx.Method()) {
			case fasthttp.MethodPost:
				CreateUser(ctx, userRepository)
			default:
				ctx.Error("Method not allowed", fasthttp.StatusMethodNotAllowed)
			}
		case "/get":
			switch string(ctx.Method()) {
			case fasthttp.MethodGet:
				GetUser(ctx, userRepository)
			default:
				ctx.Error("Method not allowed", fasthttp.StatusMethodNotAllowed)
			}
		case "/update":
			switch string(ctx.Method()) {
			case fasthttp.MethodPut:
				UpdateUser(ctx, userRepository)
			default:
				ctx.Error("Method not allowed", fasthttp.StatusMethodNotAllowed)
			}
		case "/delete":
			switch string(ctx.Method()) {
			case fasthttp.MethodDelete:
				DeleteUser(ctx, userRepository)
			default:
				ctx.Error("Method not allowed", fasthttp.StatusMethodNotAllowed)
			}
		case "/list":
			switch string(ctx.Method()) {
			case fasthttp.MethodGet:
				ListUsers(ctx, userRepository)
			default:
				ctx.Error("Method not allowed", fasthttp.StatusMethodNotAllowed)
			}
		default:
			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		}
	}

	server := fasthttp.Server{
		Handler: requestHandler,
	}

	log.Fatal(server.ListenAndServe(":8080"))
}
