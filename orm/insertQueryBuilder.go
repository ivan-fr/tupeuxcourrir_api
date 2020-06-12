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
	stmt           []interface{}
}

var singletonIQueryBuilder *InsertQueryBuilder

func (insertQueryBuilder *InsertQueryBuilder) getSQLSectionValuesToInsert(modelValue interface{}) string {
	valueOfModel := reflect.ValueOf(modelValue).Elem()

	var listToInsert = make([]interface{}, 0)

	for j := 1; j < valueOfModel.NumField(); j++ {
		if !isRelationshipField(valueOfModel.Field(j)) {
			var fieldTime, okTime = valueOfModel.Field(j).Interface().(time.Time)

			if okTime {
				switch {
				case strings.Contains(valueOfModel.Type().Field(j).Name, "CreatedAt"):
					listToInsert = append(listToInsert, "Now()")
				case fieldTime.IsZero():
					listToInsert = append(listToInsert, nil)
				default:
					listToInsert = append(listToInsert, fieldTime.Format("YYYY-MM-DD HH:MM:SS"))
				}
			} else {
				switch valueOfModel.Field(j).Kind() {
				case reflect.String:
					listToInsert = append(listToInsert, valueOfModel.Field(j).String())
				case reflect.Int:
					listToInsert = append(listToInsert, valueOfModel.Field(j).Int())
				default:
					panic("unsupported kind of field")
				}
			}
		}
	}

	var str string
	var stmt []interface{}
	str, stmt = constructSQlStmts(",", "space", listToInsert)
	insertQueryBuilder.stmt = append(insertQueryBuilder.stmt, stmt...)
	return fmt.Sprintf("(%v)", str)
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
