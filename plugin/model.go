package plugin

import (
	"fmt"
	"google.golang.org/protobuf/compiler/protogen"
	"strings"
)

type Model struct {
	*protogen.Message
	Name                     string
	TableName                string
	Fields                   []*ModelField
	HasBelongsToRelationship bool
	OmitFields               []*ModelField
	Omits                    string
}

func (m *Model) Parse() (err error) {
	m.Name = getModelNameFromMessage(m.Message)
	m.TableName = getTableNameFromMessage(m.Message)
	m.Fields = []*ModelField{}
	m.OmitFields = []*ModelField{}
	for _, field := range m.Message.Fields {
		modelField := &ModelField{Field: field}
		if err = modelField.Parse(); err != nil {
			return
		}
		if modelField.Ignore {
			continue
		}
		if modelField.HasBelongsToRelationship {
			m.HasBelongsToRelationship = true
			m.OmitFields = append(m.OmitFields, modelField)
		}
		m.Fields = append(m.Fields, modelField)
	}
	m.Omits = getOmitString(m)
	return
}

func getOmitString(m *Model) string {
	builder := strings.Builder{}
	builder.WriteString("clause.Associations")
	for _, field := range m.OmitFields {
		builder.WriteString(",")
		if field.Options.GetBelongsTo() != nil && field.Options.GetBelongsTo().Foreignkey != "" {
			builder.WriteString(fmt.Sprintf("\"%s\"", field.Options.GetBelongsTo().Foreignkey))
		} else {
			builder.WriteString(fmt.Sprintf("\"%sId\"", field.GoName))
		}
	}
	return builder.String()
}
