package controllers

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/CPU-commits/Intranet_BNews/src/aws_s3"
	"github.com/CPU-commits/Intranet_BNews/src/db"
	"github.com/CPU-commits/Intranet_BNews/src/forms"
	"github.com/CPU-commits/Intranet_BNews/src/models"
	"github.com/CPU-commits/Intranet_BNews/src/res"
	"github.com/CPU-commits/Intranet_BNews/src/services"
	"github.com/CPU-commits/Intranet_BNews/src/stack"
	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Models
var newsModel = new(models.NewsModel)
var likesModel = new(models.LikesModel)

var nats = stack.NewNats()
var aws = aws_s3.NewAWSS3()

type NewsController struct{}

type User struct {
	Name           string `json:"name" bson:"name"`
	FirstLastname  string `json:"first_lastname" bson:"first_lastname"`
	SecondLastname string `json:"second_lastname" bson:"second_lastname"`
	ID             string `json:"_id" bson:"_id"`
}

type Image struct {
	ID  string `json:"_id" bson:"_id"`
	URL string `json:"url" bson:"url"`
	Key string `bson:"key"`
}

type NewsResponse struct {
	Author     User               `json:"author" bson:"author"`
	Headline   string             `json:"headline" bson:"headline"`
	Title      string             `json:"title" bson:"title"`
	Image      Image              `json:"image" bson:"image"`
	UploadDate primitive.DateTime `json:"upload_date" bson:"upload_date"`
	UpdateDate primitive.DateTime `json:"update_date" bson:"update:date"`
	URL        string             `json:"url" bson:"url"`
	Type       string             `json:"type" bson:"type"`
	Body       string             `json:"body" bson:"body"`
	Status     bool               `json:"status" bson:"status"`
	Like       bool               `json:"like"`
	Likes      int                `json:"likes"`
	ID         string             `json:"_id" bson:"_id"`
}

func getLookupUser() bson.D {
	return bson.D{
		{
			Key: "$lookup",
			Value: bson.M{
				"from":         "users",
				"localField":   "author_id",
				"foreignField": "_id",
				"as":           "author",
				"pipeline": bson.A{
					bson.M{
						"$project": bson.M{
							"name":            1,
							"first_lastname":  1,
							"second_lastname": 1,
						},
					},
				},
			},
		},
	}
}

func getLookupFile() bson.D {
	return bson.D{
		{
			Key: "$lookup",
			Value: bson.M{
				"from":         "files",
				"localField":   "img",
				"foreignField": "_id",
				"as":           "image",
				"pipeline": bson.A{
					bson.M{
						"$project": bson.M{
							"url": 1,
							"key": 1,
						},
					},
				},
			},
		},
	}
}

func getMatchStatusTrue(newsType string) bson.D {
	return bson.D{
		{
			Key: "$match",
			Value: bson.M{
				"status": true,
				"type":   newsType,
			},
		},
	}
}

func getNews(pipeline mongo.Pipeline, requestImage bool) ([]NewsResponse, error) {
	cursor, err := newsModel.Use().Aggregate(db.Ctx, pipeline)
	if err != nil {
		return nil, err
	}
	var newsData []NewsResponse
	if err = cursor.All(db.Ctx, &newsData); err != nil {
		return nil, err
	}
	if len(newsData) == 0 {
		return nil, nil
	}
	if !newsData[0].Status {
		return newsData, nil
	}
	// Request nats
	if requestImage {
		var images []string
		for i := 0; i < len(newsData); i++ {
			images = append(images, newsData[i].Image.Key)
		}
		data, err := json.Marshal(images)
		if err != nil {
			return nil, err
		}
		msg, err := nats.Request("get_aws_token_access", data)
		if err != nil {
			return nil, err
		}
		var imagesURLs []string
		json.Unmarshal(msg.Data, &imagesURLs)
		// Add image URLs to Response
		for i := 0; i < len(newsData); i++ {
			newsData[i].Image.URL = imagesURLs[i]
		}
	}
	return newsData, nil
}

func uploadImage(file *multipart.FileHeader) (*models.FileDB, error) {
	// Upload file to S3
	result, err := aws.UploadFile(file)
	if err != nil {
		return nil, err
	}
	// Request NATS (Get id file insert)
	msg, err := nats.Request("upload_image", []byte(result.Location))
	if err != nil {
		split := strings.Split(result.Location, "/")
		key := fmt.Sprintf("news/%s", split[len(split)-1])
		aws.DeleteFile(key)
		return nil, err
	}
	// Process response NATS
	var fileDb *models.FileDB
	err = json.Unmarshal(msg.Data, &fileDb)
	if err != nil {
		return nil, err
	}
	return fileDb, nil
}

