package plugin

const GormTemplate = `
package {{ package }}

import (
)

{{ range .messages }}
type {{ structName . }} struct {
	{{- range .Fields }}
	{{ fieldComments . -}}
    {{ structField . }} {{ jsonTag . -}}
	{{ end }}
}
`
