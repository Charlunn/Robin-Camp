package main

import (
	"go_tutorial/db"
	"go_tutorial/handler"
	"go_tutorial/repository"
	"go_tutorial/service"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// =================================================================
	// 1. 初始化 (Initialization)
	// =================================================================

	// a. 初始化数据库连接
	// 我们现在只调用 db 包中的函数，main 函数不再关心具体的连接细节。
	dbConn, err := db.NewConnection()
	if err != nil {
		log.Fatalf("Could not initialize database connection: %s", err)
	}
	// 确保在程序退出时关闭数据库连接
	defer dbConn.Close()

	// =================================================================
	// 2. 依赖注入 (Dependency Injection)
	// =================================================================

	// b. 初始化 Repository
	userRepo := repository.NewMySQLUserRepository(dbConn)

	// c. 初始化 Service
	userService := service.NewUserService(userRepo)

	// d. 初始化 Handler
	userHandler := handler.NewUserHandler(userService)

	log.Println("Dependencies injected.")

	// =================================================================
	// 3. 设置路由并启动服务器 (Routing & Server Start)
	// =================================================================

	r := gin.Default()

	api := r.Group("/api/v1")
	{
		api.POST("/users", userHandler.CreateUser)
		api.GET("/users/:id", userHandler.GetUser)
	}

	log.Println("Starting server on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
