package models

type UsersRole struct {
	IdUsersRoles int `orm:"PK;SelfCOLUMN:idUsersRoles"`
	UsersIdUser  int
	RolesIdRole  int
	User         *ManyToOneRelationShip
	Role         *ManyToOneRelationShip
}

func NewUsersRole() UsersRole {
	usersRoles := UsersRole{}
	usersRoles.User = &ManyToOneRelationShip{Target: &User{}, AssociateColumn: "Users_idUser"}
	usersRoles.Role = &ManyToOneRelationShip{Target: &Role{}, AssociateColumn: "Roles_idRoles"}

	return usersRoles
}
