package orm

import (
	"database/sql"
	"fmt"
	"tupeuxcourrir_api/db"
)
import "reflect"

type InsertQueryBuilder struct {
	referenceModel interface{}
	modelValues    []interface{}
}

func (insertQueryBuilder *InsertQueryBuilder) constructSql() string {
	if len(insertQueryBuilder.modelValues) > 0 {
		return ""
	}

	var theSql = fmt.Sprintf("INSERT INTO %v VALUES",
		getModelName(getModelName(insertQueryBuilder.referenceModel)))

	var sectionValues string
	var valueOfModel reflect.Value

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
			theSql = fmt.Sprintf("%v %v", theSql, sectionValues)
		case 1 <= i && i <= (len(insertQueryBuilder.modelValues)-2):
			theSql = fmt.Sprintf("%v, %v", theSql, sectionValues)
		case i == valueOfModel.NumField()-1:
			theSql = fmt.Sprintf("%v, %v;", theSql, sectionValues)
		}
	}

	return theSql
}

func (insertQueryBuilder *InsertQueryBuilder) SetReferenceModel(model interface{}) {
	insertQueryBuilder.Clean()
	insertQueryBuilder.referenceModel = nil
	insertQueryBuilder.referenceModel = model
}

func (insertQueryBuilder *InsertQueryBuilder) Clean() {
	insertQueryBuilder.modelValues = nil
}

func (insertQueryBuilder *InsertQueryBuilder) ApplyInsert() (sql.Result, error) {
	connection := db.GetConnectionFromDB()
	defer insertQueryBuilder.Clean()
	return connection.Db.Exec(insertQueryBuilder.constructSql())
}
