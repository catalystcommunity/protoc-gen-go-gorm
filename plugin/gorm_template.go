package plugin

const GormTemplate = `
package {{ .package }}

import (
	"github.com/lib/pq"
	"github.com/samber/lo"
	"time"
)

{{ range .messages }}
type {{ gormModelName . }} struct {
	{{- range .Fields }}
	{{ fieldComments . -}}
    {{ gormModelField . -}}
	{{ end }}
}

func (m {{ gormModelName . }}) ToProto() *{{ protoMessageName . }} {
	theProto := &{{ protoMessageName . }}{}
	{{- range .Fields }}
    {{ gormModelToProtoField . -}}
	{{ end }}
	return theProto
}

func (m *{{ protoMessageName . }}) ToGormModel() {{ gormModelName . }} {
	theModel := {{ gormModelName . }}{}
	{{- range .Fields }}
	{{ protoToGormModelField . -}}
	{{ end }}
	return theModel
}

{{- end }}
`
