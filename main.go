package main

import (
	"fmt"
	"tupeuxcourrir_api/db"
	"tupeuxcourrir_api/orm"
)

func main() {
	defer db.DeferClose()

	s := orm.PutIntermediateString(",", "aggregate",
		map[string]interface{}{"COUNT__in": map[string]interface{}{
			"IdUser": []string{"3", "5"},
		}})

	fmt.Printf("p%vp\n", s)
}
