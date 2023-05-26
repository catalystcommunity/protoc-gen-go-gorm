package test

import (
	"context"
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
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"testing"
	"time"
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
	user, err := s.getPopulatedUser()
	require.NoError(s.T(), err)
	users := UserProtos{user}
	err = users.Upsert(context.Background(), db)
	expectedCreatedAt := users[0].CreatedAt
	var firstUserModel *UserGormModel
	var firstUser *User
	err = db.Joins("Company").Joins("Address").Preload(clause.Associations).First(&firstUserModel).Error
	require.NoError(s.T(), err)
	firstUser, err = firstUserModel.ToProto()
	require.NoError(s.T(), err)
	assertProtosEquality(s.T(), users[0], firstUser)
	// do an update to ensure updated at field is updated and created
	time.Sleep(2 * time.Second)
	firstUser.AnInt32 = gofakeit.Int32()
	update := proto.Clone(firstUser)
	updates := UserProtos{update.(*User)}
	err = updates.Upsert(context.Background(), db)
	require.NoError(s.T(), err)
	require.Equal(s.T(), expectedCreatedAt, updates[0].CreatedAt)
	createdAt, err := time.Parse(time.RFC3339Nano, updates[0].CreatedAt)
	require.NoError(s.T(), err)
	require.NotEqual(s.T(), createdAt.UnixMilli(), updates[0].UpdatedAt.AsTime().UnixMilli())
}

func (s *PluginSuite) TestSliceTransformers() {
	user, err := s.getPopulatedUser()
	require.NoError(s.T(), err)
	users := UserProtos{user}
	models, err := users.ToModels()
	require.NoError(s.T(), err)
	transformedThings, err := models.ToProtos()
	require.NoError(s.T(), err)
	assertProtosEquality(s.T(), users, transformedThings)
}

func assertModelsEquality(t *testing.T, expected, actual interface{}) {
	// no need to ignore id, created_at, updated_at because gorm populates them on create
	require.Empty(t, cmp.Diff(
		expected,
		actual,
		cmpopts.SortSlices(func(x, y *CommentGormModel) bool {
			return x.Name < y.Name
		}),
		cmpopts.SortSlices(func(x, y *ProfileGormModel) bool {
			return x.Name < y.Name
		}),
		cmpopts.IgnoreFields(UserGormModel{}, "AStructpb"),
	))
}

func assertProtosEquality(t *testing.T, expected, actual interface{}, ignoreFields ...string) {
	// ignoring id, created_at, updated_at, user_id because the original proto doesn't have those, but the
	// one converted from the created model does
	require.Empty(t,
		cmp.Diff(
			expected,
			actual,
			protocmp.Transform(),
			protocmp.SortRepeated(func(x, y *Comment) bool {
				return x.Name < y.Name
			}),
			protocmp.SortRepeated(func(x, y *Profile) bool {
				return x.Name < y.Name
			}),
		),
	)
}

func (s *PluginSuite) getPopulatedUser() (thing *User, err error) {
	thing = &User{}
	company, company2, company3 := &Company{}, &Company{}, &Company{}
	address := &Address{}
	comment1, comment2, comment3 := &Comment{}, &Comment{}, &Comment{}
	profile1, profile2, profile3 := &Profile{}, &Profile{}, &Profile{}
	err = gofakeit.Struct(&thing)
	require.NoError(s.T(), err)
	err = gofakeit.Struct(&company)
	require.NoError(s.T(), err)
	err = gofakeit.Struct(&company2)
	require.NoError(s.T(), err)
	err = gofakeit.Struct(&company3)
	require.NoError(s.T(), err)
	err = gofakeit.Struct(&address)
	require.NoError(s.T(), err)
	err = gofakeit.Struct(&comment1)
	require.NoError(s.T(), err)
	err = gofakeit.Struct(&comment2)
	require.NoError(s.T(), err)
	err = gofakeit.Struct(&comment3)
	require.NoError(s.T(), err)
	err = gofakeit.Struct(&profile1)
	require.NoError(s.T(), err)
	err = gofakeit.Struct(&profile2)
	require.NoError(s.T(), err)
	err = gofakeit.Struct(&profile3)
	require.NoError(s.T(), err)
	thing.Company = company
	thing.CompanyTwo = company2
	thing.CompanyThree = company3
	thing.Address = address
	thing.Comments = []*Comment{comment1, comment2, comment3}
	thing.Profiles = []*Profile{profile1, profile2, profile3}
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
	logger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,        // Disable color
		},
	)
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger})
	require.NoError(s.T(), err)
	err = db.AutoMigrate(&UserGormModel{}, &AddressGormModel{}, &CommentGormModel{})
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
