package orm

import (
	"database/sql"
	"fmt"
	"log"
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

func (iQB *InsertQueryBuilder) getSQLSectionValuesToInsert(modelValue interface{}) string {
	valueOfModel := reflect.ValueOf(modelValue).Elem()

	var sliceToInsert = make([]interface{}, 0)

	for j := 1; j < valueOfModel.NumField(); j++ {
		if !isRelationshipField(valueOfModel.Field(j)) {
			var fieldTime, okTime = valueOfModel.Field(j).Interface().(time.Time)

			if okTime {
				switch {
				case strings.Contains(valueOfModel.Type().Field(j).Name, "CreatedAt"):
					sliceToInsert = append(sliceToInsert, "Now()")
				case fieldTime.IsZero():
					sliceToInsert = append(sliceToInsert, nil)
				default:
					sliceToInsert = append(sliceToInsert, fieldTime.Format("YYYY-MM-DD HH:MM:SS"))
				}
			} else {
				switch valueOfModel.Field(j).Kind() {
				case reflect.String:
					str := valueOfModel.Field(j).String()

					if str == "" {
						sliceToInsert = append(sliceToInsert, nil)
					} else {
						sliceToInsert = append(sliceToInsert, str)
					}
				case reflect.Int:
					sliceToInsert = append(sliceToInsert, valueOfModel.Field(j).Int())
				case reflect.Bool:
					sliceToInsert = append(sliceToInsert, valueOfModel.Field(j).Bool())
				default:
					panic(valueOfModel.Field(j).Kind())
				}
			}
		}
	}

	sSA := &sQLSectionArchitecture{mode: "space", isStmts: true, intermediateString: ",", context: sliceToInsert}
	sSA.constructSQlSection()

	iQB.stmt = append(iQB.stmt, sSA.valuesFromStmts...)
	return fmt.Sprintf("(%v)", sSA.SQLSection)
}

func (iQB *InsertQueryBuilder) getSQlValuesToInsert() string {
	var theSql string

	for i, modelValue := range iQB.modelValues {

		sectionValues := iQB.getSQLSectionValuesToInsert(modelValue)

		switch {
		case 0 == i:
			theSql = fmt.Sprintf("%v %v", theSql, sectionValues)
		case 1 <= i && i <= (len(iQB.modelValues)-1):
			theSql = fmt.Sprintf("%v, %v", theSql, sectionValues)
		}
	}

	return theSql
}

func (iQB *InsertQueryBuilder) getSqlColumnNamesToInsert() string {
	typeOfRef := reflect.ValueOf(iQB.referenceModel).Elem()

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

func (iQB *InsertQueryBuilder) constructSql() string {
	if len(iQB.modelValues) == 0 {
		return ""
	}

	var theSql = fmt.Sprintf("INSERT INTO %v %v VALUES",
		getTableName(getModelName(iQB.referenceModel)),
		iQB.getSqlColumnNamesToInsert())

	_sql := fmt.Sprintf("%v %v;",
		theSql,
		iQB.getSQlValuesToInsert())
	log.Println(_sql)
	return _sql
}

func (iQB *InsertQueryBuilder) SetReferenceModel(model interface{}) *InsertQueryBuilder {
	iQB.Clean()
	iQB.referenceModel = nil
	iQB.referenceModel = model
	return iQB
}

func (iQB *InsertQueryBuilder) Clean() {
	iQB.modelValues = nil
	iQB.stmt = nil
}

func (iQB *InsertQueryBuilder) ApplyInsert() (sql.Result, error) {
	connection := db.GetConnectionFromDB()
	defer iQB.Clean()
	return connection.Db.Exec(iQB.constructSql(), iQB.stmt...)
}

func GetInsertQueryBuilder(model interface{}, modelsValues ...interface{}) *InsertQueryBuilder {
	iQB := &InsertQueryBuilder{}
	iQB.SetReferenceModel(model)
	iQB.modelValues = modelsValues
	return iQB
}
