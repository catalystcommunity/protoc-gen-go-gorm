package plugin

import (
	gorm "github.com/catalystsquad/protoc-gen-go-gorm/options"
	"google.golang.org/protobuf/compiler/protogen"
	"strings"
)

type ModelField struct {
	*protogen.Field
	ModelType                      string
	ModelSingularType              string
	Tag                            string
	Options                        *gorm.GormFieldOptions
	IsMessage                      bool
	IsRepeated                     bool
	IsTimestamp                    bool
	IsStructPb                     bool
	Comments                       string
	Ignore                         bool
	Name                           string
	ShouldGenerateBelongsToIdField bool
	HasReplaceRelationships        bool // any relationships except belongs to needs replace calls
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
	f.ModelSingularType = getModelFieldSingularType(f)
	f.Tag = getFieldTags(f)
	f.ShouldGenerateBelongsToIdField = shouldGenerateBelongsToIdField(f)
	f.HasReplaceRelationships = hasReplaceRelationships(f)
	return
}

func shouldGenerateBelongsToIdField(f *ModelField) bool {
	options := f.Options
	// no options or no belongs to means don't generate a belongs to field
	if options == nil || options.GetBelongsTo() == nil {
		return false
	}
	// there are belongs to options, loop through the fields and make sure there isn't one already
	fieldName := options.GetBelongsTo().Foreignkey
	for _, field := range f.Parent.Fields {
		if field.GoName == fieldName {
			// field is already on the message, don't generate a duplicate
			return false
		}
	}
	// belongs to id field doesn't exist, so generate it
	return true
}

func getModelFieldSingularType(field *ModelField) (fieldType string) {
	fieldType = getModelFieldType(field)
	if field.IsRepeated {
		// if it's repeated, remove the [] from the type string
		fieldType = strings.Replace(fieldType, "[]", "", 1)
	}
	return
}

func isStructPb(field *protogen.Field) bool {
	return field.Message != nil && field.Message.Desc.FullName() == "google.protobuf.Struct"
}

func getModelFieldType(field *ModelField) string {
	if field.IsTimestamp {
		g.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "time"})
		if field.IsMessage {
			g.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "github.com/samber/lo"})
			g.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "google.golang.org/protobuf/types/known/timestamppb"})
		}
		return "*time.Time"
	} else if field.IsStructPb {
		g.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "github.com/jackc/pgtype"})
		g.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "encoding/json"})
		return "*pgtype.JSONB"
	} else if field.IsMessage {
		return getMessageGormModelFieldType(field.Field)
	} else if field.Enum != nil {
		if field.IsRepeated {
			g.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "github.com/lib/pq"})
			if field.Options.EnumAsString {
				return "pq.StringArray"
			} else {
				return "pq.Int32Array"
			}
		}
		if field.Options.EnumAsString {
			return "string"
		}
		return "int"
	} else {
		return getPrimitiveGormModelFieldType(field.Field)
	}
}

func ignoreField(field *ModelField) bool {
	return field.Options != nil && field.Options.Ignore
}

func hasReplaceRelationships(field *ModelField) bool {
	return field.Options != nil && (field.Options.GetHasOne() != nil || field.Options.GetHasMany() != nil || field.Options.GetManyToMany() != nil)
}
