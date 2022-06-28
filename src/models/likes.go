package models

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const LIKES_COLLECTION = "likes"

type Likes struct {
	ID     primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	NewsID primitive.ObjectID `json:"news" bson:"news"`
	UserID primitive.ObjectID `json:"user" bson:"user"`
}

type LikesModel struct{}

func init() {
	collections, errC := DbConnect.GetCollections()
	if errC != nil {
		panic(errC)
	}
	for _, collection := range collections {
		if collection == LIKES_COLLECTION {
			return
		}
	}
	var jsonSchema = bson.M{
		"bsonType": "object",
		"required": []string{
			"news",
			"user",
		},
		"properties": bson.M{
			"news": bson.M{"bsonType": "objectId"},
			"user": bson.M{"bsonType": "objectId"},
		},
	}
	var validators = bson.M{
		"$jsonSchema": jsonSchema,
	}
	opts := &options.CreateCollectionOptions{
		Validator: validators,
	}
	err := DbConnect.CreateCollection(LIKES_COLLECTION, opts)
	if err != nil {
		panic(err)
	}
}

func (likes *LikesModel) Use() *mongo.Collection {
	return DbConnect.GetCollection(LIKES_COLLECTION)
}

func (likes *LikesModel) NewModel(userId, newsId primitive.ObjectID) *Likes {
	return &Likes{
		UserID: userId,
		NewsID: newsId,
	}
}
