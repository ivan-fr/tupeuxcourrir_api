package forms

type LoginForm struct {
	Email             string `form:"email" validate:"required,email"`
	EncryptedPassword string `form:"password" validate:"required,min=5"`
	SaveConnection    bool   `form:"saveConnection"`
}
