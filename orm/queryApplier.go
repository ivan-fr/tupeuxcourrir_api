package orm

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

type QueryApplier struct {
	model                   interface{}
	relationshipTargetOrder map[string][]interface{}
	columns                 []string
	aggregates              H
	alreadyHydrate          bool
}

func (qA *QueryApplier) getNecessaryNullFieldsForRelationshipWithOrder(relationship interface{}) []interface{} {
	valueOf := reflect.ValueOf(relationship).Elem()

	var nullFields []interface{}

	for i := 0; i < valueOf.NumField(); i++ {
		switch valueOf.Field(i).Interface().(type) {
		case string:
			nullFields = append(nullFields, sql.NullString{})
		case bool:
			nullFields = append(nullFields, sql.NullBool{})
		case int:
			nullFields = append(nullFields, sql.NullInt64{})
		case time.Time:
			nullFields = append(nullFields, sql.NullTime{})
		default:
			nullFields = append(nullFields, nil)
		}
	}

	return nullFields
}

func (qA *QueryApplier) reconstructTheMapFromHydrate(theMap H, nullFields map[string][]interface{}) {
	for fieldName, relationshipTargets := range qA.relationshipTargetOrder {
		for j := range relationshipTargets {
			currentNullFields := nullFields[fmt.Sprintf("%v_%v", fieldName, j)]
			valueOf := reflect.ValueOf(theMap[fmt.Sprintf("%v_%v", fieldName, j)]).Elem()

			for k := 0; k < valueOf.NumField(); k++ {
				if currentNullFields[k] == nil {
					continue
				}

				var nullFieldIsValid bool
				switch currentNullFields[k].(type) {
				case string:
					sN := currentNullFields[k].(sql.NullString)
					nullFieldIsValid = sN.Valid
				case bool:
					sN := currentNullFields[k].(sql.NullBool)
					nullFieldIsValid = sN.Valid
				case int:
					sN := currentNullFields[k].(sql.NullInt64)
					nullFieldIsValid = sN.Valid
				case time.Time:
					sN := currentNullFields[k].(sql.NullTime)
					nullFieldIsValid = sN.Valid
				}

				switch {
				case !nullFieldIsValid:
					theMap[fmt.Sprintf("%v_%v", fieldName, j)] = nil
				case nullFieldIsValid:
					switch currentNullFields[k].(type) {
					case string:
						sN := currentNullFields[k].(sql.NullString)
						valueOf.Field(k).SetString(sN.String)
					case bool:
						sN := currentNullFields[k].(sql.NullBool)
						valueOf.Field(k).SetBool(sN.Bool)
					case int:
						sN := currentNullFields[k].(sql.NullInt64)
						valueOf.Field(k).SetInt(sN.Int64)
					case time.Time:
						sN := currentNullFields[k].(sql.NullTime)
						valueOf.Field(k).Set(reflect.ValueOf(sN.Time))
					}
				}
			}
		}
	}
}

func (qA *QueryApplier) hydrateRelationshipsInModel(theMap H) {
	valueOf := reflect.ValueOf(qA.model).Elem()

	for fieldName := range qA.relationshipTargetOrder {
		concernField := valueOf.FieldByName(fieldName).Interface()
		switch concernField.(type) {
		case *ManyToManyRelationShip:
			relation := concernField.(*ManyToManyRelationShip)
			if relation.EffectiveIntermediateTarget != nil {
				relation.EffectiveIntermediateTarget = theMap[fmt.Sprintf("%v_%v", fieldName, 0)]
			}
			relation.EffectiveTargets = append(relation.EffectiveTargets, theMap[fmt.Sprintf("%v_%v", fieldName, 1)])
		case *OneToManyRelationShip:
			relation := concernField.(*OneToManyRelationShip)
			relation.EffectiveTargets = append(relation.EffectiveTargets, theMap[fmt.Sprintf("%v_%v", fieldName, 0)])
		case *ManyToOneRelationShip:
			relation := concernField.(*ManyToOneRelationShip)
			if relation.EffectiveTarget != nil {
				relation.EffectiveTarget = theMap[fmt.Sprintf("%v_%v", fieldName, 0)]
			}
		}
	}
}

func (qA *QueryApplier) fullHydrate(scan func(dest ...interface{}) error) error {
	addrFields, err := getAddrFieldsToScan(qA.model)

	var theRelationshipMap = make(H)
	var nullFields = make(map[string][]interface{})

	if err != nil {
		return err
	}

	for fieldName, relationshipTargets := range qA.relationshipTargetOrder {
		for j, relationshipTarget := range relationshipTargets {
			theRelationshipMap[fmt.Sprintf("%v_%v", fieldName, j)] = newModel(relationshipTarget)
			nullFields[fmt.Sprintf("%v_%v", fieldName, j)] = qA.getNecessaryNullFieldsForRelationshipWithOrder(
				theRelationshipMap[fmt.Sprintf("%v_%v", fieldName, j)])

			var addressNullFields []interface{}

			for k, nullField := range nullFields[fmt.Sprintf("%v_%v", fieldName, j)] {
				if nullField == nil {
					addr := reflect.ValueOf(theRelationshipMap[fmt.Sprintf("%v_%v", fieldName, j)]).Field(k).Addr().Interface()
					addressNullFields = append(addressNullFields, addr)
				} else {
					addressNullFields = append(addressNullFields, &nullField)
				}
			}

			addrFields = append(addrFields, addressNullFields...)
		}
	}

	err = scan(addrFields...)

	if err != nil && len(nullFields) > 0 {
		qA.reconstructTheMapFromHydrate(theRelationshipMap, nullFields)
		qA.hydrateRelationshipsInModel(theRelationshipMap)
	}

	return err
}

