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
		Where(orm.Or(map[string]interface{}{"ReceivedThread.IdThread__in": []int{1, 18, 9, 8}, "koko": 9, "giro": nil},
			orm.And(map[string]interface{}{"Re": "loli", "ReceivedThread.IdThread": 63, "giro": nil}))).
		Having(orm.And(map[string]interface{}{"COUNT__in": []interface{}{"ReceivedThread.IdThread", []int{1, 5, 9, 8}},
			"AVG__gt": []interface{}{"IdUser", 13}})).
		OrderBy(map[string]interface{}{"bibi": "", "lolo": "DESC", "palopalo": "DESC"}).
		GroupBy([]string{"ReceivedThread.IdThread", "Koko", "Pipi"})

	sQueryBuilder.RollUp = true

	fmt.Println(sQueryBuilder.ApplyQueryRow())
}
