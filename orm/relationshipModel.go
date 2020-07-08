package orm

// Model ...
type Model interface {
	PutRelationshipConfig()
}

// ManyToManyRelationShip ...
type ManyToManyRelationShip struct {
	Target                      Model `json:"-"`
	IntermediateTarget          Model `json:"-"`
	EffectiveTargets            []interface{}
	EffectiveIntermediateTarget interface{}
}

// ManyToOneRelationShip ...
type ManyToOneRelationShip struct {
	Target          Model `json:"-"`
	EffectiveTarget interface{}
	AssociateColumn string `json:"-"`
}

// OneToManyRelationShip ...
type OneToManyRelationShip struct {
	Target           Model `json:"-"`
	EffectiveTargets []interface{}
	FieldMTO         string `json:"-"`
}
