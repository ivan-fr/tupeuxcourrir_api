package orm

import (
	"fmt"
	"strings"
)

type AliasFactory struct {
	getAliasFunc func(fieldRelationshipName, targetModelName string) string
}

func (aliasFactory *AliasFactory) adaptColumnWithAlias(stringToSplit string, ptrStringToUpdate *string) bool {
	result := true
	dotSplit := strings.Split(stringToSplit, ".")
	switch len(dotSplit) {
	case 3:
		*ptrStringToUpdate = fmt.Sprintf("%v.%v",
			aliasFactory.getAliasFunc(dotSplit[0], dotSplit[1]),
			dotSplit[2])
	case 2:
		*ptrStringToUpdate = fmt.Sprintf("%v.%v",
			aliasFactory.getAliasFunc(dotSplit[0], ""),
			dotSplit[1])
	default:
		result = false
	}

	return result
}

func (aliasFactory *AliasFactory) adaptMapStringInterface(mapStringInterface map[string]interface{}, aggregate bool) {
	var mapToMerge = make(map[string]interface{})

	for key, aInterface := range mapStringInterface {

		if aggregate {
			switch aInterface.(type) {
			case string:
				theString := mapStringInterface[key].(string)
				aliasFactory.adaptColumnWithAlias(aInterface.(string), &theString)
				mapStringInterface[key] = theString
			case []interface{}:
				theSlice := aInterface.([]interface{})
				theString := theSlice[0].(string)
				aliasFactory.adaptColumnWithAlias(theSlice[0].(string), &theString)
				theSlice[0] = theString
			}
		} else {
			var theKey string

			if aliasFactory.adaptColumnWithAlias(key, &theKey) {
				mapToMerge[theKey] = mapStringInterface[key]
				delete(mapStringInterface, key)
			}
		}
	}

	for key, aInterface := range mapToMerge {
		mapStringInterface[key] = aInterface
	}
}

func (aliasFactory *AliasFactory) adaptLogical(logical *Logical, aggregate bool) {
	for _, combination := range logical.combinations {
		aliasFactory.adaptContext(combination, aggregate)
	}
}

func (aliasFactory *AliasFactory) adaptContext(context interface{}, aggregate bool) {
	switch context.(type) {
	case []string:
		sliceString := context.([]string)
		for i, valueString := range sliceString {
			aliasFactory.adaptColumnWithAlias(valueString, &valueString)
			sliceString[i] = valueString
		}
	case map[string]interface{}:
		aliasFactory.adaptMapStringInterface(context.(map[string]interface{}), aggregate)
	case *Logical:
		aliasFactory.adaptLogical(context.(*Logical), aggregate)
	}
}
