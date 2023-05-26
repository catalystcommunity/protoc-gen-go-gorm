package test

import (
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	. "github.com/catalystsquad/protoc-gen-go-gorm/example/demo"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/cockroachdb"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/testing/protocmp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	thing, err := s.getPopulatedThing()
	require.NoError(s.T(), err)
	thingModel, err := thing.ToModel()
	require.NoError(s.T(), err)
	err = db.Create(&thingModel).Error
	require.NoError(s.T(), err)
	var firstThingModel *ThingGormModel
	var firstThingProto *Thing
	err = db.Joins("BelongsTo").Joins("HasOne").Preload(clause.Associations).First(&firstThingModel).Error
	require.NoError(s.T(), err)
	assertModelsEquality(s.T(), thingModel, firstThingModel)
	firstThingProto, err = firstThingModel.ToProto()
	require.NoError(s.T(), err)
	assertProtosEquality(s.T(), thing, firstThingProto)
}

func (s *PluginSuite) TestSliceTransformers() {
	thing, err := s.getPopulatedThing()
	require.NoError(s.T(), err)
	things := ThingProtos{thing}
	models, err := things.ToModels()
	require.NoError(s.T(), err)
	transformedThings, err := models.ToProtos()
	require.NoError(s.T(), err)
	assertProtosEquality(s.T(), things, transformedThings)
}

func assertModelsEquality(t *testing.T, expected, actual interface{}) {
	// no need to ignore id, created_at, updated_at because gorm populates them on create
	require.Empty(t, cmp.Diff(
		expected,
		actual,
		cmpopts.SortSlices(func(x, y *HasManyThingGormModel) bool {
			return x.Name < y.Name
		}),
		cmpopts.SortSlices(func(x, y *ManyToManyThingGormModel) bool {
			return x.Name < y.Name
		}),
		cmpopts.IgnoreFields(ThingGormModel{}, "AStructpb"),
	))
}

func assertProtosEquality(t *testing.T, expected, actual interface{}, ignoreFields ...string) {
	// ignoring id, created_at, updated_at, thing_id because the original proto doesn't have those, but the
	// one converted from the created model does
	require.Empty(t,
		cmp.Diff(
			expected,
			actual,
			protocmp.Transform(),
			protocmp.IgnoreFields(&Thing{}, "created_at", "id", "updated_at", "belongs_to_two_id", "an_unexpected_id"),
			protocmp.IgnoreFields(&BelongsToThing{}, "created_at", "id", "updated_at"),
			protocmp.IgnoreFields(&HasOneThing{}, "created_at", "id", "updated_at", "thing_id"),
			protocmp.IgnoreFields(&HasManyThing{}, "created_at", "id", "updated_at", "thing_id"),
			protocmp.IgnoreFields(&ManyToManyThing{}, "created_at", "id", "updated_at"),
			protocmp.SortRepeated(func(x, y *HasManyThing) bool {
				return x.Name < y.Name
			}),
			protocmp.SortRepeated(func(x, y *ManyToManyThing) bool {
				return x.Name < y.Name
			}),
		),
	)
}

func (s *PluginSuite) getPopulatedThing() (thing *Thing, err error) {
	thing = &Thing{}
	belongsToThing, belongsToThingTwo, belongsToThingThree := &BelongsToThing{}, &BelongsToThing{}, &BelongsToThing{}
	hasOneThing := &HasOneThing{}
	hasManyThing1, HasManyThing2, hasManyThing3 := &HasManyThing{}, &HasManyThing{}, &HasManyThing{}
	manyToManyThing1, ManyToManyThing2, manyToManyThing3 := &ManyToManyThing{}, &ManyToManyThing{}, &ManyToManyThing{}
	err = gofakeit.Struct(&thing)
	require.NoError(s.T(), err)
	err = gofakeit.Struct(&belongsToThing)
	require.NoError(s.T(), err)
	err = gofakeit.Struct(&belongsToThingTwo)
	require.NoError(s.T(), err)
	err = gofakeit.Struct(&belongsToThingThree)
	require.NoError(s.T(), err)
	err = gofakeit.Struct(&hasOneThing)
	require.NoError(s.T(), err)
	err = gofakeit.Struct(&hasManyThing1)
	require.NoError(s.T(), err)
	err = gofakeit.Struct(&HasManyThing2)
	require.NoError(s.T(), err)
	err = gofakeit.Struct(&hasManyThing3)
	require.NoError(s.T(), err)
	err = gofakeit.Struct(&manyToManyThing1)
	require.NoError(s.T(), err)
	err = gofakeit.Struct(&ManyToManyThing2)
	require.NoError(s.T(), err)
	err = gofakeit.Struct(&manyToManyThing3)
	require.NoError(s.T(), err)
	thing.BelongsTo = belongsToThing
	thing.BelongsToTwo = belongsToThingTwo
	thing.BelongsToThree = belongsToThingThree
	thing.HasOne = hasOneThing
	thing.HasMany = []*HasManyThing{hasManyThing1, HasManyThing2, hasManyThing3}
	thing.ManyToMany = []*ManyToManyThing{manyToManyThing1, ManyToManyThing2, manyToManyThing3}
	theMap := gofakeit.Map()
	bytes, err := json.Marshal(theMap)
	err = json.Unmarshal(bytes, &thing.AStructpb)
	require.NoError(s.T(), err)
	return
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
	err = db.AutoMigrate(&ThingGormModel{}, &HasOneThingGormModel{}, &HasManyThingGormModel{})
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
