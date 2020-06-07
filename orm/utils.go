package orm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

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

func analyseSliceContext(context interface{},
	intermediateStringMap,
	mapSetterMode string,
	formats []string) string {
	var newSql string

	switch context.(type) {
	case []int:
		for i, value := range context.([]int) {
			if i > 0 {
				newSql = newSql + " "
			}
			switch mapSetterMode {
			case "space":
				newSql = analyseSpaceModeFromSlice(newSql, value, formats)
			}

			if 0 <= i && i <= (len(context.([]int))-2) &&
				(0 != len(context.([]int))-1) {
				newSql = fmt.Sprintf("%v%v", newSql, intermediateStringMap)
			}
			i++
		}
	case []string:
		for i, value := range context.([]string) {
			if i > 0 {
				newSql = newSql + " "
			}
			switch mapSetterMode {
			case "space":
				newSql = analyseSpaceModeFromSlice(newSql, value, formats)
			}

			if 0 <= i && i <= (len(context.([]string))-2) &&
				(0 != len(context.([]string))-1) {
				newSql = fmt.Sprintf("%v%v", newSql, intermediateStringMap)
			}
			i++
		}
	}

	return newSql
}

func analyseMapStringInterfaceContext(context interface{},
	intermediateStringMap,
	mapSetterMode string,
	formats []string) string {
	var newSql string
	var keySplit []string
	var columnName string
	var comparative string

	var i int
	for key, value := range context.(map[string]interface{}) {
		if i > 0 {
			newSql = newSql + " "
		}

		keySplit = strings.Split(key, "__")

		if len(keySplit) > 1 {
			columnName = strings.Join(keySplit[:len(keySplit)-1], "__")
		} else {
			columnName = keySplit[0]
		}

		for i := range formats {
			if len(keySplit) == 1 {
				comparative = getComparativeFormat("")
			} else {
				comparative = getComparativeFormat(keySplit[len(keySplit)-1])
			}

			formats[i] = fmt.Sprintf(formats[i], comparative)
			formats[i] = strings.ReplaceAll(formats[i], ".", "%v")
		}

		switch mapSetterMode {
		case "setter":
			newSql = analyseSetterMode(newSql, columnName, value, comparative, formats)
		case "aggregate":
			newSql = analyseAggregateMode(newSql, columnName, value, comparative, formats)
		case "space":
			newSql = analyseSpaceModeFromMap(newSql, columnName, value, formats)
		}

		if 0 <= i && i <= (len(context.(map[string]interface{}))-2) &&
			(0 != len(context.(map[string]interface{}))-1) {
			newSql = fmt.Sprintf("%v%v", newSql, intermediateStringMap)
		}
		i++
	}
	return newSql
}

func analyseSetterMode(sql, columnName string, value interface{}, comparative string, formats []string) string {
	var newSql string

	switch value.(type) {
	case string:
		if value.(string) == "Now()" {
			newSql = fmt.Sprintf(formats[1], sql, columnName, "Now()")
		} else {
			newSql = fmt.Sprintf(formats[0], sql, columnName, value.(string))
		}
	case int:
		newSql = fmt.Sprintf(formats[1], sql, columnName, value.(int))
	case bool:
		newSql = fmt.Sprintf(formats[1], sql, columnName, value.(bool))
	case nil:
		newSql = fmt.Sprintf(formats[1], sql, columnName, "NULL")
	case []string:
		if comparative == "IN" {
			newSql = fmt.Sprintf(formats[2], sql, columnName, putIntermediateString(
				",",
				"space",
				value.([]string)))
		} else {
			panic("incompatible comparative")
		}
	case []int:
		if comparative == "IN" {
			newSql = fmt.Sprintf(formats[2], sql, columnName, putIntermediateString(
				",",
				"space",
				value.([]int)))
		} else {
			panic("incompatible comparative")
		}
	default:
		panic("undefined type from setter")
	}

	return newSql
}

func analyseAggregateMode(sql,
	aggregateFunction string,
	value interface{},
	comparative string,
	formats []string) string {
	var newSql string

	switch value.(type) {
	case string:
		newSql = fmt.Sprintf(formats[0], sql, aggregateFunction, value.(string))
	case map[string]interface{}:
		if len(value.(map[string]interface{})) == 1 {
			for columnName, vToCompare := range value.(map[string]interface{}) {
				switch vToCompare.(type) {
				case int:
					newSql = fmt.Sprintf(formats[1], sql, aggregateFunction, columnName, vToCompare.(int))
				case string:
					newSql = fmt.Sprintf(formats[2], sql, aggregateFunction, columnName, vToCompare.(string))
				case []string:
					if comparative == "IN" {
						newSql = fmt.Sprintf(formats[3], sql, aggregateFunction, columnName, putIntermediateString(
							",",
							"space",
							vToCompare.([]string)))
					} else {
						panic("incompatible comparative")
					}
				case []int:
					if comparative == "IN" {
						newSql = fmt.Sprintf(formats[3], sql, aggregateFunction, columnName, putIntermediateString(
							",",
							"space",
							vToCompare.([]int)))
					} else {
						panic("incompatible comparative")
					}
				}
			}
		} else {
			panic("undefined configuration value")
		}
	default:
		panic("undefined type from aggregate")
	}

	return newSql
}

func analyseSpaceModeFromMap(sql,
	columnName string,
	value interface{},
	formats []string) string {
	var newSql string
	switch value.(type) {
	case string:
		if value == "" {
			panic("use a slice of your keys")
		} else {
			newSql = fmt.Sprintf(formats[0], sql, columnName, value)
		}
	default:
		panic("undefined type from space")
	}
	return newSql
}

func analyseSpaceModeFromSlice(sql,
	value interface{},
	formats []string) string {
	var newSql string
	switch value.(type) {
	case string:
		newSql = fmt.Sprintf(formats[1], sql, value)
	case int:
		newSql = fmt.Sprintf(formats[2], sql, value)
	default:
		panic("undefined type from space")
	}
	return newSql
}

func putIntermediateString(intermediateStringMap string,
	mapSetterMode string,
	context interface{}) string {

	var formats = make([]string, 0)

	switch mapSetterMode {
	case "setter":
		formats = []string{".. %v '.'", ".. %v .", ".. %v (.)"}
	case "space":
		formats = []string{"%v%v %v", "%v'%v'", "%v%v"}
	case "aggregate":
		formats = []string{"..(.)", "..(.) %v .", "..(.) %v '.'", "..(.) %v (.)"}
	default:
		panic("undefined mode from map")
	}

	switch context.(type) {
	case map[string]interface{}:
		return analyseMapStringInterfaceContext(context, intermediateStringMap, mapSetterMode, formats)
	case []string, []int:
		return analyseSliceContext(context, intermediateStringMap, mapSetterMode, formats)
	default:
		panic("undefined context type")
	}
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
