package orm

import (
	"fmt"
	"reflect"
	"tupeuxcourrir_api/db"
	"tupeuxcourrir_api/models"
)

type SelectQueryBuilder struct {
	QueryApplier
	SectionWhere  string
	SectionOrder  string
	SectionSelect string
	SectionFrom   string
	SectionLimit  string
	SectionOffset string
	SectionJoin   []string
}

func (queryBuilder *SelectQueryBuilder) putIntermediateString(baseSql *string,
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

func (queryBuilder *SelectQueryBuilder) addPrefixToSections(prefix string) {
	reflectQueryBuilder := reflect.ValueOf(queryBuilder).Elem()
	var field reflect.Value

	for i := 0; i < reflectQueryBuilder.NumField(); i++ {
		field = reflectQueryBuilder.Field(i)
		if field.Kind() == reflect.String && fmt.Sprintf("%v", field) != "" {
			field.SetString(fmt.Sprintf("%v%v", prefix, field))
		}
	}
}

func (queryBuilder *SelectQueryBuilder) constructSql() string {
	queryBuilder.SectionFrom = fmt.Sprintf("FROM %v", getTableName(getModelName(queryBuilder.model)))

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

func (queryBuilder *SelectQueryBuilder) addMTO(fieldInterface interface{}) {
	relationship := fieldInterface.(*models.ManyToOneRelationShip)
	target := reflect.ValueOf(relationship.Target)
	stringJoin := fmt.Sprintf("INNER JOIN %v ON %v.%v = %v.%v",
		getTableName(target.Elem().Type().Name()),
		getTableName(getModelName(queryBuilder.model)),
		relationship.AssociateColumn,
		getTableName(target.Elem().Type().Name()),
		getPKFieldSelfCOLUMNTagFromModel(target.Interface()))
	queryBuilder.SectionJoin = append(queryBuilder.SectionJoin, stringJoin)
}

func (queryBuilder *SelectQueryBuilder) addOTM(fieldInterface interface{}) {
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
		getTableName(target.Type().Name()),
		getTableName(getModelName(queryBuilder.model)),
		getPKFieldSelfCOLUMNTagFromModel(queryBuilder.model),
		getTableName(target.Type().Name()),
		targetAssociatedColumn)
	queryBuilder.SectionJoin = append(queryBuilder.SectionJoin, stringJoin)
}

func (queryBuilder *SelectQueryBuilder) addMTM(fieldInterface interface{}) {
	relationship := fieldInterface.(*models.ManyToManyRelationShip)
	target := reflect.ValueOf(relationship.Target)
	intermediateTarget := reflect.ValueOf(relationship.IntermediateTarget).Elem()

	stringJoin := fmt.Sprintf("LEFT JOIN %v ON %v.%v = %v.%v INNER JOIN %v ON %v.%v = %v.%v",
		getTableName(intermediateTarget.Type().Name()),
		getTableName(getModelName(queryBuilder.model)),
		getPKFieldSelfCOLUMNTagFromModel(queryBuilder.model),
		getTableName(intermediateTarget.Type().Name()),
		getAssociatedColumnFromReverse(queryBuilder.model, intermediateTarget),

		getTableName(target.Elem().Type().Name()),
		getTableName(intermediateTarget.Type().Name()),
		getAssociatedColumnFromReverse(target.Interface(), intermediateTarget),
		getTableName(target.Elem().Type().Name()),
		getPKFieldSelfCOLUMNTagFromModel(target.Interface()))
	queryBuilder.SectionJoin = append(queryBuilder.SectionJoin, stringJoin)
}

func (queryBuilder *SelectQueryBuilder) OrderBy(orderFilter map[string]string) *SelectQueryBuilder {
	sqlConstruct := "ORDER BY"

	queryBuilder.SectionOrder = queryBuilder.putIntermediateString(&sqlConstruct,
		",",
		false,
		orderFilter)

	return queryBuilder
}

func (queryBuilder *SelectQueryBuilder) Limit(limit string) *SelectQueryBuilder {
	queryBuilder.SectionLimit = fmt.Sprintf("LIMIT %v", limit)
	return queryBuilder
}

func (queryBuilder *SelectQueryBuilder) FindBy(mapFilter map[string]string) *SelectQueryBuilder {
	sqlConstruct := "WHERE"

	queryBuilder.SectionWhere = queryBuilder.putIntermediateString(&sqlConstruct,
		" and",
		true,
		mapFilter)

	return queryBuilder
}

func (queryBuilder *SelectQueryBuilder) Clear() {
	queryBuilder.SectionOffset = ""
	queryBuilder.SectionLimit = ""
	queryBuilder.SectionOrder = ""
	queryBuilder.SectionWhere = ""
	queryBuilder.SectionFrom = ""
	queryBuilder.SectionSelect = ""
}

func (queryBuilder *SelectQueryBuilder) Consider(fieldName string) {
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

func (queryBuilder *SelectQueryBuilder) ApplyQuery() ([][]ModelsOrderedToScan, error) {
	connection := db.GetConnectionFromDB()
	defer queryBuilder.Clear()

	var modelList [][]ModelsOrderedToScan
	rows, err := connection.Db.Query(queryBuilder.constructSql())

	if err == nil {
		var newModel []ModelsOrderedToScan
		for rows.Next() {
			newModel, err = queryBuilder.hydrate(rows.Scan)

			if err != nil {
				break
			}
			modelList = append(modelList, newModel)
		}
	}

	return modelList, err
}

func (queryBuilder *SelectQueryBuilder) ApplyQueryRow() ([]ModelsOrderedToScan, error) {
	connection := db.GetConnectionFromDB()
	defer queryBuilder.Clear()

	row := connection.Db.QueryRow(queryBuilder.constructSql())
	newModel, err := queryBuilder.hydrate(row.Scan)

	return newModel, err
}
