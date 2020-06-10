package main

import (
	"fmt"
	"tupeuxcourrir_api/db"
	"tupeuxcourrir_api/models"
	"tupeuxcourrir_api/orm"
)

func main() {
	defer db.DeferClose()

	sQueryBuilder := orm.GetSelectQueryBuilder(models.NewUser()).
		Consider("InitiatedThread").
		Consider("ReceivedThread").
		Consider("Roles").
		Where(map[string]interface{}{"ReceivedThread.IdThread": "lol", "koko": 9, "giro": nil}).
		Aggregate(map[string]interface{}{"COUNT": "ReceivedThread.IdThread", "AVG": "IdUser"}).
		Having(map[string]interface{}{"COUNT__lte": []interface{}{"ReceivedThread.IdThread", 10}, "AVG__gt": []interface{}{"IdUser", 13}}).
		OrderBy(map[string]interface{}{"bibi": "", "lolo": "DESC", "palopalo": "DESC"}).
		GroupBy([]string{"ReceivedThread.IdThread", "Koko", "Pipi"})

	sQueryBuilder.RollUp = true

	fmt.Println(sQueryBuilder.ConstructSql())
	fmt.Println(sQueryBuilder.GetStmts())
}
