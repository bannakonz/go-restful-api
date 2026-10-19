package controllers

import (
	"go-restful-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type Categories struct {
	DB *gorm.DB
}

type categoryResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Articles    []struct {
		Id    uint   `json:"id"`
		Title string `json:"title"`
	} `json:"articles"`
}

type createCategoryForm struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type updateCategoryForm struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type allCategoryResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (c *Categories) FindAll(ctx *gin.Context) {
	var categories []models.Category
	c.DB.Order("id DESC").Find(&categories)

	var serializedCategories []allCategoryResponse
	copier.Copy(&serializedCategories, &categories)
	ctx.JSON(http.StatusOK, gin.H{"categories": serializedCategories})

}

func (c *Categories) FindOne(ctx *gin.Context) {
	category, err := c.findCategoryById(ctx)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var serializedCategory categoryResponse
	copier.Copy(&serializedCategory, &category)
	ctx.JSON(http.StatusOK, gin.H{"category": serializedCategory})
}

func (c *Categories) Create(ctx *gin.Context) {
	var form createCategoryForm
	if err := ctx.ShouldBindJSON(&form); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	var category models.Category
	copier.Copy(&category, &form)
	if err := c.DB.Create(&category).Error; err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	var serializedCategories categoryResponse
	copier.Copy(&serializedCategories, &category)
	ctx.JSON(http.StatusCreated, gin.H{"categories": serializedCategories})
}

func (c *Categories) Update(ctx *gin.Context) {
	var form updateCategoryForm
	if err := ctx.ShouldBindJSON(&form); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	category, err := c.findCategoryById(ctx)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if err := c.DB.Model(&category).Updates(&form).Error; err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	var serializedCategory categoryResponse
	copier.Copy(&serializedCategory, &category)
	ctx.JSON(http.StatusOK, gin.H{"category": serializedCategory})

}

func (c *Categories) Delete(ctx *gin.Context) {
	category, err := c.findCategoryById(ctx)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.DB.Unscoped().Delete(&category)
	ctx.Status(http.StatusNoContent)
}

func (c *Categories) findCategoryById(ctx *gin.Context) (*models.Category, error) {
	var category models.Category
	id := ctx.Param("id")
	if err := c.DB.Preload("Articles").First(&category, id).Error; err != nil {
		return nil, err
	}

	return &category, nil

}
