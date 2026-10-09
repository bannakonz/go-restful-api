package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Serve(r *gin.Engine) {
	r.GET("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello Thailand")
	})
}
