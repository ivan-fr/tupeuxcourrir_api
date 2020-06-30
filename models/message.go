package models

import (
	"time"
	"tupeuxcourrir_api/orm"
)

type Message struct {
	IdMessage      int `orm:"PK"`
	Message        string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	UserIdUser     int
	ThreadIdThread int
	User           *orm.ManyToOneRelationShip
	Thread         *orm.ManyToOneRelationShip
}

func (message *Message) PutRelationshipConfig() {
	message.User = &orm.ManyToOneRelationShip{Target: &User{}, AssociateColumn: "UserIdUser"}
	message.Thread = &orm.ManyToOneRelationShip{Target: &Role{}, AssociateColumn: "ThreadIdThread"}
}

func NewMessage() interface{} {
	message := Message{}
	return &message
}
