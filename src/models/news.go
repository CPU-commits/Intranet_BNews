package models

import (
	"fmt"
	"time"

	"github.com/CPU-commits/Intranet_BNews/src/forms"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const NEWS_COLLECTION = "news"

type News struct {
	ID         primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	AuthorId   primitive.ObjectID `json:"author_id" bson:"author_id"`
	Title      string             `json:"title" bson:"title"`
	Headline   string             `json:"headline" bson:"headline"`
	Body       string             `json:"body" bson:"body"`
	Img        primitive.ObjectID `json:"img" bson:"img"`
	Url        string             `json:"url" bson:"url"`
	Type       string             `json:"type" bson:"type"`
	Status     bool               `json:"status" bson:"status"`
	UploadDate primitive.DateTime `json:"upload_date" bson:"upload_date"`
	UpdateDate primitive.DateTime `json:"update_date" bson:"update_date"`
}

type NewsModel struct{}

func (news *News) String() string {
	return fmt.Sprintf(
		"_id: %v, author: %v, title: %s, headline: %s, body: %s, img: %s, upload: %v",
		news.ID,
		news.AuthorId,
		news.Title,
		news.Headline,
		news.Body,
		news.Img,
		news.UpdateDate,
	)
}

func init() {
	collections, errC := DbConnect.GetCollections()
	if errC != nil {
		panic(errC)
	}
	for _, collection := range collections {
		if collection == NEWS_COLLECTION {
			return
		}
	}
	var jsonSchema = bson.M{
		"bsonType": "object",
		"required": []string{
			"author_id",
			"title",
			"headline",
			"body",
			"img",
			"url",
			"type",
			"status",
			"upload_date",
			"update_date",
		},
		"properties": bson.M{
			"author_id": bson.M{"bsonType": "objectId"},
			"title": bson.M{
				"bsonType":  "string",
				"maxLength": 100,
			},
			"headline": bson.M{
				"bsonType":  "string",
				"maxLength": 500,
			},
			"body":        bson.M{"bsonType": "string"},
			"img":         bson.M{"bsonType": "objectId"},
			"url":         bson.M{"bsonType": "string"},
			"type":        bson.M{"enum": bson.A{"student", "global"}},
			"status":      bson.M{"bsonType": "bool"},
			"upload_date": bson.M{"bsonType": "date"},
			"update_date": bson.M{"bsonType": "date"},
		},
	}
	var validators = bson.M{
		"$jsonSchema": jsonSchema,
	}
	opts := &options.CreateCollectionOptions{
		Validator: validators,
	}
	err := DbConnect.CreateCollection(NEWS_COLLECTION, opts)
	if err != nil {
		panic(err)
	}
}

func (news *NewsModel) Use() *mongo.Collection {
	return DbConnect.GetCollection(NEWS_COLLECTION)
}

func (news *NewsModel) NewModel(data forms.NewsDTO, imageId, slugNews, typeNews, authorID string) (*News, error) {
	authorObjectId, err := primitive.ObjectIDFromHex(authorID)
	if err != nil {
		return &News{}, err
	}
	imgObjectId, err := primitive.ObjectIDFromHex(imageId)
	if err != nil {
		return &News{}, err
	}
	now := primitive.NewDateTimeFromTime(time.Now())
	return &News{
		AuthorId:   authorObjectId,
		Title:      data.Title,
		Headline:   data.Headline,
		Body:       data.Body,
		Img:        imgObjectId,
		Url:        slugNews,
		Type:       typeNews,
		Status:     true,
		UploadDate: now,
		UpdateDate: now,
	}, nil
}
