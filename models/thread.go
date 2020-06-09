package models

import (
	"time"
	"tupeuxcourrir_api/orm"
)

type Thread struct {
	IdThread              int `orm:"PK"`
	CreatedAt             time.Time
	ActiveOnThread        int
	Reciprocal            bool
	InitiatorThreadIdUser int
	ReceiverThreadIdUser  int
	InitiatorThread       *orm.ManyToOneRelationShip
	ReceiverThread        *orm.ManyToOneRelationShip
	Messages              *orm.OneToManyRelationShip
}

func NewThread() *Thread {
	message := NewMessage()

	thread := Thread{}
	thread.InitiatorThread = &orm.ManyToOneRelationShip{Target: &User{}, AssociateColumn: "InitiatorThreadIdUser"}
	thread.ReceiverThread = &orm.ManyToOneRelationShip{Target: &Thread{}, AssociateColumn: "ReceiverThreadIdUser"}
	thread.Messages = &orm.OneToManyRelationShip{Target: message, FieldMTO: "Thread"}
	return &thread
}
