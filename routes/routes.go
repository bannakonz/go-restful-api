package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Serve(r *gin.Engine) {
	articlesGroup := r.Group("/api/v1/articles")

	articlesGroup.GET("", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello World")
	})

	articlesGroup.GET("/:id", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hi ID")
	})
}
