package forms

type LoginForm struct {
	Email             string `form:"email" binding:"required,email"`
	EncryptedPassword string `form:"password" binding:"required,min=5"`
	SaveConnection    bool   `form:"saveConnection"`
}
