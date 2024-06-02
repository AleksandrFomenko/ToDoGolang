package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	sqlite "qwe/db/sqlLite"
	"qwe/models"
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
		respondWithError(w, http.StatusBadRequest, "некоректный формат даты")
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
		respondWithError(w, http.StatusBadRequest, "Ошибка десериализации JSON")
		return
	}

	if req.Title == "" {
		respondWithError(w, http.StatusBadRequest, "Не указан заголовок задачи")
		return
	}

	now := time.Now()
	var taskDate time.Time

	if req.Date == "" {
		req.Date = now.Format("20060102")
		taskDate = now
	} else {
		taskDate, err = time.Parse("20060102", req.Date)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Дата представлена в неверном формате")
			return
		}
	}

	if taskDate.Before(now) || taskDate.Equal(now) {
		taskDate = now

	}
	if req.Repeat != "" {
		if _, err := services.NextDate(now, req.Date, req.Repeat); err != nil {
			respondWithError(w, http.StatusBadRequest, "Неправильное правило повторения")
			return
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
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	taskID, err := db.GetLastInsertId()
	if err != nil {
		log.Printf("ошибка с получением последнего айди: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Ошибка сервера")
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

func respondWithError(w http.ResponseWriter, status int, error string) {
	w.WriteHeader(status)
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
		respondWithError(w, http.StatusBadRequest, "Ошибка десерилизации JSON")
		return
	}

	id, err := strconv.Atoi(req.ID)
	if err != nil || id == 0 {
		respondWithError(w, http.StatusBadRequest, "Не укзан идентификатор задачи")
		return
	}

	if req.Title == "" {
		respondWithError(w, http.StatusBadRequest, "Не указан заголовок задачи")
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Ошибка при получении задачи")
		return
	}
	if task == nil {
		respondWithError(w, http.StatusBadRequest, "Задача не найдена")
		return
	}

	now := time.Now()
	var taskDate time.Time

	if req.Date == "" {
		taskDate = now
	} else {
		taskDate, err = time.Parse("20060102", req.Date)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Дата представлена в неверном формате")
			return
		}
	}

	if taskDate.Before(now) || taskDate.Equal(now) {
		if req.Repeat != "" {
			nextDateStr, err := services.NextDate(now, req.Date, req.Repeat)
			if err != nil {
				respondWithError(w, http.StatusBadRequest, "Неправильное правило повторения")
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
		respondWithError(w, http.StatusBadRequest, "Ошибка при обновлении задачи")
		return
	}

	respondWithJSON(w, struct{}{})
}
func TaskDoneHandler(w http.ResponseWriter, r *http.Request, db *sqlite.SQLiteWorker) {

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
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
	nextDateStr, err := services.NextDate(time.Now(), task.Date.Format("20060102"), task.Repeat)

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
	if r.Method != http.MethodDelete {
		respondWithError(w, http.StatusBadRequest, "Неподдерживаемый метод")
		return
	}
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		respondWithError(w, http.StatusBadRequest, "Некорректный идентификатор задачи")
		return
	}

	id, err := strconv.Atoi(idStr)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "не смог преобразовать id в инт")
		return
	}
	err = db.DeleteTask(id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Ошибка при удалении задачи")
		return
	}
	respondWithJSON(w, struct{}{})
}
