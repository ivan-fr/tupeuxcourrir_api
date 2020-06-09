package models

import "tupeuxcourrir_api/orm"

type UsersRole struct {
	IdUsersRoles int `orm:"PK"`
	UsersIdUser  int
	RolesIdRole  int
	User         *orm.ManyToOneRelationShip
	Role         *orm.ManyToOneRelationShip
}

func NewUsersRole() *UsersRole {
	usersRoles := UsersRole{}
	usersRoles.User = &orm.ManyToOneRelationShip{Target: &User{}, AssociateColumn: "UsersIdUser"}
	usersRoles.Role = &orm.ManyToOneRelationShip{Target: &Role{}, AssociateColumn: "RolesIdRole"}

	return &usersRoles
}
