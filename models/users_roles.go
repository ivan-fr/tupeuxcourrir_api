package models

type UsersRoles struct {
	IdUsersRoles int `orm:"AI;SelfCOLUMN:idUsers_roles"`
	UsersIdUser  int
	RolesIdRole  int
	User         *ManyToOneRelationShip
	Role         *ManyToOneRelationShip
}

func NewUsersRoles() UsersRoles {
	usersRoles := UsersRoles{}
	usersRoles.User = &ManyToOneRelationShip{Target: &User{}, AssociateColumn: "Users_idUser"}
	usersRoles.Role = &ManyToOneRelationShip{Target: &Role{}, AssociateColumn: "Roles_idRoles"}

	return usersRoles
}
