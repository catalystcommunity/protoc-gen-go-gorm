package plugin

import "text/template"

var messageTemplate = template.Must(template.New("message").Parse(`
type {{ .Model.Name }} struct {
	{{- range Model.Fields }}
    {{ .Comments -}}
    {{ .GoIdent.GoName }} {{ .ModelType }} {{ .Tag -}}
	{{ end }}
}

func (m *{{ .Model.Name }}) TableName() string {
	return "{{ .Model.TableName }}"
}

func (m *{{ .Model.Name }}) ToProto() *{{.GoIdent.GoName}} {
	if m == nil {
		return nil
	}
	theProto := &{{.GoIdent.GoName}}{}
	{{- range Model.Fields }}
    theProto.{{ .GoIdent.GoName }} = m.{{ .GoIdent.GoName }}
	{{ end }}
	return theProto
}

func (m *{{.GoIdent.GoName}}) ToGormModel() *{{ .Name }} {
	if m == nil {
		return nil
	}
	theModel := &{{ .Name }}{}
	{{- range Model.Fields }}
    theModel.{{ .GoIdent.GoName }} = m.{{ .GoIdent.GoName }}
	{{ end }}
	return theModel
}
`))
