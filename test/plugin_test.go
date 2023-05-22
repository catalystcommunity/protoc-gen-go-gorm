package test

import (
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	. "github.com/catalystsquad/protoc-gen-go-gorm/example/demo"
	"github.com/google/go-cmp/cmp"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/cockroachdb"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/testing/protocmp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"testing"
)

var container *gnomock.Container
var db *gorm.DB

type PluginSuite struct {
	suite.Suite
}

func TestPluginSuite(t *testing.T) {
	suite.Run(t, new(PluginSuite))
}

func (s *PluginSuite) TestPlugin() {
	var err error
	thingProto := &Thing{}
	belongsToThingProto := &BelongsToThing{}
	hasOneThingProto := &HasOneThing{}
	err = gofakeit.Struct(&thingProto)
	require.NoError(s.T(), err)
	err = gofakeit.Struct(&belongsToThingProto)
	require.NoError(s.T(), err)
	err = gofakeit.Struct(&hasOneThingProto)
	require.NoError(s.T(), err)
	thingProto.BelongsTo = belongsToThingProto
	thingProto.HasOne = hasOneThingProto
	thingModel := thingProto.ToGormModel()
	require.NoError(s.T(), err)
	err = db.Create(&thingModel).Error
	require.NoError(s.T(), err)
	var firstThing *ThingGormModel
	var firstThingProto *Thing
	err = db.Joins("BelongsTo").Joins("HasOne").First(&firstThing).Error
	require.NoError(s.T(), err)
	require.Empty(s.T(), cmp.Diff(thingModel, firstThing))
	firstThingProto = firstThing.ToProto()
	require.NoError(s.T(), err)
	require.Empty(s.T(), cmp.Diff(thingProto, firstThingProto, protocmp.Transform(), protocmp.IgnoreFields(&Thing{}, "created_at", "id", "updated_at"), protocmp.IgnoreFields(&BelongsToThing{}, "created_at", "id", "updated_at"), protocmp.IgnoreFields(&HasOneThing{}, "created_at", "id", "updated_at", "thing_id")))
}

func (s *PluginSuite) SetupSuite() {
	preset := cockroachdb.Preset()
	var err error
	portOpt := gnomock.WithCustomNamedPorts(gnomock.NamedPorts{"default": gnomock.Port{
		Protocol: "tcp",
		Port:     26257,
		HostPort: 26257,
	}})
	container, err = gnomock.Start(preset, portOpt)
	require.NoError(s.T(), err)
	dsn := fmt.Sprintf("host=%s port=%d user=root dbname=%s sslmode=disable", container.Host, container.DefaultPort(), "postgres")
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(s.T(), err)
	err = db.AutoMigrate(&ThingGormModel{}, &HasOneThingGormModel{})
	require.NoError(s.T(), err)
}

func (s *PluginSuite) TearDownSuite() {
	require.NoError(s.T(), gnomock.Stop(container))
}

func (s *PluginSuite) SetupTest() {
}

func convert(source, dest interface{}) (err error) {
	var sourceBytes []byte
	if sourceBytes, err = json.Marshal(source); err != nil {
		return
	}
	return json.Unmarshal(sourceBytes, dest)

}
