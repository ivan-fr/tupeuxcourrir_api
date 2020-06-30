package orm

type Model interface {
	PutRelationshipConfig()
}

type ManyToManyRelationShip struct {
	Target                      Model
	IntermediateTarget          Model
	EffectiveTargets            []interface{}
	EffectiveIntermediateTarget interface{}
}

type ManyToOneRelationShip struct {
	Target          Model
	EffectiveTarget interface{}
	AssociateColumn string
}

type OneToManyRelationShip struct {
	Target           Model
	EffectiveTargets []interface{}
	FieldMTO         string
}
