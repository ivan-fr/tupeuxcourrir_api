package orm

import (
	"fmt"
	"strings"
)

type ContextAdapterFactory struct {
	getAliasFunc func(fieldRelationshipName, targetModelName string) string
}

func (cAF *ContextAdapterFactory) adaptColumnWithAlias(stringToSplit string, ptrStringToUpdate *string) bool {
	result := true
	dotSplit := strings.Split(stringToSplit, ".")
	switch len(dotSplit) {
	case 3:
		*ptrStringToUpdate = fmt.Sprintf("%v.%v",
			cAF.getAliasFunc(dotSplit[0], dotSplit[1]),
			dotSplit[2])
	case 2:
		*ptrStringToUpdate = fmt.Sprintf("%v.%v",
			cAF.getAliasFunc(dotSplit[0], ""),
			dotSplit[1])
	default:
		result = false
	}

	return result
}

func (cAF *ContextAdapterFactory) adaptMapStringInterface(mapStringInterface H, aggregate bool) {
	var mapToMerge = make(H)

	for key, aInterface := range mapStringInterface {

		if aggregate {
			switch aInterface.(type) {
			case string:
				theString := mapStringInterface[key].(string)
				cAF.adaptColumnWithAlias(aInterface.(string), &theString)
				mapStringInterface[key] = theString
			case []interface{}:
				theSlice := aInterface.([]interface{})
				theString := theSlice[0].(string)
				cAF.adaptColumnWithAlias(theSlice[0].(string), &theString)
				theSlice[0] = theString
			}
		} else {
			var theKey string

			if cAF.adaptColumnWithAlias(key, &theKey) {
				mapToMerge[theKey] = mapStringInterface[key]
				delete(mapStringInterface, key)
			}
		}
	}

	for key, aInterface := range mapToMerge {
		mapStringInterface[key] = aInterface
	}
}

func (cAF *ContextAdapterFactory) adaptLogical(logical *Logical, aggregate bool) {
	for _, combination := range logical.combinations {
		cAF.adaptContext(combination, aggregate)
	}
}

func (cAF *ContextAdapterFactory) adaptContext(context interface{}, aggregate bool) {
	switch context.(type) {
	case []string:
		sliceString := context.([]string)
		for i, valueString := range sliceString {
			cAF.adaptColumnWithAlias(valueString, &valueString)
			sliceString[i] = valueString
		}
	case H:
		cAF.adaptMapStringInterface(context.(H), aggregate)
	case *Logical:
		cAF.adaptLogical(context.(*Logical), aggregate)
	}
}
