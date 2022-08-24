package services

import (
	"github.com/CPU-commits/Intranet_BNews/src/aws_s3"
	"github.com/CPU-commits/Intranet_BNews/src/models"
	"github.com/CPU-commits/Intranet_BNews/src/stack"
)

// Models
var newsModel = new(models.NewsModel)
var likesModel = new(models.LikesModel)

var nats = stack.NewNats()
var aws = aws_s3.NewAWSS3()

// Error Response
type ErrorRes struct {
	Err        error
	StatusCode int
}
