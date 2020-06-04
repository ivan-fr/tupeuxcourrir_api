package orm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func addPrefixToSections(queryBuilder interface{}, prefix string) {
	reflectQueryBuilder := reflect.ValueOf(queryBuilder).Elem()
	var field reflect.Value

	for i := 0; i < reflectQueryBuilder.NumField(); i++ {
		field = reflectQueryBuilder.Field(i)
		if field.Kind() == reflect.String && fmt.Sprintf("%v", field) != "" {
			field.SetString(fmt.Sprintf("%v%v", prefix, field))
		}
	}
}

func getTableName(name string) string {
	return strings.ToLower(fmt.Sprintf("%vs", name))
}

func putIntermediateString(baseSql *string,
	intermediateStringMap string,
	mapIsSetter bool,
	theMap map[string]interface{}) string {

	var newSql = *baseSql
	var format string
	var formatAlternative string

	if mapIsSetter {
		format = "%v %v = '%v'"
		formatAlternative = "%v %v = %v"
	} else {
		format = "%v %v %v"
		formatAlternative = "%v %v"
	}

	var i int
	for key, value := range theMap {
		switch {
		case mapIsSetter:
			switch value.(type) {
			case string:
				if value.(string) == "Now()" {
					newSql = fmt.Sprintf(formatAlternative, newSql, key, "Now()")
				} else {
					newSql = fmt.Sprintf(format, newSql, key, value.(string))
				}
			case int:
				newSql = fmt.Sprintf(formatAlternative, newSql, key, value.(int))
			case bool:
				newSql = fmt.Sprintf(formatAlternative, newSql, key, value.(bool))
			case nil:
				newSql = fmt.Sprintf(formatAlternative, newSql, key, "NULL")
			default:
				panic("undefined type")
			}
		default:
			if value == "" {
				newSql = fmt.Sprintf(formatAlternative, newSql, key)
			} else {
				newSql = fmt.Sprintf(format, newSql, key, value)
			}
		}

		if 0 <= i && i <= (len(theMap)-2) && (0 != len(theMap)-1) {
			newSql = fmt.Sprintf("%v%v", newSql, intermediateStringMap)
		}
		i++
	}

	return newSql
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
