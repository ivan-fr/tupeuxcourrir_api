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

func (thread *Thread) PutRelationshipConfig() {
	thread.InitiatorThread = &orm.ManyToOneRelationShip{Target: &User{}, AssociateColumn: "InitiatorThreadIdUser"}
	thread.ReceiverThread = &orm.ManyToOneRelationShip{Target: &Thread{}, AssociateColumn: "ReceiverThreadIdUser"}
	thread.Messages = &orm.OneToManyRelationShip{Target: &Message{}, FieldMTO: "Thread"}
}

func NewThread() interface{} {
	thread := Thread{}
	return &thread
}
