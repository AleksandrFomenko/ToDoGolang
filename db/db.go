package db

import "qwe/models"

type worker interface {
	InitializeDB(dbFile string) error
	AddTask(task *models.Task) error
	GetTasks() ([]models.Task, error)
	DeleteTask(id int) error
	GetTask(id int) (*models.Task, error)
	UpdateTask(task *models.Task) error
	CompleteTask(id int) error
	GetLastInsertId() (int, error)
}

type db struct {
	worker
}

func New(work worker) db {
	return db{worker: work}
}
