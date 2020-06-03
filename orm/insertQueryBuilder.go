package orm

import (
	"database/sql"
	"fmt"
	"time"
	"tupeuxcourrir_api/db"
)
import "reflect"

type InsertQueryBuilder struct {
	referenceModel interface{}
	modelValues    []interface{}
}

func (insertQueryBuilder *InsertQueryBuilder) getSQLSectionValuesToInsert(modelValue interface{}) string {
	valueOfModel := reflect.ValueOf(modelValue).Elem()

	sectionValues := "(NULL"
	for j := 0; j < valueOfModel.NumField(); j++ {
		if j == 0 {
			continue
		}

		var format string
		var fieldTime, okTime = valueOfModel.Field(j).Interface().(time.Time)

		if valueOfModel.Field(j).Kind() == reflect.String || okTime {
			format = "%v, '%v'"
		} else {
			format = "%v, %v"
		}

		if !isRelationshipField(valueOfModel.Field(j)) {
			if okTime {
				sectionValues = fmt.Sprintf(format, sectionValues, fieldTime.String())
			} else {
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

func (insertQueryBuilder *InsertQueryBuilder) ConstructSql() string {
	if len(insertQueryBuilder.modelValues) == 0 {
		return ""
	}

	var theSql = fmt.Sprintf("INSERT INTO %v VALUES",
		getTableName(getModelName(insertQueryBuilder.referenceModel)))

	return fmt.Sprintf("%v %v;", theSql, insertQueryBuilder.getSQlValuesToInsert())
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
	return connection.Db.Exec(insertQueryBuilder.ConstructSql())
}

func NewInsertQueryBuilder(model interface{}, modelsValues []interface{}) *InsertQueryBuilder {
	return &InsertQueryBuilder{referenceModel: model, modelValues: modelsValues}
}