func (news *NewsController) GetSingleNews(c *gin.Context) {
	slug := c.Param("slug")
	claims, _ := services.NewClaimsFromContext(c)
	// Find
	lookUpStage := getLookupFile()
	lookUpUserStage := getLookupUser()
	projectStage := bson.D{
		{
			Key: "$project",
			Value: bson.M{
				"title":       1,
				"headline":    1,
				"upload_date": 1,
				"url":         1,
				"type":        1,
				"update_date": 1,
				"status":      1,
				"body":        1,
				"image": bson.M{
					"$arrayElemAt": bson.A{
						"$image", 0,
					},
				},
				"author": bson.M{
					"$arrayElemAt": bson.A{
						"$author", 0,
					},
				},
			},
		},
	}
	matchStage := bson.D{
		{
			Key: "$match",
			Value: bson.D{
				{
					Key:   "url",
					Value: slug,
				},
			},
		},
	}
	newsData, err := getNews(mongo.Pipeline{
		matchStage,
		lookUpStage,
		lookUpUserStage,
		projectStage,
	}, true)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, res.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	if newsData == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, res.Response{
			Success: false,
			Message: "No pudimos encontrar la noticia...",
		})
		return
	}
	if !newsData[0].Status {
		c.AbortWithStatusJSON(http.StatusGone, res.Response{
			Success: false,
			Message: "Esta noticia ya no está disponible",
		})
		return
	}
	// Validate
	newsType := newsData[0].Type
	if newsType == "student" {
		if claims.UserType != models.STUDENT && claims.UserType != models.STUDENT_DIRECTIVE {
			c.AbortWithStatusJSON(http.StatusUnauthorized, res.Response{
				Success: false,
				Message: "No tienes acceso a esta noticia",
			})
			return
		}
	}
	// Response
	response := make(map[string]interface{})
	response["news"] = newsData[0]
	c.JSON(200, res.Response{
		Success: true,
		Data:    response,
	})
}

func (news *NewsController) GetNews(c *gin.Context) {
	skip := c.DefaultQuery("skip", "0")
	total := c.DefaultQuery("total", "false")
	limit := c.DefaultQuery("limit", "15")
	newsType := c.DefaultQuery("type", "global")
	skipNumber, err := strconv.Atoi(skip)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, res.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	limitNumber, err := strconv.Atoi(limit)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, res.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	sortStage := bson.D{
		{
			Key: "$sort",
			Value: bson.D{
				{Key: "upload_date", Value: -1},
			},
		},
	}
	limitStage := bson.D{
		{
			Key:   "$limit",
			Value: limitNumber,
		},
	}
	skipStage := bson.D{
		{
			Key:   "$skip",
			Value: skipNumber,
		},
	}
	lookUpStage := getLookupFile()
	lookUpUserStage := getLookupUser()
	projectStage := bson.D{
		{
			Key: "$project",
			Value: bson.M{
				"title":       1,
				"headline":    1,
				"upload_date": 1,
				"url":         1,
				"type":        1,
				"status":      1,
				"image": bson.M{
					"$arrayElemAt": bson.A{
						"$image", 0,
					},
				},
				"author": bson.M{
					"$arrayElemAt": bson.A{
						"$author", 0,
					},
				},
			},
		},
	}
	newsData, err := getNews(mongo.Pipeline{
		getMatchStatusTrue(newsType),
		sortStage,
		limitStage,
		skipStage,
		lookUpStage,
		lookUpUserStage,
		projectStage,
	}, true)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, res.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	if newsData == nil {
		response := make(map[string]interface{})
		response["news"] = newsData
		response["total"] = 0
		c.JSON(200, res.Response{
			Success: true,
			Data:    response,
		})
		return
	}
	// Get likes
	claims, _ := services.NewClaimsFromContext(c)
	userObjectID, err := primitive.ObjectIDFromHex(claims.ID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, res.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	var wg sync.WaitGroup
	for i := 0; i < len(newsData); i++ {
		wg.Add(1)
		go func(i int, wg *sync.WaitGroup) {
			defer wg.Done()
			// Get like user
			var likeData *models.Likes

			newsObjectId, _ := primitive.ObjectIDFromHex(newsData[i].ID)
			cursor := likesModel.Use().FindOne(db.Ctx, bson.D{
				{
					Key:   "user",
					Value: userObjectID,
				},
				{
					Key:   "news",
					Value: newsObjectId,
				},
			})
			cursor.Decode(&likeData)
			newsData[i].Like = (likeData != nil)
			// Get likes news
			count, err := likesModel.Use().CountDocuments(db.Ctx, bson.D{
				{
					Key:   "news",
					Value: newsObjectId,
				},
			})
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, res.Response{
					Success: false,
					Message: err.Error(),
				})
				return
			}
			newsData[i].Likes = int(count)
		}(i, &wg)
	}
	// Response
	response := make(map[string]interface{})
	if total == "true" {
		totalData, err := newsModel.Use().CountDocuments(db.Ctx, bson.D{})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, res.Response{
				Success: false,
				Message: err.Error(),
			})
			return
		}
		response["total"] = totalData
	}
	wg.Wait()
	response["news"] = newsData
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, res.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	c.JSON(200, res.Response{
		Success: true,
		Data:    response,
	})
}

