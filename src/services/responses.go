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
	ID  string `json:"_id" bson:"_id" example:"638660ca141aa4ee9faf07e8"`
	URL string `json:"url" bson:"url" example:"https://repository.com/file/$dsK2!1"`
	Key string `bson:"key" example:"$dsK2!1"`
}

type NewsResponse struct {
	Author     models.User        `json:"author,omitempty" bson:"author,omitempty" extensions:"x-omitempty"`
	Headline   string             `json:"headline" bson:"headline" example:"Example..."`
	Title      string             `json:"title" bson:"title" example:"Title !!"`
	Image      Image              `json:"image" bson:"image"`
	UploadDate primitive.DateTime `json:"upload_date" bson:"upload_date" swaggertype:"string" example:"2022-09-21T20:10:23.309+00:00"`
	UpdateDate primitive.DateTime `json:"update_date" bson:"update:date" swaggertype:"string" example:"2022-09-21T20:10:23.309+00:00"`
	URL        string             `json:"url" bson:"url" example:"title"`
	Type       string             `json:"type" bson:"type" example:"global" enum:"global,student"`
	Body       string             `json:"body" bson:"body" example:"This is a body..."`
	Status     bool               `json:"status" bson:"status"`
	Like       bool               `json:"like"`
	Likes      int                `json:"likes" example:"10"`
	ID         string             `json:"_id" bson:"_id" example:"638660ca141aa4ee9faf07e8"`
}
