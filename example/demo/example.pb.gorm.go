package example

import (
	pq "github.com/lib/pq"
	lo "github.com/samber/lo"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	time "time"
)

type ThingGormModel struct {
	// @gotags: fake:"skip"
	Id *string `gorm:"type:uuid;primaryKey;default:gen_random_uuid();" json:"id" fake:"skip"`
	// @gotags: fake:"skip"
	CreatedAt *time.Time `gorm:"default:now();" json:"createdAt" fake:"skip"`
	// @gotags: fake:"skip"
	UpdatedAt *time.Time `gorm:"default:now();" json:"updatedAt" fake:"skip"`
	// @gotags: fake:"{price:0.00,1000.00}"
	ADouble float64 `json:"aDouble" fake:"{price:0.00,1000.00}"`
	// @gotags: fake:"{price:0.00,1000.00}"
	AFloat float32 `json:"aFloat" fake:"{price:0.00,1000.00}"`
	// @gotags: fake:"{int32}"
	AnInt32 int32 `json:"anInt32" fake:"{int32}"`
	// @gotags: fake:"{number:9223372036854775807}"
	AnInt64 int64 `json:"anInt64" fake:"{number:9223372036854775807}"`
	// @gotags: fake:"{bool}"
	ABool bool `json:"aBool" fake:"{bool}"`
	// @gotags: fake:"{hackerphrase}"
	AString string `json:"aString" fake:"{hackerphrase}"`
	// @gotags: fake:"skip"
	ABytes []byte `json:"aBytes" fake:"skip"`
	// @gotags: fake:"{price:0.00,1000.00}"
	Doubles pq.Float64Array `gorm:"type:float[];" json:"doubles" fake:"{price:0.00,1000.00}"`
	// @gotags: fake:"{price:0.00,1000.00}"
	Floats pq.Float32Array `gorm:"type:float[];" json:"floats" fake:"{price:0.00,1000.00}"`
	// @gotags: fake:"{int32}"
	Int32S pq.Int32Array `gorm:"type:int[];" json:"int32s" fake:"{int32}"`
	// @gotags: fake:"{number:9223372036854775807}"
	Int64S pq.Int64Array `gorm:"type:int[];" json:"int64s" fake:"{number:9223372036854775807}"`
	// @gotags: fake:"{bool}"
	Bools pq.BoolArray `gorm:"type:bool[];" json:"bools" fake:"{bool}"`
	// @gotags: fake:"{hackerphrase}"
	Strings pq.StringArray `gorm:"type:string[];" json:"strings" fake:"{hackerphrase}"`
	// @gotags: fake:"skip"
	Bytess pq.ByteaArray `gorm:"type:bytes[];" json:"bytess" fake:"skip"`
	// @gotags: fake:"skip"
	OptionalScalarField *string `json:"optionalScalarField" fake:"skip"`
	// @gotags: fake:"skip"
	BelongsToId *string ``
	// @gotags: fake:"skip"
	BelongsTo *BelongsToThingGormModel `json:"belongsTo" fake:"skip"`
	// @gotags: fake:"skip"
	HasOne *HasOneThingGormModel `gorm:"foreignKey:ThingId;" json:"hasOne" fake:"skip"`
	// @gotags: fake:"skip"
	HasMany []*HasManyThingGormModel `gorm:"foreignKey:ThingId;" json:"hasMany" fake:"skip"`
	// @gotags: fake:"skip"
	ManyToMany []*ManyToManyThingGormModel `gorm:"many2many:things_manytomanys;" json:"manyToMany" fake:"skip"`
}

func (m *ThingGormModel) TableName() string {
	return "things"
}

func (m *ThingGormModel) ToProto() *Thing {
	if m == nil {
		return nil
	}
	theProto := &Thing{}
	theProto.Id = m.Id
	if m.CreatedAt != nil {
		theProto.CreatedAt = timestamppb.New(lo.FromPtr(m.CreatedAt))
	}
	if m.UpdatedAt != nil {
		theProto.UpdatedAt = timestamppb.New(lo.FromPtr(m.UpdatedAt))
	}
	theProto.ADouble = m.ADouble
	theProto.AFloat = m.AFloat
	theProto.AnInt32 = m.AnInt32
	theProto.AnInt64 = m.AnInt64
	theProto.ABool = m.ABool
	theProto.AString = m.AString
	theProto.ABytes = m.ABytes
	theProto.Doubles = m.Doubles
	theProto.Floats = m.Floats
	theProto.Int32S = m.Int32S
	theProto.Int64S = m.Int64S
	theProto.Bools = m.Bools
	theProto.Strings = m.Strings
	theProto.Bytess = m.Bytess
	theProto.OptionalScalarField = m.OptionalScalarField
	theProto.BelongsTo = m.BelongsTo.ToProto()
	theProto.HasOne = m.HasOne.ToProto()

	if len(m.HasMany) > 0 {
		theProto.HasMany = []*HasManyThing{}
		for _, model := range m.HasMany {
			theProto.HasMany = append(theProto.HasMany, model.ToProto())
		}
	}

	if len(m.ManyToMany) > 0 {
		theProto.ManyToMany = []*ManyToManyThing{}
		for _, model := range m.ManyToMany {
			theProto.ManyToMany = append(theProto.ManyToMany, model.ToProto())
		}
	}

	return theProto
}

