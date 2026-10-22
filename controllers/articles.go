package controllers

import (
	"errors"
	"go-restful-api/models"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type Articles struct {
	DB *gorm.DB
}

type createArticleForm struct {
	Title      string                `form:"title" binding:"required"`
	Body       string                `form:"body" binding:"required"`
	Excerpt    string                `form:"excerpt" binding:"required"`
	CategoryID uint                  `form:"category_id" binding:"required"`
	Image      *multipart.FileHeader `form:"image" binding:"required"`
}

type updateArticleForm struct {
	Title      string                `form:"title"`
	Body       string                `form:"body"`
	Excerpt    string                `form:"excerpt"`
	CategoryID uint                  `form:"category_id"`
	Image      *multipart.FileHeader `form:"image"`
}

type articleResponse struct {
	ID         uint   `json:"id"`
	Title      string `json:"title"`
	Excerpt    string `json:"excerpt"`
	Body       string `json:"body"`
	Image      string `json:"image"`
	CategoryID uint   `json:"category_id"`
	Category   struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	} `json:"category"`
	User struct {
		Name   string `json:"name"`
		Avatar string `json:"avatar"`
	} `json:"user"`
}

type createOrUpdateResponse struct {
	ID         uint   `json:"id"`
	Title      string `json:"title"`
	Excerpt    string `json:"excerpt"`
	Body       string `json:"body"`
	Image      string `json:"image"`
	CategoryID uint   `json:"category_id"`
	UserID     uint   `json:"user_id"`
}

type articlesPaging struct {
	Items  []articleResponse `json:"items"`
	Paging *pagingResult     `json:"paging"`
}

func (a *Articles) FindAll(ctx *gin.Context) {
	var articles []models.Article

	query := a.DB.Preload("User").Preload("Category").Order("id DESC")

	categoryId := ctx.Query("categoryId")

	if categoryId != "" {
		query = query.Where("category_id = ?", categoryId)
	}

	term := ctx.Query("term")
	if term != "" {
		query = query.Where("title ILIKE ?", "%"+term+"%")
	}

	pagination := pagination{
		ctx:     ctx,
		query:   query,
		records: &articles,
	}

	paging := pagination.paginate()

	var serializedArticles []articleResponse
	copier.Copy(&serializedArticles, &articles)

	ctx.JSON(http.StatusOK, gin.H{
		"articles": articlesPaging{
			Items:  serializedArticles,
			Paging: paging,
		},
	})
}

func (a *Articles) FindOne(ctx *gin.Context) {
	article, err := a.FindArticleById(ctx)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	serializedArticle := articleResponse{}
	copier.Copy(&serializedArticle, article)
	ctx.JSON(http.StatusOK, gin.H{"article": serializedArticle})
}

func (a *Articles) Create(ctx *gin.Context) {
	var form createArticleForm
	if err := ctx.ShouldBind(&form); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	var article models.Article

	user, ok := ctx.Get("sub")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	currentUser, ok := user.(*models.User)
	if !ok || currentUser == nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
		return
	}
	copier.Copy(&article, &form)
	article.UserID = currentUser.ID

	if err := a.DB.Create(&article).Error; err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := a.setArticleImage(ctx, &article); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Load associations
	if err := a.DB.Preload("User").Preload("Category").
		First(&article, article.ID).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	serializedArticle := createOrUpdateResponse{}
	copier.Copy(&serializedArticle, &article)

	ctx.JSON(http.StatusCreated, gin.H{"articles": serializedArticle})
}

func (a *Articles) Update(ctx *gin.Context) {
	var form updateArticleForm

	if err := ctx.ShouldBind(&form); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	article, err := a.FindArticleById(ctx)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Update article fields
	updates := models.Article{
		Title:      form.Title,
		Body:       form.Body,
		Excerpt:    form.Excerpt,
		CategoryID: form.CategoryID,
	}

	if err := a.DB.Model(article).Updates(updates).Error; err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	// Update image only if provided
	if form.Image != nil {
		if err := a.setArticleImage(ctx, article); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	var serializedArticle createOrUpdateResponse
	copier.Copy(&serializedArticle, article)

	ctx.JSON(http.StatusOK, gin.H{"article": serializedArticle})
}

func (a *Articles) Delete(ctx *gin.Context) {

	article, err := a.FindArticleById(ctx)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if err := a.DB.Unscoped().Delete(article).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (a *Articles) setArticleImage(
	ctx *gin.Context,
	article *models.Article,
) error {

	file, err := ctx.FormFile("image")
	if err != nil {
		return err
	}

	if file == nil {
		return errors.New("image file is required")
	}

	path := "uploads/articles/" + strconv.Itoa(int(article.ID))

	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}

	filename := path + "/" + filepath.Base(file.Filename)

	if err := ctx.SaveUploadedFile(file, filename); err != nil {
		return err
	}

	article.Image = os.Getenv("HOST") + "/" + filename

	if err := a.DB.Save(article).Error; err != nil {
		return err
	}

	return nil
}

func (a *Articles) FindArticleById(ctx *gin.Context) (*models.Article, error) {
	var article models.Article
	id := ctx.Param("id")

	if err := a.DB.Preload("User").Preload("Category").First(&article, id).Error; err != nil {
		return nil, err
	}

	return &article, nil
}
