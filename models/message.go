package models

import (
	"time"
	"tupeuxcourrir_api/orm"
)

type Message struct {
	IdMessage int `orm:"PK;SelfCOLUMN:idMessage"`
	Message   string
	CreatedAt time.Time
	UpdatedAt time.Time
	UserId    int
	ThreadId  int
	User      *orm.ManyToOneRelationShip
	Thread    *orm.ManyToOneRelationShip
}

func NewMessage() *Message {
	message := Message{}
	message.User = &orm.ManyToOneRelationShip{Target: &User{}, AssociateColumn: "User_idUser"}
	message.Thread = &orm.ManyToOneRelationShip{Target: &Role{}, AssociateColumn: "Thread_idThread"}

	return &message
}
