package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

var (
	tasks   = []Task{}
	counter = 1
	mutex   sync.Mutex
)

func main() {
	http.HandleFunc("/tasks", handleTasks)
	http.HandleFunc("/tasks/", handleTaskByID)
	log.Println("API running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		json.NewEncoder(w).Encode(tasks)
	case http.MethodPost:
		var t Task
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}
		mutex.Lock()
		t.ID = counter
		counter++
		tasks = append(tasks, t)
		mutex.Unlock()
		json.NewEncoder(w).Encode(t)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleTaskByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := strings.TrimPrefix(r.URL.Path, "/tasks/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	for i, t := range tasks {
		if t.ID == id {
			switch r.Method {
			case http.MethodPut:
				var updated Task
				if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
					http.Error(w, "Invalid input", http.StatusBadRequest)
					return
				}
				tasks[i].Title = updated.Title
				tasks[i].Done = updated.Done
				json.NewEncoder(w).Encode(tasks[i])
				return
			case http.MethodDelete:
				tasks = append(tasks[:i], tasks[i+1:]...)
				w.WriteHeader(http.StatusNoContent)
				return
			case http.MethodGet:
				json.NewEncoder(w).Encode(t)
				return
			}
		}
	}
	http.Error(w, "Task not found", http.StatusNotFound)
}
