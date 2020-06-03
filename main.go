package main

import (
	"fmt"
	"tupeuxcourrir_api/db"
	"tupeuxcourrir_api/models"
	"tupeuxcourrir_api/orm"
)

func main() {
	defer db.DeferClose()

	ref := models.NewUser()
	u := models.User{}
	p := make([]interface{}, 0)
	u.Pseudo = "ivan"
	u.Email = "troplolol"
	u.EncryptedPassword = "popop"
	u.Salt = "ja"
	u.FirstName = "popo"
	u.LastName = "bajon"
	p = append(p, &u)

	iQueryBuilder := orm.NewInsertQueryBuilder(ref, p)
	fmt.Println(iQueryBuilder.ApplyInsert())
}
