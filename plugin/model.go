package plugin

import (
	"google.golang.org/protobuf/compiler/protogen"
)

type Model struct {
	*protogen.Message
	Name      string
	TableName string
	Fields    []*ModelField
}

func (m *Model) Parse() (err error) {
	m.Name = getModelNameFromMessage(m.Message)
	m.TableName = getTableNameFromMessage(m.Message)
	m.Fields = []*ModelField{}
	for _, field := range m.Message.Fields {
		modelField := &ModelField{Field: field}
		if err = modelField.Parse(); err != nil {
			return
		}
		if modelField.Ignore {
			continue
		}
		m.Fields = append(m.Fields, modelField)
	}
	return
}
