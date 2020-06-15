package forms

type ForgotPasswordForm struct {
	Email string `form:"email" binding:"required,email"`
}
