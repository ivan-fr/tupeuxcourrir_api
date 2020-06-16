package forms

type SignUpForm struct {
	Email             string `form:"email" validate:"required,email"`
	FirstName         string `form:"firstName" validate:"required,min=3"`
	LastName          string `form:"lastName" validate:"required,min=3"`
	EncryptedPassword string `form:"password" validate:"required,min=5"`
	Pseudo            string `form:"pseudo" validate:"required,min=5"`
}
