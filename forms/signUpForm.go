package forms

type SignUpForm struct {
	Email             string `form:"email"`
	FirstName         string `form:"firstName"`
	LastName          string `form:"lastName"`
	EncryptedPassword string `form:"password"`
	Pseudo            string `form:"pseudo"`
}
