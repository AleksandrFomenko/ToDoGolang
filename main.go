package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"qwe/config"
	"qwe/db"
	sqlite "qwe/db/sqlLite"
	"qwe/handlers"
)

func main() {
	cfg := config.LoadConfig()
	sqliteWorker := &sqlite.SQLiteWorker{}
	storage := db.New(sqliteWorker)
	err := storage.InitializeDB(cfg.DBfile)
	if err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %v\n", err)
	}

	r := chi.NewRouter()
	webDir := "./web"

	r.Handle("/*", http.StripPrefix("/", http.FileServer(http.Dir(webDir))))
	r.Get("/api/nextdate", handlers.NextDateHandler)
	r.Post("/api/task", func(w http.ResponseWriter, r *http.Request) {
		handlers.AddTaskHandler(w, r, sqliteWorker)
	})
	r.Put("/api/task", func(w http.ResponseWriter, r *http.Request) {
		handlers.UpdateTaskHandler(w, r, sqliteWorker)
	})
	r.Get("/api/task", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetTaskHandler(w, r, sqliteWorker)
	})
	r.Get("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetTasksHandler(w, r, sqliteWorker)
	})
	r.Post("/api/task/done", func(w http.ResponseWriter, r *http.Request) {
		handlers.TaskDoneHandler(w, r, sqliteWorker)
	})
	r.Delete("/api/task", func(w http.ResponseWriter, r *http.Request) {
		handlers.DeleteTaskHandler(w, r, sqliteWorker)
	})

	log.Printf("Запуск сервера на порту %s...\n", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal(err)
	}
}
