package forms

type PutAddressForm struct {
	City       string `form:"city" validate:"required"`
	Street     string `form:"street" validate:"required"`
	PostalCode string `form:"postalCode" validate:"required"`
}
