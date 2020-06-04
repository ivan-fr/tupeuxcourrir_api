package orm

type ModelsOrderedToScan struct {
	ModelName string
	Model     interface{}
}

type QueryApplier struct {
	model                   interface{}
	relationshipTargetOrder []interface{}
}

func (queryApplier *QueryApplier) hydrate(scan func(dest ...interface{}) error) ([]ModelsOrderedToScan, error) {
	var newModels = []ModelsOrderedToScan{
		{getModelName(queryApplier.model), newModel(queryApplier.model)},
	}

	var addrs []interface{}
	addrFields, err := getAddrFieldsToScan(newModels[0].Model)

	if err == nil {
		for i, relationshipTarget := range queryApplier.relationshipTargetOrder {
			newModels = append(newModels,
				ModelsOrderedToScan{getModelName(relationshipTarget),
					newModel(relationshipTarget)})
			addrs, err = getAddrFieldsToScan(newModels[i+1].Model)
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
	case *ManyToManyRelationShip:
		queryApplier.relationshipTargetOrder = append(queryApplier.relationshipTargetOrder,
			relationship.(*ManyToManyRelationShip).IntermediateTarget)
		queryApplier.relationshipTargetOrder = append(queryApplier.relationshipTargetOrder,
			relationship.(*ManyToManyRelationShip).Target)
		result = true
	case *ManyToOneRelationShip:
		queryApplier.relationshipTargetOrder = append(queryApplier.relationshipTargetOrder,
			relationship.(*ManyToOneRelationShip).Target)
		result = true

	case *OneToManyRelationShip:
		queryApplier.relationshipTargetOrder = append(queryApplier.relationshipTargetOrder,
			relationship.(*OneToManyRelationShip).Target)
		result = true
	}

	return result
}

func (queryApplier *QueryApplier) Clean() {
	queryApplier.relationshipTargetOrder = nil
}
