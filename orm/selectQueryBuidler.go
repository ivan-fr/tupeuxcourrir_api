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

func (selectQueryBuilder *SelectQueryBuilder) getAlias(tableName string) string {
	return fmt.Sprintf("%v%v", tableName[0:2], len(selectQueryBuilder.relationshipTargetOrder))
}

func (selectQueryBuilder *SelectQueryBuilder) constructSql() string {
	selectQueryBuilder.SectionFrom = fmt.Sprintf("FROM %v", getTableName(getModelName(selectQueryBuilder.model)))

	addPrefixToSections(selectQueryBuilder, " ")
	selectQueryBuilder.SectionSelect = "SELECT *"

	var joins string
	for _, join := range selectQueryBuilder.SectionJoin {
		joins = fmt.Sprintf("%v %v", joins, join)
	}

	return fmt.Sprintf("%v%v%v%v%v%v%v;", selectQueryBuilder.SectionSelect,
		selectQueryBuilder.SectionFrom,
		joins,
		selectQueryBuilder.SectionWhere,
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
		getPKFieldSelfCOLUMNTagFromModel(target.Interface()))
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
		getPKFieldSelfCOLUMNTagFromModel(selectQueryBuilder.model),
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
		getPKFieldSelfCOLUMNTagFromModel(selectQueryBuilder.model),
		selectQueryBuilder.getAlias(getTableName(intermediateTarget.Type().Name())),
		getAssociatedColumnFromReverse(selectQueryBuilder.model, intermediateTarget),

		getTableName(target.Elem().Type().Name()),
		selectQueryBuilder.getAlias(getTableName(target.Elem().Type().Name())),
		selectQueryBuilder.getAlias(getTableName(intermediateTarget.Type().Name())),
		getAssociatedColumnFromReverse(target.Interface(), intermediateTarget),
		selectQueryBuilder.getAlias(getTableName(target.Elem().Type().Name())),
		getPKFieldSelfCOLUMNTagFromModel(target.Interface()))
	selectQueryBuilder.SectionJoin = append(selectQueryBuilder.SectionJoin, stringJoin)
}

func (selectQueryBuilder *SelectQueryBuilder) OrderBy(orderFilter map[string]string) *SelectQueryBuilder {
	sqlConstruct := "ORDER BY"

	selectQueryBuilder.SectionOrder = putIntermediateString(&sqlConstruct,
		",",
		false,
		orderFilter)

	return selectQueryBuilder
}

func (selectQueryBuilder *SelectQueryBuilder) Limit(limit string) *SelectQueryBuilder {
	selectQueryBuilder.SectionLimit = fmt.Sprintf("LIMIT %v", limit)
	return selectQueryBuilder
}

func (selectQueryBuilder *SelectQueryBuilder) FindBy(mapFilter map[string]string) *SelectQueryBuilder {
	sqlConstruct := "WHERE"

	selectQueryBuilder.SectionWhere = putIntermediateString(&sqlConstruct,
		" and",
		true,
		mapFilter)

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

func (selectQueryBuilder *SelectQueryBuilder) ApplyQuery() ([][]*ModelsScanned, error) {
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
	connection := db.GetConnectionFromDB()
	defer selectQueryBuilder.Clean()

	row := connection.Db.QueryRow(selectQueryBuilder.constructSql())
	modelsList, err := selectQueryBuilder.hydrate(row.Scan)

	return modelsList, err
}

func NewSelectQueryBuilder(model interface{}) *SelectQueryBuilder {
	return &SelectQueryBuilder{QueryApplier: QueryApplier{model: model}}
}
