package middlewares

import (
	"net/http"

	"github.com/CPU-commits/Intranet_BNews/src/res"
	"github.com/CPU-commits/Intranet_BNews/src/services"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func JWTMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := services.VerifyToken(ctx.Request)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, res.Response{
				Success: false,
				Message: err.Error(),
			})
			return
		}
		if _, ok := token.Claims.(jwt.Claims); !ok && token.Valid {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, res.Response{
				Success: false,
				Message: "Unauthorized",
			})
			return
		}
		metadata, err := services.ExtractTokenMetadata(token)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, res.Response{
				Success: false,
				Message: err.Error(),
			})
			return
		}
		ctx.Set("user", metadata)
		ctx.Next()
	}
}
