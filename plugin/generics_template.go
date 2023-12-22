package plugin

import "text/template"

var genericsTemplate = template.Must(template.New("generics").Funcs(genericsTemplateFuncs).Parse(`
{{ $messages := .messages }}
// Protos is a union of other types that defines which types may be used in generic functions
type Protos interface {
	{{ range $i, $m := .messages -}}
	{{ if not .Ignore -}}
	*{{ .GoIdent.GoName }} {{- if pipe $i $messages -}} | {{- end -}}
	{{ end -}}
	{{ end }}
	GetProtoId() *string
	SetProtoId(string)
}

// Models is a union of other types that defines which types may be used in generic functions
type Models interface {
	{{ range $i, $m := .messages -}}
	{{ if not .Ignore -}}
	*{{ .GoIdent.GoName }}GormModel {{- if pipe $i $messages -}} | {{- end -}}
	{{ end -}}
	{{ end }}
	GetModelId() *string
	SetModelId(string)
	New() interface{}
}

// Proto[M Models] is an interface type that defines behavior for the implementer of a given Models type
type Proto[M Models] interface {
	GetProtoId() *string
	SetProtoId(string)
	ToModel() (M, error)
}

// Model[P Protos] is an interface type that defines behavior for the implementer of a given Protos type
type Model[P Protos] interface {
	ToProto() (P, error)
}

// ToModels converts an array of protos to an array of gorm db models by calling the proto's ToModel method
func ToModels[P Protos, M Models](protos interface{}) ([]M, error) {
	converted := ConvertProtosToProtosM[P, M](protos)
	models := []M{}
	for _, proto := range converted {
		model, err := proto.ToModel()
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	return models, nil
}

// ConvertProtosToProtosM converts a given slice of protos into an array of the Proto interface type, which can then
// leverage the interface methods
func ConvertProtosToProtosM[P Protos, M Models](protos interface{}) []Proto[M] {
	assertedProtos := protos.([]P)
	things := make([]Proto[M], len(assertedProtos))
	for i, v := range assertedProtos {
		things[i] = ConvertProtoToProtosM[P, M](v)
	}
	return things
}

// ConvertProtoToProtosM converts a single proto to a Proto[M]
func ConvertProtoToProtosM[P Protos, M Models](proto interface{}) Proto[M] {
	return any(proto).(Proto[M])
}

// ConvertProtosToProtosM converts a given slice of protos into an array of the Proto interface type, which can then
// leverage the interface methods
func ConvertModelsToModelsP[P Protos, M Models](models interface{}) []Model[P] {
	assertedModels := models.([]M)
	things := make([]Model[P], len(assertedModels))
	for i, v := range assertedModels {
		things[i] = ConvertModelToModelP[P, M](v)
	}
	return things
}

// ConvertProtoToProtosM converts a single proto to a Proto[M]
func ConvertModelToModelP[P Protos, M Models](model interface{}) Model[P] {
	return any(model).(Model[P])
}

// ToProtos converts an array of models into an array of protos by calling the model's ToProto method
func ToProtos[P Protos, M Models](models interface{}) ([]P, error) {
	converted := ConvertModelsToModelsP[P, M](models)
	protos := []P{}
	for _, model := range converted {
		proto, err := model.ToProto()
		if err != nil {
			return nil, err
		}
		protos = append(protos, proto)
	}
	return protos, nil
}

// Upsert is a generic function that will upsert any of the generated protos, returning the upserted models. Upsert
// excludes all associations, and uses an on conflict clause to handle upsert. A function may be provided to be executed
// during the transaction. The function is executed after the upsert. If the function returns an error, the transaction
// will be rolled back.
func Upsert[P Protos, M Models](ctx context.Context, db *gorm.DB, protos interface{}) ([]M, error) {
	converted := ConvertProtosToProtosM[P, M](protos)
	if len(converted) > 0 {
		models := []M{}
		for _, proto := range converted {
			if proto.GetProtoId() == nil {
				proto.SetProtoId(uuid.New().String())
			}
			model, err := proto.ToModel()
			if err != nil {
				return nil, err
			}
			models = append(models, model)
		}
		session := db.Session(&gorm.Session{})
		err := session.
			// on conflict, update all fields
			Clauses(clause.OnConflict{
				UpdateAll: true,
			}).
			// exclude associations from upsert
			Omit(clause.Associations).
			Create(&models).Error

		return models, err
	}
	return nil, nil
}

// Delete is a generic function that will delete any of the generated protos. A function may be provided to be executed
// during the transaction. The function is executed after the delete. If the function returns an error, the transaction
// will be rolled back.
func Delete[M Models](ctx context.Context, db *gorm.DB, ids []string) ([]M, error) {
	if len(ids) > 0 {
		session := db.Session(&gorm.Session{})
		models := []M{}
		err := session.Where("id in ?", ids).Delete(&models).Error
		return models, err
	}
	return nil, nil
}

// List lists the given model type
func List[M Models](ctx context.Context, db *gorm.DB, limit, offset int, orderBy string, preloads []string) ([]M, error) {
	session := db.Session(&gorm.Session{}).WithContext(ctx)
	// set limit
	if limit > 0 {
		session = session.Limit(limit)
	}
	// set offset
	if offset > 0 {
		session = session.Offset(offset)
	}
	// set preloads
	for _, preload := range preloads {
		session = session.Preload(preload)
	}
	// set order by
	if orderBy != "" {
		session = session.Order(orderBy)
	}
	// execute
	var models []M
	err := session.Find(&models).Error
	return models, err
}

// GetByIds gets the given model type by id
func GetByIds[M Models](ctx context.Context, db *gorm.DB, ids []string, preloads []string) ([]M, error) {
	session := db.Session(&gorm.Session{}).WithContext(ctx)
	// set preloads
	for _, preload := range preloads {
		session = session.Preload(preload)
	}
	models := []M{}
	err := session.Where("id in ?", ids).Find(&models).Error
	return models, err
}

// ManyToManyAssociations is a sync map with helper functions. I'm using a sync.map so that it's thread safe, and
// a struct to allow us to easily define behavior we can use elsewhere
type ManyToManyAssociations struct {
	data sync.Map
}

func (m *ManyToManyAssociations) Associations() map[string][]string {
	associations := map[string][]string{}
	m.data.Range(func(key, value any) bool {
		associations[key.(string)] = value.([]string)
		return true
	})
	return associations
}

func (m *ManyToManyAssociations) AddAssociation(modelId, associatedId string) {
	var associations []string
	val, ok := m.data.Load(modelId)
	if ok {
		associations = val.([]string)
		associations = append(associations, associatedId)
	} else {
		associations = []string{associatedId}
	}
	m.data.Store(modelId, associations)
}

func AssociateManyToMany[L Models, R Models](ctx context.Context, db *gorm.DB, associations *ManyToManyAssociations, associationName string) error {
	session := db.Session(&gorm.Session{})
	session = session.Clauses(clause.OnConflict{DoNothing: true})
	for id, associatedIds := range associations.Associations() {
		var associations []R
		var temp L
		model := temp.New().(L)
		model.SetModelId(id)
		for _, id := range associatedIds {
			var associatedTemp R
			associatedModel := associatedTemp.New().(R)
			associatedModel.SetModelId(id)
			associations = append(associations, associatedModel)
		}
		err := session.Model(&model).Association(associationName).Append(&associations)
		if err != nil {
			return err
		}
	}
	return nil
}

func DissociateManyToMany[L Models, R Models](ctx context.Context, db *gorm.DB, associations *ManyToManyAssociations, associationName string) error {
	session := db.Session(&gorm.Session{})
	for id, associatedIds := range associations.Associations() {
		var associations []R
		var temp L
		model := temp.New().(L)
		model.SetModelId(id)
		for _, id := range associatedIds {
			var associatedTemp R
			associatedModel := associatedTemp.New().(R)
			associatedModel.SetModelId(id)
			associations = append(associations, associatedModel)
		}
		txErr := session.Model(&model).Association(associationName).Delete(&associations)
		if txErr != nil {
			return txErr
		}
	}
	return nil
}
`))
