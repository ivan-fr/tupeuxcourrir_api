package orm

import (
	"fmt"
	"reflect"
	"strings"
)

type sQLSectionArchitecture struct {
	SQLSection      string
	valuesFromStmts []interface{}
	formats         []string

	intermediateString string
	mode               string
	isStmts            bool
	context            interface{}
}

func (sSA *sQLSectionArchitecture) analyseSliceContext() {
	valueOfContext := reflect.ValueOf(sSA.context)

	for i := 0; i < valueOfContext.Len(); i++ {
		if i > 0 {
			sSA.SQLSection = sSA.SQLSection + " "
		}

		switch sSA.mode {
		case "space":
			sSA.analyseSpaceModeFromSlice(valueOfContext.Index(i).Interface())
		}

		sSA.putIntermediate(i)
	}
}

func (sSA *sQLSectionArchitecture) addStmt(stmt interface{}) {
	if !sSA.isStmts {
		panic("the program try to put a stmt whereas the sSA have isStmts to false")
	}

	sSA.valuesFromStmts = append(sSA.valuesFromStmts, stmt)
}

func (sSA *sQLSectionArchitecture) putIntermediate(index int) {
	valueOfContext := reflect.ValueOf(sSA.context)
	if 0 <= index && index <= (valueOfContext.Len()-2) && (0 != valueOfContext.Len()-1) {
		sSA.SQLSection = fmt.Sprintf("%v%v", sSA.SQLSection, sSA.intermediateString)
	}
}

func (sSA *sQLSectionArchitecture) getComparative(keySplit []string) string {
	if len(keySplit) == 1 {
		return getComparativeFormat("")
	}

	return getComparativeFormat(keySplit[len(keySplit)-1])
}

func (sSA *sQLSectionArchitecture) getEffectiveFormats(comparative string) []string {
	sliceToStorage := make([]string, 0)

	if sSA.mode != "space" {
		for i := range sSA.formats {
			if strings.Contains(sSA.formats[i], "%v") {
				sliceToStorage = append(sliceToStorage, fmt.Sprintf(sSA.formats[i], comparative))
			} else {
				sliceToStorage = append(sliceToStorage, sSA.formats[i])
			}
			sliceToStorage[i] = strings.ReplaceAll(sliceToStorage[i], ".", "%v")
		}
	} else {
		sliceToStorage = sSA.formats
	}

	return sliceToStorage
}

func (sSA *sQLSectionArchitecture) analyseMapStringInterfaceContext() {
	var keySplit []string
	var columnName string

	var i int
	for key, value := range sSA.context.(H) {
		if i > 0 {
			sSA.SQLSection = sSA.SQLSection + " "
		}

		keySplit = strings.Split(key, "__")

		comparative := sSA.getComparative(keySplit)
		effectiveFormat := sSA.getEffectiveFormats(comparative)

		if len(keySplit) > 1 {
			columnName = strings.Join(keySplit[:len(keySplit)-1], "__")
		} else {
			columnName = keySplit[0]
		}

		switch sSA.mode {
		case "setter":
			sSA.analyseSetterMode(columnName, value, comparative, effectiveFormat)
		case "aggregate":
			sSA.analyseAggregateMode(columnName, value, comparative, effectiveFormat)
		case "space":
			sSA.analyseSpaceModeFromMap(columnName, value, effectiveFormat)
		default:
			panic("undefined mode")
		}

		sSA.putIntermediate(i)
		i++
	}
}

func (sSA *sQLSectionArchitecture) analyseSetterMode(columnName string, value interface{}, comparative string, formats []string) {
	var checkSlice = false

	switch value.(type) {
	case string:
		if value.(string) == "Now()" {
			sSA.SQLSection = fmt.Sprintf(formats[0], sSA.SQLSection, columnName, value.(string))
		} else {
			sSA.SQLSection = fmt.Sprintf(formats[0], sSA.SQLSection, columnName, "?")
			sSA.addStmt(value.(string))
		}
	case int:
		sSA.SQLSection = fmt.Sprintf(formats[0], sSA.SQLSection, columnName, "?")
		sSA.addStmt(value.(int))
	case bool:
		sSA.SQLSection = fmt.Sprintf(formats[0], sSA.SQLSection, columnName, "?")
		sSA.addStmt(value.(bool))
	case nil:
		formats[0] = strings.Replace(formats[0], "=", "IS", 1)
		sSA.SQLSection = fmt.Sprintf(formats[0], sSA.SQLSection, columnName, "NULL")
	default:
		checkSlice = true
	}

	if checkSlice && comparative == "IN" {
		valueOfValue := reflect.ValueOf(value)
		if valueOfValue.Type().Kind() == reflect.Slice {
			sSASub := &sQLSectionArchitecture{mode: "space",
				isStmts:            true,
				intermediateString: ",",
				context:            valueOfValue.Interface()}
			sSASub.constructSQlSection()

			sSA.SQLSection = fmt.Sprintf(formats[1], sSA.SQLSection, columnName, sSASub.SQLSection)
			sSA.valuesFromStmts = append(sSA.valuesFromStmts, sSASub.valuesFromStmts...)
		}
	}
}

