package services

import (
	"fmt"
	"net/http"

	"github.com/CPU-commits/Intranet_BNews/src/db"
	"github.com/CPU-commits/Intranet_BNews/src/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var likesService *LikesServices

type LikesServices struct{}

func (l *LikesServices) LikeNews(
	idNews string,
	claims *Claims,
) *ErrorRes {
	newsObjectId, err := primitive.ObjectIDFromHex(idNews)
	if err != nil {
		return &ErrorRes{
			StatusCode: http.StatusBadRequest,
			Err:        err,
		}
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
		return &ErrorRes{
			Err:        fmt.Errorf("Noticia no encontrada"),
			StatusCode: http.StatusNotFound,
		}
	}
	// Toogle like
	userObjectID, err := primitive.ObjectIDFromHex(claims.ID)
	if err != nil {
		return &ErrorRes{
			StatusCode: http.StatusBadRequest,
			Err:        err,
		}
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
			return &ErrorRes{
				StatusCode: http.StatusBadRequest,
				Err:        err,
			}
		}
	} else {
		likesModel.Use().FindOneAndDelete(db.Ctx, bson.D{
			{
				Key:   "_id",
				Value: hasLike.ID,
			},
		})
	}
	return nil
}

func NewLikesService() *LikesServices {
	if likesService == nil {
		likesService = &LikesServices{}
	}
	return likesService
}
