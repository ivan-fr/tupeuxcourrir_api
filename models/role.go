package models

type Role struct {
	IdRoles int `orm:"PK;SelfCOLUMN:idRoles"`
	Role    string
	Users   *ManyToManyRelationShip
}

func NewRole() Role {
	usersRoles := NewUsersRole()
	role := Role{}
	role.Users = &ManyToManyRelationShip{Target: &Thread{}, IntermediateTarget: &usersRoles}

	return role
}