func (sSA *sQLSectionArchitecture) analyseAggregateMode(aggregateFunction string, value interface{}, comparative string, formats []string) {
	if _, ok := value.(string); ok {
		sSA.SQLSection = fmt.Sprintf(formats[0], sSA.SQLSection, aggregateFunction, value.(string))
	} else {
		valueOfValue := reflect.ValueOf(value)
		if valueOfValue.Type().Kind() == reflect.Slice && valueOfValue.Len() == 2 {
			columnName := valueOfValue.Index(0).Interface().(string)
			vToCompare := valueOfValue.Index(1).Interface()
			checkSlice := false
			switch vToCompare.(type) {
			case int:
				sSA.SQLSection = fmt.Sprintf(formats[1], sSA.SQLSection, aggregateFunction, columnName, "?")
				sSA.addStmt(vToCompare.(int))
			case nil:
				if comparative == "=" {
					formats[1] = strings.Replace(formats[1], "=", "IS", 1)
				}
				sSA.SQLSection = fmt.Sprintf(formats[1], sSA.SQLSection, aggregateFunction, columnName, "NULL")
			case string:
				sSA.SQLSection = fmt.Sprintf(formats[1], sSA.SQLSection, aggregateFunction, columnName, "?")
				sSA.addStmt(vToCompare.(string))
			default:
				checkSlice = true
			}

			if comparative == "IN" && checkSlice {
				valueOfVToCompare := valueOfValue.Index(1).Interface()
				if reflect.TypeOf(valueOfVToCompare).Kind() == reflect.Slice {
					sSASub := &sQLSectionArchitecture{mode: "space",
						isStmts:            true,
						intermediateString: ",",
						context:            valueOfVToCompare}
					sSASub.constructSQlSection()

					sSA.SQLSection = fmt.Sprintf(formats[2], sSA.SQLSection, aggregateFunction, columnName, sSASub.SQLSection)
					sSA.valuesFromStmts = append(sSA.valuesFromStmts, sSASub.valuesFromStmts...)
				}
			}
		}
	}
}

func (sSA *sQLSectionArchitecture) analyseSpaceModeFromMap(columnName string, value interface{}, formats []string) {
	switch value.(type) {
	case string:
		if value == "" {
			sSA.SQLSection = fmt.Sprintf(formats[1], sSA.SQLSection, columnName)
		} else {
			sSA.SQLSection = fmt.Sprintf(formats[0], sSA.SQLSection, columnName, value.(string))
		}
	default:
		panic("undefined type from space")
	}
}

func (sSA *sQLSectionArchitecture) analyseSpaceModeFromSlice(value interface{}) {
	switch value.(type) {
	case string:
		switch value.(string) {
		case "Now()":
			sSA.SQLSection = fmt.Sprintf(sSA.formats[1], sSA.SQLSection, "Now()")
		default:
			if sSA.isStmts {
				sSA.SQLSection = fmt.Sprintf(sSA.formats[1], sSA.SQLSection, "?")
				sSA.addStmt(value.(string))
			} else {
				sSA.SQLSection = fmt.Sprintf(sSA.formats[1], sSA.SQLSection, value.(string))
			}
		}
	case nil:
		sSA.SQLSection = fmt.Sprintf(sSA.formats[1], sSA.SQLSection, "NULL")
	case int:
		sSA.SQLSection = fmt.Sprintf(sSA.formats[1], sSA.SQLSection, "?")
		sSA.addStmt(value.(int))
	case bool:
		sSA.SQLSection = fmt.Sprintf(sSA.formats[1], sSA.SQLSection, "?")
		sSA.addStmt(value.(bool))
	default:
		panic("undefined type from space")
	}
}

func (sSA *sQLSectionArchitecture) setFormatsFromMode() {
	switch sSA.mode {
	case "setter":
		sSA.formats = []string{".. %v .", ".. %v (.)"}
	case "space":
		sSA.formats = []string{"%v%v %v", "%v%v"}
	case "aggregate":
		sSA.formats = []string{"..(.)", "..(.) %v .", "..(.) %v (.)"}
	default:
		panic("undefined mode from map")
	}
}

func (sSA *sQLSectionArchitecture) constructSQlSection() {
	sSA.setFormatsFromMode()

	switch sSA.context.(type) {
	case H:
		sSA.analyseMapStringInterfaceContext()
	case []interface{}, []string, []int:
		sSA.analyseSliceContext()
	default:
		panic("undefined context type")
	}
}
