package orm

import (
	"fmt"
	"reflect"
	"tupeuxcourrir_api/db"
)

type SelectQueryBuilder struct {
	QueryApplier
	SectionSelect    string
	SectionAggregate string
	SectionWhere     string
	SectionOrder     string
	SectionFrom      string
	SectionLimit     string
	SectionOffset    string
	SectionJoin      []string
	SectionGroupBy   string
	rollUp           bool
	root             bool
}

var singletonSQueryBuilder *SelectQueryBuilder

func (selectQueryBuilder *SelectQueryBuilder) getAlias(tableName string) string {
	return fmt.Sprintf("%v%v", tableName[0:2], len(selectQueryBuilder.relationshipTargetOrder))
}

func (selectQueryBuilder *SelectQueryBuilder) constructSql() string {
	switch {
	case selectQueryBuilder.SectionSelect == "" && selectQueryBuilder.SectionAggregate == "":
		selectQueryBuilder.SectionSelect = "SELECT *"
	case selectQueryBuilder.SectionAggregate != "" && selectQueryBuilder.SectionSelect == "":
		selectQueryBuilder.SectionSelect = "SELECT"
	}

	selectQueryBuilder.SectionFrom = fmt.Sprintf("FROM %v", getTableName(getModelName(selectQueryBuilder.model)))

	addPrefixToSections(selectQueryBuilder, " ", 1)

	var joins string
	for _, join := range selectQueryBuilder.SectionJoin {
		joins = fmt.Sprintf("%v %v", joins, join)
	}

	var withRollUp string
	if selectQueryBuilder.rollUp {
		withRollUp = " WITH ROLLUP"
	}

	return fmt.Sprintf("%v%v%v%v%v%v%v%v%v%v;",
		selectQueryBuilder.SectionSelect,
		selectQueryBuilder.SectionAggregate,
		selectQueryBuilder.SectionFrom,
		joins,
		selectQueryBuilder.SectionWhere,
		selectQueryBuilder.SectionGroupBy,
		withRollUp,
		selectQueryBuilder.SectionOrder,
		selectQueryBuilder.SectionLimit,
		selectQueryBuilder.SectionOffset)
}

func (selectQueryBuilder *SelectQueryBuilder) addMTO(fieldInterface interface{}) {
	relationship := fieldInterface.(*ManyToOneRelationShip)
	target := reflect.ValueOf(relationship.Target)
	stringJoin := fmt.Sprintf("INNER JOIN %v %v ON %v.%v = %v.%v",
		getTableName(target.Elem().Type().Name()),
		selectQueryBuilder.getAlias(getTableName(target.Elem().Type().Name())),
		getTableName(getModelName(selectQueryBuilder.model)),
		relationship.AssociateColumn,
		selectQueryBuilder.getAlias(getTableName(target.Elem().Type().Name())),
		getPKFieldNameFromModel(target.Interface()))
	selectQueryBuilder.SectionJoin = append(selectQueryBuilder.SectionJoin, stringJoin)
}

func (selectQueryBuilder *SelectQueryBuilder) addOTM(fieldInterface interface{}) {
	relationship := fieldInterface.(*OneToManyRelationShip)
	target := reflect.ValueOf(relationship.Target).Elem()

	var targetAssociatedColumn string

	if relationship.FieldMTO != "" {
		targetMTO := target.FieldByName(relationship.FieldMTO).Interface().(*ManyToOneRelationShip)
		targetAssociatedColumn = targetMTO.AssociateColumn
	} else {
		targetAssociatedColumn = getAssociatedColumnFromReverse(selectQueryBuilder.model, target)
	}

	stringJoin := fmt.Sprintf("LEFT JOIN %v %v ON %v.%v = %v.%v",
		getTableName(target.Type().Name()),
		selectQueryBuilder.getAlias(getTableName(target.Type().Name())),
		getTableName(getModelName(selectQueryBuilder.model)),
		getPKFieldNameFromModel(selectQueryBuilder.model),
		selectQueryBuilder.getAlias(getTableName(target.Type().Name())),
		targetAssociatedColumn)
	selectQueryBuilder.SectionJoin = append(selectQueryBuilder.SectionJoin, stringJoin)
}

