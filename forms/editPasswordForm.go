package forms

type EditPasswordForm struct {
	EncryptedPassword string `schema:"password" validate:"required,min=5"`
	ConfirmPassword   string `schema:"confirmPassword" validate:"required,min=5"`
}
