package routes

import (
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

type article struct {
	ID    uint   `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
	Image string `json:"image"`
}

type createArticleForm struct {
	Title string                `form:"title" binding:"required"`
	Body  string                `form:"body" binding:"required"`
	Image *multipart.FileHeader `form:"image" binding:"required"`
}

func Serve(r *gin.Engine) {
	articles := []article{
		{ID: 1, Title: "Title#1", Body: "Body#1"},
		{ID: 2, Title: "Title#2", Body: "Body#2"},
		{ID: 3, Title: "Title#3", Body: "Body#3"},
		{ID: 4, Title: "Title#4", Body: "Body#4"},
		{ID: 5, Title: "Title#5", Body: "Body#5"},
	}
	articlesGroup := r.Group("/api/v1/articles")

	articlesGroup.GET("", func(ctx *gin.Context) {
		result := articles
		limit := ctx.Query("limit")
		if limit != "" {
			n, _ := strconv.Atoi(limit)
			result = result[:n]
		}

		ctx.JSON(http.StatusOK, gin.H{"articles": result})
	})

	articlesGroup.GET("/:id", func(ctx *gin.Context) {
		id, _ := strconv.Atoi(ctx.Param("id"))
		for _, item := range articles {
			if item.ID == uint(id) {
				ctx.JSON(http.StatusOK, gin.H{"articles": item})
				return
			}
		}
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Articles Not Found"})
	})

	articlesGroup.POST("", func(ctx *gin.Context) {
		var form createArticleForm
		err := ctx.ShouldBind(&form)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		a := article{
			ID:    uint(len(articles)) + 1,
			Title: form.Title,
			Body:  form.Body,
		}

		// Get file
		file, _ := ctx.FormFile("image")

		// Create Path
		// ID=8 -> uploads/articles/8/image.png
		path := "uploads/articles/" + strconv.Itoa(int(a.ID)) // สร้าง path
		os.MkdirAll(path, 0755)                               // สร้าง folder เช่น folder ที่ 8

		// Upload file
		filename := path + "/" + file.Filename
		if err := ctx.SaveUploadedFile(file, filename); err != nil { // dst(destination คือ ไฟล์ที่เราจะสร้างขึ้นมา คือการเอา path ต่อกับ file name

		}

		// Attach File to article
		a.Image = os.Getenv("HOST") + "/" + filename
		articles = append(articles, a)

		ctx.JSON(http.StatusCreated, gin.H{"articles": a})

	})
}
