package main

import (
	"fmt"
	"tupeuxcourrir_api/db"
	"tupeuxcourrir_api/models"
	"tupeuxcourrir_api/orm"
)

func main() {
	defer db.DeferClose()

	uQueryBuilder := orm.NewUpdateQueryBuilder(models.NewUser())
	uQueryBuilder.Where(map[string]interface{}{"IdUser": 9})

	fmt.Println(uQueryBuilder.ConstructSql())
}
