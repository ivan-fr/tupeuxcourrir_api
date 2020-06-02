package main

import (
	"fmt"
	"tupeuxcourrir_api/db"
	"tupeuxcourrir_api/models"
	"tupeuxcourrir_api/orm"
)

func main() {
	defer db.DeferClose()

	var u = models.NewUser()
	sQueryBuilder := orm.NewSelectQueryBuilder(u).
		FindBy(map[string]string{"id": "lol", "koko": "popo", "giro": "pipo"}).
		Consider("InitiatedThread").
		Consider("ReceivedThread").
		Consider("Roles").
		OrderBy(map[string]string{"bibi": "", "lolo": "DESC", "palopalo": "DESC"})
	fmt.Println(sQueryBuilder.ConstructSql())
}
