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

type UpdateQueryBuilder struct {
	referenceModel   interface{}
	SectionWhere     string
	SectionWhereStmt []interface{}

	SectionSet     string
	SectionSetStmt []interface{}

	SectionIncrement     string
	SectionIncrementStmt []interface{}
}

func (uQB *UpdateQueryBuilder) valueToInsertFromStringCase(str string) interface{} {
	if str == "" {
		return nil
	} else {
		return str
	}
}

func (uQB *UpdateQueryBuilder) valueToInsertFromTimeCase(_time time.Time) interface{} {
	switch {
	case _time.IsZero():
		return nil
	default:
		return _time
	}
}

func (uQB *UpdateQueryBuilder) valueToInsertFromStructCase(fieldName string, value interface{}) interface{} {
	switch value.(type) {
	case time.Time:
		if strings.Contains(fieldName, "UpdatedAt") {
			return "Now()"
		}
		return uQB.valueToInsertFromTimeCase(value.(time.Time))
	case sql.NullTime:
		if strings.Contains(fieldName, "UpdatedAt") {
			return "Now()"
		}

		nullTime := value.(sql.NullTime)
		if nullTime.Valid {
			_time, _ := nullTime.Value()
			return uQB.valueToInsertFromTimeCase(_time.(time.Time))
		}
	case sql.NullString:
		nullStr := value.(sql.NullString)

		if nullStr.Valid {
			_str, _ := nullStr.Value()
			return uQB.valueToInsertFromStringCase(_str.(string))
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
		panic("only accept sql.Null* for struct")
	}

	return nil
}

func (uQB *UpdateQueryBuilder) getSetSectionFromRef() {
	valueOfRef := reflect.ValueOf(uQB.referenceModel).Elem()
	var mapFilter = make(H)

	for j := 1; j < valueOfRef.NumField(); j++ {
		if !isRelationshipField(valueOfRef.Field(j)) {
			switch valueOfRef.Field(j).Kind() {
			case reflect.Struct:
				mapFilter[valueOfRef.Type().Field(j).Name] = uQB.valueToInsertFromStructCase(
					valueOfRef.Type().Field(j).Name,
					valueOfRef.Field(j).Interface())
			default:
				mapFilter[valueOfRef.Type().Field(j).Name] = valueOfRef.Field(j).Interface()
			}
		}
	}

	sSA := &sQLSectionArchitecture{mode: "fullSetter", intermediateString: ",", context: mapFilter, isStmts: true}
	sSA.constructSQlSection()
	uQB.SectionSetStmt = sSA.valuesFromStmts
	uQB.SectionSet = fmt.Sprintf("SET %v", sSA.SQLSection)
}

func (uQB *UpdateQueryBuilder) Increment(incrementMap H) *UpdateQueryBuilder {
	sSA := &sQLSectionArchitecture{intermediateString: ",", isStmts: true, mode: "increment", context: incrementMap}
	sSA.constructSQlSection()

	uQB.SectionIncrement = sSA.SQLSection
	uQB.SectionIncrementStmt = sSA.valuesFromStmts

	return uQB
}

func (uQB *UpdateQueryBuilder) ConstructSql() string {
	var theSQL = fmt.Sprintf("UPDATE %v",
		getTableName(getModelName(uQB.referenceModel)))

	if uQB.SectionWhere == "" {
		pkFieldName := getPKFieldNameFromModel(uQB.referenceModel)
		pkFieldValue := reflect.ValueOf(uQB.referenceModel).FieldByName(pkFieldName)

		if pkFieldValue.IsZero() {
			panic("your model haven't already an ID")
		}

		uQB.Where(And(H{pkFieldName: pkFieldValue}))
	}

	if uQB.SectionSet == "" {
		uQB.getSetSectionFromRef()
	}

	addPrefixToSections(uQB, " ", 0)

	_sql := fmt.Sprintf("%v%v%v%v;",
		theSQL,
		uQB.SectionSet,
		uQB.SectionIncrement,
		uQB.SectionWhere)
	log.Println(_sql)
	return _sql
}

func (uQB *UpdateQueryBuilder) Where(logical *Logical) *UpdateQueryBuilder {
	var str string
	str, uQB.SectionWhereStmt = logical.GetSentence("setter")
	uQB.SectionWhere = fmt.Sprintf("WHERE %v", str)

	return uQB
}

func (uQB *UpdateQueryBuilder) SetReferenceModel(model interface{}) *UpdateQueryBuilder {
	uQB.Clean()
	uQB.referenceModel = nil
	uQB.referenceModel = model
	return uQB
}

func (uQB *UpdateQueryBuilder) Clean() {
	uQB.SectionWhere = ""
	uQB.SectionSet = ""
	uQB.SectionWhereStmt = nil
	uQB.SectionSetStmt = nil
}

func (uQB *UpdateQueryBuilder) GetStmts() []interface{} {
	var stmtsInterface = make([]interface{}, 0)
	stmtsInterface = append(stmtsInterface, uQB.SectionSetStmt...)
	stmtsInterface = append(stmtsInterface, uQB.SectionIncrementStmt...)
	stmtsInterface = append(stmtsInterface, uQB.SectionWhereStmt...)
	return stmtsInterface
}

func (uQB *UpdateQueryBuilder) ApplyUpdate() (sql.Result, error) {
	connection := db.GetConnectionFromDB()
	return connection.Db.Exec(uQB.ConstructSql(), uQB.GetStmts()...)
}

func GetUpdateQueryBuilder(model interface{}) *UpdateQueryBuilder {
	uQB := &UpdateQueryBuilder{}
	uQB.SetReferenceModel(model)
	return uQB
}
