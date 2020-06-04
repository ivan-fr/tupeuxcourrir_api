package models

import "tupeuxcourrir_api/orm"

type Role struct {
	IdRoles int `orm:"PK"`
	Role    string
	Users   *orm.ManyToManyRelationShip
}

func NewRole() *Role {
	usersRoles := NewUsersRole()
	role := Role{}
	role.Users = &orm.ManyToManyRelationShip{Target: &User{}, IntermediateTarget: usersRoles}

	return &role
}
