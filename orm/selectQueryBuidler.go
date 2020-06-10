package orm

import (
	"fmt"
	"reflect"
	"strings"
	"tupeuxcourrir_api/db"
)

type SelectQueryBuilder struct {
	QueryApplier
	SectionSelect     string
	SectionSelectStmt []interface{}

	SectionAggregate     string
	SectionAggregateStmt []interface{}

	SectionWhere     string
	SectionWhereStmt []interface{}

	SectionOrder string

	SectionFrom   string
	SectionLimit  string
	SectionOffset string

	SectionJoin []string

	SectionGroupBy string

	SectionHaving     string
	SectionHavingStmt []interface{}

	RollUp bool
}

var singletonSQueryBuilder *SelectQueryBuilder

func (selectQueryBuilder *SelectQueryBuilder) adaptColumnWithAlias(context interface{}, aggregate bool) {
	switch context.(type) {
	case []string:
		sliceString := context.([]string)
		for i, valueString := range sliceString {
			if dotSplit := strings.Split(valueString, "."); len(dotSplit) == 3 {
				sliceString[i] = fmt.Sprintf("%v.%v",
					selectQueryBuilder.getAlias(dotSplit[0], dotSplit[1]),
					dotSplit[2])
			}
		}
	case map[string]interface{}:
		mapStringInterface := context.(map[string]interface{})

		for key, aInterface := range mapStringInterface {

			if aggregate {
				switch aInterface.(type) {
				case string:
					if dotSplit := strings.Split(aInterface.(string), "."); len(dotSplit) == 3 {
						mapStringInterface[key] = fmt.Sprintf("%v.%v",
							selectQueryBuilder.getAlias(dotSplit[0], dotSplit[1]),
							dotSplit[2])
					}
				case []interface{}:
					theSlice := aInterface.([]interface{})
					if dotSplit := strings.Split(theSlice[0].(string), "."); len(dotSplit) == 3 {
						theSlice[0] = fmt.Sprintf("%v.%v",
							selectQueryBuilder.getAlias(dotSplit[0], dotSplit[1]),
							dotSplit[2])
					}
				}
			} else {
				var theKey string

				if dotSplit := strings.Split(key, "."); len(dotSplit) == 3 {
					theKey = fmt.Sprintf("%v.%v",
						selectQueryBuilder.getAlias(dotSplit[0], dotSplit[1]),
						dotSplit[2])
					mapStringInterface[theKey] = mapStringInterface[key]
					delete(mapStringInterface, key)
				}
			}
		}
	}
}

func (selectQueryBuilder *SelectQueryBuilder) getAlias(fieldRelationshipName, targetModelName string) string {
	reflectValueOf := reflect.ValueOf(selectQueryBuilder.relationshipTargetOrder)
	relationshipTargets := reflectValueOf.MapIndex(reflect.ValueOf(fieldRelationshipName))

	var sliceIndex int

	for i := 0; i < relationshipTargets.Len(); i++ {
		if getModelName(relationshipTargets.Index(i).Interface()) == targetModelName {
			sliceIndex = i + 1
			break
		}
	}

	if sliceIndex == 0 {
		panic("undefined targetModelName in relationship")
	}

	return fmt.Sprintf("%v%v%v", getAbbreviation(fieldRelationshipName),
		getAbbreviation(targetModelName),
		sliceIndex)
}

func (selectQueryBuilder *SelectQueryBuilder) ConstructSql() string {
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
	if selectQueryBuilder.RollUp {
		withRollUp = " WITH ROLLUP"
	}

	return fmt.Sprintf("%v%v%v%v%v%v%v%v%v%v%v;",
		selectQueryBuilder.SectionSelect,
		selectQueryBuilder.SectionAggregate,
		selectQueryBuilder.SectionFrom,
		joins,
		selectQueryBuilder.SectionWhere,
		selectQueryBuilder.SectionGroupBy,
		withRollUp,
		selectQueryBuilder.SectionHaving,
		selectQueryBuilder.SectionOrder,
		selectQueryBuilder.SectionLimit,
		selectQueryBuilder.SectionOffset)
}

func (selectQueryBuilder *SelectQueryBuilder) GetStmts() []interface{} {
	var stmtsInterface = make([]interface{}, 0)
	stmtsInterface = append(stmtsInterface, selectQueryBuilder.SectionSelectStmt...)
	stmtsInterface = append(stmtsInterface, selectQueryBuilder.SectionAggregateStmt...)
	stmtsInterface = append(stmtsInterface, selectQueryBuilder.SectionWhereStmt...)
	stmtsInterface = append(stmtsInterface, selectQueryBuilder.SectionHavingStmt...)
	return stmtsInterface
}

