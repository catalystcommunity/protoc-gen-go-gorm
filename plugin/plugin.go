package plugin

import (
	"bytes"
	"google.golang.org/protobuf/reflect/protoreflect"
	"strings"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

type Builder struct {
	plugin         *protogen.Plugin
	messages       map[string]struct{}
	currentFile    string
	currentPackage string
	dbEngine       int
	stringEnums    bool
	suppressWarn   bool
}

// I can't find where the constant is for this in protogen, so I'm putting it here
const SUPPORTS_OPTIONAL_FIELDS = 1

var templateFuncs = map[string]any{}

func New(opts protogen.Options, request *pluginpb.CodeGeneratorRequest) (*Builder, error) {
	plugin, err := opts.New(request)
	if err != nil {
		return nil, err
	}
	plugin.SupportedFeatures = SUPPORTS_OPTIONAL_FIELDS
	builder := &Builder{
		plugin:   plugin,
		messages: make(map[string]struct{}),
	}

	params := parseParameter(request.GetParameter())

	if strings.EqualFold(params["enums"], "string") {
		builder.stringEnums = true
	}

	if _, ok := params["quiet"]; ok {
		builder.suppressWarn = true
	}

	return builder, nil
}

func parseParameter(param string) map[string]string {
	paramMap := make(map[string]string)

	params := strings.Split(param, ",")
	for _, param := range params {
		if strings.Contains(param, "=") {
			kv := strings.Split(param, "=")
			paramMap[kv[0]] = kv[1]
			continue
		}
		paramMap[param] = ""
	}

	return paramMap
}

func (b *Builder) Generate() (response *pluginpb.CodeGeneratorResponse, err error) {
	for _, protoFile := range b.plugin.Files {
		var tpl *template.Template
		templateFuncs["package"] = func() string { return string(protoFile.GoPackageName) }
		if tpl, err = template.New("gorm").Funcs(templateFuncs).Parse(GormTemplate); err != nil {
			return
		}
		fileName := protoFile.GeneratedFilenamePrefix + ".pb.gorm.go"
		g := b.plugin.NewGeneratedFile(fileName, ".")
		var data bytes.Buffer
		templateMap := map[string]any{
			"messages": protoFile.Messages,
		}
		if err = tpl.Execute(&data, templateMap); err != nil {
			return
		}
		if _, err = g.Write(data.Bytes()); err != nil {
			return
		}
	}
	response = b.plugin.Response()
	return
}

var gormTypeMap = map[protoreflect.Kind]string{
	protoreflect.StringKind:   "text",
	protoreflect.BoolKind:     "boolean",
	protoreflect.EnumKind:     "int",
	protoreflect.Int32Kind:    "int",
	protoreflect.Sint32Kind:   "int",
	protoreflect.Uint32Kind:   "int",
	protoreflect.Int64Kind:    "string",
	protoreflect.Sint64Kind:   "string",
	protoreflect.Uint64Kind:   "string",
	protoreflect.Sfixed32Kind: "int",
	protoreflect.Fixed32Kind:  "int",
	protoreflect.Sfixed64Kind: "string",
	protoreflect.Fixed64Kind:  "string",
	protoreflect.FloatKind:    "number",
	protoreflect.DoubleKind:   "number",
	protoreflect.BytesKind:    "blob",
}

var goTypeMap = map[protoreflect.Kind]string{
	protoreflect.BoolKind:     "bool",
	protoreflect.EnumKind:     "int",
	protoreflect.Int32Kind:    "int32",
	protoreflect.Sint32Kind:   "int32",
	protoreflect.Uint32Kind:   "uint32",
	protoreflect.Int64Kind:    "int64",
	protoreflect.Sint64Kind:   "int64",
	protoreflect.Uint64Kind:   "uint64",
	protoreflect.Sfixed32Kind: "int32",
	protoreflect.Fixed32Kind:  "uint32",
	protoreflect.FloatKind:    "float32",
	protoreflect.Sfixed64Kind: "int64",
	protoreflect.Fixed64Kind:  "uint64",
	protoreflect.DoubleKind:   "float64",
	protoreflect.StringKind:   "string",
	protoreflect.BytesKind:    "[]byte",
}
