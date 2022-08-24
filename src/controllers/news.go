package controllers

import (
	"net/http"

	"github.com/CPU-commits/Intranet_BNews/src/forms"
	"github.com/CPU-commits/Intranet_BNews/src/res"
	"github.com/CPU-commits/Intranet_BNews/src/services"
	"github.com/gin-gonic/gin"
)

// Services
var newsService = services.NewNewsService()
var likesService = services.NewLikesService()

type NewsController struct{}

func init() {
	UploadNews()
}

// Nats
func UploadNews() {
	newsService.UploadNews()
}

// API
func (n *NewsController) GetSingleNews(c *gin.Context) {
	slug := c.Param("slug")
	claims, _ := services.NewClaimsFromContext(c)
	// Find
	news, err := newsService.GetSingleNews(slug, claims)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, res.Response{
			Message: err.Err.Error(),
			Success: false,
		})
	}
	// Response
	response := make(map[string]interface{})
	response["news"] = news
	c.JSON(200, res.Response{
		Success: true,
		Data:    response,
	})
}

func (n *NewsController) GetNews(c *gin.Context) {
	claims, _ := services.NewClaimsFromContext(c)
	skip := c.DefaultQuery("skip", "0")
	total := c.DefaultQuery("total", "false")
	limit := c.DefaultQuery("limit", "15")
	newsType := c.DefaultQuery("type", "global")
	// Get
	news, totalData, err := newsService.GetNews(
		skip,
		total == "true",
		limit,
		newsType,
		claims,
	)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, res.Response{
			Success: false,
			Message: err.Err.Error(),
		})
		return
	}
	// Response
	response := make(map[string]interface{})
	response["news"] = news
	response["total"] = totalData
	c.JSON(200, res.Response{
		Success: true,
		Data:    response,
	})
}

func (news *NewsController) NewNews(c *gin.Context) {
	var data forms.NewsDTO
	claims, _ := services.NewClaimsFromContext(c)

	if err := c.ShouldBind(&data); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, res.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	// Get file from form
	file, err := c.FormFile("img")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, res.Response{
			Success: false,
			Message: "Ha ocurrido un error tratando de leer el archivo",
		})
		return
	}
	uploadedNews, errRes := newsService.NewNews(data, file, claims)
	if errRes != nil {
		c.AbortWithStatusJSON(errRes.StatusCode, res.Response{
			Success: false,
			Message: errRes.Err.Error(),
		})
		return
	}
	c.JSON(200, res.Response{
		Success: true,
		Data: gin.H{
			"news": uploadedNews,
		},
	})
}

func (news *NewsController) LikeNews(c *gin.Context) {
	idNews := c.Param("id")
	claims, _ := services.NewClaimsFromContext(c)
	// Get news
	err := likesService.LikeNews(idNews, claims)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, res.Response{
			Success: false,
			Message: err.Err.Error(),
		})
		return
	}
	c.JSON(200, res.Response{
		Success: true,
	})
}

func (news *NewsController) UpdateNews(c *gin.Context) {
	// Data
	var data forms.UpdateNewsDTO
	id := c.Param("id")
	claims, _ := services.NewClaimsFromContext(c)

	if err := c.ShouldBind(&data); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, res.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	// Update
	newsData, errRes := newsService.UpdateNews(data, id, claims)
	if errRes != nil {
		c.AbortWithStatusJSON(errRes.StatusCode, res.Response{
			Success: false,
			Message: errRes.Err.Error(),
		})
		return
	}

	// Response
	response := make(map[string]interface{})
	response["news"] = newsData

	c.JSON(200, res.Response{
		Success: true,
		Data:    response,
	})
}

func (news *NewsController) DeleteNews(c *gin.Context) {
	id := c.Param("id")
	claims, _ := services.NewClaimsFromContext(c)
	// Delete
	err := newsService.DeleteNews(id, claims)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, res.Response{
			Success: false,
			Message: err.Err.Error(),
		})
		return
	}
	c.JSON(200, res.Response{
		Success: true,
	})
}