func (news *NewsController) NewNews(c *gin.Context) {
	var data forms.NewsDTO

	if c.ShouldBind(&data) != nil {
		c.AbortWithStatusJSON(406, res.Response{
			Success: false,
			Message: "Provide relevante fields",
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
	// Validate unique slug
	var findNews *models.News
	slugNews := slug.MakeLang(data.Title, "es")
	cursor := newsModel.Use().FindOne(db.Ctx, bson.D{
		primitive.E{
			Key:   "url",
			Value: slugNews,
		},
	})
	cursor.Decode(&findNews)
	if findNews != nil {
		c.AbortWithStatusJSON(http.StatusConflict, res.Response{
			Success: false,
			Message: "El titulo de la noticia ya está en uso",
		})
		return
	}
	// Upload image
	fileDb, err := uploadImage(file)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, res.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	// Upload news
	var newsType string
	claims, _ := services.NewClaimsFromContext(c)
	if claims.UserType == models.STUDENT_DIRECTIVE {
		newsType = "student"
	} else {
		newsType = "global"
	}
	newsData, err := newsModel.NewModel(data, fileDb.ID.OID, slugNews, newsType, claims.ID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, res.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	uploadedNews, err := newsModel.Use().InsertOne(db.Ctx, newsData)
	if err != nil {
		c.AbortWithStatusJSON(502, res.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	// Notify news
	nats.PublishEncode("notify/global", &res.Notify{
		Title: data.Title,
		Link:  fmt.Sprintf("/noticias/%s", newsData.Url),
		Img:   fileDb.Key,
		Type:  newsType,
	})
	c.JSON(200, res.Response{
		Success: true,
		Data: gin.H{
			"news": uploadedNews,
		},
	})
}

func (news *NewsController) LikeNews(c *gin.Context) {
	idNews := c.Param("id")
	// Get news
	newsObjectId, err := primitive.ObjectIDFromHex(idNews)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, res.Response{
			Success: true,
			Message: err.Error(),
		})
		return
	}

	var newsData *models.News
	opts := options.FindOne().SetProjection(bson.D{
		{
			Key:   "_id",
			Value: 1,
		},
	})
	cursor := newsModel.Use().FindOne(db.Ctx, bson.D{
		{
			Key:   "_id",
			Value: newsObjectId,
		},
		{
			Key:   "status",
			Value: true,
		},
	}, opts)
	cursor.Decode(&newsData)
	if newsData == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, res.Response{
			Success: false,
			Message: "News not found",
		})
		return
	}
	// Toogle like
	claims, _ := services.NewClaimsFromContext(c)
	userObjectID, err := primitive.ObjectIDFromHex(claims.ID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, res.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	var hasLike *models.Likes
	cursor = likesModel.Use().FindOne(db.Ctx, bson.D{
		{
			Key:   "user",
			Value: userObjectID,
		},
		{
			Key:   "news",
			Value: newsObjectId,
		},
	})
	cursor.Decode(&hasLike)
	if hasLike == nil {
		like := likesModel.NewModel(userObjectID, newsObjectId)
		_, err := likesModel.Use().InsertOne(db.Ctx, like)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, res.Response{
				Success: false,
				Message: err.Error(),
			})
			return
		}
	} else {
		likesModel.Use().FindOneAndDelete(db.Ctx, bson.D{
			{
				Key:   "_id",
				Value: hasLike.ID,
			},
		})
	}
	c.JSON(200, res.Response{
		Success: true,
	})
}

func (news *NewsController) UpdateNews(c *gin.Context) {
	// Data
	var data forms.UpdateNewsDTO

	id := c.Param("id")
	idObjectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, res.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	if c.ShouldBind(&data) != nil {
		c.AbortWithStatusJSON(406, res.Response{
			Success: false,
			Message: "Provide relevante fields",
		})
		return
	}
	// Get news
	var findNews *models.News
	cursorNews := newsModel.Use().FindOne(db.Ctx, bson.D{
		{
			Key:   "_id",
			Value: idObjectId,
		},
	})
	err = cursorNews.Decode(&findNews)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, res.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	// Verify identity
	claims, _ := services.NewClaimsFromContext(c)
	if findNews.Type == "global" && claims.UserType != models.DIRECTIVE && claims.UserType != models.DIRECTOR || findNews.Type == "student" && findNews.Type != models.STUDENT_DIRECTIVE {
		c.AbortWithStatusJSON(http.StatusUnauthorized, res.Response{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}
	// Update data
	update := bson.D{
		{
			Key:   "update_date",
			Value: primitive.NewDateTimeFromTime(time.Now()),
		},
	}
	if data.Img != nil {
		fileDb, err := uploadImage(data.Img)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, res.Response{
				Success: false,
				Message: err.Error(),
			})
			return
		}
		imgObjectId, err := primitive.ObjectIDFromHex(fileDb.ID.OID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, res.Response{
				Success: false,
				Message: err.Error(),
			})
			return
		}
		update = append(update, primitive.E{
			Key:   "img",
			Value: imgObjectId,
		})
	}
	if data.Body != "" {
		update = append(update, primitive.E{
			Key:   "body",
			Value: data.Body,
		})
	}
	if data.Headline != "" {
		update = append(update, primitive.E{
			Key:   "headline",
			Value: data.Headline,
		})
	}
	if data.Title != "" {
		update = append(update, primitive.E{
			Key:   "title",
			Value: data.Title,
		})
	}
	// Update news
	var newsData *models.News
	cursor := newsModel.Use().FindOneAndUpdate(
		db.Ctx, bson.D{
			{
				Key:   "_id",
				Value: idObjectId,
			},
		},
		bson.D{
			{
				Key:   "$set",
				Value: update,
			},
		},
	)
	err = cursor.Decode(&news)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, res.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	response := make(map[string]interface{})
	response["news"] = newsData

	c.JSON(200, res.Response{
		Success: true,
		Data:    response,
	})
}

