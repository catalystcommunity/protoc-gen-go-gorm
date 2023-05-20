package plugin

import (
	"bytes"
	"fmt"
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

const protoTimestampType = "timestamppb.Timestamp"
const gormModelTimestampType = "time.Time"

// I can't find where the constant is for this in protogen, so I'm putting it here
const SUPPORTS_OPTIONAL_FIELDS = 1

var templateFuncs = map[string]any{
	"protoMessageName":      protoMessageName,
	"fieldComments":         fieldComments,
	"gormModelField":        gormModelField,
	"gormModelToProtoField": gormModelToProtoField,
	"protoToGormModelField": protoToGormModelField,
	"fieldGoType":           fieldGoType,
	"fieldGoIdent":          fieldGoIdent,
	"gormModelName":         gormModelName,
}

var g *protogen.GeneratedFile

func protoMessageName(message *protogen.Message) protoreflect.Name {
	return message.Desc.Name()
}

func gormModelName(message *protogen.Message) string {
	return fmt.Sprintf("%sGormModel", protoMessageName(message))
}

func fieldComments(field *protogen.Field) string {
	return field.Comments.Leading.String() + field.Comments.Trailing.String()
}

func gormModelField(field *protogen.Field) string {
	return fmt.Sprintf("%s %s %s", fieldGoName(field), getGormModelFieldType(field), getFieldTags(field))
}

func pointer(field *protogen.Field) string {
	if isOptional(field) || isMessageType(field) {
		return "*"
	}
	return ""
}

func slice(field *protogen.Field) (slice string) {
	if isRepeated(field) {
		slice = "[]"
	}
	return
}

func fieldGoName(field *protogen.Field) string {
	return field.GoName
}

func getGormModelFieldType(field *protogen.Field) string {
	fieldType := ""
	slice := slice(field)
	pointer := pointer(field)
	if isTimestampType(field) {
		fieldType = gormModelTimestampType
	} else {
		fieldType = gormTypeMap[fieldKind(field)]
	}
	return fmt.Sprintf("%s%s%s", slice, pointer, fieldType)
}

func fieldKind(field *protogen.Field) protoreflect.Kind {
	return field.Desc.Kind()
}

func getFieldTags(field *protogen.Field) string {
	return fmt.Sprintf("`%s %s`", getGormFieldTag(field), getJsonFieldTag(field))
}

func getGormFieldTag(field *protogen.Field) string {
	tag := "gorm:\""
	if isIdField(field) {
		tag += "primaryKey;default:gen_random_uuid()"
	} else if isTimestampType(field) {
		tag += "default:now()"
	}
	return tag + "\""
}

func isIdField(field *protogen.Field) bool {
	return strings.ToLower(string(field.Desc.Name())) == "id"
}

func getJsonFieldTag(field *protogen.Field) string {
	return fmt.Sprintf(`json:"%s"`, field.Desc.JSONName())
}

func gormModelToProtoField(field *protogen.Field) string {
	fieldName := fieldGoName(field)
	fieldType := fieldGoType(field)
	if isTimestampType(field) {
		return fmt.Sprintf(`if m.%s != nil {
			theProto.%s = timestamppb.New(lo.FromPtr(m.%s))
		}`, fieldName, fieldName, fieldName)
	} else if isPrimitiveType(field) {
		return fmt.Sprintf("theProto.%s = m.%s", fieldName, fieldName)
	} else {
		// message type means we need to convert messages to protos using their toproto method
		if isRepeated(field) {
			// repeated means loop through and append
			return fmt.Sprintf(`
				theProto.%s = []%s{}
				for _, message := range m.%s {
					theProto.%s = append(theProto.%s, message.ToProto())
				}
			`, fieldName, fieldType, fieldName, fieldName, fieldName)
		} else {
			// not repeated, simply call toProto on the field
			return fmt.Sprintf("theProto.%s = m.%s.ToProto()", fieldName, fieldName)
		}
	}
}

func protoToGormModelField(field *protogen.Field) string {
	return ""
}

func isRepeated(field *protogen.Field) bool {
	return field.Desc.IsList()
}

// isPrimitiveType returns true if the field is a go primitive type. This is accomplished by getting the field primitive type
// and returning true if a primitive type was returned, or false if no type was returned
func isPrimitiveType(field *protogen.Field) bool {
	return fieldPrimitiveType(field) != ""
}

// isMessageType returns true if the field kind is protoreflect.MessageKind
func isMessageType(field *protogen.Field) bool {
	return fieldKind(field) == protoreflect.MessageKind
}

func isOptional(field *protogen.Field) bool {
	return field.Desc.HasOptionalKeyword()
}

// fieldPrimitiveType gets the field's primitive type from the go type map, returning an empty string if the field's
// type is not primitive
func fieldPrimitiveType(field *protogen.Field) string {
	return goTypeMap[fieldKind(field)]
}

// fieldGoType returns the go type of the field. It checks first for a primitive type, if no primitive type is returned
// then the message's name is returned as the type
func fieldGoType(field *protogen.Field) (typ string) {
	if typ = fieldPrimitiveType(field); typ != "" {
		return
	}
	return string(field.Message.Desc.Name())
}

func fieldGoIdent(field *protogen.Field) string {
	if isMessageType(field) {
		return getImport(field)
	}
	return ""
}

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
		g = b.plugin.NewGeneratedFile(fileName, ".")
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

func getImport(field *protogen.Field) string {
	return g.QualifiedGoIdent(field.Message.GoIdent)
}

func isTimestampType(field *protogen.Field) bool {
	return isMessageType(field) && getImport(field) == protoTimestampType
}
