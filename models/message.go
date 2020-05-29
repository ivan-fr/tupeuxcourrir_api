package models

import "time"

type Message struct {
	IdMessage int
	Message   string
	CreatedAt time.Time
	UpdatedAt time.Time
	UserId    int
	ThreadId  int
	User      *ManyToOneRelationShip
	Thread    *ManyToOneRelationShip
}

func NewMessage() Message {
	message := Message{}
	message.User = &ManyToOneRelationShip{Target: &User{}, AssociateColumn: "User_idUser"}
	message.Thread = &ManyToOneRelationShip{Target: &Role{}, AssociateColumn: "Thread_idThread"}

	return message
}
