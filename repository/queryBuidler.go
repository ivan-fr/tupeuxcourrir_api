package repository

import (
	"fmt"
	"reflect"
	"strings"
	"tupeuxcourrir_api/db"
	"tupeuxcourrir_api/models"
)

type QueryBuilder struct {
	QueryApplier
	SectionWhere  string
	SectionOrder  string
	SectionSelect string
	SectionFrom   string
	SectionLimit  string
	SectionOffset string
	SectionJoin   []string
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
	queryBuilder.SectionFrom = fmt.Sprintf("FROM %v", queryBuilder.getTableName(queryBuilder.getModelName()))

	queryBuilder.addPrefixToSections(" ")
	queryBuilder.SectionSelect = "SELECT *"

	var joins string
	for _, join := range queryBuilder.SectionJoin {
		joins = fmt.Sprintf("%v %v", joins, join)
	}

	return fmt.Sprintf("%v%v%v%v%v%v%v;", queryBuilder.SectionSelect,
		queryBuilder.SectionFrom,
		joins,
		queryBuilder.SectionWhere,
		queryBuilder.SectionOrder,
		queryBuilder.SectionLimit,
		queryBuilder.SectionOffset)
}

func (queryBuilder *QueryBuilder) getTableName(name string) string {
	return strings.ToLower(fmt.Sprintf("%vs", name))
}

func (queryBuilder *QueryBuilder) addMTO(fieldInterface interface{}) {
	relationship := fieldInterface.(*models.ManyToOneRelationShip)
	target := reflect.ValueOf(relationship.Target)
	stringJoin := fmt.Sprintf("INNER JOIN %v ON %v.%v = %v.%v",
		queryBuilder.getTableName(target.Elem().Type().Name()),
		queryBuilder.getTableName(queryBuilder.getModelName()),
		relationship.AssociateColumn,
		queryBuilder.getTableName(target.Elem().Type().Name()),
		getPKFieldSelfCOLUMNTagFromModel(target.Interface()))
	queryBuilder.SectionJoin = append(queryBuilder.SectionJoin, stringJoin)
}

func (queryBuilder *QueryBuilder) addOTM(fieldInterface interface{}) {
	relationship := fieldInterface.(*models.OneToManyRelationShip)
	target := reflect.ValueOf(relationship.Target).Elem()

	var targetAssociatedColumn string

	if relationship.FieldMTO != "" {
		targetMTO := target.FieldByName(relationship.FieldMTO).Interface().(*models.ManyToOneRelationShip)
		targetAssociatedColumn = targetMTO.AssociateColumn
	} else {
		targetAssociatedColumn = getAssociatedColumnFromReverse(queryBuilder.model, target)
	}

	stringJoin := fmt.Sprintf("LEFT JOIN %v ON %v.%v = %v.%v",
		queryBuilder.getTableName(target.Type().Name()),
		queryBuilder.getTableName(queryBuilder.getModelName()),
		getPKFieldSelfCOLUMNTagFromModel(queryBuilder.model),
		queryBuilder.getTableName(target.Type().Name()),
		targetAssociatedColumn)
	queryBuilder.SectionJoin = append(queryBuilder.SectionJoin, stringJoin)
}

func (queryBuilder *QueryBuilder) addMTM(fieldInterface interface{}) {
	relationship := fieldInterface.(*models.ManyToManyRelationShip)
	target := reflect.ValueOf(relationship.Target)
	intermediateTarget := reflect.ValueOf(relationship.IntermediateTarget).Elem()

	stringJoin := fmt.Sprintf("LEFT JOIN %v ON %v.%v = %v.%v INNER JOIN %v ON %v.%v = %v.%v",
		queryBuilder.getTableName(intermediateTarget.Type().Name()),
		queryBuilder.getTableName(queryBuilder.getModelName()),
		getPKFieldSelfCOLUMNTagFromModel(queryBuilder.model),
		queryBuilder.getTableName(intermediateTarget.Type().Name()),
		getAssociatedColumnFromReverse(queryBuilder.model, intermediateTarget),

		queryBuilder.getTableName(target.Elem().Type().Name()),
		queryBuilder.getTableName(intermediateTarget.Type().Name()),
		getAssociatedColumnFromReverse(target.Interface(), intermediateTarget),
		queryBuilder.getTableName(target.Elem().Type().Name()),
		getPKFieldSelfCOLUMNTagFromModel(target.Interface()))
	queryBuilder.SectionJoin = append(queryBuilder.SectionJoin, stringJoin)
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

func (queryBuilder *QueryBuilder) Consider(fieldName string) {
	reflectQueryBuilder := reflect.ValueOf(queryBuilder.model).Elem()
	fieldInterface := reflectQueryBuilder.FieldByName(fieldName).Interface()

	if queryBuilder.addRelationship(fieldInterface) {
		switch fieldInterface.(type) {
		case *models.ManyToOneRelationShip:
			queryBuilder.addMTO(fieldInterface)
		case *models.OneToManyRelationShip:
			queryBuilder.addOTM(fieldInterface)
		case *models.ManyToManyRelationShip:
			queryBuilder.addMTM(fieldInterface)
		}
	}
}

func (queryBuilder *QueryBuilder) ApplyQuery() ([]interface{}, error) {
	connection := db.GetConnectionFromDB()
	defer queryBuilder.Clear()

	var modelList []interface{}
	rows, err := connection.Db.Query(queryBuilder.constructSql())

	if err == nil {
		var newModel interface{}
		for rows.Next() {
			newModel, err = queryBuilder.hydrateOne(rows.Scan)

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
