package controllers

import (
	"errors"
	"go-restful-api/config"
	"go-restful-api/models"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type Auth struct {
	DB *gorm.DB
}

type authForm struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type updateProfileForm struct {
	Email  string                `form:"email"`
	Name   string                `form:"name"`
	Avatar *multipart.FileHeader `form:"avatar"`
}

type authResponse struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
}

type userResponse struct {
	ID     uint   `json:"id"`
	Email  string `json:"email"`
	Avatar string `json:"avatar"`
	Name   string `json:"name"`
	Role   string `json:"role"`
}

/*
GetProfile
 1. /auth/profile => JWT => sub (UserId) => user
 2. GET Profile ดึงค่า user จาก context
*/
func (a *Auth) GetProfile(ctx *gin.Context) {

	sub, _ := ctx.Get("sub")
	user := sub.(*models.User)

	var serializedUser userResponse
	copier.Copy(&serializedUser, &user)
	ctx.JSON(http.StatusOK, gin.H{"user": serializedUser})
}

func (a *Auth) Signup(ctx *gin.Context) {
	var form authForm
	if err := ctx.ShouldBindJSON(&form); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	copier.Copy(&user, &form)

	hashedPassword, err := user.GenerateEncryptedPassword()
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": err.Error(),
		})
		return
	}
	user.Password = hashedPassword

	if err := a.DB.Create(&user).Error; err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	var serializedUser authResponse
	copier.Copy(&serializedUser, &user)
	ctx.JSON(http.StatusCreated, gin.H{"user": serializedUser})
}

func (a *Auth) UpdateProfile(ctx *gin.Context) {
	var form updateProfileForm

	// เปลี่ยนจาก ShouldBindJSON เป็น ShouldBind
	if err := ctx.ShouldBind(&form); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": err.Error(),
		})
		return
	}

	sub, exists := ctx.Get("sub")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, ok := sub.(*models.User)
	if !ok || user == nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Update เฉพาะ Email และ Name
	if err := a.DB.Model(user).Updates(models.User{
		Email: form.Email,
		Name:  form.Name,
	}).Error; err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": err.Error(),
		})
		return
	}

	if form.Email != "" {
		user.Email = form.Email
	}
	if form.Name != "" {
		user.Name = form.Name
	}

	// Upload avatar
	if err := setUserImage(ctx, user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	var serializedUser userResponse
	copier.Copy(&serializedUser, user)

	ctx.JSON(http.StatusOK, gin.H{
		"user": serializedUser,
	})
}

func setUserImage(ctx *gin.Context, user *models.User) error {
	file, err := ctx.FormFile("avatar")

	if errors.Is(err, http.ErrMissingFile) {
		return nil
	}
	if err != nil {
		return err
	}

	oldAvatar := user.Avatar

	path := "uploads/users/" + strconv.Itoa(int(user.ID))

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	filename := path + "/" + filepath.Base(file.Filename)

	if err := ctx.SaveUploadedFile(file, filename); err != nil {
		return err
	}

	db := config.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}

	host := strings.TrimRight(os.Getenv("HOST"), "/")
	user.Avatar = host + "/" + filename

	if err := db.Save(user).Error; err != nil {
		return err
	}

	// ลบรูปเดิมหลังจากบันทึกรูปใหม่สำเร็จ
	if oldAvatar != "" {
		oldPath := strings.TrimPrefix(oldAvatar, host+"/")

		if strings.HasPrefix(oldPath, path+"/") && oldPath != filename {
			_ = os.Remove(oldPath)
		}
	}

	return nil
}
