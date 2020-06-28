package orm

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"
)

func getAbbreviation(str string) string {
	var abbr = make([]string, 0)

	for _, char := range str {
		if unicode.IsUpper(char) && unicode.IsLetter(char) {
			abbr = append(abbr, fmt.Sprintf("%c", char))
		}
	}

	return strings.Join(abbr, "")
}

func addPrefixToSections(queryBuilder interface{}, prefix string, startIndexStringField int) {
	reflectQueryBuilder := reflect.ValueOf(queryBuilder).Elem()
	var field reflect.Value

	iStringField := 0
	for i := 0; i < reflectQueryBuilder.NumField(); i++ {
		field = reflectQueryBuilder.Field(i)
		if field.Kind() == reflect.String {
			if fmt.Sprintf("%v", field) != "" && iStringField >= startIndexStringField {
				field.SetString(fmt.Sprintf("%v%v", prefix, field))
			}
			iStringField++
		}
	}
}

func getTableName(name string) string {
	return strings.ToLower(fmt.Sprintf("%vs", name))
}

func getPKFieldNameFromModel(model interface{}) string {
	reflectModel := reflect.TypeOf(model).Elem()
	var field reflect.StructField

	var ormTags []string
	var isPk bool

	for i := 0; i < reflectModel.NumField(); i++ {
		field = reflectModel.Field(i)
		if v, ok := field.Tag.Lookup("orm"); ok {
			ormTags = strings.Split(v, ";")

			for _, vOfData := range ormTags {
				if vOfData == "PK" {
					isPk = true
					break
				}
			}

			if isPk {
				return field.Name
			}
		}
	}

	panic("no pk model tag")
}

func getAssociatedColumnFromReverse(target interface{}, targetStructFields reflect.Value) string {
	var associatedColumn string
	typeOfApplierQueryModel := reflect.TypeOf(target).Elem()
	var field reflect.Value
	for i := 0; i < targetStructFields.NumField(); i++ {
		field = targetStructFields.Field(i)
		if field.Kind() == reflect.Ptr {
			if v, ok := field.Interface().(*ManyToOneRelationShip); ok {
				if reflect.TypeOf(v.Target).Elem() == typeOfApplierQueryModel {
					associatedColumn = v.AssociateColumn
					break
				}
			}
		}
	}

	return associatedColumn
}

func newModel(model interface{}) interface{} {
	modelValue := reflect.TypeOf(model)
	if modelValue.Kind() == reflect.Ptr {
		modelValue = modelValue.Elem()
	}

	return reflect.New(modelValue).Interface()
}

func getAddrFieldsToScan(model interface{}) ([]interface{}, error) {
	reflectModel := reflect.ValueOf(model)
	if reflectModel.Kind() != reflect.Ptr {
		return make([]interface{}, 0, 0),
			errors.New("must pass a pointer, not a value")
	}

	reflectModel = reflectModel.Elem()
	fieldsTab := make([]interface{}, 0)
	var field reflect.Value

	for i := 0; i < reflectModel.NumField(); i++ {
		field = reflectModel.Field(i)
		if !isRelationshipField(field) {
			fieldsTab = append(fieldsTab, field.Addr().Interface())
		}
	}

	return fieldsTab, nil
}

func isRelationshipField(field reflect.Value) bool {
	_, ok := field.Interface().(*ManyToManyRelationShip)
	_, ok1 := field.Interface().(*OneToManyRelationShip)
	_, ok2 := field.Interface().(*ManyToOneRelationShip)

	return ok || ok1 || ok2
}

func getModelName(model interface{}) string {
	typeof := reflect.TypeOf(model)

	if typeof.Kind() == reflect.Ptr {
		typeof = typeof.Elem()
	}

	return typeof.Name()
}

func getComparativeFormat(comparative string) string {
	switch comparative {
	case "":
		return "="
	case "in":
		return "IN"
	case "gt":
		return ">"
	case "gte":
		return ">="
	case "lt":
		return "<"
	case "lte":
		return "<="
	default:
		panic("undefined format")
	}
}

func getNullFieldInstance(field interface{}) interface{} {
	switch field.(type) {
	case string:
		return sql.NullString{}
	case bool:
		return sql.NullBool{}
	case int:
		return sql.NullInt64{}
	case float64:
		return sql.NullFloat64{}
	case time.Time:
		return sql.NullTime{}
	default:
		return nil
	}
}
