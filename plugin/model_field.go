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
	IsRepeated  bool
	IsTimestamp bool
	IsStructPb  bool
	Comments    string
	Ignore      bool
	Name        string
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
	f.Name = f.Field.GoName
	f.IsMessage = isMessage(f.Field)
	f.IsRepeated = isRepeated(f.Field)
	f.IsTimestamp = isTimestamp(f.Field)
	f.IsStructPb = isStructPb(f.Field)
	f.Comments = f.Field.Comments.Leading.String() + f.Field.Comments.Trailing.String()
	f.ModelType = getModelFieldType(f)
	f.Tag = getFieldTags(f.Field)
	return
}

func isStructPb(field *protogen.Field) bool {
	return field.Message != nil && field.Message.Desc.FullName() == "google.protobuf.Struct"
}

func getModelFieldType(field *ModelField) string {
	if field.IsTimestamp {
		g.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "time"})
		g.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "github.com/samber/lo"})
		g.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "google.golang.org/protobuf/types/known/timestamppb"})
		return "*time.Time"
	} else if field.IsStructPb {
		g.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "google.golang.org/protobuf/types/known/structpb"})
		return "*map[string]interface{}"
	} else if field.IsMessage {
		return getMessageGormModelFieldType(field.Field)
	} else {
		return getPrimitiveGormModelFieldType(field.Field)
	}
}

func ignoreField(field *ModelField) bool {
	return field.Options != nil && field.Options.Ignore
}
