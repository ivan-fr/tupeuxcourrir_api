package models

type Role struct {
	IdRoles int `orm:"AI;SelfCOLUMN:idRoles"`
	Role    string
	Users   *ManyToManyRelationShip
}

func NewRole() Role {
	usersRoles := NewUsersRoles()
	role := Role{}
	role.Users = &ManyToManyRelationShip{Target: &Thread{}, IntermediateTarget: &usersRoles}

	return role
}
