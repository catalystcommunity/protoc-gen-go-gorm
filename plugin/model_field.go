package plugin

import (
	gorm "github.com/catalystsquad/protoc-gen-go-gorm/options"
	"google.golang.org/protobuf/compiler/protogen"
)

type ModelField struct {
	*protogen.Field
	ModelType   string
	Tag         string
	Options     *gorm.GormFieldOptions
	IsMessage   bool
	IsTimestamp bool
	Comments    string
	Ignore      bool
}

func (f *ModelField) Parse() (err error) {
	// parse options first
	f.Options = getFieldOptions(f.Field)
	// set ignore
	f.Ignore = ignoreField(f)
	// if ignore then stop parsing and return, field should be ignored
	if f.Ignore {
		return
	}
	f.IsMessage = isMessage(f.Field)
	f.IsTimestamp = isTimestampType(f.Field)
	f.Comments = f.Field.Comments.Leading.String() + f.Field.Comments.Trailing.String()
	f.ModelType = getModelFieldType(f)
	return
}

func getModelFieldType(field *ModelField) string {
	if field.IsTimestamp {
		return "*time.Time"
	} else if field.IsMessage {
		return getMessageGormModelFieldType(field.Field)
	} else {
		return getPrimitiveGormModelFieldType(field.Field)
	}
}

func ignoreField(field *ModelField) bool {
	return field.Options != nil && field.Options.Ignore
}
