package repository

import (
	"reflect"
	"tupeuxcourrir_api/models"
)

type QueryApplier struct {
	model                   interface{}
	relationshipTargetOrder []interface{}
}

func (queryApplier *QueryApplier) getModelName() string {
	modelName := reflect.TypeOf(queryApplier.model).Elem().Name()
	return modelName
}

func (queryApplier *QueryApplier) hydrate(scan func(dest ...interface{}) error) (interface{}, error) {
	var newModel = newModel(queryApplier.model)
	addrFields, err := getAddrFieldsToScan(&newModel)

	if err == nil {

		for _, relationshipTarget := range queryApplier.relationshipTargetOrder {

		}

		err = scan(addrFields...)
	}

	return newModel, err
}

func (queryApplier *QueryApplier) addRelationship(relationship interface{}) bool {
	result := false

	switch relationship.(type) {
	case *models.ManyToManyRelationShip:
		queryApplier.relationshipTargetOrder = append(queryApplier.relationshipTargetOrder,
			relationship.(*models.ManyToManyRelationShip).IntermediateTarget)
		queryApplier.relationshipTargetOrder = append(queryApplier.relationshipTargetOrder,
			relationship.(*models.ManyToManyRelationShip).Target)
		result = true
	case *models.ManyToOneRelationShip:
		queryApplier.relationshipTargetOrder = append(queryApplier.relationshipTargetOrder,
			relationship.(*models.ManyToOneRelationShip).Target)
		result = true

	case *models.OneToManyRelationShip:
		queryApplier.relationshipTargetOrder = append(queryApplier.relationshipTargetOrder,
			relationship.(*models.OneToManyRelationShip).Target)
		result = true
	}

	return result
}
