package orm

type ManyToManyRelationShip struct {
	Target                      interface{}
	IntermediateTarget          interface{}
	EffectiveTargets            []interface{}
	EffectiveIntermediateTarget interface{}
}

type ManyToOneRelationShip struct {
	Target          interface{}
	EffectiveTarget interface{}
	AssociateColumn string
}

type OneToManyRelationShip struct {
	Target           interface{}
	EffectiveTargets []interface{}
	FieldMTO         string
}
