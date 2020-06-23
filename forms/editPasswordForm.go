package forms

type EditPasswordForm struct {
	EncryptedPassword string `form:"password" validate:"required,min=5"`
	ConfirmPassword   string `form:"confirmPassword" validate:"required,min=5"`
}
