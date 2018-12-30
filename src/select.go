package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/todo")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer db.Close()
	// make sure connection is available
	err = db.Ping()
	if err != nil {
		fmt.Println(err.Error())
	}

	type Todo struct {
		Id       int
		Title    string
		Detail   string
		Deadline string
		Status   string
	}

	var (
		todo  Todo
		todos []Todo
	)
	rows, err := db.Query("select id, title, detail, deadline, status from todo;")
	if err != nil {
		fmt.Print(err.Error())
	}
	for rows.Next() {
		err = rows.Scan(&todo.Id, &todo.Title, &todo.Detail, &todo.Deadline, &todo.Status)
		todos = append(todos, todo)
		if err != nil {
			fmt.Print(err.Error())
		}
	}
	defer rows.Close()

	fmt.Println("Person Table successfully migrated.... ", len(todos))

}