func (selectQueryBuilder *SelectQueryBuilder) addMTM(fieldInterface interface{}) {
	relationship := fieldInterface.(*ManyToManyRelationShip)
	target := reflect.ValueOf(relationship.Target)
	intermediateTarget := reflect.ValueOf(relationship.IntermediateTarget).Elem()

	stringJoin := fmt.Sprintf("LEFT JOIN %v %v ON %v.%v = %v.%v INNER JOIN %v %v ON %v.%v = %v.%v",
		getTableName(intermediateTarget.Type().Name()),
		selectQueryBuilder.getAlias(getTableName(intermediateTarget.Type().Name())),
		getTableName(getModelName(selectQueryBuilder.model)),
		getPKFieldNameFromModel(selectQueryBuilder.model),
		selectQueryBuilder.getAlias(getTableName(intermediateTarget.Type().Name())),
		getAssociatedColumnFromReverse(selectQueryBuilder.model, intermediateTarget),

		getTableName(target.Elem().Type().Name()),
		selectQueryBuilder.getAlias(getTableName(target.Elem().Type().Name())),
		selectQueryBuilder.getAlias(getTableName(intermediateTarget.Type().Name())),
		getAssociatedColumnFromReverse(target.Interface(), intermediateTarget),
		selectQueryBuilder.getAlias(getTableName(target.Elem().Type().Name())),
		getPKFieldNameFromModel(target.Interface()))
	selectQueryBuilder.SectionJoin = append(selectQueryBuilder.SectionJoin, stringJoin)
}

func (selectQueryBuilder *SelectQueryBuilder) OrderBy(orderFilter map[string]interface{}) *SelectQueryBuilder {
	selectQueryBuilder.SectionOrder = fmt.Sprintf("ORDER BY %v", putIntermediateString(
		",",
		"space",
		orderFilter))

	return selectQueryBuilder
}

func (selectQueryBuilder *SelectQueryBuilder) Limit(limit string) *SelectQueryBuilder {
	selectQueryBuilder.SectionLimit = fmt.Sprintf("LIMIT %v", limit)
	return selectQueryBuilder
}

func (selectQueryBuilder *SelectQueryBuilder) FindBy(mapFilter map[string]interface{}) *SelectQueryBuilder {
	selectQueryBuilder.SectionWhere = fmt.Sprintf("WHERE %v", putIntermediateString(
		" and",
		"setter",
		mapFilter))

	return selectQueryBuilder
}

func (selectQueryBuilder *SelectQueryBuilder) SetModel(model interface{}) {
	selectQueryBuilder.Clean()
	selectQueryBuilder.model = nil
	selectQueryBuilder.model = model
}

func (selectQueryBuilder *SelectQueryBuilder) Clean() {
	selectQueryBuilder.SectionOffset = ""
	selectQueryBuilder.SectionLimit = ""
	selectQueryBuilder.SectionOrder = ""
	selectQueryBuilder.SectionWhere = ""
	selectQueryBuilder.SectionFrom = ""
	selectQueryBuilder.SectionSelect = ""
	selectQueryBuilder.SectionAggregate = ""
	selectQueryBuilder.SectionGroupBy = ""
	selectQueryBuilder.rollUp = false
	selectQueryBuilder.QueryApplier.Clean()
}

func (selectQueryBuilder *SelectQueryBuilder) Consider(fieldName string) *SelectQueryBuilder {
	reflectQueryBuilder := reflect.ValueOf(selectQueryBuilder.model).Elem()
	fieldInterface := reflectQueryBuilder.FieldByName(fieldName).Interface()

	if selectQueryBuilder.addRelationship(fieldInterface) {
		switch fieldInterface.(type) {
		case *ManyToOneRelationShip:
			selectQueryBuilder.addMTO(fieldInterface)
		case *OneToManyRelationShip:
			selectQueryBuilder.addOTM(fieldInterface)
		case *ManyToManyRelationShip:
			selectQueryBuilder.addMTM(fieldInterface)
		}
	}

	return selectQueryBuilder
}

