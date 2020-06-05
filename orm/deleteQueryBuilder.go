package orm

import (
	"database/sql"
	"fmt"
	"tupeuxcourrir_api/db"
)

type DeleteQueryBuilder struct {
	referenceModel interface{}
	SectionWhere   string
}

var singletonDQueryBuilder *DeleteQueryBuilder

func (deleteQueryBuilder *DeleteQueryBuilder) ConstructSql() string {
	var theSql = fmt.Sprintf("DELETE FROM %v",
		getTableName(getModelName(deleteQueryBuilder.referenceModel)))

	if deleteQueryBuilder.SectionWhere == "" {
		panic("no where section")
	}

	addPrefixToSections(deleteQueryBuilder, " ")

	return fmt.Sprintf("%v%v;",
		theSql,
		deleteQueryBuilder.SectionWhere)
}

func (deleteQueryBuilder *DeleteQueryBuilder) Where(mapFilter map[string]interface{}) *DeleteQueryBuilder {
	sqlConstruct := "WHERE"

	deleteQueryBuilder.SectionWhere = putIntermediateString(&sqlConstruct,
		" and",
		true,
		mapFilter)

	return deleteQueryBuilder
}

func (deleteQueryBuilder *DeleteQueryBuilder) SetReferenceModel(model interface{}) *DeleteQueryBuilder {
	deleteQueryBuilder.Clean()
	deleteQueryBuilder.referenceModel = nil
	deleteQueryBuilder.referenceModel = model
	return deleteQueryBuilder
}

func (deleteQueryBuilder *DeleteQueryBuilder) Clean() {
	deleteQueryBuilder.SectionWhere = ""
}

func (deleteQueryBuilder *DeleteQueryBuilder) ApplyDelete() (sql.Result, error) {
	connection := db.GetConnectionFromDB()
	return connection.Db.Exec(deleteQueryBuilder.ConstructSql())
}

func GetDeleteQueryBuilder(model interface{}) *DeleteQueryBuilder {
	if singletonDQueryBuilder == nil {
		singletonDQueryBuilder = &DeleteQueryBuilder{}
	}

	singletonDQueryBuilder.SetReferenceModel(model)
	return singletonDQueryBuilder
}
