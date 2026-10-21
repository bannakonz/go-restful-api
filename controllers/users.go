package controllers

import (
	"errors"
	"go-restful-api/config"
	"go-restful-api/models"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type Users struct { // เป็นตัวแทนของ controller
	DB *gorm.DB
}

/*
ไว้ให้แอดมิน create user
*/
type createUserForm struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
}

/*
omitempty ไม่ต้อง validate กฎของ field อื่น ถ้าใส่เข้ามาถึง validate
*/
type updateUserForm struct {
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" binding:"omitempty,min=8"`
	Name     string `json:"name"`
}

type userResponse struct {
	ID     uint   `json:"id"`
	Email  string `json:"email"`
	Avatar string `json:"avatar"`
	Name   string `json:"name"`
	Role   string `json:"role"`
}

type usersPaging struct {
	Items  []userResponse `json:"items"`
	Paging *pagingResult  `json:"paging"`
}

func (u *Users) FindAll(ctx *gin.Context) {
	var users []models.User
	query := u.DB.Order("id DESC")

	pagination := pagination{
		ctx:     ctx,
		query:   query,
		records: &users,
	}

	paging := pagination.paginate()

	var serializedUser []userResponse
	copier.Copy(&serializedUser, &users)
	ctx.JSON(http.StatusOK, gin.H{
		"users": usersPaging{
			Items:  serializedUser,
			Paging: paging,
		},
	})
}

func (u *Users) Create(ctx *gin.Context) {
	var form createUserForm
	if err := ctx.ShouldBindJSON(&form); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": err.Error(),
		})
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

	if err := u.DB.Create(&user).Error; err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": err.Error(),
		})
		return
	}

	var serializedUser userResponse
	copier.Copy(&serializedUser, &user)
	ctx.JSON(http.StatusCreated, gin.H{
		"user": serializedUser,
	})
}

func (u *Users) Update(ctx *gin.Context) {
	var form updateUserForm
	if err := ctx.ShouldBindJSON(&form); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": err.Error(),
		})
		return
	}
	user, err := u.findUserById(ctx)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	updates := models.User{
		Email: form.Email,
		Name:  form.Name,
	}

	if form.Password != "" {
		passwordUser := models.User{
			Password: form.Password,
		}

		hashedPassword, err := passwordUser.GenerateEncryptedPassword()
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{
				"error": err.Error(),
			})
			return
		}

		updates.Password = hashedPassword
	}

	// Update only the provided non-empty fields.
	if err := u.DB.Model(user).Updates(updates).Error; err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{
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

func (u *Users) Delete(ctx *gin.Context) {
	user, err := u.findUserById(ctx)

	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := u.DB.Unscoped().Delete(user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (u *Users) Promote(ctx *gin.Context) {
	user, err := u.findUserById(ctx)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	user.Promote()

	if err := u.DB.Save(user).Error; err != nil {
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

func (u *Users) Demote(ctx *gin.Context) {
	user, err := u.findUserById(ctx)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}
	user.Demote()

	if err := u.DB.Save(user).Error; err != nil {
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

func (u *Users) FindOne(ctx *gin.Context) {
	user, err := u.findUserById(ctx)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
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

func (u *Users) findUserById(ctx *gin.Context) (*models.User, error) {
	var user models.User
	id := ctx.Param("id")
	if err := u.DB.First(&user, id).Error; err != nil {
		return nil, err
	}

	return &user, nil
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
