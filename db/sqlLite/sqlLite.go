package sqlite

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"qwe/models"
)

type SQLiteWorker struct {
	DB *sql.DB
}

func (s *SQLiteWorker) InitializeDB(dbFile string) error {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return err
	}

	s.DB = db

	query := `
    CREATE TABLE IF NOT EXISTS scheduler (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        date TEXT,
        title TEXT,
        comment TEXT,
        repeat TEXT
    );`
	_, err = s.DB.Exec(query)
	if err != nil {
		return err
	}

	log.Println("База данных успешно инициализирована")
	return nil
}

func (s *SQLiteWorker) AddTask(task *models.Task) error {
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	_, err := s.DB.Exec(query, task.Date.Format("20060102"), task.Title, task.Comment, task.Repeat)
	return fmt.Errorf("ошибка Add Task : %w", err)
}

func (s *SQLiteWorker) GetTasks() ([]models.Task, error) {
	query := `SELECT id, date, title, comment, repeat FROM scheduler`
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка GetTasks : %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		var dateStr string
		if err := rows.Scan(&task.ID, &dateStr, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, err
		}
		task.Date, _ = time.Parse("20060102", dateStr)
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка GetTasks : %w", err)
	}

	return tasks, nil
}

func (s *SQLiteWorker) DeleteTask(id int) error {
	query := `DELETE FROM scheduler WHERE id = ?`
	_, err := s.DB.Exec(query, id)
	return fmt.Errorf("DeleteTask : %w", err)
}

func (s *SQLiteWorker) GetTask(id int) (*models.Task, error) {
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
	row := s.DB.QueryRow(query, id)
	var task models.Task
	var dateStr string
	if err := row.Scan(&task.ID, &dateStr, &task.Title, &task.Comment, &task.Repeat); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка GetTask : %w", err)
	}
	var err error
	task.Date, err = time.Parse("20060102", dateStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка GetTask: %w", err)
	}

	return &task, nil
}

func (s *SQLiteWorker) UpdateTask(task *models.Task) error {
	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	_, err := s.DB.Exec(query, task.Date.Format("20060102"), task.Title, task.Comment, task.Repeat, task.ID)
	return fmt.Errorf("ошибка UpdateTask: %w", err)
}

func (s *SQLiteWorker) CompleteTask(id int) error {
	query := `UPDATE scheduler SET repeat = '' WHERE id = ?`
	_, err := s.DB.Exec(query, id)
	return fmt.Errorf("ошибка CompleteTask: %w", err)
}

func (s *SQLiteWorker) GetLastInsertId() (int, error) {
	var id int
	err := s.DB.QueryRow("SELECT last_insert_rowid()").Scan(&id)
	return id, fmt.Errorf("ошибка GetLastInsertId: %w", err)
}
