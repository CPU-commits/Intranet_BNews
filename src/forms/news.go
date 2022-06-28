package forms

import "mime/multipart"

type NewsDTO struct {
	Title    string                `form:"title" binding:"required,min=3,max=100"`
	Headline string                `form:"headline" binding:"required,min=3,max=500"`
	Body     string                `form:"body" binding:"required"`
	Img      *multipart.FileHeader `form:"img" binding:"required,file"`
}

type UpdateNewsDTO struct {
	Title    string                `form:"title" binding:"omitempty,min=3,max=100"`
	Headline string                `form:"headline" binding:"omitempty,min=3,max=500"`
	Body     string                `form:"body" binding:"omitempty"`
	Img      *multipart.FileHeader `form:"img" binding:"omitempty,file"`
}
