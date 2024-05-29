package handlers

import (
	"encoding/json"
	"net/http"
	sqlite "qwe/db/sqlLite"
	"qwe/models"
	"sort"
	"strconv"
	"time"

	"qwe/services"
)

// omitempty - если поле пустое, оно будет опущено из сериализации
type AddTaskRequest struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat,omitempty"`
}

type AddTaskResponse struct {
	ID    int    `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")

	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		http.Error(w, "некоректный формат даты", http.StatusBadRequest)
		return
	}

	nextDate, err := services.NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Write([]byte(nextDate))
}

func AddTaskHandler(w http.ResponseWriter, r *http.Request, db *sqlite.SQLiteWorker) {
	var req AddTaskRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondWithError(w, "Ошибка десериализации JSON")
		return
	}

	if req.Title == "" {
		respondWithError(w, "Не указан заголовок задачи")
		return
	}

	now := time.Now()
	var taskDate time.Time

	if req.Date == "" {
		taskDate = now
	} else {
		taskDate, err = time.Parse("20060102", req.Date)
		if err != nil {
			respondWithError(w, "Дата представлена в неверном формате")
			return
		}
	}

	if taskDate.Before(now) || taskDate.Equal(now) {
		if req.Repeat != "" {
			nextDateStr, err := services.NextDate(now, req.Date, req.Repeat)
			if err != nil {
				respondWithError(w, "Неправильное правило повторения")
				return
			}
			nextDate, _ := time.Parse("20060102", nextDateStr)
			if nextDate.Before(now.AddDate(0, 0, 1)) {
				taskDate = now
			} else {
				taskDate = nextDate
			}
		} else {
			taskDate = now
		}
	}

	task := models.Task{
		Date:    taskDate,
		Title:   req.Title,
		Comment: req.Comment,
		Repeat:  req.Repeat,
	}

	err = db.AddTask(&task)
	if err != nil {
		respondWithError(w, err.Error())
		return
	}

	taskID, err := db.GetLastInsertId()
	if err != nil {
		respondWithError(w, err.Error())
		return
	}

	respondWithJSON(w, AddTaskResponse{ID: taskID})
}

func respondWithJSON(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if payload == nil {
		payload = map[string]interface{}{}
	}
	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, error string) {
	respondWithJSON(w, map[string]string{"error": error})
}

type TaskResponse struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type TasksResponse struct {
	Tasks []TaskResponse `json:"tasks"`
	Error string         `json:"error,omitempty"`
}

func GetTasksHandler(w http.ResponseWriter, r *http.Request, db *sqlite.SQLiteWorker) {
	tasks, err := db.GetTasks()
	if err != nil {
		respondWithJSON(w, TasksResponse{Error: "Ошибка получения задач"})
		return
	}

	if tasks == nil {
		respondWithJSON(w, TasksResponse{Tasks: []TaskResponse{}})
		return
	}

	var taskResponses []TaskResponse
	for _, task := range tasks {
		taskResponses = append(taskResponses, TaskResponse{
			ID:      strconv.Itoa(task.ID),
			Date:    task.Date.Format("20060102"),
			Title:   task.Title,
			Comment: task.Comment,
			Repeat:  task.Repeat,
		})
	}

	sort.Slice(taskResponses, func(i, j int) bool {
		return taskResponses[i].Date < taskResponses[j].Date
	})

	respondWithJSON(w, TasksResponse{Tasks: taskResponses})
}

func GetTaskHandler(w http.ResponseWriter, r *http.Request, db *sqlite.SQLiteWorker) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		respondWithJSON(w, AddTaskResponse{Error: "Не указан идентификатор"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithJSON(w, AddTaskResponse{Error: "Неверный формат идентификатора"})
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		respondWithJSON(w, AddTaskResponse{Error: "Ошибка при получении задачи"})
		return
	}

	if task == nil {
		respondWithJSON(w, AddTaskResponse{Error: "Задача не найдена"})
		return
	}

	respondWithJSON(w, TaskResponse{
		ID:      strconv.Itoa(task.ID),
		Date:    task.Date.Format("20060102"),
		Title:   task.Title,
		Comment: task.Comment,
		Repeat:  task.Repeat,
	})
}

func UpdateTaskHandler(w http.ResponseWriter, r *http.Request, db *sqlite.SQLiteWorker) {
	var req AddTaskRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondWithError(w, "Ошибка десерилизации JSON")
		return
	}

	id, err := strconv.Atoi(req.ID)
	if err != nil || id == 0 {
		respondWithError(w, "Не укзан идентификатор задачи")
		return
	}

	if req.Title == "" {
		respondWithError(w, "Не указан заголовок задачи")
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		respondWithError(w, "Ошибка при получении задачи")
		return
	}
	if task == nil {
		respondWithError(w, "Задача не найдена")
		return
	}

	now := time.Now()
	var taskDate time.Time

	if req.Date == "" {
		taskDate = now
	} else {
		taskDate, err = time.Parse("20060102", req.Date)
		if err != nil {
			respondWithError(w, "Дата представлена в неверном формате")
			return
		}
	}

	if taskDate.Before(now) || taskDate.Equal(now) {
		if req.Repeat != "" {
			nextDateStr, err := services.NextDate(now, req.Date, req.Repeat)
			if err != nil {
				respondWithError(w, "Неправильное правило повторения")
				return
			}
			nextDate, _ := time.Parse("20060102", nextDateStr)
			if nextDate.Before(now.AddDate(0, 0, 1)) {
				taskDate = now
			} else {
				taskDate = nextDate
			}
		} else {
			taskDate = now
		}
	}

	task = &models.Task{
		ID:      id,
		Date:    taskDate,
		Title:   req.Title,
		Comment: req.Comment,
		Repeat:  req.Repeat,
	}

	err = db.UpdateTask(task)
	if err != nil {
		respondWithError(w, "Ошибка при обновлении задачи")
		return
	}

	respondWithJSON(w, struct{}{})
}
func TaskDoneHandler(w http.ResponseWriter, r *http.Request, db *sqlite.SQLiteWorker) {

	idStr := r.URL.Query().Get("id")
	if idStr == "" || idStr == "<nil>" {
		respondWithJSON(w, map[string]string{"error": "Некорректный идентификатор задачи"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithJSON(w, map[string]string{"error": "Некорректный идентификатор задачи"})
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		respondWithJSON(w, map[string]string{"error": "Задача не найдена"})
		return
	}

	if task == nil {
		respondWithJSON(w, map[string]string{"error": "Задача не найдена"})
		return
	}

	if task.Repeat == "" {
		err = db.DeleteTask(id)
		if err != nil {
			respondWithJSON(w, map[string]string{"error": "Ошибка при удалении задачи"})
			return
		}
		respondWithJSON(w, struct{}{})
		return
	}
	nextDateStr, err := services.NextDate(task.Date, task.Date.Format("20060102"), task.Repeat)
	if err != nil {
		respondWithJSON(w, map[string]string{"error": "Ошибка при вычислении следующей даты"})
		return
	}
	nextDate, err := time.Parse("20060102", nextDateStr)
	if err != nil {
		respondWithJSON(w, map[string]string{"error": "Ошибка при парсинге следующей даты"})
		return
	}

	task.Date = nextDate
	err = db.UpdateTask(task)
	if err != nil {
		respondWithJSON(w, map[string]string{"error": "Ошибка при обновлении задачи"})
		return
	}
	respondWithJSON(w, struct{}{})
}
func DeleteTaskHandler(w http.ResponseWriter, r *http.Request, db *sqlite.SQLiteWorker) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" || idStr == "<nil>" {
		respondWithError(w, "Некорректный идентификатор задачи")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, "Некорректный идентификатор задачи")
		return
	}
	err = db.DeleteTask(id)
	if err != nil {
		respondWithError(w, "Ошибка при удалении задачи")
		return
	}
	respondWithJSON(w, struct{}{}) // Передаем nil для отправки пустого JSON объекта {}
}
