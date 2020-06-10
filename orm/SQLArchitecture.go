package orm

import (
	"fmt"
	"reflect"
	"strings"
)

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
	formats []string,
	stmts bool) (string, []interface{}) {

	var newSql string
	var valueStmts []interface{}
	var valueStmt interface{}

	valueOfContext := reflect.ValueOf(context)

	if valueOfContext.Type().Kind() == reflect.Slice {
		for i := 0; i < valueOfContext.Len(); i++ {
			if i > 0 {
				newSql = newSql + " "
			}
			switch mapSetterMode {
			case "space":
				if stmts {
					newSql, valueStmt = analyseSpaceModeFromSlice(newSql, valueOfContext.Index(i).Interface(), formats)
					putStmtToASlice(&valueStmts, valueStmt)
				} else {
					newSql = analyseSpaceModeFromSliceNoStmt(newSql, valueOfContext.Index(i).Interface(), formats)
				}
			}

			if 0 <= i && i <= (valueOfContext.Len()-2) &&
				(0 != valueOfContext.Len()-1) {
				newSql = fmt.Sprintf("%v%v", newSql, intermediateStringMap)
			}
		}
	}

	return newSql, valueStmts
}

func putStmtToASlice(slice *[]interface{}, stmt interface{}) {
	if stmt == nil {
		return
	}

	reflectValueOfStmt := reflect.ValueOf(stmt)

	if !reflectValueOfStmt.IsZero() {
		if reflectValueOfStmt.Type().Kind() == reflect.Slice {
			for i := 0; i < reflectValueOfStmt.Len(); i++ {
				if !reflectValueOfStmt.Index(i).IsZero() {
					*slice = append(*slice, reflectValueOfStmt.Index(i).Interface())
				}
			}
		} else {
			*slice = append(*slice, stmt)
		}
	}
}

func analyseMapStringInterfaceContext(context interface{},
	intermediateStringMap,
	mapSetterMode string,
	formats []string) (string, []interface{}) {

	var effectiveFormat []string
	var newSql string
	var keySplit []string
	var columnName string
	var comparative string
	var valueStmts []interface{}
	var valueStmt interface{}

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

		effectiveFormat = make([]string, 0)
		if mapSetterMode != "space" {
			for i := range formats {
				if len(keySplit) == 1 {
					comparative = getComparativeFormat("")
				} else {
					comparative = getComparativeFormat(keySplit[len(keySplit)-1])
				}
				if strings.Contains(formats[i], "%v") {
					effectiveFormat = append(effectiveFormat, fmt.Sprintf(formats[i], comparative))
				} else {
					effectiveFormat = append(effectiveFormat, formats[i])
				}
				effectiveFormat[i] = strings.ReplaceAll(effectiveFormat[i], ".", "%v")
			}
		} else {
			effectiveFormat = formats
		}

		switch mapSetterMode {
		case "setter":
			newSql, valueStmt = analyseSetterMode(newSql, columnName, value, comparative, effectiveFormat)
			putStmtToASlice(&valueStmts, valueStmt)
		case "aggregate":
			newSql, valueStmt = analyseAggregateMode(newSql, columnName, value, comparative, effectiveFormat)
			putStmtToASlice(&valueStmts, valueStmt)
		case "space":
			newSql, valueStmt = analyseSpaceModeFromMap(newSql, columnName, value, effectiveFormat)
			putStmtToASlice(&valueStmts, valueStmt)
		default:
			panic("undefined mode")
		}

		if 0 <= i && i <= (len(context.(map[string]interface{}))-2) &&
			(0 != len(context.(map[string]interface{}))-1) {
			newSql = fmt.Sprintf("%v%v", newSql, intermediateStringMap)
		}
		i++
	}
	return newSql, valueStmts
}

func analyseSetterMode(sql, columnName string, value interface{}, comparative string, formats []string) (string, interface{}) {
	var checkSlice = false

	switch value.(type) {
	case string:
		if value.(string) == "Now()" {
			return fmt.Sprintf(formats[0], sql, columnName, value.(string)), nil
		} else {
			return fmt.Sprintf(formats[0], sql, columnName, "?"), value.(string)
		}
	case int:
		return fmt.Sprintf(formats[0], sql, columnName, "?"), value.(int)
	case bool:
		return fmt.Sprintf(formats[0], sql, columnName, "?"), value.(bool)
	case nil:
		formats[0] = strings.Replace(formats[0], "=", "IS", 1)
		return fmt.Sprintf(formats[0], sql, columnName, "NULL"), nil
	default:
		checkSlice = true
	}

	if checkSlice && comparative == "IN" {
		valueOfValue := reflect.ValueOf(value)
		if valueOfValue.Type().Kind() == reflect.Slice {
			str, stmts := ConstructSQlStmts(",", "space", valueOfValue.Interface())
			return fmt.Sprintf(formats[1], sql, columnName, str), stmts
		}
	}

	panic("unsupported/wrong value type from setter")
}

