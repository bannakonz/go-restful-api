package middleware

import (
	"go-restful-api/models"
	"net/http"

	"github.com/casbin/casbin/v3"
	"github.com/gin-gonic/gin"
)

func Authorize() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Authentication must run before Authorization.
		user, ok := ctx.Get("sub")
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			return
		}

		// Convert interface{} to *models.User.
		currentUser, ok := user.(*models.User)
		if !ok || currentUser == nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid user",
			})
			return
		}

		// Initialize Casbin enforcer.
		enforcer, err := casbin.NewEnforcer(
			"config/acl_model.conf",
			"config/policy.csv",
		)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		// Check role, resource path, and HTTP method.
		allowed, err := enforcer.Enforce(
			currentUser,
			ctx.Request.URL.Path,
			ctx.Request.Method,
		)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		if !allowed {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "you are not allowed to access this resource",
			})
			return
		}

		// Continue to the next middleware or controller.
		ctx.Next()
	}
}
