package models

import (
	"database/sql"
	"tupeuxcourrir_api/orm"
)

type User struct {
	IdUser                   int `orm:"PK"`
	Email                    string
	EncryptedPassword        string
	FirstName                string
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
	InitiatedThreads         *orm.OneToManyRelationShip
	ReceivedThreads          *orm.OneToManyRelationShip
}

func NewUser() *User {
	usersRoles := NewUsersRole()
	thread := NewThread()
	user := User{}

	user.Roles = &orm.ManyToManyRelationShip{Target: &Role{}, IntermediateTarget: usersRoles}
	user.InitiatedThreads = &orm.OneToManyRelationShip{Target: thread, FieldMTO: "InitiatorThread"}
	user.ReceivedThreads = &orm.OneToManyRelationShip{Target: thread, FieldMTO: "ReceiverThread"}

	return &user
}
