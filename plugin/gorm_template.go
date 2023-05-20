package plugin

const GormTemplate = `
package {{ package }}

import (
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/timestamppb"
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
	model := {{ gormModelName . }}{}
	{{- range .Fields }}
	{{ protoToGormModelField . -}}
	{{ end }}
	return model
}

{{- end }}
`
