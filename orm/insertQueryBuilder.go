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

func (iQB *InsertQueryBuilder) valueToInsertFromStringCase(str string) interface{} {
	if str == "" {
		return nil
	} else {
		return str
	}
}

func (iQB *InsertQueryBuilder) valueToInsertFromTimeCase(fieldName string, time time.Time) interface{} {
	switch {
	case strings.Contains(fieldName, "CreatedAt"):
		return "Now()"
	case time.IsZero():
		return nil
	default:
		return time.Format("YYYY-MM-DD HH:MM:SS")
	}
}

func (iQB *InsertQueryBuilder) valueToInsertFromStructCase(fieldName string, value interface{}) interface{} {
	switch value.(type) {
	case sql.NullTime:
		nullTime := value.(sql.NullTime)
		if nullTime.Valid {
			_time, _ := nullTime.Value()
			return iQB.valueToInsertFromTimeCase(fieldName, _time.(time.Time))
		}
	case sql.NullString:
		nullStr := value.(sql.NullString)

		if nullStr.Valid {
			_str, _ := nullStr.Value()
			return iQB.valueToInsertFromStringCase(_str.(string))
		}
	case sql.NullInt64:
		nullInt := value.(sql.NullInt64)

		if nullInt.Valid {
			_int, _ := nullInt.Value()
			return _int.(int)
		}
	case sql.NullBool:
		nullBool := value.(sql.NullBool)

		if nullBool.Valid {
			_bool, _ := nullBool.Value()
			return _bool.(bool)
		}
	default:
		panic("only accept sql.Null* for struct Time")
	}

	return nil
}

func (iQB *InsertQueryBuilder) getSQLSectionValuesToInsert(modelValue interface{}) string {
	valueOfModel := reflect.ValueOf(modelValue).Elem()

	var sliceToInsert = make([]interface{}, 0)

	for j := 1; j < valueOfModel.NumField(); j++ {
		if !isRelationshipField(valueOfModel.Field(j)) {
			var fieldTime, okTime = valueOfModel.Field(j).Interface().(time.Time)

			if okTime {
				sliceToInsert = append(sliceToInsert, iQB.valueToInsertFromTimeCase(
					valueOfModel.Type().Field(j).Name, fieldTime))
			} else {
				switch valueOfModel.Field(j).Kind() {
				case reflect.String:
					sliceToInsert = append(sliceToInsert,
						iQB.valueToInsertFromStringCase(valueOfModel.Field(j).String()))
				case reflect.Int:
					sliceToInsert = append(sliceToInsert, valueOfModel.Field(j).Int())
				case reflect.Bool:
					sliceToInsert = append(sliceToInsert, valueOfModel.Field(j).Bool())
				case reflect.Struct:
					sliceToInsert = append(sliceToInsert,
						iQB.valueToInsertFromStructCase(valueOfModel.Type().Field(j).Name,
							valueOfModel.Field(j).Interface()))
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
			theSql = fmt.Sprintf("%v", sectionValues)
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
