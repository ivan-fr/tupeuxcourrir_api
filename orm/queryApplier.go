package orm

import "fmt"

type QueryApplier struct {
	model                   interface{}
	relationshipTargetOrder map[string][]interface{}
	columns                 []string
	aggregates              H
}

func (queryApplier *QueryApplier) hydrate(scan func(dest ...interface{}) error) (H, error) {
	var theMap = H{
		getModelName(queryApplier.model): newModel(queryApplier.model),
	}

	var address []interface{}
	var subErr error
	addrFields, err := getAddrFieldsToScan(theMap[getModelName(queryApplier.model)])

	if err == nil {
		i := 0
		for fieldName, relationshipTargets := range queryApplier.relationshipTargetOrder {
			for j, relationshipTarget := range relationshipTargets {
				theMap[fmt.Sprintf("%v_%v", fieldName, j)] = newModel(relationshipTarget)
				address, subErr = getAddrFieldsToScan(theMap[fmt.Sprintf("%v_%v", fieldName, j)])
				addrFields = append(addrFields, address...)

				if subErr != nil {
					break
				}
				i++
			}
		}

		if subErr == nil {
			err = scan(addrFields...)
		}
	}

	return theMap, err
}

func (queryApplier *QueryApplier) addRelationship(fieldName string, relationship interface{}) bool {
	result := true

	switch relationship.(type) {
	case *ManyToManyRelationShip:
		queryApplier.relationshipTargetOrder[fieldName] = append(queryApplier.relationshipTargetOrder[fieldName],
			relationship.(*ManyToManyRelationShip).IntermediateTarget)
		queryApplier.relationshipTargetOrder[fieldName] = append(queryApplier.relationshipTargetOrder[fieldName],
			relationship.(*ManyToManyRelationShip).Target)
	case *ManyToOneRelationShip:
		queryApplier.relationshipTargetOrder[fieldName] = append(queryApplier.relationshipTargetOrder[fieldName],
			relationship.(*ManyToOneRelationShip).Target)
	case *OneToManyRelationShip:
		queryApplier.relationshipTargetOrder[fieldName] = append(queryApplier.relationshipTargetOrder[fieldName],
			relationship.(*OneToManyRelationShip).Target)
	default:
		result = false
	}

	return result
}

func (queryApplier *QueryApplier) Clean() {
	queryApplier.relationshipTargetOrder = make(map[string][]interface{})
	queryApplier.columns = make([]string, 0)
	queryApplier.aggregates = make(H)
}