func analyseAggregateMode(sql, aggregateFunction string, value interface{}, comparative string, formats []string) (string, interface{}) {
	if _, ok := value.(string); ok {
		return fmt.Sprintf(formats[0], sql, aggregateFunction, value.(string)), nil
	} else {
		valueOfValue := reflect.ValueOf(value)
		if valueOfValue.Type().Kind() == reflect.Slice && valueOfValue.Len() == 2 {
			columnName := valueOfValue.Index(0).Interface().(string)
			vToCompare := valueOfValue.Index(1).Interface()
			checkSlice := false
			switch vToCompare.(type) {
			case int:
				return fmt.Sprintf(formats[1], sql, aggregateFunction, columnName, "?"), vToCompare.(int)
			case nil:
				if comparative == "=" {
					formats[1] = strings.Replace(formats[1], "=", "IS", 1)
				}
				return fmt.Sprintf(formats[1], sql, aggregateFunction, columnName, "NULL"), nil
			case string:
				return fmt.Sprintf(formats[1], sql, aggregateFunction, columnName, "?"), vToCompare.(string)
			default:
				checkSlice = true
			}

			if comparative == "IN" && checkSlice {
				valueOfVToCompare := reflect.ValueOf(vToCompare)
				if valueOfVToCompare.Type().Kind() == reflect.Slice {
					str, stmts := ConstructSQlStmts(",", "space", valueOfVToCompare.Interface())
					return fmt.Sprintf(formats[1], sql, columnName, str), stmts
				}
			}
		}
	}

	panic("Wrong configuration")
}

func analyseSpaceModeFromMap(sql, columnName string, value interface{}, formats []string) (string, interface{}) {
	switch value.(type) {
	case string:
		if value == "" {
			return fmt.Sprintf(formats[1], sql, columnName), nil
		} else {
			return fmt.Sprintf(formats[0], sql, columnName, value.(string)), nil
		}
	default:
		panic("undefined type from space")
	}
}

func analyseSpaceModeFromSlice(sql,
	value interface{},
	formats []string) (string, interface{}) {
	switch value.(type) {
	case string:
		switch value.(string) {
		case "Now()":
			return fmt.Sprintf(formats[1], sql, "Now()"), nil
		default:
			return fmt.Sprintf(formats[1], sql, "?"), value.(string)
		}
	case nil:
		return fmt.Sprintf(formats[1], sql, "NULL"), nil
	case int:
		return fmt.Sprintf(formats[1], sql, "?"), value.(int)
	default:
		panic("undefined type from space")
	}
}

func analyseSpaceModeFromSliceNoStmt(sql,
	value interface{},
	formats []string) string {
	switch value.(type) {
	case string:
		return fmt.Sprintf(formats[1], sql, value.(string))
	default:
		panic("undefined type from space")
	}
}

func getFormatsMode(mapSetterMode string) []string {
	switch mapSetterMode {
	case "setter":
		return []string{".. %v .", ".. %v (.)"}
	case "space":
		return []string{"%v%v %v", "%v%v"}
	case "aggregate":
		return []string{"..(.)", "..(.) %v .", "..(.) %v (.)"}
	default:
		panic("undefined mode from map")
	}
}

func ConstructSQlStmts(intermediateStringMap string,
	mapSetterMode string,
	context interface{}) (string, []interface{}) {

	var formats = getFormatsMode(mapSetterMode)

	switch context.(type) {
	case map[string]interface{}:
		return analyseMapStringInterfaceContext(context, intermediateStringMap, mapSetterMode, formats)
	case []string, []int:
		return analyseSliceContext(context, intermediateStringMap, mapSetterMode, formats, true)
	default:
		panic("undefined context type")
	}
}

func ConstructSQlSpaceNoStmts(intermediateStringMap string,
	context interface{}) string {

	var formats = getFormatsMode("space")

	switch context.(type) {
	case []string, []int:
		str, _ := analyseSliceContext(context, intermediateStringMap, "space", formats, false)
		return str
	default:
		panic("undefined context type")
	}
}
