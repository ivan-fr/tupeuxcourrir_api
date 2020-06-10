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
		Where(map[string]interface{}{"id": "lol", "koko": 9, "giro": nil}).
		Aggregate(map[string]interface{}{"COUNT": "*", "AVG": "IdUser"}).
		Having(map[string]interface{}{"COUNT__lte": []interface{}{"*", 10}, "AVG__gt": []interface{}{"IdUser", 13}}).
		Consider("InitiatedThread").
		Consider("ReceivedThread").
		Consider("Roles").
		OrderBy(map[string]interface{}{"bibi": "", "lolo": "DESC", "palopalo": "DESC"}).
		GroupBy([]string{"IdUser", "Koko", "Pipi"})
	sQueryBuilder.RollUp = true

	fmt.Println(sQueryBuilder.ConstructSql())
	fmt.Println(sQueryBuilder.GetStmts())

	sQueryBuilder.Clean()

	sQueryBuilder = sQueryBuilder.
		Where(map[string]interface{}{"id": "lol", "koko": 9, "giro": nil}).
		Aggregate(map[string]interface{}{"COUNT": "*", "AVG": "IdUser"}).
		Having(map[string]interface{}{"COUNT__lte": []interface{}{"*", 10}, "AVG__gt": []interface{}{"IdUser", 13}}).
		Consider("InitiatedThread").
		Consider("ReceivedThread").
		Consider("Roles").
		OrderBy(map[string]interface{}{"bibi": "", "lolo": "DESC", "palopalo": "DESC"}).
		GroupBy([]string{"IdUser", "Koko", "Pipi"})
	fmt.Println(sQueryBuilder.ConstructSql())
	fmt.Println(sQueryBuilder.GetStmts())
}
