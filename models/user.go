package models

import (
	"time"
	"tupeuxcourrir_api/orm"
)

type User struct {
	IdUser                   int `orm:"PK;SelfCOLUMN:idUser"`
	Email                    string
	EncryptedPassword        string
	Salt                     string
	FirstName                string
	LastName                 string
	Pseudo                   string
	PhotoPath                string
	City                     string
	Street                   string
	PostalCode               string
	CheckedEmail             bool
	SentValidateMailAt       time.Time
	SentChangePasswordMailAt time.Time
	CreatedAt                time.Time
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
