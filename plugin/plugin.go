package plugin

import (
	"bytes"
	"errors"
	"fmt"
	gorm "github.com/catalystsquad/protoc-gen-go-gorm/options"
	"github.com/stoewer/go-strcase"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
	"strings"
	"text/template"
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

const protoTimestampTypeGoName = "Timestamp"
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
	"tableName":             tableName,
}

var g *protogen.GeneratedFile

func (b *Builder) Generate() (response *pluginpb.CodeGeneratorResponse, err error) {
	for _, protoFile := range b.plugin.Files {
		// make sure all field types are supported
		if err = fileIsSupported(protoFile); err != nil {
			return
		}
		// template the proto file
		if err = b.handleFile(protoFile); err != nil {
			return
		}
	}
	// no errors, set and return the response
	response = b.plugin.Response()
	return
}

func (b *Builder) handleFile(file *protogen.File) (err error) {
	if fileHasOrmableMessages(file) {
		// create new generated file
		g = b.plugin.NewGeneratedFile(fileName(file), ".")
		outputPackage(file)
		for _, message := range file.Messages {
			if err = handleMessage(message); err != nil {
				return
			}
		}
	}
	return
}

func outputPackage(file *protogen.File) {
	g.P(fmt.Sprintf("package %s", file.GoPackageName))
}

func handleMessage(message *protogen.Message) (err error) {
	if messageIsOrmable(message) {
		var tpl *template.Template
		var buffer bytes.Buffer
		// create template and parse template file
		if tpl, err = template.New("gorm").Funcs(templateFuncs).Parse(GormTemplate); err != nil {
			return
		}
		// execute template
		data := map[string]interface{}{"message": message}
		if err = tpl.Execute(&buffer, data); err != nil {
			return
		}
		// write the templated buffer to the generated file
		if _, err = g.Write(buffer.Bytes()); err != nil {
			return
		}
	}

	return
}

func fileName(file *protogen.File) string {
	return file.GeneratedFilenamePrefix + ".pb.gorm.go"
}

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
	if isMessage(field) {
		return getMessageGormModelField(field)
	}
	return getPrimitiveGormModelField(field)
}

func getPrimitiveGormModelField(field *protogen.Field) string {
	return fmt.Sprintf("%s%s %s %s", fieldComments(field), getPrimitiveGormModelFieldName(field), getPrimitiveGormModelFieldType(field), getFieldTags(field))
}

func getMessageGormModelField(field *protogen.Field) (modelField string) {
	fieldName := getMessageGormModelFieldName(field)
	fieldType := getMessageGormModelFieldType(field)
	fieldTags := getFieldTags(field)
	options := getFieldOptions(field)
	if !isTimestampType(field) && options != nil {
		if options.GetBelongsTo() != nil {
			modelField = getGormModelFieldBelongsToField(field)
		}
	}
	modelField = fmt.Sprintf("%s%s%s %s %s", modelField, fieldComments(field), fieldName, fieldType, fieldTags)
	return
}

func getGormModelFieldBelongsToField(field *protogen.Field) (belongsToField string) {
	return fmt.Sprintf("%s%sId *string `` \n", fieldComments(field), fieldGoName(field))
}

func getGormModelFieldHasOneField(field *protogen.Field) (belongsToField string) {
	return fmt.Sprintf("%s%sId *string `` \n", fieldComments(field), fieldGoName(field))
}

func pointer(field *protogen.Field) string {
	if !isRepeated(field) && (isOptional(field) || isMessage(field)) {
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

func getPrimitiveGormModelFieldName(field *protogen.Field) string {
	return fieldGoName(field)
}

func getMessageGormModelFieldName(field *protogen.Field) string {
	return fieldGoName(field)
}

func getPrimitiveGormModelFieldType(field *protogen.Field) (fieldType string) {
	pointer := pointer(field)
	if isRepeated(field) {
		g.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "github.com/lib/pq"})
		fieldType = gormArrayTypeMap[fieldKind(field)]
	} else {
		fieldType = gormTypeMap[fieldKind(field)]
	}
	return fmt.Sprintf("%s%s", pointer, fieldType)
}

