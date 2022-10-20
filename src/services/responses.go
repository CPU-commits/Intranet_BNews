package services

import (
	"github.com/CPU-commits/Intranet_BNews/src/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UploadNewsNats struct {
	Title    string `json:"title"`
	Headline string `json:"headline"`
	Body     string `json:"body"`
	Author   string `json:"author"`
	Img      string `json:"img"`
}

type Image struct {
	ID  string `json:"_id" bson:"_id"`
	URL string `json:"url" bson:"url"`
	Key string `bson:"key"`
}

type NewsResponse struct {
	Author     models.User        `json:"author,omitempty" bson:"author,omitempty"`
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
