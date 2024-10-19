package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
)

type Task struct {
	Rowid       int
	Title       string
	Description string
}

type TasksRepository struct {
	DB *sql.DB
}

func (repo TasksRepository) getById(rowid int) Task {
	rows, err := repo.DB.Query(
		fmt.Sprintf(
			"SELECT title, description FROM tasks Where rowid=%d;",
			rowid,
		),
	)

	if err != nil {
		panic(err)
	}

	defer rows.Close()

	for rows.Next() {
		var (
			rowid       int
			title       string
			description string
		)
		rows.Scan(&rowid, &title, &description)

		rows.Close()
		return Task{
			Title:       title,
			Description: description,
		}
	}

	panic(fmt.Sprintf("Task with id %d not fonud", rowid))
}

func (repo TasksRepository) add(title string, desc string) (Task, error) {
	result, err := repo.DB.Exec(
		fmt.Sprintf(
			`INSERT INTO tasks (title, description) VALUES ("%s", "%s")`,
			title, desc,
		),
	)
	if err != nil {
		return Task{}, err
	}

	insertedId, err := result.LastInsertId()

	if err != nil {
		return Task{}, err
	}

	return repo.getById(int(insertedId)), nil
}

func (repo TasksRepository) deleteTaskById(taskId int) error {
	result, err := repo.DB.Exec(
		fmt.Sprintf(`DELETE FROM tasks WHERE rowid=%d`, taskId),
	)

	if err != nil {
		return err
	}

	affectedId, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affectedId == 0 {
		return errors.New("Not task effected! Trying to delete task with id " + strconv.Itoa(taskId))
	}

	return nil
}

func (repo TasksRepository) fetchAll() []Task {
	rows, err := repo.DB.Query("SELECT rowid, title, description FROM tasks;")

	if err != nil {
		panic(err)
	}

	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var (
			rowid       int
			title       string
			description string
		)
		rows.Scan(&rowid, &title, &description)

		tasks = append(
			tasks,
			Task{
				Rowid:       rowid,
				Title:       title,
				Description: description,
			},
		)
	}

	return tasks
}
