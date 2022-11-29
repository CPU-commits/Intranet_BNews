package forms

import "mime/multipart"

type NewsDTO struct {
	Title    string                `form:"title" binding:"required,min=3,max=100" validate:"required" minimum:"3" maximum:"100"`
	Headline string                `form:"headline" binding:"required,min=3,max=500" validate:"required" minimum:"3" maximum:"500"`
	Body     string                `form:"body" binding:"required" validate:"required"`
	Img      *multipart.FileHeader `form:"img" binding:"required,file" validate:"required" swaggertype:"string" format:"binary"`
}

type UpdateNewsDTO struct {
	Title    string                `form:"title" binding:"omitempty,min=3,max=100" validate:"optional" minimum:"3" maximum:"100"`
	Headline string                `form:"headline" binding:"omitempty,min=3,max=500" validate:"optional" minimum:"3" maximum:"500"`
	Body     string                `form:"body" binding:"omitempty" validate:"optional"`
	Img      *multipart.FileHeader `form:"img" binding:"omitempty,file" validate:"optional" swaggertype:"string" format:"binary"`
}
