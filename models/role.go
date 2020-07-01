package models

import "tupeuxcourrir_api/orm"

type Role struct {
	IdRoles int `orm:"PK"`
	Role    string
	Users   *orm.ManyToManyRelationShip
}

func (role *Role) PutRelationshipConfig() {
	role.Users = &orm.ManyToManyRelationShip{Target: &User{}, IntermediateTarget: &UsersRole{}}
}

func NewRole() *Role {
	role := Role{}
	return &role
}
