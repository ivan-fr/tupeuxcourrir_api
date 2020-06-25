package forms

import "mime/multipart"

type PutPhotoForm struct {
	Photo *multipart.FileHeader `form:"photoFile" validate:"required"`
}