func (m *Thing) ToGormModel() *ThingGormModel {
	if m == nil {
		return nil
	}
	theModel := &ThingGormModel{}
	theModel.Id = m.Id
	if m.CreatedAt != nil {
		theModel.CreatedAt = lo.ToPtr(m.CreatedAt.AsTime())
	}
	if m.UpdatedAt != nil {
		theModel.UpdatedAt = lo.ToPtr(m.UpdatedAt.AsTime())
	}
	theModel.ADouble = m.ADouble
	theModel.AFloat = m.AFloat
	theModel.AnInt32 = m.AnInt32
	theModel.AnInt64 = m.AnInt64
	theModel.ABool = m.ABool
	theModel.AString = m.AString
	theModel.ABytes = m.ABytes
	theModel.Doubles = m.Doubles
	theModel.Floats = m.Floats
	theModel.Int32S = m.Int32S
	theModel.Int64S = m.Int64S
	theModel.Bools = m.Bools
	theModel.Strings = m.Strings
	theModel.Bytess = m.Bytess
	theModel.OptionalScalarField = m.OptionalScalarField
	theModel.BelongsTo = m.BelongsTo.ToGormModel()
	theModel.HasOne = m.HasOne.ToGormModel()

	if len(m.HasMany) > 0 {
		theModel.HasMany = []*HasManyThingGormModel{}
		for _, message := range m.HasMany {
			theModel.HasMany = append(theModel.HasMany, message.ToGormModel())
		}
	}

	if len(m.ManyToMany) > 0 {
		theModel.ManyToMany = []*ManyToManyThingGormModel{}
		for _, message := range m.ManyToMany {
			theModel.ManyToMany = append(theModel.ManyToMany, message.ToGormModel())
		}
	}

	return theModel
}

type BelongsToThingGormModel struct {
	// @gotags: fake:"skip"
	Id *string `gorm:"type:uuid;primaryKey;default:gen_random_uuid();" json:"id" fake:"skip"`
	// @gotags: fake:"skip"
	CreatedAt *time.Time `gorm:"default:now();" json:"createdAt" fake:"skip"`
	// @gotags: fake:"skip"
	UpdatedAt *time.Time `gorm:"default:now();" json:"updatedAt" fake:"skip"`
	// @gotags: fake:"{name}"
	Name string `json:"name" fake:"{name}"`
}

func (m *BelongsToThingGormModel) TableName() string {
	return "belongs_to_things"
}

func (m *BelongsToThingGormModel) ToProto() *BelongsToThing {
	if m == nil {
		return nil
	}
	theProto := &BelongsToThing{}
	theProto.Id = m.Id
	if m.CreatedAt != nil {
		theProto.CreatedAt = timestamppb.New(lo.FromPtr(m.CreatedAt))
	}
	if m.UpdatedAt != nil {
		theProto.UpdatedAt = timestamppb.New(lo.FromPtr(m.UpdatedAt))
	}
	theProto.Name = m.Name
	return theProto
}

func (m *BelongsToThing) ToGormModel() *BelongsToThingGormModel {
	if m == nil {
		return nil
	}
	theModel := &BelongsToThingGormModel{}
	theModel.Id = m.Id
	if m.CreatedAt != nil {
		theModel.CreatedAt = lo.ToPtr(m.CreatedAt.AsTime())
	}
	if m.UpdatedAt != nil {
		theModel.UpdatedAt = lo.ToPtr(m.UpdatedAt.AsTime())
	}
	theModel.Name = m.Name
	return theModel
}

type HasOneThingGormModel struct {
	// @gotags: fake:"skip"
	Id *string `gorm:"type:uuid;primaryKey;default:gen_random_uuid();" json:"id" fake:"skip"`
	// @gotags: fake:"skip"
	CreatedAt *time.Time `gorm:"default:now();" json:"createdAt" fake:"skip"`
	// @gotags: fake:"skip"
	UpdatedAt *time.Time `gorm:"default:now();" json:"updatedAt" fake:"skip"`
	// @gotags: fake:"{name}"
	Name string `json:"name" fake:"{name}"`
	// @gotags: fake:"skip"
	ThingId *string `json:"thingId" fake:"skip"`
}

func (m *HasOneThingGormModel) TableName() string {
	return "has_one_things"
}

