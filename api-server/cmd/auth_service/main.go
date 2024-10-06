package main

import (
	"matching-service/api-server/internal/handlers"
	"matching-service/api-server/internal/middleware"
	"matching-service/api-server/internal/repository"
	"matching-service/api-server/internal/services"
	"matching-service/api-server/pkg/database"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag/example/basic/docs"
)

// @title          Matching Service API
// @version        1.0
// @description    This is a sample server for a matching service.
// @termsOfService http://swagger.io/terms/

// @contact.name  API Support
// @contact.url   http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url  http://www.apache.org/licenses/LICENSE-2.0.html

// @host     localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	database.InitDB()

	db := database.GetDB()

	userRepo := repository.NewUserRepo(db)
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

	r := gin.Default()

	// Swagger documentation route
	docs.SwaggerInfo.BasePath = "/api/v1"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/api/v1")

	{
		v1.POST("/register", userHandler.Register)
		v1.POST("/login", userHandler.Login)

		authorized := v1.Group("/")
		authorized.Use(middleware.AuthMiddleware())
		{
			authorized.GET("/profile", userHandler.GetProfile)
		}
	}

	r.Run(":8080")
}
