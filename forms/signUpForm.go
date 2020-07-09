package forms

type SignUpForm struct {
	Email             string `schema:"email" validate:"required,email"`
	FirstName         string `schema:"firstName" validate:"required,min=3"`
	LastName          string `schema:"lastName" validate:"required,min=3"`
	EncryptedPassword string `schema:"password" validate:"required,min=5"`
	Pseudo            string `schema:"pseudo" validate:"required,min=5"`
}
