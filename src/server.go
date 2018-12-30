package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/todo")
	if err != nil {
		fmt.Print(err.Error())
	}
	defer db.Close()
	// make sure connection is available
	err = db.Ping()
	if err != nil {
		fmt.Print(err.Error())
	}
	type Todo struct {
		Id       int
		Title    string
		Detail   string
		Deadline string
		Status   string
	}
	router := gin.Default()

	// GET a todo detail
	router.GET("/todo/:id", func(c *gin.Context) {
		var (
			todo   Todo
			result gin.H
		)
		id := c.Param("id")
		row := db.QueryRow("select id, title, detail, deadline, status from todo where id = ?;", id)
		err = row.Scan(&todo.Id, &todo.Title, &todo.Detail, &todo.Deadline, &todo.Status)
		if err != nil {
			// If no results send null
			result = gin.H{
				"result": nil,
				"count":  0,
			}
		} else {
			result = gin.H{
				"result": todo,
				"count":  1,
			}
		}
		c.JSON(http.StatusOK, result)
	})

	// GET all todos
	router.GET("/todos", func(c *gin.Context) {
		var (
			todo  Todo
			todos []Todo
		)
		rows, err := db.Query("select id, title, detail, deadline, status from todo where status = '0';")
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
		fmt.Println("todolist.... ", len(todos))

		c.JSON(http.StatusOK, gin.H{
			"result": todos,
			"count":  len(todos),
		})
	})

	// POST new todo details
	router.POST("/todo", func(c *gin.Context) {
		var buffer bytes.Buffer
		title := c.PostForm("title")
		detail := c.PostForm("detail")
		deadline := c.PostForm("deadline")
		status := c.PostForm("status")
		stmt, err := db.Prepare("insert into todo (title, detail, deadline, status) values(?,?,?,?);")
		if err != nil {
			fmt.Print(err.Error())
		}
		_, err = stmt.Exec(title, detail, deadline, status)

		if err != nil {
			fmt.Print(err.Error())
		}

		// Fastest way to append strings
		buffer.WriteString(title)
		buffer.WriteString(" ")
		buffer.WriteString(detail)
		buffer.WriteString(" ")
		buffer.WriteString(deadline)
		buffer.WriteString(" ")
		buffer.WriteString(status)
		defer stmt.Close()
		name := buffer.String()
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf(" %s successfully created", name),
		})
	})

	// PUT - update a todo details
	router.PUT("/todo", func(c *gin.Context) {
		// id := c.Query("id")
		id := c.PostForm("id")
		title := c.PostForm("title")
		detail := c.PostForm("detail")
		deadline := c.PostForm("deadline")
		status := c.PostForm("status")
		fmt.Printf("id: ", id)
		stmt, err := db.Prepare("update todo set title=?, detail=?, deadline=?, status=? where id=?;")
		if err != nil {
			fmt.Print(err.Error())
		}
		_, err = stmt.Exec(title, detail, deadline, status, id)
		if err != nil {
			fmt.Print(err.Error())
		}

		// Fastest way to append strings
		var buffer bytes.Buffer
		buffer.WriteString(id)
		buffer.WriteString(" ")
		buffer.WriteString(title)
		buffer.WriteString(" ")
		buffer.WriteString(detail)
		buffer.WriteString(" ")
		buffer.WriteString(deadline)
		buffer.WriteString(" ")
		buffer.WriteString(status)
		defer stmt.Close()
		name := buffer.String()
		fmt.Printf("name: ", name)
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Successfully updated todo %s", name),
		})
	})

	// Delete resources
	router.DELETE("/todo", func(c *gin.Context) {
		id := c.Query("id")
		stmt, err := db.Prepare("delete from todo where id=?;")
		if err != nil {
			fmt.Print(err.Error())
		}
		_, err = stmt.Exec(id)
		if err != nil {
			fmt.Print(err.Error())
		}
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Successfully deleted todo: %s", id),
		})
	})
	router.Run(":3000")
}
