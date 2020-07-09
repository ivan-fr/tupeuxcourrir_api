package forms

type PutAddressForm struct {
	City       string `schema:"city" validate:"required"`
	Street     string `schema:"street" validate:"required"`
	PostalCode string `schema:"postalCode" validate:"required"`
}
