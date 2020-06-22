package orm

import (
	"database/sql"
	"reflect"
	"time"
)

func adaptNullTypeStruct(fieldDest interface{}, fieldForm reflect.Value) {
	switch fieldDest.(type) {
	case *sql.NullTime:
		nullTime := fieldDest.(*sql.NullTime)
		nullTime.Valid = !(fieldForm.IsZero() || fieldForm.Interface().(time.Time).IsZero())
		nullTime.Time = fieldForm.Interface().(time.Time)
	case *sql.NullString:
		nullStr := fieldDest.(*sql.NullString)
		nullStr.Valid = !fieldForm.IsZero()
		nullStr.String = fieldForm.String()
	case *sql.NullInt64:
		nullInt := fieldDest.(*sql.NullInt64)
		nullInt.Valid = !fieldForm.IsZero()
		nullInt.Int64 = fieldForm.Int()
	case *sql.NullBool:
		nullBool := fieldDest.(*sql.NullBool)
		nullBool.Valid = !fieldForm.IsZero()
		nullBool.Bool = fieldForm.Bool()
	default:
		panic("only accept sql.Null* for struct")
	}
}

func BindForm(destModel, form interface{}) {
	valueOfForm := reflect.ValueOf(form).Elem()
	valueOfDest := reflect.ValueOf(destModel).Elem()

	for i := 0; i < valueOfForm.NumField(); i++ {
		modelFieldConcern := valueOfDest.FieldByName(valueOfForm.Type().Field(i).Name)
		if !modelFieldConcern.IsValid() {
			continue
		}

		switch modelFieldConcern.Kind() {
		case reflect.Struct:
			adaptNullTypeStruct(modelFieldConcern.Addr().Interface(), valueOfForm.Field(i))
		default:
			modelFieldConcern.Set(valueOfForm.Field(i))
		}
	}
}
