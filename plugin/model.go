package plugin

import (
	gorm "github.com/catalystsquad/protoc-gen-go-gorm/options"
	"google.golang.org/protobuf/compiler/protogen"
)

type Model struct {
	*protogen.Message
	Name      string
	TableName string
	Fields    []*ModelField
	Options   *gorm.GormMessageOptions
}

func (m *Model) Parse() (err error) {
	// parse options first
	m.Options = getMessageOptions(m.Message)
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
