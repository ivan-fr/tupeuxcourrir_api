package orm

import (
	"database/sql"
	"fmt"
	"tupeuxcourrir_api/db"
)

type DeleteQueryBuilder struct {
	referenceModel   interface{}
	SectionWhere     string
	SectionWhereStmt []interface{}
}

var singletonDQueryBuilder *DeleteQueryBuilder

func (dQB *DeleteQueryBuilder) ConstructSql() string {
	var theSql = fmt.Sprintf("DELETE FROM %v",
		getTableName(getModelName(dQB.referenceModel)))

	if dQB.SectionWhere == "" {
		panic("no where section")
	}

	addPrefixToSections(dQB, " ", 0)

	return fmt.Sprintf("%v%v;",
		theSql,
		dQB.SectionWhere)
}

func (dQB *DeleteQueryBuilder) Where(mapFilter map[string]interface{}) *DeleteQueryBuilder {
	sSA := &sQLSectionArchitecture{mode: "setter", isStmts: true, intermediateString: " and", context: mapFilter}
	sSA.constructSQlSection()

	dQB.SectionWhereStmt = sSA.valuesFromStmts
	dQB.SectionWhere = fmt.Sprintf("WHERE %v", sSA.SQLSection)

	return dQB
}

func (dQB *DeleteQueryBuilder) SetReferenceModel(model interface{}) *DeleteQueryBuilder {
	dQB.Clean()
	dQB.referenceModel = nil
	dQB.referenceModel = model
	return dQB
}

func (dQB *DeleteQueryBuilder) Clean() {
	dQB.SectionWhere = ""
}

func (dQB *DeleteQueryBuilder) ApplyDelete() (sql.Result, error) {
	connection := db.GetConnectionFromDB()
	return connection.Db.Exec(dQB.ConstructSql(), dQB.SectionWhereStmt...)
}

func GetDeleteQueryBuilder(model interface{}) *DeleteQueryBuilder {
	if singletonDQueryBuilder == nil {
		singletonDQueryBuilder = &DeleteQueryBuilder{}
	}

	singletonDQueryBuilder.SetReferenceModel(model)
	return singletonDQueryBuilder
}
