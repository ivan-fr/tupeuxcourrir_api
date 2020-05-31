package models

import (
	"time"
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
	Roles                    *ManyToManyRelationShip
	InitiatedThread          *OneToManyRelationShip
	ReceivedThread           *OneToManyRelationShip
}

func NewUser() User {
	usersRoles := NewUsersRole()
	thread := NewThread()

	user := User{}
	user.Roles = &ManyToManyRelationShip{Target: &Role{}, IntermediateTarget: &usersRoles}
	user.InitiatedThread = &OneToManyRelationShip{Target: &thread, FieldMTO: "InitiatorThread"}
	user.ReceivedThread = &OneToManyRelationShip{Target: &thread, FieldMTO: "ReceiverThread"}

	return user
}
