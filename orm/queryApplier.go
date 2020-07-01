package orm

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

type aggregate struct {
	column    string
	aggregate string
	value     interface{}
}

type QueryApplier struct {
	model               Model
	relationshipTargets map[string][]interface{}
	orderConsideration  []string
	columns             []string
	aggregates          H
	EffectiveAggregates []*aggregate
	EffectiveModel      Model
}

func (qA *QueryApplier) getNecessaryNullFieldsForRelationshipWithOrder(relationship interface{}) []interface{} {
	valueOf := reflect.ValueOf(relationship).Elem()

	var nullFields []interface{}

	for i := 0; i < valueOf.NumField(); i++ {
		nullFields = append(nullFields, getNullFieldInstance(valueOf.Field(i).Interface()))
	}

	return nullFields
}

func (qA *QueryApplier) mergePartialRelationshipModelsFromNullFields(theRelationshipMap H, nullFields H) {
	for _, column := range qA.columns {
		splitDot := strings.Split(column, ".")

		switch len(splitDot) {
		case 2:
			if nullField, ok := nullFields[column]; ok {
				if isAValidNullField(nullField) {
					theField := reflect.ValueOf(theRelationshipMap[fmt.Sprintf("%v_%v", splitDot[0], 0)]).Elem().FieldByName(splitDot[1])
					setNullFieldToAField(nullField, theField)
				}
			}
		case 3:
			sliceIndex := qA.getIndexOfWantedModelFromRelationshipTargets(splitDot[0], splitDot[1])
			relationshipName := fmt.Sprintf("%v_%v", splitDot[0], sliceIndex)

			if nullField, ok := nullFields[column]; ok {
				if isAValidNullField(nullField) {
					theField := reflect.ValueOf(theRelationshipMap[relationshipName]).Elem().FieldByName(splitDot[1])
					setNullFieldToAField(nullField, theField)
				}
			}
		}
	}
}

func (qA *QueryApplier) mergeRelationshipModelsFromNullFields(theMap H, nullFields map[string][]interface{}) {
	for _, fieldName := range qA.orderConsideration {
		for j := range qA.relationshipTargets[fieldName] {
			currentNullFields := nullFields[fmt.Sprintf("%v_%v", fieldName, j)]
			valueOfRelationshipModel := reflect.ValueOf(theMap[fmt.Sprintf("%v_%v", fieldName, j)]).Elem()

			for k := 0; k < valueOfRelationshipModel.NumField(); k++ {
				if currentNullFields[k] == nil {
					continue
				}

				nullFieldIsValid := isAValidNullField(currentNullFields[k])

				switch {
				case !nullFieldIsValid:
					log.Println(fmt.Sprintf("%v_%v", fieldName, j))
					theMap[fmt.Sprintf("%v_%v", fieldName, j)] = nil
				case nullFieldIsValid:
					setNullFieldToAField(currentNullFields[k], valueOfRelationshipModel.Field(k))
				}
			}
		}
	}
}

func (qA *QueryApplier) hydrateRelationshipsInModel(theMap H) {
	valueOfMainModel := reflect.ValueOf(qA.EffectiveModel).Elem()

	for _, fieldName := range qA.orderConsideration {
		concernField := valueOfMainModel.FieldByName(fieldName).Interface()
		switch concernField.(type) {
		case *ManyToManyRelationShip:
			relation := concernField.(*ManyToManyRelationShip)
			if relation.EffectiveIntermediateTarget != nil && theMap[fmt.Sprintf("%v_%v", fieldName, 0)] != nil {
				relation.EffectiveIntermediateTarget = theMap[fmt.Sprintf("%v_%v", fieldName, 0)]
			}
			if theMap[fmt.Sprintf("%v_%v", fieldName, 1)] != nil {
				relation.EffectiveTargets = append(relation.EffectiveTargets, theMap[fmt.Sprintf("%v_%v", fieldName, 1)])
			}
		case *OneToManyRelationShip:
			relation := concernField.(*OneToManyRelationShip)
			if theMap[fmt.Sprintf("%v_%v", fieldName, 0)] != nil {
				relation.EffectiveTargets = append(relation.EffectiveTargets, theMap[fmt.Sprintf("%v_%v", fieldName, 0)])
			}
		case *ManyToOneRelationShip:
			relation := concernField.(*ManyToOneRelationShip)
			if relation.EffectiveTarget != nil && theMap[fmt.Sprintf("%v_%v", fieldName, 0)] != nil {
				relation.EffectiveTarget = theMap[fmt.Sprintf("%v_%v", fieldName, 0)]
			}
		}
	}
}

func (qA *QueryApplier) fullHydrate(scan func(dest ...interface{}) error) error {
	addrFields, err := getAddrFieldsToScan(qA.EffectiveModel)

	var theRelationshipMap = make(H)
	var nullFields = make(map[string][]interface{})

	if err != nil {
		return err
	}

	for _, fieldName := range qA.orderConsideration {
		for j, relationshipTarget := range qA.relationshipTargets[fieldName] {
			theRelationshipMap[fmt.Sprintf("%v_%v", fieldName, j)] = newModel(relationshipTarget)
			nullFields[fmt.Sprintf("%v_%v", fieldName, j)] = qA.getNecessaryNullFieldsForRelationshipWithOrder(
				theRelationshipMap[fmt.Sprintf("%v_%v", fieldName, j)])

			var addressNullFields []interface{}

			for k, ptrNullField := range nullFields[fmt.Sprintf("%v_%v", fieldName, j)] {
				if ptrNullField == nil {
					theField := reflect.ValueOf(theRelationshipMap[fmt.Sprintf("%v_%v", fieldName, j)]).Elem().Field(k)
					if !isRelationshipField(theField) {
						addr := theField.Addr().Interface()
						addressNullFields = append(addressNullFields, addr)
					}
				} else {
					addressNullFields = append(addressNullFields, ptrNullField)
				}
			}

			addrFields = append(addrFields, addressNullFields...)
		}
	}

	err = scan(addrFields...)

	if err == nil && len(nullFields) > 0 {
		qA.mergeRelationshipModelsFromNullFields(theRelationshipMap, nullFields)
		qA.hydrateRelationshipsInModel(theRelationshipMap)
	}

	return err
}

