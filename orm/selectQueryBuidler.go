package orm

import (
	"fmt"
	"reflect"
	"tupeuxcourrir_api/db"
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

func (queryBuilder *SelectQueryBuilder) getAlias(tableName string) string {
	return fmt.Sprintf("%v%v", tableName[0:2], len(queryBuilder.relationshipTargetOrder))
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
	relationship := fieldInterface.(*ManyToOneRelationShip)
	target := reflect.ValueOf(relationship.Target)
	stringJoin := fmt.Sprintf("INNER JOIN %v %v ON %v.%v = %v.%v",
		getTableName(target.Elem().Type().Name()),
		queryBuilder.getAlias(getTableName(target.Elem().Type().Name())),
		getTableName(getModelName(queryBuilder.model)),
		relationship.AssociateColumn,
		queryBuilder.getAlias(getTableName(target.Elem().Type().Name())),
		getPKFieldSelfCOLUMNTagFromModel(target.Interface()))
	queryBuilder.SectionJoin = append(queryBuilder.SectionJoin, stringJoin)
}

func (queryBuilder *SelectQueryBuilder) addOTM(fieldInterface interface{}) {
	relationship := fieldInterface.(*OneToManyRelationShip)
	target := reflect.ValueOf(relationship.Target).Elem()

	var targetAssociatedColumn string

	if relationship.FieldMTO != "" {
		targetMTO := target.FieldByName(relationship.FieldMTO).Interface().(*ManyToOneRelationShip)
		targetAssociatedColumn = targetMTO.AssociateColumn
	} else {
		targetAssociatedColumn = getAssociatedColumnFromReverse(queryBuilder.model, target)
	}

	stringJoin := fmt.Sprintf("LEFT JOIN %v %v ON %v.%v = %v.%v",
		getTableName(target.Type().Name()),
		queryBuilder.getAlias(getTableName(target.Type().Name())),
		getTableName(getModelName(queryBuilder.model)),
		getPKFieldSelfCOLUMNTagFromModel(queryBuilder.model),
		queryBuilder.getAlias(getTableName(target.Type().Name())),
		targetAssociatedColumn)
	queryBuilder.SectionJoin = append(queryBuilder.SectionJoin, stringJoin)
}

func (queryBuilder *SelectQueryBuilder) addMTM(fieldInterface interface{}) {
	relationship := fieldInterface.(*ManyToManyRelationShip)
	target := reflect.ValueOf(relationship.Target)
	intermediateTarget := reflect.ValueOf(relationship.IntermediateTarget).Elem()

	stringJoin := fmt.Sprintf("LEFT JOIN %v %v ON %v.%v = %v.%v INNER JOIN %v %v ON %v.%v = %v.%v",
		getTableName(intermediateTarget.Type().Name()),
		queryBuilder.getAlias(getTableName(intermediateTarget.Type().Name())),
		getTableName(getModelName(queryBuilder.model)),
		getPKFieldSelfCOLUMNTagFromModel(queryBuilder.model),
		queryBuilder.getAlias(getTableName(intermediateTarget.Type().Name())),
		getAssociatedColumnFromReverse(queryBuilder.model, intermediateTarget),

		getTableName(target.Elem().Type().Name()),
		queryBuilder.getAlias(getTableName(target.Elem().Type().Name())),
		queryBuilder.getAlias(getTableName(intermediateTarget.Type().Name())),
		getAssociatedColumnFromReverse(target.Interface(), intermediateTarget),
		queryBuilder.getAlias(getTableName(target.Elem().Type().Name())),
		getPKFieldSelfCOLUMNTagFromModel(target.Interface()))
	queryBuilder.SectionJoin = append(queryBuilder.SectionJoin, stringJoin)
}

func (queryBuilder *SelectQueryBuilder) OrderBy(orderFilter map[string]string) *SelectQueryBuilder {
	sqlConstruct := "ORDER BY"

	queryBuilder.SectionOrder = putIntermediateString(&sqlConstruct,
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

	queryBuilder.SectionWhere = putIntermediateString(&sqlConstruct,
		" and",
		true,
		mapFilter)

	return queryBuilder
}

func (queryBuilder *SelectQueryBuilder) SetModel(model interface{}) {
	queryBuilder.Clean()
	queryBuilder.model = nil
	queryBuilder.model = model
}

func (queryBuilder *SelectQueryBuilder) Clean() {
	queryBuilder.SectionOffset = ""
	queryBuilder.SectionLimit = ""
	queryBuilder.SectionOrder = ""
	queryBuilder.SectionWhere = ""
	queryBuilder.SectionFrom = ""
	queryBuilder.SectionSelect = ""
	queryBuilder.QueryApplier.Clean()
}

func (queryBuilder *SelectQueryBuilder) Consider(fieldName string) *SelectQueryBuilder {
	reflectQueryBuilder := reflect.ValueOf(queryBuilder.model).Elem()
	fieldInterface := reflectQueryBuilder.FieldByName(fieldName).Interface()

	if queryBuilder.addRelationship(fieldInterface) {
		switch fieldInterface.(type) {
		case *ManyToOneRelationShip:
			queryBuilder.addMTO(fieldInterface)
		case *OneToManyRelationShip:
			queryBuilder.addOTM(fieldInterface)
		case *ManyToManyRelationShip:
			queryBuilder.addMTM(fieldInterface)
		}
	}

	return queryBuilder
}

func (queryBuilder *SelectQueryBuilder) ApplyQuery() ([][]ModelsOrderedToScan, error) {
	connection := db.GetConnectionFromDB()
	defer queryBuilder.Clean()

	var modelsMatrix [][]ModelsOrderedToScan
	rows, err := connection.Db.Query(queryBuilder.constructSql())

	if err == nil {
		var modelsList []ModelsOrderedToScan
		for rows.Next() {
			modelsList, err = queryBuilder.hydrate(rows.Scan)

			if err != nil {
				break
			}
			modelsMatrix = append(modelsMatrix, modelsList)
		}
	}

	return modelsMatrix, err
}

func (queryBuilder *SelectQueryBuilder) ApplyQueryRow() ([]ModelsOrderedToScan, error) {
	connection := db.GetConnectionFromDB()
	defer queryBuilder.Clean()

	row := connection.Db.QueryRow(queryBuilder.constructSql())
	modelsList, err := queryBuilder.hydrate(row.Scan)

	return modelsList, err
}

func NewSelectQueryBuilder(model interface{}) *SelectQueryBuilder {
	return &SelectQueryBuilder{QueryApplier: QueryApplier{model: model}}
}
