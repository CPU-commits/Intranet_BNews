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
// GetSingleNews godoc
// @Summary Get a single news
// @Description Get a single news
// @Tags news
// @Accept json
// @Produce json
// @Param slug path string true "News slug"
// @Success 200 {object} res.Response{body=smaps.SingleNewsMap}
// @Failure 404 {object} res.Response{} "No pudimos encontrar la noticia..."
// @Failure 410 {object} res.Response{} "Esta noticia ya no está disponible"
// @Failure 401 {object} res.Response{} "No tienes acceso a esta noticia"
// @Router /get_single_news/{slug} [get]
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

// GetNews godoc
// @Summary Get news
// @Description Get news
// @Tags news
// @Accept json
// @Produce json
// @Param skip query integer false "Default 0"
// @Param total query bool false "Defaul false"
// @Paraam limit query integer false "Default 15"
// @Param type query string false "Default global -> Values: global || student"
// @Success 200 {object} res.Response{body=smaps.NewsMap}
// @Failure 503 {object} res.Response{} "StatusServiceUnavailable"
// @Failure 400 {object} res.Response{} "Bad query param"
// @Router /get_news [get]
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

// NewNews godoc
// @Summary New news
// @Description New news
// @Tags news
// @Accept mpfd
// @Produce json
// @Param data formData forms.NewsDTO true "News"
// @Success 201 {object} res.Response{body=smaps.SingleNewsMap}
// @Failure 400 {object} res.Response{} "Bad body"
// @Failure 400 {object} res.Response{} "El titulo de la noticia ya está en uso"
// @Failure 401 {object} res.Response{} "Unauthorized"
// @Failure 503 {object} res.Response{} "Service Unavailable - NATS || DB Service Unavailable"
// @Router /new_news [post]
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
	c.JSON(201, res.Response{
		Success: true,
		Data: gin.H{
			"news": uploadedNews,
		},
	})
}

// LikeNews godoc
// @Summary Like news
// @Description Toggle Like news
// @Tags news
// @Accept json
// @Produce json
// @Param idNews path string true "MongoID"
// @Success 200 {object} res.Response{} ""
// @Failure 400 {object} res.Response{} "Bad path param"
// @Failure 404 {object} res.Response{} "Noticia no encontrada"
// @Failure 503 {object} res.Response{} "Service Unavailable - NATS || DB Service Unavailable"
// @Router /like_news/{idNews} [post]
func (news *NewsController) LikeNews(c *gin.Context) {
	idNews := c.Param("idNews")
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

// UpadteNews godoc
// @Summary Update news
// @Description Update news
// @Tags news
// @Accept json
// @Produce json
// @Param idNews path string true "MongoID"
// @Param data body forms.UpdateNewsDTO true "Update"
// @Success 200 {object} res.Response{body=smaps.SingleNewsMap} ""
// @Failure 401 {object} res.Response{} "Unauthorized"
// @Failure 400 {object} res.Response{} "Bad path || body param"
// @Failure 404 {object} res.Response{} "Noticia no encontrada"
// @Router /update_news/{idNews} [put]
func (news *NewsController) UpdateNews(c *gin.Context) {
	// Data
	var data forms.UpdateNewsDTO
	id := c.Param("idNews")
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

// @Summary Update news
// @Description Update news
// @Tags news
// @Accept json
// @Produce json
// @Param idNews path string true "MongoID"
// @Success 200 {object} res.Response{} ""
// @Failure 401 {object} res.Response{} "Unauthorized"
// @Failure 400 {object} res.Response{} "Bad path || body param"
// @Failure 404 {object} res.Response{} "Noticia no encontrada"
// @Router /delete_news/{idNews} [delete]
func (news *NewsController) DeleteNews(c *gin.Context) {
	id := c.Param("idNews")
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