func (selectQueryBuilder *SelectQueryBuilder) addMTO(fieldName string, fieldInterface interface{}) {
	relationship := fieldInterface.(*ManyToOneRelationShip)
	target := reflect.ValueOf(relationship.Target)

	aliasTarget := selectQueryBuilder.getAlias(fieldName, target.Elem().Type().Name())

	stringJoin := fmt.Sprintf("INNER JOIN %v %v ON %v.%v = %v.%v",
		getTableName(target.Elem().Type().Name()),
		aliasTarget,
		getTableName(getModelName(selectQueryBuilder.model)),
		relationship.AssociateColumn,
		aliasTarget,
		getPKFieldNameFromModel(target.Interface()))
	selectQueryBuilder.SectionJoin = append(selectQueryBuilder.SectionJoin, stringJoin)
}

func (selectQueryBuilder *SelectQueryBuilder) addOTM(fieldName string, fieldInterface interface{}) {
	relationship := fieldInterface.(*OneToManyRelationShip)
	target := reflect.ValueOf(relationship.Target).Elem()

	var targetAssociatedColumn string

	if relationship.FieldMTO != "" {
		targetMTO := target.FieldByName(relationship.FieldMTO).Interface().(*ManyToOneRelationShip)
		targetAssociatedColumn = targetMTO.AssociateColumn
	} else {
		targetAssociatedColumn = getAssociatedColumnFromReverse(selectQueryBuilder.model, target)
	}

	aliasTarget := selectQueryBuilder.getAlias(fieldName, target.Type().Name())

	stringJoin := fmt.Sprintf("LEFT JOIN %v %v ON %v.%v = %v.%v",
		getTableName(target.Type().Name()),
		aliasTarget,
		getTableName(getModelName(selectQueryBuilder.model)),
		getPKFieldNameFromModel(selectQueryBuilder.model),
		aliasTarget,
		targetAssociatedColumn)
	selectQueryBuilder.SectionJoin = append(selectQueryBuilder.SectionJoin, stringJoin)
}

func (selectQueryBuilder *SelectQueryBuilder) addMTM(fieldName string, fieldInterface interface{}) {
	relationship := fieldInterface.(*ManyToManyRelationShip)
	target := reflect.ValueOf(relationship.Target)
	intermediateTarget := reflect.ValueOf(relationship.IntermediateTarget).Elem()

	aliasIntermediateTarget := selectQueryBuilder.getAlias(fieldName, intermediateTarget.Type().Name())
	aliasTarget := selectQueryBuilder.getAlias(fieldName, target.Elem().Type().Name())

	stringJoin := fmt.Sprintf("LEFT JOIN %v %v ON %v.%v = %v.%v INNER JOIN %v %v ON %v.%v = %v.%v",
		getTableName(intermediateTarget.Type().Name()),
		aliasIntermediateTarget,
		getTableName(getModelName(selectQueryBuilder.model)),
		getPKFieldNameFromModel(selectQueryBuilder.model),
		aliasIntermediateTarget,
		getAssociatedColumnFromReverse(selectQueryBuilder.model, intermediateTarget),

		getTableName(target.Elem().Type().Name()),
		aliasTarget,
		aliasIntermediateTarget,
		getAssociatedColumnFromReverse(target.Interface(), intermediateTarget),
		aliasTarget,
		getPKFieldNameFromModel(target.Interface()))
	selectQueryBuilder.SectionJoin = append(selectQueryBuilder.SectionJoin, stringJoin)
}

func (selectQueryBuilder *SelectQueryBuilder) OrderBy(orderFilter map[string]interface{}) *SelectQueryBuilder {
	selectQueryBuilder.adaptColumnWithAlias(orderFilter, false)
	var str string
	str, _ = ConstructSQlStmts(
		",",
		"space",
		orderFilter)
	selectQueryBuilder.SectionOrder = fmt.Sprintf("ORDER BY %v", str)

	return selectQueryBuilder
}

func (selectQueryBuilder *SelectQueryBuilder) Limit(limit string) *SelectQueryBuilder {
	selectQueryBuilder.SectionLimit = fmt.Sprintf("LIMIT %v", limit)
	return selectQueryBuilder
}

func (selectQueryBuilder *SelectQueryBuilder) Offset(offset string) *SelectQueryBuilder {
	selectQueryBuilder.SectionOffset = fmt.Sprintf("OFFSET %v", offset)
	return selectQueryBuilder
}

func (selectQueryBuilder *SelectQueryBuilder) Where(mapFilter map[string]interface{}) *SelectQueryBuilder {
	selectQueryBuilder.adaptColumnWithAlias(mapFilter, false)

	var str string
	str, selectQueryBuilder.SectionWhereStmt = ConstructSQlStmts(
		" and",
		"setter",
		mapFilter)
	selectQueryBuilder.SectionWhere = fmt.Sprintf("WHERE %v", str)

	return selectQueryBuilder
}

