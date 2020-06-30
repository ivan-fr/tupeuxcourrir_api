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

func (sQB *SelectQueryBuilder) getAlias(fieldRelationshipName, wantedModel string) string {
	var sliceIndex = -1

	switch {
	case 1 == len(sQB.relationshipTargets[fieldRelationshipName]) && wantedModel == "":
		sliceIndex = 0
		wantedModel = getModelName(sQB.relationshipTargets[fieldRelationshipName][0])
	case wantedModel != "":
		sliceIndex = sQB.getIndexOfWantedModelFromRelationshipTargets(fieldRelationshipName, wantedModel)
	}

	if sliceIndex == -1 {
		panic("undefined wantedModel in relationship")
	}

	return fmt.Sprintf("%v%v%v", getAbbreviation(fieldRelationshipName),
		getAbbreviation(wantedModel),
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

	_sql := fmt.Sprintf("%v%v%v%v%v%v%v%v%v%v%v;",
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
	log.Println(_sql)
	return _sql
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

func (sQB *SelectQueryBuilder) SetModel(model Model) {
	sQB.Clean()
	model.PutRelationshipConfig()
	sQB.model = model
	sQB.QueryApplier.EffectiveModel = nil
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

func (sQB *SelectQueryBuilder) ApplyQuery() error {
	if sQB.SectionSelect != "" || sQB.SectionAggregate != "" {
		panic("configuration not supported")
	}

	connection := db.GetConnectionFromDB()
	defer sQB.Clean()

	rows, err := connection.Db.Query(sQB.constructSql(), sQB.getStmts()...)

	if err == nil {
		for rows.Next() {
			err = sQB.hydrate(rows.Scan)

			if err != nil {
				break
			}
		}
	}

	return err
}

func (sQB *SelectQueryBuilder) ApplyQueryRow() error {
	if sQB.SectionSelect != "" || sQB.SectionAggregate != "" {
		panic("configuration not supported")
	}

	connection := db.GetConnectionFromDB()
	defer sQB.Clean()

	row := connection.Db.QueryRow(sQB.constructSql(), sQB.getStmts()...)
	err := sQB.hydrate(row.Scan)

	return err
}

func GetSelectQueryBuilder(model Model) *SelectQueryBuilder {
	sQB := &SelectQueryBuilder{}
	sQB.aliasFactory = &ContextAdapterFactory{getAliasFunc: sQB.getAlias}
	sQB.SetModel(model)
	return sQB
}
