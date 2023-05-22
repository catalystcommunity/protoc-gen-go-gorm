package plugin

const GormTemplate = `
type {{ gormModelName .message }} struct {
	{{- range .message.Fields }}
    {{ gormModelField . -}}
	{{ end }}
}

func (m *{{ gormModelName .message }}) TableName() string {
	return {{ tableName .message }}
}

func (m *{{ gormModelName .message }}) ToProto() *{{ protoMessageName .message }} {
	if m == nil {
		return nil
	}
	theProto := &{{ protoMessageName .message }}{}
	{{- range .message.Fields }}
    {{ gormModelToProtoField . -}}
	{{ end }}
	return theProto
}

func (m *{{ protoMessageName .message }}) ToGormModel() *{{ gormModelName .message }} {
	if m == nil {
		return nil
	}
	theModel := &{{ gormModelName .message }}{}
	{{- range .message.Fields }}
	{{ protoToGormModelField . -}}
	{{ end }}
	return theModel
}
`
