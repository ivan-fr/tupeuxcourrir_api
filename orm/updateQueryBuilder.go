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
	referenceModel interface{}
	SectionWhere   string
	SectionSet     string
}

func (updateQueryBuilder *UpdateQueryBuilder) getSetSectionFromRef() {
	sqlConstruct := "SET"

	valueOfRef := reflect.ValueOf(updateQueryBuilder.referenceModel).Elem()
	var mapFilter = make(map[string]string)

	for i := 0; i < valueOfRef.NumField(); i++ {
		if i == 0 {
			continue
		}

		if !isRelationshipField(valueOfRef.Field(i)) {
			timeField, okTime := valueOfRef.Field(i).Interface().(time.Time)
			if okTime {
				switch {
				case strings.Contains(valueOfRef.Type().Field(i).Name, "UpdatedAt"):
					mapFilter[valueOfRef.Type().Field(i).Name] = "Now()"
				case timeField.IsZero():
					mapFilter[valueOfRef.Type().Field(i).Name] = "NULL"
				default:
					mapFilter[valueOfRef.Type().Field(i).Name] = timeField.Format("YYYY-MM-DD HH:MM:SS")
				}
			} else {
				mapFilter[valueOfRef.Type().Field(i).Name] = valueOfRef.Field(i).String()
			}
		}
	}

	updateQueryBuilder.SectionSet = putIntermediateString(
		&sqlConstruct,
		" and",
		true,
		mapFilter)
}

func (updateQueryBuilder *UpdateQueryBuilder) Where(mapFilter map[string]string) *UpdateQueryBuilder {
	sqlConstruct := "WHERE"

	updateQueryBuilder.SectionWhere = putIntermediateString(&sqlConstruct,
		" and",
		true,
		mapFilter)

	return updateQueryBuilder
}

func (updateQueryBuilder *UpdateQueryBuilder) constructSql() string {
	var theSql = fmt.Sprintf("UPDATE %v",
		getTableName(getModelName(updateQueryBuilder.referenceModel)))

	if updateQueryBuilder.SectionWhere == "" {
		panic("no where section")
	}

	if updateQueryBuilder.SectionSet == "" {
		updateQueryBuilder.getSetSectionFromRef()
	}

	addPrefixToSections(updateQueryBuilder, " ")

	return fmt.Sprintf("%v%v%v;",
		theSql,
		updateQueryBuilder.SectionSet,
		updateQueryBuilder.SectionWhere)
}

func (updateQueryBuilder *UpdateQueryBuilder) SetReferenceModel(model interface{}) *UpdateQueryBuilder {
	updateQueryBuilder.referenceModel = nil
	updateQueryBuilder.referenceModel = model
	return updateQueryBuilder
}

func (updateQueryBuilder *UpdateQueryBuilder) ApplyUpdate() (sql.Result, error) {
	connection := db.GetConnectionFromDB()
	return connection.Db.Exec(updateQueryBuilder.constructSql())
}

func NewUpdateQueryBuilder(model interface{}) *UpdateQueryBuilder {
	return &UpdateQueryBuilder{referenceModel: model}
}
