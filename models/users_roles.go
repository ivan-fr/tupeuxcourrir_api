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
	usersRoles.User = &orm.ManyToOneRelationShip{Target: &User{}, AssociateColumn: "Users_idUser"}
	usersRoles.Role = &orm.ManyToOneRelationShip{Target: &Role{}, AssociateColumn: "Roles_idRoles"}

	return &usersRoles
}
