package models

import (
	"time"
	"tupeuxcourrir_api/orm"
)

type Thread struct {
	IdThread              int `orm:"PK;SelfCOLUMN:idThread"`
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
	thread.InitiatorThread = &orm.ManyToOneRelationShip{Target: &User{}, AssociateColumn: "initiator_thread_idUser"}
	thread.ReceiverThread = &orm.ManyToOneRelationShip{Target: &Thread{}, AssociateColumn: "receiver_thread_idUser"}
	thread.Messages = &orm.OneToManyRelationShip{Target: message, FieldMTO: "Thread"}
	return &thread
}
