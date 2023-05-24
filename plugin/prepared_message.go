package plugin

import (
	"fmt"
	"github.com/stoewer/go-strcase"
	"google.golang.org/protobuf/compiler/protogen"
)

type PreparedMessage struct {
	*protogen.Message
	*Model
	PluginOptions
	Ignore bool
}

func (pm *PreparedMessage) Parse() (err error) {
	// parse options first
	pm.Options = getMessageOptions(pm.Message)
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
	return fmt.Sprintf(`"%ss"`, strcase.SnakeCase(message.GoIdent.GoName))
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
	return pm.Options != nil && !pm.Options.Ormable
}
