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

func (user *User) PutRelationshipConfig() {
	user.Roles = &orm.ManyToManyRelationShip{Target: &Role{}, IntermediateTarget: &UsersRole{}}
	user.InitiatedThreads = &orm.OneToManyRelationShip{Target: &Thread{}, FieldMTO: "InitiatorThread"}
	user.ReceivedThreads = &orm.OneToManyRelationShip{Target: &Thread{}, FieldMTO: "ReceiverThread"}
}

func NewUser() *User {
	user := User{}
	return &user
}