func (news *NewsController) DeleteNews(c *gin.Context) {
	id := c.Param("id")
	idObjectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, res.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	// Get news
	lookUpStage := getLookupFile()
	lookUpUserStage := getLookupUser()
	projectStage := bson.D{
		{
			Key: "$project",
			Value: bson.M{
				"title":       1,
				"headline":    1,
				"upload_date": 1,
				"url":         1,
				"type":        1,
				"update_date": 1,
				"body":        1,
				"image": bson.M{
					"$arrayElemAt": bson.A{
						"$image", 0,
					},
				},
				"author": bson.M{
					"$arrayElemAt": bson.A{
						"$author", 0,
					},
				},
			},
		},
	}
	matchStage := bson.D{
		{
			Key: "$match",
			Value: bson.D{
				{
					Key:   "_id",
					Value: idObjectId,
				},
				{
					Key:   "status",
					Value: true,
				},
			},
		},
	}
	newsData, err := getNews(mongo.Pipeline{
		matchStage,
		lookUpStage,
		lookUpUserStage,
		projectStage,
	}, false)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, res.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	if newsData == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, res.Response{
			Success: false,
			Message: "No existe la noticia",
		})
		return
	}
	claims, _ := services.NewClaimsFromContext(c)
	if newsData[0].Type == "global" && (claims.UserType != models.DIRECTIVE && claims.UserType != models.DIRECTOR) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, res.Response{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}
	if newsData[0].Type == "student" && claims.UserType != models.STUDENT_DIRECTIVE {
		c.AbortWithStatusJSON(http.StatusUnauthorized, res.Response{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}
	// Delete image
	_, err = nats.Request("delete_image", []byte(newsData[0].Image.ID))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, res.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	// Delete news
	newsModel.Use().FindOneAndUpdate(
		db.Ctx,
		bson.D{
			{
				Key:   "_id",
				Value: idObjectId,
			},
		},
		bson.D{
			{
				Key: "$set",
				Value: bson.D{
					{
						Key:   "status",
						Value: false,
					},
					{
						Key:   "body",
						Value: "",
					},
				},
			},
		},
	)
	c.JSON(200, res.Response{
		Success: true,
	})
}
