package repository

import (
	"errors"
	"reflect"
)

type QueryApplier struct {
	model             interface{}
	relationShipOrder []interface{}
}

func (queryApplier *QueryApplier) getAddrFieldsToScan(model interface{}) ([]interface{}, error) {
	reflectModel := reflect.ValueOf(model)
	if reflectModel.Kind() != reflect.Ptr {
		return make([]interface{}, 0, 0),
			errors.New("must pass a pointer, not a value")
	}

	reflectModel = reflectModel.Elem()
	fieldsTab := make([]interface{}, reflectModel.NumField())

	for i := 0; i < reflectModel.NumField(); i++ {
		fieldsTab[i] = reflectModel.Field(i).Addr()
	}

	return fieldsTab, nil
}

func (queryApplier *QueryApplier) appendRelationShip(relationShip interface{}) {
	queryApplier.relationShipOrder = append(queryApplier.relationShipOrder, relationShip)
}

func (queryApplier *QueryApplier) newModel() interface{} {
	modelValue := reflect.ValueOf(queryApplier.model)
	if modelValue.Kind() == reflect.Ptr {
		modelValue = reflect.Indirect(modelValue)
	} else {
		panic("the model passed to this queryBuilder must be a pointer")
	}
	return reflect.New(modelValue.Type()).Interface().(interface{})
}

func (queryApplier *QueryApplier) hydrateOne(scan func(dest ...interface{}) error) (interface{}, error) {
	var newModel = queryApplier.newModel()
	addrFields, err := queryApplier.getAddrFieldsToScan(newModel)

	if err == nil {
		err = scan(addrFields...)
	}

	return newModel, err
}
