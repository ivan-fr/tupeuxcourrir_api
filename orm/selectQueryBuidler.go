package orm

import (
	"fmt"
	"log"
	"reflect"
	"tupeuxcourrir_api/db"
)

type SelectQueryBuilder struct {
	QueryApplier
	aliasFactory *ContextAdapterFactory

	SectionSelect string

	SectionAggregate string

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

func (sQB *SelectQueryBuilder) getAlias(fieldRelationshipName, targetModelName string) string {
	reflectValueOf := reflect.ValueOf(sQB.relationshipTargetOrder)
	relationshipTargets := reflectValueOf.MapIndex(reflect.ValueOf(fieldRelationshipName))

	var sliceIndex int

	switch {
	case 1 == relationshipTargets.Len() && targetModelName == "":
		sliceIndex = 1
		targetModelName = getModelName(relationshipTargets.Index(0).Interface())
	case targetModelName != "":
		for i := 0; i < relationshipTargets.Len(); i++ {
			if getModelName(relationshipTargets.Index(i).Interface()) == targetModelName {
				sliceIndex = i + 1
				break
			}
		}
	}

	if sliceIndex == 0 {
		panic("undefined targetModelName in relationship")
	}

	return fmt.Sprintf("%v%v%v", getAbbreviation(fieldRelationshipName),
		getAbbreviation(targetModelName),
		sliceIndex)
}

func (sQB *SelectQueryBuilder) constructSql() string {
	switch {
	case sQB.SectionSelect == "" && sQB.SectionAggregate == "":
		sQB.SectionSelect = "SELECT *"
	case sQB.SectionAggregate != "" && sQB.SectionSelect == "":
		sQB.SectionSelect = "SELECT"
	}

	sQB.SectionFrom = fmt.Sprintf("FROM %v", getTableName(getModelName(sQB.model)))

	addPrefixToSections(sQB, " ", 1)

	var joins string
	for _, join := range sQB.SectionJoin {
		joins = fmt.Sprintf("%v %v", joins, join)
	}

	var withRollUp string
	if sQB.RollUp {
		withRollUp = " WITH ROLLUP"
	}

	return fmt.Sprintf("%v%v%v%v%v%v%v%v%v%v%v;",
		sQB.SectionSelect,
		sQB.SectionAggregate,
		sQB.SectionFrom,
		joins,
		sQB.SectionWhere,
		sQB.SectionGroupBy,
		withRollUp,
		sQB.SectionHaving,
		sQB.SectionOrder,
		sQB.SectionLimit,
		sQB.SectionOffset)
}

func (sQB *SelectQueryBuilder) getStmts() []interface{} {
	var stmtsInterface = make([]interface{}, 0)
	stmtsInterface = append(stmtsInterface, sQB.SectionWhereStmt...)
	stmtsInterface = append(stmtsInterface, sQB.SectionHavingStmt...)
	return stmtsInterface
}

func (sQB *SelectQueryBuilder) addMTO(fieldName string, fieldInterface interface{}) {
	relationship := fieldInterface.(*ManyToOneRelationShip)
	target := reflect.ValueOf(relationship.Target)

	aliasTarget := sQB.getAlias(fieldName, target.Elem().Type().Name())

	stringJoin := fmt.Sprintf("INNER JOIN %v %v ON %v.%v = %v.%v",
		getTableName(target.Elem().Type().Name()),
		aliasTarget,
		getTableName(getModelName(sQB.model)),
		relationship.AssociateColumn,
		aliasTarget,
		getPKFieldNameFromModel(target.Interface()))
	sQB.SectionJoin = append(sQB.SectionJoin, stringJoin)
}

func (sQB *SelectQueryBuilder) addOTM(fieldName string, fieldInterface interface{}) {
	relationship := fieldInterface.(*OneToManyRelationShip)
	target := reflect.ValueOf(relationship.Target).Elem()

	var targetAssociatedColumn string

	if relationship.FieldMTO != "" {
		targetMTO := target.FieldByName(relationship.FieldMTO).Interface().(*ManyToOneRelationShip)
		targetAssociatedColumn = targetMTO.AssociateColumn
	} else {
		targetAssociatedColumn = getAssociatedColumnFromReverse(sQB.model, target)
	}

	aliasTarget := sQB.getAlias(fieldName, target.Type().Name())

	stringJoin := fmt.Sprintf("LEFT JOIN %v %v ON %v.%v = %v.%v",
		getTableName(target.Type().Name()),
		aliasTarget,
		getTableName(getModelName(sQB.model)),
		getPKFieldNameFromModel(sQB.model),
		aliasTarget,
		targetAssociatedColumn)
	sQB.SectionJoin = append(sQB.SectionJoin, stringJoin)
}

func (sQB *SelectQueryBuilder) addMTM(fieldName string, fieldInterface interface{}) {
	relationship := fieldInterface.(*ManyToManyRelationShip)
	target := reflect.ValueOf(relationship.Target)
	intermediateTarget := reflect.ValueOf(relationship.IntermediateTarget).Elem()

	aliasIntermediateTarget := sQB.getAlias(fieldName, intermediateTarget.Type().Name())
	aliasTarget := sQB.getAlias(fieldName, target.Elem().Type().Name())

	stringJoin := fmt.Sprintf("LEFT JOIN %v %v ON %v.%v = %v.%v INNER JOIN %v %v ON %v.%v = %v.%v",
		getTableName(intermediateTarget.Type().Name()),
		aliasIntermediateTarget,
		getTableName(getModelName(sQB.model)),
		getPKFieldNameFromModel(sQB.model),
		aliasIntermediateTarget,
		getAssociatedColumnFromReverse(sQB.model, intermediateTarget),

		getTableName(target.Elem().Type().Name()),
		aliasTarget,
		aliasIntermediateTarget,
		getAssociatedColumnFromReverse(target.Interface(), intermediateTarget),
		aliasTarget,
		getPKFieldNameFromModel(target.Interface()))
	sQB.SectionJoin = append(sQB.SectionJoin, stringJoin)
}

func (sQB *SelectQueryBuilder) OrderBy(orderFilter H) *SelectQueryBuilder {
	sQB.aliasFactory.adaptContext(orderFilter, false)
	sSA := &sQLSectionArchitecture{intermediateString: ",", isStmts: true, mode: "space", context: orderFilter}
	sSA.constructSQlSection()

	sQB.SectionOrder = fmt.Sprintf("ORDER BY %v", sSA.SQLSection)

	return sQB
}

func (sQB *SelectQueryBuilder) Limit(limit string) *SelectQueryBuilder {
	sQB.SectionLimit = fmt.Sprintf("LIMIT %v", limit)
	return sQB
}

func (sQB *SelectQueryBuilder) Offset(offset string) *SelectQueryBuilder {
	sQB.SectionOffset = fmt.Sprintf("OFFSET %v", offset)
	return sQB
}

func (sQB *SelectQueryBuilder) Where(logical *Logical) *SelectQueryBuilder {
	sQB.aliasFactory.adaptContext(logical, false)

	var str string
	str, sQB.SectionWhereStmt = logical.GetSentence("setter")
	sQB.SectionWhere = fmt.Sprintf("WHERE %v", str)

	return sQB
}

func (sQB *SelectQueryBuilder) SetModel(model interface{}) {
	sQB.Clean()
	sQB.model = nil
	sQB.model = model
}

func (sQB *SelectQueryBuilder) Clean() {
	reflectQueryBuilder := reflect.ValueOf(sQB).Elem()

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
	sQB.QueryApplier.Clean()
}

func (sQB *SelectQueryBuilder) Consider(fieldName string) *SelectQueryBuilder {
	reflectQueryBuilder := reflect.ValueOf(sQB.model).Elem()
	fieldInterface := reflectQueryBuilder.FieldByName(fieldName).Interface()

	if sQB.addRelationship(fieldName, fieldInterface) {
		switch fieldInterface.(type) {
		case *ManyToOneRelationShip:
			sQB.addMTO(fieldName, fieldInterface)
		case *OneToManyRelationShip:
			sQB.addOTM(fieldName, fieldInterface)
		case *ManyToManyRelationShip:
			sQB.addMTM(fieldName, fieldInterface)
		}
	}

	return sQB
}

func (sQB *SelectQueryBuilder) GroupBy(columns []string) *SelectQueryBuilder {
	sQB.aliasFactory.adaptContext(columns, false)

	sSA := &sQLSectionArchitecture{intermediateString: ",", isStmts: false, mode: "space", context: columns}
	sSA.constructSQlSection()

	sQB.SectionGroupBy = fmt.Sprintf("GROUP BY %v", sSA.SQLSection)

	return sQB
}

func (sQB *SelectQueryBuilder) Select(columns []string) *SelectQueryBuilder {
	sQB.columns = make([]string, 0)

	for _, val := range columns {
		sQB.columns = append(sQB.columns, val)
	}

	sQB.aliasFactory.adaptContext(columns, false)

	sSA := &sQLSectionArchitecture{intermediateString: ",", isStmts: false, mode: "space", context: columns}
	sSA.constructSQlSection()
	sQB.SectionSelect = fmt.Sprintf("SELECT %v", sSA.SQLSection)

	return sQB
}

func (sQB *SelectQueryBuilder) Aggregate(aggregateMap H) *SelectQueryBuilder {
	sQB.aggregates = make(H)

	for key, val := range aggregateMap {
		sQB.aggregates[key] = val
	}

	sQB.aliasFactory.adaptContext(aggregateMap, true)

	sSA := &sQLSectionArchitecture{intermediateString: ",", isStmts: false, mode: "aggregate", context: aggregateMap}
	sSA.constructSQlSection()

	sQB.SectionAggregate = sSA.SQLSection

	return sQB
}

func (sQB *SelectQueryBuilder) Having(logical *Logical) *SelectQueryBuilder {
	sQB.aliasFactory.adaptContext(logical, true)

	var str string
	str, sQB.SectionHavingStmt = logical.GetSentence("aggregate")
	sQB.SectionHaving = fmt.Sprintf("HAVING %v", str)

	return sQB
}

func (sQB *SelectQueryBuilder) ApplyQueryToSlice() (H, error) {
	if sQB.SectionSelect == "" && sQB.SectionAggregate == "" {
		panic("configuration not supported")
	}

	connection := db.GetConnectionFromDB()
	defer sQB.Clean()

	row := connection.Db.QueryRow(sQB.constructSql(), sQB.getStmts()...)

	var columnsResult = make([]interface{}, len(sQB.columns)+len(sQB.aggregates))

	var addrColumnsResult []interface{}
	for i := 0; i < len(columnsResult); i++ {
		addrColumnsResult = append(addrColumnsResult, columnsResult[i])
	}

	err := row.Scan(addrColumnsResult...)

	var mapColumnsResult = make(H)

	i := 0
	for _, value := range sQB.columns {
		mapColumnsResult[value] = columnsResult[i]
		i++
	}

	for key, value := range sQB.aggregates {
		mapColumnsResult[fmt.Sprintf("%v(%v)", key, value)] = columnsResult[i]
		i++
	}

	return mapColumnsResult, err
}

func (sQB *SelectQueryBuilder) ApplyQuery() ([][]*ModelsScanned, error) {
	if sQB.SectionSelect != "" || sQB.SectionAggregate != "" {
		panic("configuration not supported")
	}

	connection := db.GetConnectionFromDB()
	defer sQB.Clean()

	var modelsMatrix [][]*ModelsScanned
	rows, err := connection.Db.Query(sQB.constructSql(), sQB.getStmts()...)

	if err == nil {
		var modelsList []*ModelsScanned
		for rows.Next() {
			modelsList, err = sQB.hydrate(rows.Scan)

			if err != nil {
				break
			}
			modelsMatrix = append(modelsMatrix, modelsList)
		}
	}

	return modelsMatrix, err
}

func (sQB *SelectQueryBuilder) ApplyQueryRow() ([]*ModelsScanned, error) {
	if sQB.SectionSelect != "" || sQB.SectionAggregate != "" {
		panic("configuration not supported")
	}

	connection := db.GetConnectionFromDB()
	defer sQB.Clean()

	sql := sQB.constructSql()
	log.Println(sql)

	row := connection.Db.QueryRow(sql, sQB.getStmts()...)
	modelsList, err := sQB.hydrate(row.Scan)

	return modelsList, err
}

func GetSelectQueryBuilder(model interface{}) *SelectQueryBuilder {
	if singletonIQueryBuilder == nil {
		singletonSQueryBuilder = &SelectQueryBuilder{}
		singletonSQueryBuilder.aliasFactory = &ContextAdapterFactory{getAliasFunc: singletonSQueryBuilder.getAlias}
	}

	singletonSQueryBuilder.SetModel(model)
	return singletonSQueryBuilder
}
