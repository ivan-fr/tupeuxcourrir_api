package orm

import (
	"tupeuxcourrir_api/models"
)

type ModelsOrderedToScan struct {
	modelName string
	model     interface{}
}

type QueryApplier struct {
	model                   interface{}
	relationshipTargetOrder []interface{}
}

func (queryApplier *QueryApplier) hydrate(scan func(dest ...interface{}) error) ([]ModelsOrderedToScan, error) {
	var newModels = []ModelsOrderedToScan{
		{getModelName(queryApplier.model), newModel(queryApplier.model)},
	}

	var considerateModel interface{} = newModels[0].model
	var addrs []interface{}
	addrFields, err := getAddrFieldsToScan(&considerateModel)

	if err == nil {
		for i, relationshipTarget := range queryApplier.relationshipTargetOrder {
			newModels = append(newModels,
				ModelsOrderedToScan{getModelName(relationshipTarget),
					newModel(relationshipTarget)})
			considerateModel = newModels[i+1].model
			addrs, err = getAddrFieldsToScan(&considerateModel)
			addrFields = append(addrFields, addrs...)

			if err != nil {
				break
			}
		}

		addrs = nil

		if err == nil {
			err = scan(addrFields...)
		}
	}

	return newModels, err
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
