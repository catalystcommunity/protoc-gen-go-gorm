package weaviate

import (
	"context"
	"fmt"
	"github.com/catalystsquad/app-utils-go/errorutils"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/data"
	"github.com/weaviate/weaviate/entities/models"
	"strconv"
)

type WeaviateFileOptions struct {
}

func (s WeaviateFileOptions) ToProto() string {
	theProto := WeaviateFileOptions{}
}

func (s WeaviateFileOptions) WeaviateClassName() string {
	return "WeaviateFileOptions"
}

func (s WeaviateFileOptions) WeaviateClassSchema() models.Class {
	return models.Class{
		Class:      s.WeaviateClassName(),
		Properties: s.WeaviateClassSchemaProperties(),
	}
}

func (s WeaviateFileOptions) WeaviateClassSchemaProperties() []*models.Property {
	return []*models.Property{}
}

func (s WeaviateFileOptions) Data() map[string]interface{} {
	data := map[string]interface{}{}

	data = s.addCrossReferenceData(data)

	return data
}

func (s WeaviateFileOptions) addCrossReferenceData(data map[string]interface{}) map[string]interface{} {
	return data
}

func (s WeaviateFileOptions) Create(ctx context.Context, client *weaviate.Client, consistencyLevel string) (*data.ObjectWrapper, error) {
	return client.Data().Creator().
		WithClassName(s.WeaviateClassName()).
		WithProperties(s.Data()).
		WithID(s.Id).
		WithConsistencyLevel(consistencyLevel).
		Do(ctx)
}

func (s WeaviateFileOptions) Update(ctx context.Context, client *weaviate.Client, consistencyLevel string) error {
	return client.Data().Updater().
		WithClassName(s.WeaviateClassName()).
		WithID(s.Id).
		WithProperties(s.Data()).
		WithConsistencyLevel(consistencyLevel).
		Do(ctx)
}

func (s WeaviateFileOptions) Delete(ctx context.Context, client *weaviate.Client, consistencyLevel string) error {
	return client.Data().Deleter().
		WithClassName(s.WeaviateClassName()).
		WithID(s.Id).
		WithConsistencyLevel(consistencyLevel).
		Do(ctx)
}

func (s WeaviateFileOptions) EnsureClass(client *weaviate.Client) {
	ensureClass(client, s.WeaviateClassSchema())
}

type WeaviateMessageOptions struct {
	Ignore bool `json:"ignore"`
}

func (s WeaviateMessageOptions) ToProto() string {
	theProto := WeaviateMessageOptions{}
	theProto.Ignore = s.Ignore

}

func (s WeaviateMessageOptions) WeaviateClassName() string {
	return "WeaviateMessageOptions"
}

func (s WeaviateMessageOptions) WeaviateClassSchema() models.Class {
	return models.Class{
		Class:      s.WeaviateClassName(),
		Properties: s.WeaviateClassSchemaProperties(),
	}
}

func (s WeaviateMessageOptions) WeaviateClassSchemaProperties() []*models.Property {
	return []*models.Property{{
		Name:     "ignore",
		DataType: []string{"boolean"},
	},
	}
}

func (s WeaviateMessageOptions) Data() map[string]interface{} {
	data := map[string]interface{}{
		"ignore": s.Ignore,
	}

	data = s.addCrossReferenceData(data)

	return data
}

func (s WeaviateMessageOptions) addCrossReferenceData(data map[string]interface{}) map[string]interface{} {
	return data
}

func (s WeaviateMessageOptions) Create(ctx context.Context, client *weaviate.Client, consistencyLevel string) (*data.ObjectWrapper, error) {
	return client.Data().Creator().
		WithClassName(s.WeaviateClassName()).
		WithProperties(s.Data()).
		WithID(s.Id).
		WithConsistencyLevel(consistencyLevel).
		Do(ctx)
}

func (s WeaviateMessageOptions) Update(ctx context.Context, client *weaviate.Client, consistencyLevel string) error {
	return client.Data().Updater().
		WithClassName(s.WeaviateClassName()).
		WithID(s.Id).
		WithProperties(s.Data()).
		WithConsistencyLevel(consistencyLevel).
		Do(ctx)
}

func (s WeaviateMessageOptions) Delete(ctx context.Context, client *weaviate.Client, consistencyLevel string) error {
	return client.Data().Deleter().
		WithClassName(s.WeaviateClassName()).
		WithID(s.Id).
		WithConsistencyLevel(consistencyLevel).
		Do(ctx)
}