func (selectQueryBuilder *SelectQueryBuilder) Select(columns []string) *SelectQueryBuilder {
	selectQueryBuilder.columns = columns

	var mapColumns = make(map[string]interface{})

	for _, column := range columns {
		mapColumns[column] = ""
	}

	selectQueryBuilder.SectionSelect = fmt.Sprintf("SELECT %v",
		putIntermediateString(",", "space", mapColumns))

	return selectQueryBuilder
}

func (selectQueryBuilder *SelectQueryBuilder) Aggregate(aggregateMap map[string]interface{}) *SelectQueryBuilder {
	selectQueryBuilder.aggregates = nil

	for key := range aggregateMap {
		selectQueryBuilder.aggregates = append(selectQueryBuilder.aggregates, key)
	}

	selectQueryBuilder.SectionAggregate = putIntermediateString(
		",", "aggregate", aggregateMap)

	return selectQueryBuilder
}

func (selectQueryBuilder *SelectQueryBuilder) ApplyQueryToSlice() (map[string]interface{}, error) {
	defer selectQueryBuilder.Clean()
	connection := db.GetConnectionFromDB()
	row := connection.Db.QueryRow(selectQueryBuilder.constructSql())

	var columnsResult = make([]interface{}, len(selectQueryBuilder.columns)+len(selectQueryBuilder.aggregates))

	var addrColumnsResult []interface{}
	for i := 0; i < len(columnsResult); i++ {
		addrColumnsResult = append(addrColumnsResult, &columnsResult[i])
	}

	err := row.Scan(addrColumnsResult...)

	var mapColumnsResult = make(map[string]interface{})

	i := 0
	for _, value := range selectQueryBuilder.columns {
		mapColumnsResult[value] = columnsResult[i]
		i++
	}

	for _, value := range selectQueryBuilder.aggregates {
		mapColumnsResult[value] = columnsResult[i]
		i++
	}

	return mapColumnsResult, err
}

func (selectQueryBuilder *SelectQueryBuilder) ApplyQuery() ([][]*ModelsScanned, error) {
	if selectQueryBuilder.SectionSelect != "" || selectQueryBuilder.SectionAggregate != "" {
		panic("configuration not supported")
	}

	connection := db.GetConnectionFromDB()
	defer selectQueryBuilder.Clean()

	var modelsMatrix [][]*ModelsScanned
	rows, err := connection.Db.Query(selectQueryBuilder.constructSql())

	if err == nil {
		var modelsList []*ModelsScanned
		for rows.Next() {
			modelsList, err = selectQueryBuilder.hydrate(rows.Scan)

			if err != nil {
				break
			}
			modelsMatrix = append(modelsMatrix, modelsList)
		}
	}

	return modelsMatrix, err
}

func (selectQueryBuilder *SelectQueryBuilder) ApplyQueryRow() ([]*ModelsScanned, error) {
	if selectQueryBuilder.SectionSelect != "" || selectQueryBuilder.SectionAggregate != "" {
		panic("configuration not supported")
	}

	connection := db.GetConnectionFromDB()
	defer selectQueryBuilder.Clean()

	row := connection.Db.QueryRow(selectQueryBuilder.constructSql())
	modelsList, err := selectQueryBuilder.hydrate(row.Scan)

	return modelsList, err
}

func GetSubSelectQueryBuilder(model interface{}) *SelectQueryBuilder {
	subSQueryBuilder := &SelectQueryBuilder{root: false}
	subSQueryBuilder.SetModel(model)
	return subSQueryBuilder
}

func GetSelectQueryBuilder(model interface{}) *SelectQueryBuilder {
	if singletonIQueryBuilder == nil {
		singletonSQueryBuilder = &SelectQueryBuilder{root: true}
	}

	singletonSQueryBuilder.SetModel(model)
	return singletonSQueryBuilder
}
