package orm

import "fmt"
import "reflect"

type InsertQueryBuilder struct {
	referenceModel interface{}
	modelValues    []interface{}
}

func (insertQueryBuilder *InsertQueryBuilder) constructSql() (string, error) {
	var sql = fmt.Sprintf("INSERT INTO %v VALUES",
		getModelName(getModelName(insertQueryBuilder.referenceModel)))

	var sectionValues string
	var valueOfModel reflect.Value
	var err error

	for i, modelValue := range insertQueryBuilder.modelValues {
		valueOfModel = reflect.ValueOf(modelValue).Elem()
		sectionValues = "(NULL"
		for j := 0; j < valueOfModel.NumField(); j++ {
			var format string
			if valueOfModel.Field(i).Kind() == reflect.String {
				format = "%v, '%v'"
			} else {
				format = "%v, %v"
			}
			if !isRelationshipField(valueOfModel.Field(j)) {
				switch {
				case 0 <= j && j <= (valueOfModel.NumField()-2):
					sectionValues = fmt.Sprintf(format, sectionValues, valueOfModel.Field(j))
				case j == valueOfModel.NumField()-1:
					sectionValues = fmt.Sprintf(format+")", sectionValues, valueOfModel.Field(j))
				}
			}
		}

		switch {
		case 0 == i:
			sql = fmt.Sprintf("%v %v", sql, sectionValues)
		case 1 <= i && i <= (len(insertQueryBuilder.modelValues)-2):
			sql = fmt.Sprintf("%v, %v", sql, sectionValues)
		case i == valueOfModel.NumField()-1:
			sql = fmt.Sprintf("%v, %v;", sql, sectionValues)
		}
	}

	return sql, err
}
