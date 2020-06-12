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

var singletonUQueryBuilder *UpdateQueryBuilder

func (updateQueryBuilder *UpdateQueryBuilder) getSetSectionFromRef() {
	valueOfRef := reflect.ValueOf(updateQueryBuilder.referenceModel).Elem()
	var mapFilter = make(map[string]interface{})

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

	var str string
	str, updateQueryBuilder.SectionSetStmt = constructSQlStmts(
		",",
		"setter",
		mapFilter)
	updateQueryBuilder.SectionSet = fmt.Sprintf("SET %v", str)
}

func (updateQueryBuilder *UpdateQueryBuilder) ConstructSql() string {
	var theSql = fmt.Sprintf("UPDATE %v",
		getTableName(getModelName(updateQueryBuilder.referenceModel)))

	if updateQueryBuilder.SectionWhere == "" {
		panic("no where section")
	}

	if updateQueryBuilder.SectionSet == "" {
		updateQueryBuilder.getSetSectionFromRef()
	}

	addPrefixToSections(updateQueryBuilder, " ", 0)

	return fmt.Sprintf("%v%v%v;",
		theSql,
		updateQueryBuilder.SectionSet,
		updateQueryBuilder.SectionWhere)
}

func (updateQueryBuilder *UpdateQueryBuilder) Where(mapFilter map[string]interface{}) *UpdateQueryBuilder {
	var str string
	str, updateQueryBuilder.SectionWhereStmt = constructSQlStmts(
		" and",
		"setter",
		mapFilter)
	updateQueryBuilder.SectionWhere = fmt.Sprintf("WHERE %v", str)

	return updateQueryBuilder
}

func (updateQueryBuilder *UpdateQueryBuilder) SetReferenceModel(model interface{}) *UpdateQueryBuilder {
	updateQueryBuilder.Clean()
	updateQueryBuilder.referenceModel = nil
	updateQueryBuilder.referenceModel = model
	return updateQueryBuilder
}

func (updateQueryBuilder *UpdateQueryBuilder) Clean() {
	updateQueryBuilder.SectionWhere = ""
	updateQueryBuilder.SectionSet = ""
}

func (updateQueryBuilder *UpdateQueryBuilder) ApplyUpdate() (sql.Result, error) {
	connection := db.GetConnectionFromDB()
	return connection.Db.Exec(updateQueryBuilder.ConstructSql())
}

func GetUpdateQueryBuilder(model interface{}) *UpdateQueryBuilder {
	if singletonUQueryBuilder == nil {
		singletonUQueryBuilder = &UpdateQueryBuilder{}
	}

	singletonUQueryBuilder.SetReferenceModel(model)
	return singletonUQueryBuilder
}
