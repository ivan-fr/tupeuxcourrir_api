package orm

import (
	"fmt"
	"reflect"
)

type Logical struct {
	intermediateString string
	combinations       []interface{}
}

func (logical *Logical) factorLogical(interfaces ...interface{}) {
	usedMultipleLogical := false

	for _, aInterface := range interfaces {
		switch aInterface.(type) {
		case []string, map[string]interface{}:
			if reflect.ValueOf(aInterface).Len() == 1 {
				panic("you should use Logical.Single method instead")
			}

			if usedMultipleLogical {
				panic("only one []string or map[string]interface{}")
			} else {
				usedMultipleLogical = true
			}
		case *Logical:
		default:
			panic("undefined configuration")
		}

		logical.combinations = append(logical.combinations, aInterface)
	}
}

func (logical *Logical) GetSentence(mapSetterMode string) (string, []interface{}) {
	var sentences []string
	var stmts []interface{}

	for _, combination := range logical.combinations {
		switch combination.(type) {
		case *Logical:
			aLogical := combination.(*Logical)
			str, stmtsPart := aLogical.GetSentence(mapSetterMode)
			stmts = append(stmts, stmtsPart...)
			sentences = append(sentences, fmt.Sprintf("(%v)", str))
		default:
			sSA := &SQLSectionArchitecture{intermediateString: logical.intermediateString,
				isStmts: true,
				mode:    mapSetterMode,
				context: combination}
			sSA.constructSQlSection()
			stmts = append(stmts, sSA.valuesFromStmts...)
			sentences = append(sentences, sSA.SQLSection)
		}
	}

	sSA := &SQLSectionArchitecture{intermediateString: logical.intermediateString,
		isStmts: false,
		mode:    "space",
		context: sentences}
	sSA.constructSQlSection()
	return sSA.SQLSection, stmts
}

func Or(interfaces ...interface{}) *Logical {
	logical := &Logical{intermediateString: " OR"}
	logical.factorLogical(interfaces...)
	return logical
}

func And(interfaces ...interface{}) *Logical {
	logical := &Logical{intermediateString: " AND"}
	logical.factorLogical(interfaces...)
	return logical
}

func XOr(interfaces ...interface{}) *Logical {
	logical := &Logical{intermediateString: " XOR"}
	logical.factorLogical(interfaces...)
	return logical
}
