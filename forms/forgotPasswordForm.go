package forms

type ForgotPasswordForm struct {
	Email string `schema:"email" validate:"required,email"`
}