func (s WeaviateMessageOptions) EnsureClass(client *weaviate.Client) {
	ensureClass(client, s.WeaviateClassSchema())
}

type WeaviateFieldOptions struct {
	Ignore bool `json:"ignore"`
}

func (s WeaviateFieldOptions) ToProto() string {
	theProto := WeaviateFieldOptions{}
	theProto.Ignore = s.Ignore

}

func (s WeaviateFieldOptions) WeaviateClassName() string {
	return "WeaviateFieldOptions"
}

func (s WeaviateFieldOptions) WeaviateClassSchema() models.Class {
	return models.Class{
		Class:      s.WeaviateClassName(),
		Properties: s.WeaviateClassSchemaProperties(),
	}
}

func (s WeaviateFieldOptions) WeaviateClassSchemaProperties() []*models.Property {
	return []*models.Property{{
		Name:     "ignore",
		DataType: []string{"boolean"},
	},
	}
}

func (s WeaviateFieldOptions) Data() map[string]interface{} {
	data := map[string]interface{}{
		"ignore": s.Ignore,
	}

	data = s.addCrossReferenceData(data)

	return data
}

func (s WeaviateFieldOptions) addCrossReferenceData(data map[string]interface{}) map[string]interface{} {
	return data
}

func (s WeaviateFieldOptions) Create(ctx context.Context, client *weaviate.Client, consistencyLevel string) (*data.ObjectWrapper, error) {
	return client.Data().Creator().
		WithClassName(s.WeaviateClassName()).
		WithProperties(s.Data()).
		WithID(s.Id).
		WithConsistencyLevel(consistencyLevel).
		Do(ctx)
}

func (s WeaviateFieldOptions) Update(ctx context.Context, client *weaviate.Client, consistencyLevel string) error {
	return client.Data().Updater().
		WithClassName(s.WeaviateClassName()).
		WithID(s.Id).
		WithProperties(s.Data()).
		WithConsistencyLevel(consistencyLevel).
		Do(ctx)
}

func (s WeaviateFieldOptions) Delete(ctx context.Context, client *weaviate.Client, consistencyLevel string) error {
	return client.Data().Deleter().
		WithClassName(s.WeaviateClassName()).
		WithID(s.Id).
		WithConsistencyLevel(consistencyLevel).
		Do(ctx)
}

func (s WeaviateFieldOptions) EnsureClass(client *weaviate.Client) {
	ensureClass(client, s.WeaviateClassSchema())
}

func ensureClass(client *weaviate.Client, class models.Class) {
	var exists bool
	if exists = classExists(client, class.Class); exists {
		updateClass(client, class)
	} else {
		createClass(client, class)
	}
}

func updateClass(client *weaviate.Client, class models.Class) {
	var fetchedClass *models.Class
	if fetchedClass = getClass(client, class.Class); fetchedClass == nil {
		return
	}
	for _, property := range class.Properties {
		// continue on error, weaviate doesn't support updating property data types so we don't try to do that on startup
		// because it requires reindexing and is non trivial
		if containsProperty(fetchedClass.Properties, property) {
			continue
		}
		createProperty(client, class.Class, property)
	}
}

func createProperty(client *weaviate.Client, className string, property *models.Property) {
	err := client.Schema().PropertyCreator().WithClassName(className).WithProperty(property).Do(context.Background())
	errorutils.LogOnErr(logrus.WithFields(logrus.Fields{"class_name": className, "property_name": property.Name, "property_data_type": property.DataType}), "error creating property", err)
	return
}

func getClass(client *weaviate.Client, name string) (class *models.Class) {
	var err error
	class, err = client.Schema().ClassGetter().WithClassName(name).Do(context.Background())
	errorutils.LogOnErr(logrus.WithField("class_name", name), "error getting class", err)
	return
}

func createClass(client *weaviate.Client, class models.Class) {
	// all classes use contextionary
	class.Vectorizer = "text2vec-contextionary"
	err := client.Schema().ClassCreator().WithClass(&class).Do(context.Background())
	errorutils.LogOnErr(logrus.WithField("class_name", class.Class), "error creating class", err)
}

func classExists(client *weaviate.Client, name string) (exists bool) {
	var err error
	exists, err = client.Schema().ClassExistenceChecker().WithClassName(name).Do(context.Background())
	errorutils.LogOnErr(logrus.WithField("class_name", name), "error checking class existence", err)
	return
}

func containsProperty(source []*models.Property, property *models.Property) bool {
	// todo maybe: use a map/set to avoid repeated loops
	return lo.ContainsBy(source, func(item *models.Property) bool {
		return item.Name == property.Name
	})
}
