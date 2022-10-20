package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/CPU-commits/Intranet_BNews/src/controllers"
	"github.com/CPU-commits/Intranet_BNews/src/middlewares"
	"github.com/CPU-commits/Intranet_BNews/src/res"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("No .env file found")
	}
}

func Init() {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Server Internal Error: %s", err))
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, res.Response{
			Success: false,
			Message: "Server Internal Error",
		})
	}))
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"*"},
		AllowHeaders: []string{"*"},
	}))
	// Routes
	news := router.Group("/api/news", middlewares.JWTMiddleware())
	{
		// Init controllers
		newsController := new(controllers.NewsController)
		// Define routes
		news.GET("/get_news", newsController.GetNews)
		news.GET("/get_single_news/:slug", newsController.GetSingleNews)
		news.POST(
			"/new_news",
			middlewares.RolesMiddleware(),
			newsController.NewNews,
		)
		news.POST("/like_news/:id", newsController.LikeNews)
		news.PUT(
			"/update_news/:id",
			middlewares.RolesMiddleware(),
			newsController.UpdateNews,
		)
		news.DELETE(
			"/delete_news/:id",
			middlewares.RolesMiddleware(),
			newsController.DeleteNews,
		)
	}
	// No route
	router.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(404, res.Response{
			Success: false,
			Message: "Not found",
		})
	})
	// Init server
	if err := router.Run(); err != nil {
		log.Fatalf("Error init server")
	}
}
