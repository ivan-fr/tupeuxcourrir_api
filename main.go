package main

import (
	"fmt"
	"tupeuxcourrir_api/db"
	"tupeuxcourrir_api/models"
	"tupeuxcourrir_api/orm"
)

func main() {
	defer db.DeferClose()

	sQueryBuilder := orm.NewSelectQueryBuilder(models.NewUser()).
		Consider("InitiatedThread").
		Consider("ReceivedThread").
		Consider("Roles")
	fmt.Println(sQueryBuilder.ApplyQueryRow())
}