func getMessageGormModelFieldType(field *protogen.Field) (fieldType string) {
	pointer := pointer(field)
	goType := gormModelName(field.Message)
	if isTimestampType(field) {
		g.QualifiedGoIdent(protogen.GoIdent{
			GoName:       "",
			GoImportPath: "time",
		})
		g.QualifiedGoIdent(protogen.GoIdent{
			GoName:       "",
			GoImportPath: "google.golang.org/protobuf/types/known/timestamppb",
		})
		goType = gormModelTimestampType
	}
	if isRepeated(field) {
		fieldType = fmt.Sprintf("%s[]*%s", pointer, goType)
	} else {
		fieldType = fmt.Sprintf("*%s", goType)
	}
	return
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
		tag += "type:uuid;primaryKey;default:gen_random_uuid();"
	} else if isTimestampType(field) {
		tag += "default:now()"
	} else if isRepeated(field) && !isMessage(field) {
		tag += fmt.Sprintf("type:%s", gormTagTypeMap[fieldKind(field)])
	}
	options := getFieldOptions(field)
	if options != nil && (options.GetHasOne() != nil || options.GetHasMany() != nil) {
		tag += fmt.Sprintf("foreignKey:%sId", protoMessageName(field.Parent))
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
	if isMessage(field) {
		return getGormModelToProtoMessageField(field)
	}
	return getGormModelToProtoPrimitiveField(field)
	//fieldName := fieldGoName(field)
	//fieldType := fieldGoType(field)
	//if isTimestampType(field) {
	//	return fmt.Sprintf(`if m.%s != nil {
	//		theProto.%s = timestamppb.New(lo.FromPtr(m.%s))
	//	}`, fieldName, fieldName, fieldName)
	//} else if isPrimitiveType(field) {
	//	return fmt.Sprintf("theProto.%s = m.%s", fieldName, fieldName)
	//} else {
	//	// message type means we need to convert messages to protos using their toproto method
	//	if isRepeated(field) {
	//		// repeated means loop through and append
	//		return fmt.Sprintf(`
	//			theProto.%s = []%s{}
	//			for _, message := range m.%s {
	//				theProto.%s = append(theProto.%s, message.ToProto())
	//			}
	//		`, fieldName, fieldType, fieldName, fieldName, fieldName)
	//	} else {
	//		// not repeated, simply call toProto on the field
	//		return fmt.Sprintf("theProto.%s = m.%s.ToProto()", fieldName, fieldName)
	//	}
	//}
}

func getGormModelToProtoMessageField(field *protogen.Field) string {
	fieldName := fieldGoName(field)
	if isTimestampType(field) {
		g.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "github.com/samber/lo"})
		return fmt.Sprintf(`if m.%s != nil {
			theProto.%s = timestamppb.New(lo.FromPtr(m.%s))
		}`, fieldName, fieldName, fieldName)
	}
	if isRepeated(field) {
		return fmt.Sprintf(`
	if len(m.%s) > 0 {
		theProto.%s = []*%s{}
		for _, model := range m.%s {
			theProto.%s = append(theProto.%s, model.ToProto())
		}
	}
`, fieldName, fieldName, fieldGoType(field), fieldName, fieldName, fieldName)
	}
	return fmt.Sprintf("theProto.%s = m.%s.ToProto()", fieldName, fieldName)
}

func getGormModelToProtoPrimitiveField(field *protogen.Field) string {
	fieldName := fieldGoName(field)
	return fmt.Sprintf("theProto.%s = m.%s", fieldName, fieldName)
}

func protoToGormModelField(field *protogen.Field) string {
	fieldName := fieldGoName(field)
	if isTimestampType(field) {
		return fmt.Sprintf(`if m.%s != nil {
			theModel.%s = lo.ToPtr(m.%s.AsTime())
		}`, fieldName, fieldName, fieldName)
	} else if isPrimitiveType(field) {
		return fmt.Sprintf("theModel.%s = m.%s", fieldName, fieldName)
	} else {
		// message type means we need to convert messages to protos using their toGormModel method
		if isRepeated(field) {
			fieldType := gormModelName(field.Message)
			// repeated means loop through and append
			return fmt.Sprintf(`
				if len(m.%s) > 0 {
				theModel.%s = []*%s{}
				for _, message := range m.%s {
					theModel.%s = append(theModel.%s, message.ToGormModel())
				}
				}
				
			`, fieldName, fieldName, fieldType, fieldName, fieldName, fieldName)
		} else {
			// not repeated, simply call toGormModel on the field
			return fmt.Sprintf("theModel.%s = m.%s.ToGormModel()", fieldName, fieldName)
		}
	}
}

func isRepeated(field *protogen.Field) bool {
	return field.Desc.IsList()
}

// isPrimitiveType returns true if the field is a go primitive type. This is accomplished by getting the field primitive type
// and returning true if a primitive type was returned, or false if no type was returned
func isPrimitiveType(field *protogen.Field) bool {
	return fieldPrimitiveType(field) != ""
}

