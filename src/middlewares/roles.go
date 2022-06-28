package middlewares

import (
	"net/http"

	"github.com/CPU-commits/Intranet_BNews/src/models"
	"github.com/CPU-commits/Intranet_BNews/src/res"
	"github.com/CPU-commits/Intranet_BNews/src/services"
	"github.com/gin-gonic/gin"
)

func RolesMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims, _ := services.NewClaimsFromContext(ctx)
		if claims.UserType == models.TEACHER || claims.UserType == models.ATTORNEY || claims.UserType == models.STUDENT {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, &res.Response{
				Success: false,
				Message: "Unauthorized",
			})
			return
		}
		ctx.Next()
	}
}
