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
	t := models.User{}
	p := make([]interface{}, 0)
	u.Pseudo = "ivan"
	u.Email = "tropjpopijlololl"
	u.EncryptedPassword = "popop"
	u.Salt = "ja"
	u.FirstName = "popo"
	u.LastName = "bajon"
	t.Pseudo = "ivan"
	t.Email = "pokpopk"
	t.EncryptedPassword = "popop"
	t.Salt = "ja"
	t.FirstName = "popo"
	t.LastName = "bajon"
	p = append(p, &u, &t)

	iQueryBuilder := orm.NewInsertQueryBuilder(ref, p)
	fmt.Println(iQueryBuilder.ApplyInsert())
}
