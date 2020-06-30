package models

import "tupeuxcourrir_api/orm"

type UsersRole struct {
	IdUsersRoles int `orm:"PK"`
	UsersIdUser  int
	RolesIdRole  int
	User         *orm.ManyToOneRelationShip
	Role         *orm.ManyToOneRelationShip
}

func (usersRoles *UsersRole) PutRelationshipConfig() {
	usersRoles.User = &orm.ManyToOneRelationShip{Target: &User{}, AssociateColumn: "UsersIdUser"}
	usersRoles.Role = &orm.ManyToOneRelationShip{Target: &Role{}, AssociateColumn: "RolesIdRole"}
}

func NewUsersRole() orm.Model {
	usersRoles := UsersRole{}
	return &usersRoles
}
