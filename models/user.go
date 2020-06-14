package models

import (
	"database/sql"
	"tupeuxcourrir_api/orm"
)

type User struct {
	IdUser                   sql.NullInt64 `orm:"PK"`
	Email                    string        `form:"email"`
	EncryptedPassword        string        `form:"password"`
	FirstName                string        `form:"firstName"`
	LastName                 string        `form:"lastName"`
	Pseudo                   string        `form:"pseudo"`
	PhotoPath                sql.NullString
	City                     sql.NullString `form:"city"`
	Street                   sql.NullString `form:"street"`
	PostalCode               sql.NullString `form:"postalCode"`
	CheckedEmail             bool
	SentValidateMailAt       sql.NullTime
	SentChangePasswordMailAt sql.NullTime
	CreatedAt                sql.NullTime
	Roles                    *orm.ManyToManyRelationShip
	InitiatedThread          *orm.OneToManyRelationShip
	ReceivedThread           *orm.OneToManyRelationShip
}

func NewUser() *User {
	usersRoles := NewUsersRole()
	thread := NewThread()
	user := User{}

	user.Roles = &orm.ManyToManyRelationShip{Target: &Role{}, IntermediateTarget: usersRoles}
	user.InitiatedThread = &orm.OneToManyRelationShip{Target: thread, FieldMTO: "InitiatorThread"}
	user.ReceivedThread = &orm.OneToManyRelationShip{Target: thread, FieldMTO: "ReceiverThread"}

	return &user
}
