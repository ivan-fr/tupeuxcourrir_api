package main

import (
	"fmt"
	"tupeuxcourrir_api/db"
	"tupeuxcourrir_api/models"
	"tupeuxcourrir_api/orm"
)

func main() {
	defer db.DeferClose()
	defer db.DeferClose()

	sQueryBuilder := orm.GetSelectQueryBuilder(models.NewUser()).
		FindBy(map[string]interface{}{"id": "lol", "koko": "popo", "giro": nil}).
		Consider("InitiatedThread").
		Consider("ReceivedThread").
		Consider("Roles").
		OrderBy(map[string]interface{}{"bibi": "", "lolo": "DESC", "palopalo": "DESC"})
	fmt.Println(sQueryBuilder.ConstructSql())
	fmt.Println(sQueryBuilder.GetStmts())
}
