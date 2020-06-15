package utils

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"strings"
)

type FieldError struct {
	Err validator.FieldError
}

func (fE *FieldError) String() string {
	var sb strings.Builder

	sb.WriteString("validation failed on field '" + fE.Err.Field() + "'")
	sb.WriteString(", condition: " + fE.Err.ActualTag())

	// Print condition parameters, e.g. oneof=red blue -> { red blue }
	if fE.Err.Param() != "" {
		sb.WriteString(" { " + fE.Err.Param() + " }")
	}

	if fE.Err.Value() != nil && fE.Err.Value() != "" {
		sb.WriteString(fmt.Sprintf(", actual: %v", fE.Err.Value()))
	}

	return sb.String()
}
