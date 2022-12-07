package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/CPU-commits/Intranet_BNews/src/controllers"
	"github.com/CPU-commits/Intranet_BNews/src/docs"
	"github.com/CPU-commits/Intranet_BNews/src/middlewares"
	"github.com/CPU-commits/Intranet_BNews/src/res"
	"github.com/CPU-commits/Intranet_BNews/src/settings"
	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/secure"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"     // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
	"go.uber.org/zap"
)

func keyFunc(c *gin.Context) string {
	return c.ClientIP()
}

func ErrorHandler(c *gin.Context, info ratelimit.Info) {
	c.JSON(http.StatusTooManyRequests, &res.Response{
		Success: false,
		Message: "Too many requests. Try again in" + time.Until(info.ResetTime).String(),
	})
}

var settingsData = settings.GetSettings()

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("No .env file found")
	}
}

func Init() {
	router := gin.New()
	// Proxies
	router.SetTrustedProxies([]string{"localhost"})
	// Zap looger
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	router.Use(ginzap.GinzapWithConfig(logger, &ginzap.Config{
		TimeFormat: time.RFC3339,
		UTC:        true,
		SkipPaths:  []string{"/api/c/classroom/swagger"},
	}))
	router.Use(ginzap.RecoveryWithZap(logger, true))

	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Server Internal Error: %s", err))
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, res.Response{
			Success: false,
			Message: "Server Internal Error",
		})
	}))

	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Server Internal Error: %s", err))
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, res.Response{
			Success: false,
			Message: "Server Internal Error",
		})
	}))
	// Docs
	docs.SwaggerInfo.BasePath = "/api/c/classroom"
	docs.SwaggerInfo.Version = "v1"
	docs.SwaggerInfo.Host = "localhost:8080"
	// CORS
	httpOrigin := "http://" + settingsData.CLIENT_URL
	httpsOrigin := "https://" + settingsData.CLIENT_URL

	fmt.Printf("settingsData.CLIENT_URL: %v\n", settingsData.CLIENT_URL)
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{httpOrigin, httpsOrigin},
		AllowMethods:     []string{"GET", "OPTIONS", "PUT", "DELETE", "POST"},
		AllowCredentials: true,
		AllowWebSockets:  false,
		AllowHeaders:     []string{"*"},
		MaxAge:           12 * time.Hour,
	}))
	// Secure
	sslUrl := "ssl." + settingsData.CLIENT_URL
	secureConfig := secure.Config{
		SSLHost:              sslUrl,
		STSSeconds:           315360000,
		STSIncludeSubdomains: true,
		FrameDeny:            true,
		ContentTypeNosniff:   true,
		BrowserXssFilter:     true,
		IENoOpen:             true,
		ReferrerPolicy:       "strict-origin-when-cross-origin",
		SSLProxyHeaders: map[string]string{
			"X-Fowarded-Proto": "https",
		},
	}
	/*if settingsData.NODE_ENV == "prod" {
		secureConfig.AllowedHosts = []string{
			settingsData.CLIENT_URL,
			sslUrl,
		}
	}*/
	router.Use(secure.New(secureConfig))
	// Rate limit
	store := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  time.Second,
		Limit: 7,
	})
	mw := ratelimit.RateLimiter(store, &ratelimit.Options{
		ErrorHandler: ErrorHandler,
		KeyFunc:      keyFunc,
	})
	router.Use(mw)
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
		news.POST("/like_news/:idNews", newsController.LikeNews)
		news.PUT(
			"/update_news/:idNews",
			middlewares.RolesMiddleware(),
			newsController.UpdateNews,
		)
		news.DELETE(
			"/delete_news/:idNews",
			middlewares.RolesMiddleware(),
			newsController.DeleteNews,
		)
	}
	// Route docs
	router.GET("/api/news/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
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
