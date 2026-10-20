package middleware

import (
	"errors"
	"go-restful-api/config"
	"go-restful-api/models"
	"log"
	"os"

	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	jwt2 "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type login struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

var identityKey = "sub"

func Authenticate() *jwt.GinJWTMiddleware {
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Key: []byte(os.Getenv("SECRET_KEY")),

		IdentityKey: identityKey,
		// Login -> user
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var form login
			var user models.User

			if err := c.ShouldBindJSON(&form); err != nil {
				return nil, jwt.ErrMissingLoginValues
			}

			db := config.GetDB()

			err := db.Where("email = ?", form.Email).First(&user).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, jwt.ErrFailedAuthentication
			}

			if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.Password)); err != nil {
				return nil, jwt.ErrFailedAuthentication
			}

			return &user, nil
		},

		// user => payload(sub) => jwt
		PayloadFunc: func(data interface{}) jwt2.MapClaims {
			if v, ok := data.(*models.User); ok {
				claims := jwt2.MapClaims{
					identityKey: v.ID,
				}

				return claims
			}

			return jwt2.MapClaims{}
		},

		// handle error
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{"error": message})
		},
	})

	if err != nil {
		log.Fatal("JWT error: " + err.Error())
	}

	if err := authMiddleware.MiddlewareInit(); err != nil {
		log.Fatal("JWT init error: " + err.Error())
	}

	return authMiddleware
}
