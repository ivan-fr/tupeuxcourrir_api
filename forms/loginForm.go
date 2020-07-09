package forms

type LoginForm struct {
	Email             string `schema:"email" validate:"required,email"`
	EncryptedPassword string `schema:"password" validate:"required,min=5"`
	SaveConnection    bool   `schema:"saveConnection"`
}