func (qA *QueryApplier) partialHydrate(scan func(dest ...interface{}) error) error {
	reflectModel := reflect.ValueOf(qA.model).Elem()
	var relationshipMap = make(H)

	var addrFields []interface{}
	for _, column := range qA.columns {
		splitDot := strings.Split(column, ".")

		switch len(splitDot) {
		case 1:
			addrFields = append(addrFields, reflectModel.FieldByName(splitDot[0]).Addr().Interface())
		case 2:
			if _, ok := relationshipMap[splitDot[0]]; !ok {
				relationshipMap[splitDot[0]] = newModel(qA.relationshipTargetOrder[splitDot[0]][0])
			}

			addrFields = append(addrFields,
				reflect.ValueOf(relationshipMap[splitDot[0]]).Elem().FieldByName(splitDot[1]).Addr().Interface())
		case 3:
			relationshipName := fmt.Sprintf("%v<sub>%v", splitDot[0], splitDot[1])
			if _, ok := relationshipMap[relationshipName]; !ok {
				sliceIndex := -1
				for i, targetModel := range qA.relationshipTargetOrder[splitDot[0]] {
					if getModelName(targetModel) == splitDot[1] {
						sliceIndex = i
						break
					}
				}

				relationshipMap[relationshipName] = newModel(qA.relationshipTargetOrder[splitDot[0]][sliceIndex])
			}

			addrFields = append(addrFields,
				reflect.ValueOf(relationshipMap[relationshipName]).Elem().FieldByName(splitDot[1]).Addr().Interface())
		}
	}

	for _, column := range qA.aggregates {
		splitDot := strings.Split(column.(string), ".")

		switch len(splitDot) {
		case 1:
			addrFields = append(addrFields, reflectModel.FieldByName(splitDot[0]).Addr().Interface())
		case 2:
			if _, ok := relationshipMap[splitDot[0]]; !ok {
				relationshipMap[splitDot[0]] = newModel(qA.relationshipTargetOrder[splitDot[0]][0])
			}

			addrFields = append(addrFields,
				reflect.ValueOf(relationshipMap[splitDot[0]]).Elem().FieldByName(splitDot[1]).Addr().Interface())
		case 3:
			relationshipName := fmt.Sprintf("%v<sub>%v", splitDot[0], splitDot[1])
			if _, ok := relationshipMap[relationshipName]; !ok {
				sliceIndex := -1
				for i, targetModel := range qA.relationshipTargetOrder[splitDot[0]] {
					if getModelName(targetModel) == splitDot[1] {
						sliceIndex = i
						break
					}
				}

				relationshipMap[relationshipName] = newModel(qA.relationshipTargetOrder[splitDot[0]][sliceIndex])
			}

			addrFields = append(addrFields,
				reflect.ValueOf(relationshipMap[relationshipName]).Elem().FieldByName(splitDot[1]).Addr().Interface())
		}
	}

	err := scan(addrFields...)

	return err
}

func (qA *QueryApplier) hydrate(scan func(dest ...interface{}) error) error {
	var err error

	if qA.alreadyHydrate {
		return errors.New("the model is already hydrate")
	}

	switch {
	case len(qA.columns) == 0 && len(qA.aggregates) == 0:
		err = qA.fullHydrate(scan)
	default:
		err = qA.partialHydrate(scan)
	}

	if err != nil {
		qA.alreadyHydrate = true
	}

	return err
}

func (qA *QueryApplier) addRelationship(fieldName string, relationship interface{}) bool {
	result := true

	switch relationship.(type) {
	case *ManyToManyRelationShip:
		qA.relationshipTargetOrder[fieldName] = append(qA.relationshipTargetOrder[fieldName],
			relationship.(*ManyToManyRelationShip).IntermediateTarget)
		qA.relationshipTargetOrder[fieldName] = append(qA.relationshipTargetOrder[fieldName],
			relationship.(*ManyToManyRelationShip).Target)
	case *ManyToOneRelationShip:
		qA.relationshipTargetOrder[fieldName] = append(qA.relationshipTargetOrder[fieldName],
			relationship.(*ManyToOneRelationShip).Target)
	case *OneToManyRelationShip:
		qA.relationshipTargetOrder[fieldName] = append(qA.relationshipTargetOrder[fieldName],
			relationship.(*OneToManyRelationShip).Target)
	default:
		result = false
	}

	return result
}

func (qA *QueryApplier) Clean() {
	qA.relationshipTargetOrder = make(map[string][]interface{})
	qA.columns = make([]string, 0)
	qA.aggregates = make(H)
}
