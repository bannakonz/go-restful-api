package controllers

import (
	"go-restful-api/models"
	"mime/multipart"

	"github.com/gin-gonic/gin"
)

type Articles struct {
}

type createArticleForm struct {
	Title string                `form:"title" binding:"required"`
	Body  string                `form:"body" binding:"required"`
	Image *multipart.FileHeader `form:"image" binding:"required"`
}

var articles []models.Article = []models.Article{
	{ID: 1, Title: "Title#1", Body: "Body#1"},
	{ID: 2, Title: "Title#2", Body: "Body#2"},
	{ID: 3, Title: "Title#3", Body: "Body#3"},
	{ID: 4, Title: "Title#4", Body: "Body#4"},
	{ID: 5, Title: "Title#5", Body: "Body#5"},
}

func (a *Articles) FindAll(ctx *gin.Context) {

}

func (a *Articles) FindOne(ctx *gin.Context) { // Receiver ส่งแค่ memory address มา ทำให้ประสิทธิภาพดี
}

func (a *Articles) Create(ctx *gin.Context) {


}
