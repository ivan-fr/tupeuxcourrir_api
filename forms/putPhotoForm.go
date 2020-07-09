package forms

import "mime/multipart"

type PutPhotoForm struct {
	Photo *multipart.FileHeader `schema:"photoFile" validate:"required"`
}
