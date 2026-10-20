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
	
	authGroupController := controllers.Auth{DB: db}

	{
		authGroup.POST("/signup", authGroupController.Signup)
		authGroup.POST("/sign-in", middleware.Authenticate().LoginHandler)
		authGroup.GET("/profile", authenticate, authGroupController.GetProfile)
	}

	articleController := controllers.Articles{DB: db}
	{
		articlesGroup.GET("", articleController.FindAll)
		articlesGroup.GET("/:id", articleController.FindOne)
		articlesGroup.POST("", articleController.Create)
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