func (m *HasOneThingGormModel) ToProto() *HasOneThing {
	if m == nil {
		return nil
	}
	theProto := &HasOneThing{}
	theProto.Id = m.Id
	if m.CreatedAt != nil {
		theProto.CreatedAt = timestamppb.New(lo.FromPtr(m.CreatedAt))
	}
	if m.UpdatedAt != nil {
		theProto.UpdatedAt = timestamppb.New(lo.FromPtr(m.UpdatedAt))
	}
	theProto.Name = m.Name
	theProto.ThingId = m.ThingId
	return theProto
}

func (m *HasOneThing) ToGormModel() *HasOneThingGormModel {
	if m == nil {
		return nil
	}
	theModel := &HasOneThingGormModel{}
	theModel.Id = m.Id
	if m.CreatedAt != nil {
		theModel.CreatedAt = lo.ToPtr(m.CreatedAt.AsTime())
	}
	if m.UpdatedAt != nil {
		theModel.UpdatedAt = lo.ToPtr(m.UpdatedAt.AsTime())
	}
	theModel.Name = m.Name
	theModel.ThingId = m.ThingId
	return theModel
}

type HasManyThingGormModel struct {
	// @gotags: fake:"skip"
	Id *string `gorm:"type:uuid;primaryKey;default:gen_random_uuid();" json:"id" fake:"skip"`
	// @gotags: fake:"skip"
	CreatedAt *time.Time `gorm:"default:now();" json:"createdAt" fake:"skip"`
	// @gotags: fake:"skip"
	UpdatedAt *time.Time `gorm:"default:now();" json:"updatedAt" fake:"skip"`
	// @gotags: fake:"{name}"
	Name string `json:"name" fake:"{name}"`
	// @gotags: fake:"skip"
	ThingId *string `json:"thingId" fake:"skip"`
}

func (m *HasManyThingGormModel) TableName() string {
	return "has_many_things"
}

func (m *HasManyThingGormModel) ToProto() *HasManyThing {
	if m == nil {
		return nil
	}
	theProto := &HasManyThing{}
	theProto.Id = m.Id
	if m.CreatedAt != nil {
		theProto.CreatedAt = timestamppb.New(lo.FromPtr(m.CreatedAt))
	}
	if m.UpdatedAt != nil {
		theProto.UpdatedAt = timestamppb.New(lo.FromPtr(m.UpdatedAt))
	}
	theProto.Name = m.Name
	theProto.ThingId = m.ThingId
	return theProto
}

func (m *HasManyThing) ToGormModel() *HasManyThingGormModel {
	if m == nil {
		return nil
	}
	theModel := &HasManyThingGormModel{}
	theModel.Id = m.Id
	if m.CreatedAt != nil {
		theModel.CreatedAt = lo.ToPtr(m.CreatedAt.AsTime())
	}
	if m.UpdatedAt != nil {
		theModel.UpdatedAt = lo.ToPtr(m.UpdatedAt.AsTime())
	}
	theModel.Name = m.Name
	theModel.ThingId = m.ThingId
	return theModel
}

type ManyToManyThingGormModel struct {
	// @gotags: fake:"skip"
	Id *string `gorm:"type:uuid;primaryKey;default:gen_random_uuid();" json:"id" fake:"skip"`
	// @gotags: fake:"skip"
	CreatedAt *time.Time `gorm:"default:now();" json:"createdAt" fake:"skip"`
	// @gotags: fake:"skip"
	UpdatedAt *time.Time `gorm:"default:now();" json:"updatedAt" fake:"skip"`
	// @gotags: fake:"{name}"
	Name string `json:"name" fake:"{name}"`
}

func (m *ManyToManyThingGormModel) TableName() string {
	return "many_to_many_things"
}

func (m *ManyToManyThingGormModel) ToProto() *ManyToManyThing {
	if m == nil {
		return nil
	}
	theProto := &ManyToManyThing{}
	theProto.Id = m.Id
	if m.CreatedAt != nil {
		theProto.CreatedAt = timestamppb.New(lo.FromPtr(m.CreatedAt))
	}
	if m.UpdatedAt != nil {
		theProto.UpdatedAt = timestamppb.New(lo.FromPtr(m.UpdatedAt))
	}
	theProto.Name = m.Name
	return theProto
}

func (m *ManyToManyThing) ToGormModel() *ManyToManyThingGormModel {
	if m == nil {
		return nil
	}
	theModel := &ManyToManyThingGormModel{}
	theModel.Id = m.Id
	if m.CreatedAt != nil {
		theModel.CreatedAt = lo.ToPtr(m.CreatedAt.AsTime())
	}
	if m.UpdatedAt != nil {
		theModel.UpdatedAt = lo.ToPtr(m.UpdatedAt.AsTime())
	}
	theModel.Name = m.Name
	return theModel
}
