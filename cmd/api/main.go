package main

import (
	"log"
	"net/http"
	"os"

	"github.com/rest-from-zero/internal/database"
	"github.com/rest-from-zero/internal/handlers"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://taskuser:taskpass@localhost:5432/tasksdb"
	}
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	log.Printf("Running server on port %s", serverPort)

	db, err := database.Connect(databaseURL)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	log.Println("Connected to DB successfully")

	taskStore := database.NewTaskStore(db)
	handlers := handlers.NewHandler(taskStore)

	mux := http.NewServeMux()

	mux.HandleFunc("/tasks", methodHandler(handlers.GetAllTasks, "GET"))
	mux.HandleFunc("/tasks/create", methodHandler(handlers.CreateTask, "POST"))
	mux.HandleFunc("/tasks/", taskIDHandler(handlers))

	loggdMux := logginMiddleware(mux)

	serverAddr := ":" + serverPort

	err = http.ListenAndServe(serverAddr, loggdMux)
	if err != nil {
		log.Fatal(err)
	}
}

func methodHandler(handlerFunc http.HandlerFunc, allowedMethod string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != allowedMethod {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
		handlerFunc(w, r)
	}
}

func taskIDHandler(handlers *handlers.Handlers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetTask(w, r)
		case http.MethodPut:
			handlers.UpdateTask(w, r)
		case http.MethodDelete:
			handlers.DeleteTask(w, r)
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	}
}

func logginMiddleware(next *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
