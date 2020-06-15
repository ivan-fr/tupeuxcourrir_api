package models

import (
	"database/sql"
	"tupeuxcourrir_api/orm"
)

type User struct {
	IdUser                   int `orm:"PK"`
	Email                    string
	EncryptedPassword        string
	FirstName                sql.NullString
	LastName                 string
	Pseudo                   string
	PhotoPath                sql.NullString
	City                     sql.NullString
	Street                   sql.NullString
	PostalCode               sql.NullString
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
