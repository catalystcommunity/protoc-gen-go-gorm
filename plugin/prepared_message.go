package plugin

import (
	"fmt"
	gorm "github.com/catalystsquad/protoc-gen-go-gorm/options"
	"github.com/stoewer/go-strcase"
	"google.golang.org/protobuf/compiler/protogen"
)

type PreparedMessage struct {
	*protogen.Message
	*Model
	PluginOptions
	Options *gorm.GormMessageOptions
	Ignore  bool
}

func (pm *PreparedMessage) Parse() (err error) {
	// parse options first
	options := getMessageOptions(pm.Message)
	pm.Options = options
	// set ignore
	pm.Ignore = ignoreMessage(pm)
	// if ignore then stop parsing and return, field should be ignored
	if pm.Ignore {
		return
	}
	model := &Model{Message: pm.Message}
	if err = model.Parse(); err != nil {
		return
	}
	pm.Model = model
	return
}

func getModelNameFromMessage(message *protogen.Message) string {
	return fmt.Sprintf("%sGormModel", message.GoIdent.GoName)
}

func getTableNameFromMessage(message *protogen.Message) string {
	options := getMessageOptions(message)
	if options != nil && options.Table != "" {
		return options.Table
	}
	return pluralizer.Plural(strcase.SnakeCase(message.GoIdent.GoName))
}

func prepareMessages(messages []*protogen.Message, opts PluginOptions) (preparedMessages []*PreparedMessage, err error) {
	preparedMessages = []*PreparedMessage{}
	for _, message := range messages {
		preparedMessage := &PreparedMessage{Message: message}
		if err = preparedMessage.Parse(); err != nil {
			return
		}
		if preparedMessage.Ignore {
			continue
		}
		preparedMessages = append(preparedMessages, preparedMessage)
	}
	return
}

func ignoreMessage(pm *PreparedMessage) bool {
	return pm.Options == nil || !pm.Options.Ormable
}
