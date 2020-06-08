package main

import (
	"fmt"
	"tupeuxcourrir_api/db"
	"tupeuxcourrir_api/orm"
)

func main() {
	defer db.DeferClose()

	s := orm.PutIntermediateString(",", "setter",
		map[string]interface{}{"COUNT__in": []interface{}{"3", 5, "Now()", nil}, "lol__lte": "55"})

	fmt.Printf("p%vp\n", s)
}
