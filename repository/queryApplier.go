package repository

import (
	"errors"
	"reflect"
	"tupeuxcourrir_api/models"
)

type QueryApplier struct {
	model             interface{}
	relationshipOrder []interface{}
}

func (queryApplier *QueryApplier) getModelName() string {
	modelName := reflect.TypeOf(queryApplier.model).Elem().Name()
	return modelName
}

func (queryApplier *QueryApplier) getAddrFieldsToScan(model interface{}) ([]interface{}, error) {
	reflectModel := reflect.ValueOf(model)
	if reflectModel.Kind() != reflect.Ptr {
		return make([]interface{}, 0, 0),
			errors.New("must pass a pointer, not a value")
	}

	reflectModel = reflectModel.Elem()
	fieldsTab := make([]interface{}, reflectModel.NumField())
	var field reflect.Value

	for i := 0; i < reflectModel.NumField(); i++ {
		field = reflectModel.Field(i)
		_, ok := field.Interface().(*models.ManyToOneRelationShip)
		_, ok1 := field.Interface().(*models.OneToManyRelationShip)
		_, ok2 := field.Interface().(*models.ManyToOneRelationShip)

		if !ok && !ok1 && !ok2 {
			fieldsTab = append(fieldsTab, field.Addr())
		}
	}

	return fieldsTab, nil
}

func (queryApplier *QueryApplier) newModel() interface{} {
	modelValue := reflect.ValueOf(queryApplier.model)
	if modelValue.Kind() == reflect.Ptr {
		modelValue = reflect.Indirect(modelValue)
	} else {
		panic("the model passed to this queryBuilder must be a pointer")
	}
	return reflect.New(modelValue.Type()).Interface()
}

func (queryApplier *QueryApplier) hydrateOne(scan func(dest ...interface{}) error) (interface{}, error) {
	var newModel = queryApplier.newModel()
	addrFields, err := queryApplier.getAddrFieldsToScan(&newModel)

	if err == nil {
		err = scan(addrFields...)
	}

	return newModel, err
}

func (queryApplier *QueryApplier) addRelationship(relationship interface{}) bool {
	result := false

	switch relationship.(type) {
	case *models.ManyToManyRelationShip, *models.ManyToOneRelationShip, *models.OneToManyRelationShip:
		result = true
		queryApplier.relationshipOrder = append(queryApplier.relationshipOrder, relationship)
	}

	return result
}