func (selectQueryBuilder *SelectQueryBuilder) SetModel(model interface{}) {
	selectQueryBuilder.Clean()
	selectQueryBuilder.model = nil
	selectQueryBuilder.model = model
}

func (selectQueryBuilder *SelectQueryBuilder) Clean() {
	reflectQueryBuilder := reflect.ValueOf(selectQueryBuilder).Elem()

	for i := 0; i < reflectQueryBuilder.NumField(); i++ {
		switch reflectQueryBuilder.Field(i).Kind() {
		case reflect.String:
			reflectQueryBuilder.Field(i).SetString("")
		case reflect.Slice:
			reflectQueryBuilder.Field(i).SetLen(0)
			reflectQueryBuilder.Field(i).SetCap(0)
			reflectQueryBuilder.Field(i).Set(reflect.Zero(reflectQueryBuilder.Type().Field(i).Type))
		case reflect.Bool:
			reflectQueryBuilder.Field(i).SetBool(false)
		}
	}
	selectQueryBuilder.QueryApplier.Clean()
}

func (selectQueryBuilder *SelectQueryBuilder) Consider(fieldName string) *SelectQueryBuilder {
	reflectQueryBuilder := reflect.ValueOf(selectQueryBuilder.model).Elem()
	fieldInterface := reflectQueryBuilder.FieldByName(fieldName).Interface()

	if selectQueryBuilder.addRelationship(fieldName, fieldInterface) {
		switch fieldInterface.(type) {
		case *ManyToOneRelationShip:
			selectQueryBuilder.addMTO(fieldName, fieldInterface)
		case *OneToManyRelationShip:
			selectQueryBuilder.addOTM(fieldName, fieldInterface)
		case *ManyToManyRelationShip:
			selectQueryBuilder.addMTM(fieldName, fieldInterface)
		}
	}

	return selectQueryBuilder
}

func (selectQueryBuilder *SelectQueryBuilder) GroupBy(columns []string) *SelectQueryBuilder {
	selectQueryBuilder.adaptColumnWithAlias(columns, false)

	selectQueryBuilder.SectionGroupBy = fmt.Sprintf("GROUP BY %v",
		ConstructSQlSpaceNoStmts(",", columns))

	return selectQueryBuilder
}

func (selectQueryBuilder *SelectQueryBuilder) Select(columns []string) *SelectQueryBuilder {
	selectQueryBuilder.adaptColumnWithAlias(columns, false)

	selectQueryBuilder.columns = columns

	var str string
	str, selectQueryBuilder.SectionSelectStmt = ConstructSQlStmts(",", "space", columns)
	selectQueryBuilder.SectionSelect = fmt.Sprintf("SELECT %v", str)

	return selectQueryBuilder
}

func (selectQueryBuilder *SelectQueryBuilder) Aggregate(aggregateMap map[string]interface{}) *SelectQueryBuilder {
	selectQueryBuilder.adaptColumnWithAlias(aggregateMap, true)

	selectQueryBuilder.aggregates = aggregateMap
	var str string
	str, selectQueryBuilder.SectionAggregateStmt = ConstructSQlStmts(
		",", "aggregate", aggregateMap)
	selectQueryBuilder.SectionAggregate = str

	return selectQueryBuilder
}

func (selectQueryBuilder *SelectQueryBuilder) Having(aggregateMap map[string]interface{}) *SelectQueryBuilder {
	selectQueryBuilder.adaptColumnWithAlias(aggregateMap, true)

	var str string
	str, selectQueryBuilder.SectionHavingStmt = ConstructSQlStmts(
		" and", "aggregate", aggregateMap)
	selectQueryBuilder.SectionHaving = fmt.Sprintf("HAVING %v", str)

	return selectQueryBuilder
}

func (selectQueryBuilder *SelectQueryBuilder) ApplyQueryToSlice() (map[string]interface{}, error) {
	defer selectQueryBuilder.Clean()
	connection := db.GetConnectionFromDB()
	row := connection.Db.QueryRow(selectQueryBuilder.ConstructSql())

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

	for key := range selectQueryBuilder.aggregates {
		mapColumnsResult[key] = columnsResult[i]
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
	rows, err := connection.Db.Query(selectQueryBuilder.ConstructSql())

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

	row := connection.Db.QueryRow(selectQueryBuilder.ConstructSql())
	modelsList, err := selectQueryBuilder.hydrate(row.Scan)

	return modelsList, err
}

func GetSelectQueryBuilder(model interface{}) *SelectQueryBuilder {
	if singletonIQueryBuilder == nil {
		singletonSQueryBuilder = &SelectQueryBuilder{}
	}

	singletonSQueryBuilder.SetModel(model)
	return singletonSQueryBuilder
}
