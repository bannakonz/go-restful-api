package middleware

import (
	"go-restful-api/config"
	"go-restful-api/models"
	"log"
	"os"
	"strconv"

	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	jwt2 "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
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

		TokenLookup:   "header:Authorization",
		TokenHeadName: "Bearer",

		// JWT -> sub (User ID) -> User
		IdentityHandler: func(c *gin.Context) interface{} {
			var user models.User

			claims := jwt.ExtractClaims(c)

			idString, ok := claims[identityKey].(string)
			if !ok {
				return nil
			}

			id, err := strconv.ParseUint(idString, 10, 64)
			if err != nil || id == 0 {
				return nil
			}

			db := config.GetDB()
			if db == nil {
				return nil
			}

			if err := db.First(&user, id).Error; err != nil {
				return nil
			}

			return &user
		},

		Authorizer: func(c *gin.Context, data interface{}) bool {
			user, ok := data.(*models.User)
			return ok && user != nil && user.ID != 0
		},

		// Login -> User
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var form login
			var user models.User

			if err := c.ShouldBindJSON(&form); err != nil {
				return nil, jwt.ErrMissingLoginValues
			}

			db := config.GetDB()
			if db == nil {
				return nil, jwt.ErrFailedAuthentication
			}

			err := db.Where("email = ?", form.Email).
				First(&user).Error

			if err != nil {
				return nil, jwt.ErrFailedAuthentication
			}

			if err := bcrypt.CompareHashAndPassword(
				[]byte(user.Password),
				[]byte(form.Password),
			); err != nil {
				return nil, jwt.ErrFailedAuthentication
			}

			return &user, nil
		},

		// User -> JWT Payload (sub)
		PayloadFunc: func(data interface{}) jwt2.MapClaims {
			if user, ok := data.(*models.User); ok {
				return jwt2.MapClaims{
					identityKey: strconv.FormatUint(
						uint64(user.ID), 10,
					),
				}
			}

			return jwt2.MapClaims{}
		},

		// Handle error
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"error": message,
			})
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
