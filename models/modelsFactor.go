package models

type ManyToManyRelationShip struct {
	Target             interface{}
	IntermediateTarget interface{}
}

type ManyToOneRelationShip struct {
	Target          interface{}
	AssociateColumn string
}

type OneToManyRelationShip struct {
	Target   interface{}
	FieldMTO string
}