// isMessage returns true if the field kind is protoreflect.MessageKind
func isMessage(field *protogen.Field) bool {
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
	if isMessage(field) && field.Message != nil {
		return field.Message.GoIdent.String()
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

var supportedTypes = map[protoreflect.Kind]bool{
	protoreflect.BoolKind:    true,
	protoreflect.EnumKind:    true,
	protoreflect.Int32Kind:   true,
	protoreflect.Int64Kind:   true,
	protoreflect.FloatKind:   true,
	protoreflect.DoubleKind:  true,
	protoreflect.StringKind:  true,
	protoreflect.BytesKind:   true,
	protoreflect.MessageKind: true,
}

var gormTypeMap = map[protoreflect.Kind]string{
	protoreflect.BoolKind:   "bool",
	protoreflect.EnumKind:   "int",
	protoreflect.Int32Kind:  "int32",
	protoreflect.Int64Kind:  "int64",
	protoreflect.FloatKind:  "float32",
	protoreflect.DoubleKind: "float64",
	protoreflect.StringKind: "string",
	protoreflect.BytesKind:  "[]byte",
}

var gormArrayTypeMap = map[protoreflect.Kind]string{
	protoreflect.BoolKind:   "pq.BoolArray",
	protoreflect.EnumKind:   "pq.Int32Array",
	protoreflect.Int32Kind:  "pq.Int32Array",
	protoreflect.FloatKind:  "pq.Float32Array",
	protoreflect.Int64Kind:  "pq.Int64Array",
	protoreflect.DoubleKind: "pq.Float64Array",
	protoreflect.StringKind: "pq.StringArray",
	protoreflect.BytesKind:  "pq.ByteaArray",
}

var gormTagTypeMap = map[protoreflect.Kind]string{
	protoreflect.BoolKind:   "bool[]",
	protoreflect.EnumKind:   "int[]",
	protoreflect.Int32Kind:  "int[]",
	protoreflect.FloatKind:  "float[]",
	protoreflect.Int64Kind:  "int[]",
	protoreflect.DoubleKind: "float[]",
	protoreflect.StringKind: "string[]",
	protoreflect.BytesKind:  "bytes[]",
}

var goTypeMap = map[protoreflect.Kind]string{
	protoreflect.BoolKind:   "bool",
	protoreflect.EnumKind:   "int",
	protoreflect.Int32Kind:  "int32",
	protoreflect.Int64Kind:  "int64",
	protoreflect.FloatKind:  "float32",
	protoreflect.DoubleKind: "float64",
	protoreflect.StringKind: "string",
	protoreflect.BytesKind:  "[]byte",
}

func isTimestampType(field *protogen.Field) bool {
	return isMessage(field) && field.Message != nil && field.Message.GoIdent.GoName == protoTimestampTypeGoName
}

func fileIsSupported(file *protogen.File) (err error) {
	for _, message := range file.Messages {
		if messageIsOrmable(message) {
			for _, field := range message.Fields {
				if err = fieldTypeIsSupported(field); err != nil {
					return
				}
			}
		}
	}
	return
}

func fieldTypeIsSupported(field *protogen.Field) (err error) {
	fieldKind := fieldKind(field)
	if !supportedTypes[fieldKind] {
		err = errors.New(fmt.Sprintf("field %s is of unsupported type: %s", field.GoIdent.String(), fieldKind))
	}
	return
}

func fileHasOrmableMessages(file *protogen.File) bool {
	for _, message := range file.Messages {
		if messageIsOrmable(message) {
			return true
		}
	}
	return false
}

func messageIsOrmable(message *protogen.Message) bool {
	options := getMessageOptions(message)
	return options != nil && options.Ormable
}

func getMessageOptions(message *protogen.Message) *gorm.GormMessageOptions {
	options := message.Desc.Options()
	if options == nil {
		return nil
	}
	v := proto.GetExtension(options, gorm.E_Opts)
	if v == nil {
		return nil
	}

	opts, ok := v.(*gorm.GormMessageOptions)
	if !ok {
		return nil
	}

	return opts
}

func getFieldOptions(field *protogen.Field) *gorm.GormFieldOptions {
	options := field.Desc.Options().(*descriptorpb.FieldOptions)
	if options == nil {
		return &gorm.GormFieldOptions{}
	}

	v := proto.GetExtension(options, gorm.E_Field)
	if v == nil {
		return nil
	}

	opts, ok := v.(*gorm.GormFieldOptions)
	if !ok {
		return nil
	}

	return opts
}

func tableName(message *protogen.Message) string {
	options := getMessageOptions(message)
	if options != nil && options.Table != "" {
		return options.Table
	}
	return fmt.Sprintf(`"%ss"`, strcase.SnakeCase(message.GoIdent.GoName))
}
