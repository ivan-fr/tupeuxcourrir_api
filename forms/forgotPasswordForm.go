package forms

type ForgotPasswordForm struct {
	Email string `form:"email" validate:"required,email"`
}
