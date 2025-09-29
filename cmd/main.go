package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/yatiac/go-todo-api/models"
)

var userCache = make(map[int]models.User)

var cacheMutex sync.RWMutex

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	mux.HandleFunc("POST /users", createUser)
	mux.HandleFunc("GET /users/{id}", findUserById)
	mux.HandleFunc("DELETE /users/{id}", deleteUser)

	fmt.Println("Starting server at port 5050")
	http.ListenAndServe(":5050", mux)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "My Todo app")
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if user.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	userCache[len(userCache)+1] = user
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func findUserById(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	cacheMutex.RLock()
	user, ok := userCache[id]
	cacheMutex.RUnlock()

	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	j, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusFound)
	w.Write(j)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	_, ok := userCache[id]
	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	delete(userCache, id)
	w.WriteHeader(http.StatusNoContent)
}
