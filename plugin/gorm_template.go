package plugin

import "text/template"

var messageTemplate = template.Must(template.New("message").Funcs(templateFuncs).Parse(`
type {{ .Model.Name }} struct {
	{{- range .Model.Fields }}
    {{ if .Options.GetBelongsTo }}
    {{ .GoName }}Id *string {{ emptyTag }}
    {{ end }}
    {{ .Comments -}}
    {{ .GoName }} {{ .ModelType }} {{ .Tag -}}
	{{ end }}
}

func (m *{{ .Model.Name }}) TableName() string {
	return "{{ .Model.TableName }}"
}

func (m *{{ .Model.Name }}) ToProto() (theProto *{{.GoIdent.GoName}}, err error) {
	if m == nil {
		return
	}
	theProto = &{{.GoIdent.GoName}}{}
	{{ range .Model.Fields }}
    {{ if .IsTimestamp }}
    if m.{{ .GoName }} != nil {
		theProto.{{ .GoName }} = timestamppb.New(*m.{{ .GoName }})
	}
    {{ else if .IsStructPb }}
	if m.{{ .GoName }} != nil {
		if theProto.{{ .GoName }}, err = structpb.NewStruct(*m.{{ .GoName }}); err != nil {
			return
		}
	}
    {{ else if and .IsMessage (eq .IsRepeated false) }}
	if theProto.{{ .GoName }}, err = m.{{ .GoName }}.ToProto(); err != nil {
		return
	}
    {{ else if and .IsMessage .IsRepeated }}
	if len(m.{{ .GoName }}) > 0 {
		theProto.{{ .GoName }} = []*{{ .Message.Desc.Name }}{}
        for _, item := range m.{{ .GoName }} {
			if {{ .GoName }}Proto, err := item.ToProto(); err != nil {
				return 
			} else {
				theProto.{{ .GoName }} = append(theProto.{{ .GoName }}, {{ .GoName }}Proto)
			}	
		}
	}
    {{ else }}
    theProto.{{ .GoName }} = m.{{ .GoName }}
    {{ end }}
	{{ end }}
	return
}

func (p *{{.GoIdent.GoName}}) ToModel() (theModel *{{ .Model.Name }}, err error) {
	if p == nil {
		return
	}
	theModel = &{{ .Model.Name }}{}
	{{ range .Model.Fields }}
    {{ if .IsTimestamp }}
    if p.{{ .GoName }} != nil {
		theModel.{{ .GoName }} = lo.ToPtr(p.{{ .GoName }}.AsTime())
	}
    {{ else if .IsStructPb }}
	if p.{{ .GoName }} != nil {
		theModel.{{ .GoName }} = lo.ToPtr(p.{{ .GoName }}.AsMap())
	}
	{{ else if and .IsMessage (eq .IsRepeated false) }}
	if theModel.{{ .GoName }}, err = p.{{ .GoName }}.ToModel(); err != nil {
		return
	}
	{{ else if and .IsMessage .IsRepeated }}
	if len(p.{{ .GoName }}) > 0 {
		theModel.{{ .GoName }} = {{ .ModelType }}{}
        for _, item := range p.{{ .GoName }} {
			if {{ .GoName }}Model, err := item.ToModel(); err != nil {
				return 
			} else {
				theModel.{{ .GoName }} = append(theModel.{{ .GoName }}, {{ .GoName }}Model)
			}	
		}
	}
    {{ else }}
    theModel.{{ .GoName }} = p.{{ .GoName }}
    {{ end }}
	{{ end }}
	return
}
`))
