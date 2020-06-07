package orm

type ModelsScanned struct {
	ModelName string
	Model     interface{}
}

type QueryApplier struct {
	model                   interface{}
	relationshipTargetOrder []interface{}
	columns                 []string
	aggregates              []string
}

func (queryApplier *QueryApplier) hydrate(scan func(dest ...interface{}) error) ([]*ModelsScanned, error) {
	var listModels = []*ModelsScanned{
		{getModelName(queryApplier.model), newModel(queryApplier.model)},
	}

	var addrs []interface{}
	addrFields, err := getAddrFieldsToScan(listModels[0].Model)

	if err == nil {
		for i, relationshipTarget := range queryApplier.relationshipTargetOrder {
			listModels = append(listModels,
				&ModelsScanned{getModelName(relationshipTarget),
					newModel(relationshipTarget)})
			addrs, err = getAddrFieldsToScan(listModels[i+1].Model)
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

	return listModels, err
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
	queryApplier.relationshipTargetOrder = make([]interface{}, 0)
	queryApplier.columns = make([]string, 0)
	queryApplier.aggregates = make([]string, 0)
}
