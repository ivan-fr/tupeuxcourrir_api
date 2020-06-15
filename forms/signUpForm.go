package forms

type SignUpForm struct {
	Email             string `form:"email" binding:"required,email"`
	FirstName         string `form:"firstName" binding:"required,min=5"`
	LastName          string `form:"lastName" binding:"required,min=5"`
	EncryptedPassword string `form:"password" binding:"required,min=5"`
	Pseudo            string `form:"pseudo" binding:"required,min=5"`
}
