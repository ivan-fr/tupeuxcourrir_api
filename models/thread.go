package models

import "time"

type Thread struct {
	IdThread              int `orm:"PK;SelfCOLUMN:idThread"`
	CreatedAt             time.Time
	ActiveOnThread        int
	Reciprocal            bool
	InitiatorThreadIdUser int
	ReceiverThreadIdUser  int
	InitiatorThread       *ManyToOneRelationShip
	ReceiverThread        *ManyToOneRelationShip
	Messages              *OneToManyRelationShip
}

func NewThread() Thread {
	message := NewMessage()

	thread := Thread{}
	thread.InitiatorThread = &ManyToOneRelationShip{Target: &User{}, AssociateColumn: "initiator_thread_idUser"}
	thread.ReceiverThread = &ManyToOneRelationShip{Target: &Thread{}, AssociateColumn: "receiver_thread_idUser"}
	thread.Messages = &OneToManyRelationShip{Target: &message, FieldMTO: "Thread"}
	return thread
}