func (qA *QueryApplier) getIndexOfWantedModelFromRelationshipTargets(fieldNameOfRelationship, wantedModel string) int {
	sliceIndex := -1
	for i, targetModel := range qA.relationshipTargets[fieldNameOfRelationship] {
		if getModelName(targetModel) == wantedModel {
			sliceIndex = i
			break
		}
	}

	return sliceIndex
}

func (qA *QueryApplier) partialHydrate(scan func(dest ...interface{}) error) error {
	reflectModel := reflect.ValueOf(qA.EffectiveModel).Elem()
	var theRelationshipMap = make(H)
	var nullFields = make(H)

	var addrFields []interface{}
	for _, column := range qA.columns {
		splitDot := strings.Split(column, ".")

		switch len(splitDot) {
		case 1:
			addrFields = append(addrFields, reflectModel.FieldByName(splitDot[0]).Addr().Interface())
		case 2:
			if _, ok := theRelationshipMap[fmt.Sprintf("%v_%v", splitDot[0], 0)]; !ok {
				theRelationshipMap[fmt.Sprintf("%v_%v", splitDot[0], 0)] = newModel(qA.relationshipTargets[splitDot[0]][0])
			}

			theField := reflect.ValueOf(theRelationshipMap[fmt.Sprintf("%v_%v", splitDot[0], 0)]).Elem().FieldByName(splitDot[1])
			if ptrNullField := getNullFieldInstance(theField.Interface()); ptrNullField == nil {
				addrFields = append(addrFields, theField.Addr().Interface())
			} else {
				nullFields[column] = ptrNullField
				addrFields = append(addrFields, nullFields[column])
			}
		case 3:
			sliceIndex := qA.getIndexOfWantedModelFromRelationshipTargets(splitDot[0], splitDot[1])
			if sliceIndex == -1 {
				panic("the target model doesn't exist")
			}

			relationshipName := fmt.Sprintf("%v_%v", splitDot[0], sliceIndex)
			if _, ok := theRelationshipMap[relationshipName]; !ok {
				theRelationshipMap[relationshipName] = newModel(qA.relationshipTargets[splitDot[0]][sliceIndex])
			}

			theField := reflect.ValueOf(theRelationshipMap[relationshipName]).Elem().FieldByName(splitDot[1])
			if ptrNullField := getNullFieldInstance(theField.Interface()); ptrNullField == nil {
				addrFields = append(addrFields, theField.Addr().Interface())
			} else {
				nullFields[column] = ptrNullField
				addrFields = append(addrFields, nullFields[column])
			}
		}
	}

	for aAggregate, column := range qA.aggregates {
		qA.EffectiveAggregates = append(qA.EffectiveAggregates,
			&aggregate{column: column.(string), aggregate: aAggregate})
		addr := &qA.EffectiveAggregates[len(qA.EffectiveAggregates)-1].value
		addrFields = append(addrFields, addr)
	}

	err := scan(addrFields...)

	if err == nil && len(nullFields) > 0 {
		qA.mergePartialRelationshipModelsFromNullFields(theRelationshipMap, nullFields)
		qA.hydrateRelationshipsInModel(theRelationshipMap)
	}

	return err
}

func (qA *QueryApplier) hydrate(scan func(dest ...interface{}) error) error {
	var err error

	qA.EffectiveModel = newModel(qA.model).(Model)
	qA.EffectiveModel.PutRelationshipConfig()

	switch {
	case len(qA.columns) == 0 && len(qA.aggregates) == 0:
		err = qA.fullHydrate(scan)
	default:
		err = qA.partialHydrate(scan)
	}

	return err
}

func (qA *QueryApplier) addRelationship(fieldName string, relationship interface{}) bool {
	result := true

	switch relationship.(type) {
	case *ManyToManyRelationShip:
		MTM := relationship.(*ManyToManyRelationShip)

		MTM.IntermediateTarget.PutRelationshipConfig()
		MTM.Target.PutRelationshipConfig()

		qA.relationshipTargets[fieldName] = append(qA.relationshipTargets[fieldName],
			MTM.IntermediateTarget)
		qA.relationshipTargets[fieldName] = append(qA.relationshipTargets[fieldName],
			MTM.Target)
	case *ManyToOneRelationShip:
		MTO := relationship.(*ManyToOneRelationShip)
		MTO.Target.PutRelationshipConfig()

		qA.relationshipTargets[fieldName] = append(qA.relationshipTargets[fieldName],
			MTO.Target)
	case *OneToManyRelationShip:
		OTM := relationship.(*OneToManyRelationShip)
		OTM.Target.PutRelationshipConfig()

		qA.relationshipTargets[fieldName] = append(qA.relationshipTargets[fieldName],
			OTM.Target)
	default:
		result = false
	}

	if result {
		qA.orderConsideration = append(qA.orderConsideration, fieldName)
	}

	return result
}

func (qA *QueryApplier) Clean() {
	qA.relationshipTargets = make(map[string][]interface{})
	qA.orderConsideration = make([]string, 0)
	qA.columns = make([]string, 0)
	qA.EffectiveAggregates = make([]*aggregate, 0)
	qA.aggregates = make(H)
}
