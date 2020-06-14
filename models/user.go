package models

import (
	"time"
	"tupeuxcourrir_api/orm"
)

type User struct {
	IdUser                   int    `orm:"PK"`
	Email                    string `form:"email"`
	EncryptedPassword        string `form:"password"`
	FirstName                string `form:"firstName"`
	LastName                 string `form:"lastName"`
	Pseudo                   string `form:"pseudo"`
	PhotoPath                string
	City                     string `form:"city"`
	Street                   string `form:"street"`
	PostalCode               string `form:"postalCode"`
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
