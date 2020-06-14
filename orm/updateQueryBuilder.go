package orm

import (
	"database/sql"
	"fmt"
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
}

func (uQB *UpdateQueryBuilder) getSetSectionFromRef() {
	valueOfRef := reflect.ValueOf(uQB.referenceModel).Elem()
	var mapFilter = make(H)

	for i := 1; i < valueOfRef.NumField(); i++ {

		if !isRelationshipField(valueOfRef.Field(i)) {
			timeField, okTime := valueOfRef.Field(i).Interface().(time.Time)
			if okTime {
				switch {
				case strings.Contains(valueOfRef.Type().Field(i).Name, "UpdatedAt"):
					mapFilter[valueOfRef.Type().Field(i).Name] = "Now()"
				case timeField.IsZero():
					mapFilter[valueOfRef.Type().Field(i).Name] = nil
				default:
					mapFilter[valueOfRef.Type().Field(i).Name] = timeField.Format("YYYY-MM-DD HH:MM:SS")
				}
			} else {
				mapFilter[valueOfRef.Type().Field(i).Name] = valueOfRef.Field(i).Interface()
			}
		}
	}

	sSA := &sQLSectionArchitecture{mode: "setter", intermediateString: ",", context: mapFilter, isStmts: true}
	sSA.constructSQlSection()
	uQB.SectionSetStmt = sSA.valuesFromStmts
	uQB.SectionSet = fmt.Sprintf("SET %v", sSA.SQLSection)
}

func (uQB *UpdateQueryBuilder) ConstructSql() string {
	var theSql = fmt.Sprintf("UPDATE %v",
		getTableName(getModelName(uQB.referenceModel)))

	if uQB.SectionWhere == "" {
		panic("no where section")
	}

	if uQB.SectionSet == "" {
		uQB.getSetSectionFromRef()
	}

	addPrefixToSections(uQB, " ", 0)

	return fmt.Sprintf("%v%v%v;",
		theSql,
		uQB.SectionSet,
		uQB.SectionWhere)
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
