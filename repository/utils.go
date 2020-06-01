package repository

import (
	"reflect"
	"strings"
	"tupeuxcourrir_api/models"
)

func getPKFieldSelfCOLUMNTagFromModel(model interface{}) string {
	reflectModel := reflect.TypeOf(model).Elem()
	var field reflect.StructField

	var ormTags []string
	var isPk bool

	for i := 0; i < reflectModel.NumField(); i++ {
		field = reflectModel.Field(i)
		if v, ok := field.Tag.Lookup("orm"); ok {
			ormTags = strings.Split(v, ";")

			for _, vOfData := range ormTags {
				if vOfData == "PK" {
					isPk = true
					break
				}
			}

			if isPk {
				break
			}
		}
	}

	if isPk {
		for _, vOfData := range ormTags {
			if strings.Contains(vOfData, "SelfCOLUMN") {
				return strings.Split(vOfData, ":")[1]
			}
		}
	}

	panic("no self column in pk model tag")
}

func getAssociatedColumnFromReverse(target interface{}, targetStructFields reflect.Value) string {
	var associatedColumn string
	typeOfApplierQueryModel := reflect.TypeOf(target).Elem()
	var field reflect.Value
	for i := 0; i < targetStructFields.NumField(); i++ {
		field = targetStructFields.Field(i)
		if field.Kind() == reflect.Ptr {
			if v, ok := field.Interface().(*models.ManyToOneRelationShip); ok {
				if reflect.TypeOf(v.Target).Elem() == typeOfApplierQueryModel {
					associatedColumn = v.AssociateColumn
					break
				}
			}
		}
	}

	return associatedColumn
}
