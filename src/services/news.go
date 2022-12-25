package services

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/CPU-commits/Intranet_BNews/src/db"
	"github.com/CPU-commits/Intranet_BNews/src/forms"
	"github.com/CPU-commits/Intranet_BNews/src/models"
	"github.com/CPU-commits/Intranet_BNews/src/res"
	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var newsService *NewsService

type NewsService struct{}

func (news *NewsService) getLookupUser() bson.D {
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

func (news *NewsService) getLookupFile() bson.D {
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

func (news *NewsService) getMatchStatusTrue(newsType string) bson.D {
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

func uploadImage(file *multipart.FileHeader) (*models.FileDB, error) {
	// Upload file to S3
	_, key, err := aws.UploadFile(file)
	if err != nil {
		return nil, err
	}
	// Request NATS (Get id file insert)
	msg, err := nats.Request("upload_image", []byte(key))
	if err != nil {
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

func (news *NewsService) getNews(pipeline mongo.Pipeline, requestImage bool) ([]NewsResponse, error) {
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

func (n *NewsService) GetSingleNews(slug string, claims *Claims) (*NewsResponse, *ErrorRes) {
	lookUpStage := n.getLookupFile()
	lookUpUserStage := n.getLookupUser()
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
	newsData, err := n.getNews(mongo.Pipeline{
		matchStage,
		lookUpStage,
		lookUpUserStage,
		projectStage,
	}, true)
	if err != nil {
		return nil, &ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	if newsData == nil {
		return nil, &ErrorRes{
			Err:        fmt.Errorf("no pudimos encontrar la noticia"),
			StatusCode: http.StatusNotFound,
		}
	}
	if !newsData[0].Status {
		return nil, &ErrorRes{
			Err:        fmt.Errorf("esta noticia ya no está disponible"),
			StatusCode: http.StatusGone,
		}
	}
	// Validate
	newsType := newsData[0].Type
	if newsType == "student" {
		if claims.UserType != models.STUDENT && claims.UserType != models.STUDENT_DIRECTIVE {
			return nil, &ErrorRes{
				Err:        fmt.Errorf("no tienes acceso a esta noticia"),
				StatusCode: http.StatusUnauthorized,
			}
		}
	}
	return &newsData[0], nil
}

func (n *NewsService) GetNews(
	skip string,
	total bool,
	limit string,
	newsType string,
	claims *Claims,
) ([]NewsResponse, int, *ErrorRes) {
	// Recovery if close channel
	defer func() {
		recovery := recover()
		if recovery != nil {
			fmt.Printf("A channel closed")
		}
	}()

	skipNumber, err := strconv.Atoi(skip)
	if err != nil {
		return nil, 0, &ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	limitNumber, err := strconv.Atoi(limit)
	if err != nil {
		return nil, 0, &ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
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
	lookUpStage := n.getLookupFile()
	lookUpUserStage := n.getLookupUser()
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
	newsData, err := n.getNews(mongo.Pipeline{
		n.getMatchStatusTrue(newsType),
		sortStage,
		limitStage,
		skipStage,
		lookUpStage,
		lookUpUserStage,
		projectStage,
	}, true)
	if err != nil {
		return nil, 0, &ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	if newsData == nil {
		return newsData, 0, nil
	}
	// Get likes
	userObjectID, err := primitive.ObjectIDFromHex(claims.ID)
	if err != nil {
		return nil, 0, &ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}

	var wg sync.WaitGroup
	c := make(chan (int), 10)
	for i := 0; i < len(newsData); i++ {
		wg.Add(1)
		c <- 1

		go func(i int, wg *sync.WaitGroup, errRet *error) {
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
				*errRet = err
			}
			newsData[i].Likes = int(count)
			<-c
		}(i, &wg, &err)
	}
	wg.Wait()
	if err != nil {
		return nil, 0, &ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	var totalData int64
	if total {
		totalData, err = newsModel.Use().CountDocuments(db.Ctx, bson.D{})
		if err != nil {
			return nil, 0, &ErrorRes{
				Err:        err,
				StatusCode: http.StatusBadRequest,
			}
		}
	}
	return newsData, int(totalData), nil
}

func (n *NewsService) NewNews(
	news forms.NewsDTO,
	file *multipart.FileHeader,
	claims *Claims,
) (primitive.ObjectID, *ErrorRes) {
	// Validate unique slug
	var findNews *models.News
	slugNews := slug.MakeLang(news.Title, "es")
	cursor := newsModel.Use().FindOne(db.Ctx, bson.D{
		{
			Key:   "url",
			Value: slugNews,
		},
	})
	cursor.Decode(&findNews)
	if findNews != nil {
		return primitive.NilObjectID, &ErrorRes{
			Err:        fmt.Errorf("el titulo de la noticia ya está en uso"),
			StatusCode: http.StatusConflict,
		}
	}
	// Upload image
	fileDb, err := uploadImage(file)
	if err != nil {
		return primitive.NilObjectID, &ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	// Upload news
	var newsType string
	if claims.UserType == models.STUDENT_DIRECTIVE {
		newsType = "student"
	} else {
		newsType = "global"
	}
	newsData, err := newsModel.NewModel(news, fileDb.ID.OID, slugNews, newsType, claims.ID)
	if err != nil {
		return primitive.NilObjectID, &ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	uploadedNews, err := newsModel.Use().InsertOne(db.Ctx, newsData)
	if err != nil {
		return primitive.NilObjectID, &ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	// Notify news
	nats.PublishEncode("notify/global", &res.Notify{
		Title: news.Title,
		Link:  fmt.Sprintf("/noticias/%s", newsData.Url),
		Img:   fileDb.Key,
		Type:  newsType,
	})
	return uploadedNews.InsertedID.(primitive.ObjectID), nil
}

func (n *NewsService) UpdateNews(
	data forms.UpdateNewsDTO,
	id string,
	claims *Claims,
) (*models.News, *ErrorRes) {
	idObjectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, &ErrorRes{
			Err:        fmt.Errorf("noticia no encontrada"),
			StatusCode: http.StatusNotFound,
		}
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
		return nil, &ErrorRes{
			Err:        fmt.Errorf("noticia no encontrada"),
			StatusCode: http.StatusNotFound,
		}
	}
	// Verify identity
	if findNews.Type == "global" && (claims.UserType != models.DIRECTIVE && claims.UserType != models.DIRECTOR) {
		return nil, &ErrorRes{
			Err:        fmt.Errorf("Unauthorized"),
			StatusCode: http.StatusUnauthorized,
		}
	}
	if findNews.Type == "student" && claims.UserType != models.STUDENT_DIRECTIVE {
		return nil, &ErrorRes{
			Err:        fmt.Errorf("Unauthorized"),
			StatusCode: http.StatusUnauthorized,
		}
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
			return nil, &ErrorRes{
				Err:        err,
				StatusCode: http.StatusNotFound,
			}
		}
		imgObjectId, err := primitive.ObjectIDFromHex(fileDb.ID.OID)
		if err != nil {
			return nil, &ErrorRes{
				Err:        err,
				StatusCode: http.StatusNotFound,
			}
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
	err = cursor.Decode(&newsData)
	if err != nil {
		return nil, &ErrorRes{
			Err:        err,
			StatusCode: http.StatusNotFound,
		}
	}
	return newsData, nil
}

func (n *NewsService) DeleteNews(
	id string,
	claims *Claims,
) *ErrorRes {
	idObjectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return &ErrorRes{
			Err:        err,
			StatusCode: http.StatusNotFound,
		}
	}
	// Get news
	lookUpStage := n.getLookupFile()
	lookUpUserStage := n.getLookupUser()
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
	newsData, err := n.getNews(mongo.Pipeline{
		matchStage,
		lookUpStage,
		lookUpUserStage,
		projectStage,
	}, false)
	if err != nil {
		return &ErrorRes{
			Err:        err,
			StatusCode: http.StatusNotFound,
		}
	}
	if newsData == nil {
		return &ErrorRes{
			Err:        fmt.Errorf("no existe la noticia"),
			StatusCode: http.StatusNotFound,
		}
	}
	if newsData[0].Type == "global" && (claims.UserType != models.DIRECTIVE && claims.UserType != models.DIRECTOR) {
		return &ErrorRes{
			Err:        fmt.Errorf("Unauthorized"),
			StatusCode: http.StatusUnauthorized,
		}
	}
	if newsData[0].Type == "student" && claims.UserType != models.STUDENT_DIRECTIVE {
		return &ErrorRes{
			Err:        fmt.Errorf("Unauthorized"),
			StatusCode: http.StatusUnauthorized,
		}
	}
	// Delete image
	_, err = nats.Request("delete_image", []byte(newsData[0].Image.ID))
	if err != nil {
		return &ErrorRes{
			Err:        err,
			StatusCode: http.StatusNotFound,
		}
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
	return nil
}

func NewNewsService() *NewsService {
	if newsService == nil {
		newsService = &NewsService{}
	}
	return newsService
}
