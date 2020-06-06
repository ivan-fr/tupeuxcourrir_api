package orm

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
	"tupeuxcourrir_api/db"
)

type InsertQueryBuilder struct {
	referenceModel interface{}
	modelValues    []interface{}
}

var singletonIQueryBuilder *InsertQueryBuilder

func (insertQueryBuilder *InsertQueryBuilder) getSQLSectionValuesToInsert(modelValue interface{}) string {
	valueOfModel := reflect.ValueOf(modelValue).Elem()
	var format1, format2, format string

	sectionValues := "("
	for j := 1; j < valueOfModel.NumField(); j++ {
		format1 = "%v"
		if j == 1 {
			format1 += "'%v'"
		} else {
			format1 += ", '%v'"
		}

		format2 = "%v"
		if j == 1 {
			format2 += "%v"
		} else {
			format2 += ", %v"
		}

		if !isRelationshipField(valueOfModel.Field(j)) {
			var fieldTime, okTime = valueOfModel.Field(j).Interface().(time.Time)

			if okTime {
				var valueToInsert string
				switch {
				case strings.Contains(valueOfModel.Type().Field(j).Name, "CreatedAt"):
					format = format2
					valueToInsert = "Now()"
				case fieldTime.IsZero():
					format = format2
					valueToInsert = "NULL"
				default:
					format = format1
					valueToInsert = fieldTime.Format("YYYY-MM-DD HH:MM:SS")
				}
				sectionValues = fmt.Sprintf(format, sectionValues, valueToInsert)
			} else {
				if valueOfModel.Field(j).Kind() == reflect.String {
					format = format1
				} else {
					format = format2
				}

				sectionValues = fmt.Sprintf(format, sectionValues, valueOfModel.Field(j))
			}
		}
	}
	return sectionValues + ")"
}

func (insertQueryBuilder *InsertQueryBuilder) getSQlValuesToInsert() string {
	var theSql string

	for i, modelValue := range insertQueryBuilder.modelValues {

		sectionValues := insertQueryBuilder.getSQLSectionValuesToInsert(modelValue)

		switch {
		case 0 == i:
			theSql = fmt.Sprintf("%v %v", theSql, sectionValues)
		case 1 <= i && i <= (len(insertQueryBuilder.modelValues)-1):
			theSql = fmt.Sprintf("%v, %v", theSql, sectionValues)
		}
	}

	return theSql
}

func (insertQueryBuilder *InsertQueryBuilder) getSqlColumnNamesToInsert() string {
	typeOfRef := reflect.ValueOf(insertQueryBuilder.referenceModel).Elem()

	sectionColumn := "("
	for i := 1; i < typeOfRef.NumField(); i++ {
		var format string

		if i == 1 {
			format = "%v%v"
		} else {
			format = "%v, %v"
		}

		if !isRelationshipField(typeOfRef.Field(i)) {
			sectionColumn = fmt.Sprintf(format, sectionColumn, typeOfRef.Type().Field(i).Name)
		}
	}
	sectionColumn += ")"

	return sectionColumn
}

func (insertQueryBuilder *InsertQueryBuilder) constructSql() string {
	if len(insertQueryBuilder.modelValues) == 0 {
		return ""
	}

	var theSql = fmt.Sprintf("INSERT INTO %v %v VALUES",
		getTableName(getModelName(insertQueryBuilder.referenceModel)),
		insertQueryBuilder.getSqlColumnNamesToInsert())

	return fmt.Sprintf("%v %v;",
		theSql,
		insertQueryBuilder.getSQlValuesToInsert())
}

func (insertQueryBuilder *InsertQueryBuilder) SetReferenceModel(model interface{}) *InsertQueryBuilder {
	insertQueryBuilder.Clean()
	insertQueryBuilder.referenceModel = nil
	insertQueryBuilder.referenceModel = model
	return insertQueryBuilder
}

func (insertQueryBuilder *InsertQueryBuilder) Clean() {
	insertQueryBuilder.modelValues = nil
}

func (insertQueryBuilder *InsertQueryBuilder) ApplyInsert() (sql.Result, error) {
	connection := db.GetConnectionFromDB()
	defer insertQueryBuilder.Clean()
	return connection.Db.Exec(insertQueryBuilder.constructSql())
}

func GetInsertQueryBuilder(model interface{}, modelsValues []interface{}) *InsertQueryBuilder {
	if singletonIQueryBuilder == nil {
		singletonIQueryBuilder = &InsertQueryBuilder{}
	}

	singletonIQueryBuilder.SetReferenceModel(model)
	singletonIQueryBuilder.modelValues = modelsValues
	return singletonIQueryBuilder
}
