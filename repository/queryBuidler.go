package repository

import (
	"fmt"
	"reflect"
	"tupeuxcourrir_api/db"
)

type QueryBuilder struct {
	QueryApplier
	SectionWhere  string
	SectionOrder  string
	SectionSelect string
	SectionFrom   string
	SectionLimit  string
	SectionOffset string
	SectionJoin   string
}

func (queryBuilder *QueryBuilder) putIntermediateString(baseSql *string,
	intermediateStringMap string,
	mapIsSetter bool,
	theMap map[string]string) string {

	var newSql = *baseSql
	var format string

	if mapIsSetter {
		format = "%v %v='%v'"
	} else {
		format = "%v %v %v"
	}

	var i int
	for key, value := range theMap {
		newSql = fmt.Sprintf(format, newSql, key, value)
		if 0 <= i && i <= (len(theMap)-2) {
			newSql = fmt.Sprintf("%v%v", newSql, intermediateStringMap)
		}
		i++
	}

	return newSql
}

func (queryBuilder *QueryBuilder) addPrefixToSections(prefix string) {
	reflectQueryBuilder := reflect.ValueOf(queryBuilder).Elem()
	var field reflect.Value

	for i := 0; i < reflectQueryBuilder.NumField(); i++ {
		field = reflectQueryBuilder.Field(i)
		if field.Kind() == reflect.String && fmt.Sprintf("%v", field) != "" {
			field.SetString(fmt.Sprintf("%v%v", prefix, field))
		}
	}
}

func (queryBuilder *QueryBuilder) constructSql() string {
	modelName := reflect.TypeOf(queryBuilder.model).Name()
	queryBuilder.SectionFrom = fmt.Sprintf("FROM %vs", modelName)

	if queryBuilder.SectionSelect == "" {
		queryBuilder.SectionSelect = "SELECT *"
	}

	queryBuilder.addPrefixToSections(" ")

	return fmt.Sprintf("%v%v%v%v%v%v;", queryBuilder.SectionSelect,
		queryBuilder.SectionFrom,
		queryBuilder.SectionWhere,
		queryBuilder.SectionOrder,
		queryBuilder.SectionLimit,
		queryBuilder.SectionOffset)
}

func (queryBuilder *QueryBuilder) OrderBy(orderFilter map[string]string) *QueryBuilder {
	sqlConstruct := "ORDER BY"

	queryBuilder.SectionOrder = queryBuilder.putIntermediateString(&sqlConstruct,
		",",
		false,
		orderFilter)

	return queryBuilder
}

func (queryBuilder *QueryBuilder) Limit(limit string) *QueryBuilder {
	queryBuilder.SectionLimit = fmt.Sprintf("LIMIT %v", limit)
	return queryBuilder
}

func (queryBuilder *QueryBuilder) FindBy(mapFilter map[string]string) *QueryBuilder {
	sqlConstruct := "WHERE"

	queryBuilder.SectionWhere = queryBuilder.putIntermediateString(&sqlConstruct,
		" and",
		true,
		mapFilter)

	return queryBuilder
}

func (queryBuilder *QueryBuilder) Clear() {
	queryBuilder.SectionOffset = ""
	queryBuilder.SectionLimit = ""
	queryBuilder.SectionOrder = ""
	queryBuilder.SectionWhere = ""
	queryBuilder.SectionFrom = ""
	queryBuilder.SectionSelect = ""
}

func (queryBuilder *QueryBuilder) ApplyQuery() ([]interface{}, error) {
	connection := db.GetConnectionFromDB()
	defer queryBuilder.Clear()

	var modelList []interface{}
	rows, err := connection.Db.Query(queryBuilder.constructSql())
	if err == nil {
		for rows.Next() {
			newModel, err := queryBuilder.hydrateOne(rows.Scan)

			if err != nil {
				break
			}
			modelList = append(modelList, newModel)
		}
	}

	return modelList, err
}

func (queryBuilder *QueryBuilder) ApplyQueryRow() (interface{}, error) {
	connection := db.GetConnectionFromDB()
	defer queryBuilder.Clear()

	row := connection.Db.QueryRow(queryBuilder.constructSql())

	newModel, err := queryBuilder.hydrateOne(row.Scan)

	return newModel, err
}
