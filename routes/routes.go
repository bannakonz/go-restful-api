package routes

import (
	"go-restful-api/config"
	"go-restful-api/controllers"
	"go-restful-api/middleware"

	"github.com/gin-gonic/gin"
)

func Serve(r *gin.Engine) {
	db := config.GetDB()
	v1 := r.Group("/api/v1")

	// middleware ของตัว gin , อ่านค่า jwt จาก header แล้วเอา jwt มาดึงในส่วน payload หาค่าของ sub แล้ว -> หา user -> แล้วคืน user กลับออกมาในส่วนของ context
	authenticate := middleware.Authenticate().MiddlewareFunc()

	authGroup := v1.Group("auth")
	articlesGroup := v1.Group("articles")

	authController := controllers.Auth{DB: db}

	{
		authGroup.POST("/signup", authController.Signup)
		authGroup.POST("/sign-in", middleware.Authenticate().LoginHandler)
		authGroup.GET("/profile", authenticate, authController.GetProfile)
		authGroup.PATCH("/profile", authenticate, authController.UpdateProfile)
	}

	userGroup := v1.Group("users")
	userController := controllers.Users{DB: db}

	{
		userGroup.GET("", userController.FindAll)
		userGroup.POST("", userController.Create)
		userGroup.GET("/:id", userController.FindOne)
		userGroup.PATCH("/:id", userController.Update)
		userGroup.DELETE("/:id", userController.Delete)
		userGroup.PATCH("/:id/promote", userController.Promote)
		userGroup.PATCH("/:id/demote", userController.Demote)
	}

	articleController := controllers.Articles{DB: db}
	{
		articlesGroup.GET("", articleController.FindAll)
		articlesGroup.GET("/:id", articleController.FindOne)
		articlesGroup.POST("", authenticate, articleController.Create)
		articlesGroup.PATCH("/:id", articleController.Update)
		articlesGroup.DELETE("/:id", articleController.Delete)
	}

	categoriesGroup := v1.Group("categories")
	categoryController := controllers.Categories{DB: db}
	{
		categoriesGroup.GET("", categoryController.FindAll)
		categoriesGroup.GET("/:id", categoryController.FindOne)
		categoriesGroup.POST("", categoryController.Create)
		categoriesGroup.PATCH("/:id", categoryController.Update)
		categoriesGroup.DELETE("/:id", categoryController.Delete)
	}

}
