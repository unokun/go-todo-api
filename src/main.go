package main

import (
	"github.com/gin-gonic/gin"
	"github.com/unokun/go-todo-api/src/controller"
)

func main() {
	router := gin.Default()

    v1 := router.Group("/todo/api/v1")
    {
        v1.GET("/tasks", controller.TasksGET)
        v1.POST("/tasks", controller.TaskPOST)
        v1.PATCH("/tasks/:id", controller.TaskPATCH)
        v1.DELETE("/tasks/:id", controller.TaskDELETE)
    }
    // nginxのreverse proxy設定
	router.Run(":9000")
}
