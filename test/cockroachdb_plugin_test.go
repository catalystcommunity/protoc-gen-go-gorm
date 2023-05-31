package test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	. "github.com/catalystsquad/protoc-gen-go-gorm/example/cockroachdb"
	"github.com/google/go-cmp/cmp"
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

var cockroachdbContainer *gnomock.Container
var cockroachdbDb *gorm.DB

type PluginSuite struct {
	suite.Suite
}

func TestPluginSuite(t *testing.T) {
	suite.Run(t, new(PluginSuite))
}

func (s *PluginSuite) TestPlugin() {
	user, err := getPopulatedUser(s.T())
	require.NoError(s.T(), err)
	users := UserProtos{user}
	err = users.Save(context.Background(), cockroachdbDb, nil, nil, false)
	require.NoError(s.T(), err)
	upsertedUser := users[0]
	require.NotNil(s.T(), upsertedUser.Company)
	require.NotNil(s.T(), upsertedUser.CompanyTwo)
	require.NotNil(s.T(), upsertedUser.CompanyThree)
	require.NotNil(s.T(), upsertedUser.Address)
	require.Greater(s.T(), len(upsertedUser.Comments), 0)
	require.Greater(s.T(), len(upsertedUser.Profiles), 0)
	expectedCreatedAt := users[0].CreatedAt
	var firstUserModel *UserGormModel
	var firstUser *User
	err = cockroachdbDb.Preload(clause.Associations).First(&firstUserModel).Error
	require.NoError(s.T(), err)
	firstUser, err = firstUserModel.ToProto()
	require.NoError(s.T(), err)
	assertProtosEquality(s.T(), users[0], firstUser)
	// do an update to ensure updated at field is updated and created
	oldInt32 := firstUser.AnInt32
	newInt32 := gofakeit.Int32()
	require.NotEqual(s.T(), oldInt32, newInt32)
	firstUser.AnInt32 = newInt32
	update := proto.Clone(firstUser)
	updates := UserProtos{update.(*User)}
	updates[0].Company.Name = "derp"
	err = updates.Save(context.Background(), cockroachdbDb, nil, nil, false)
	require.NoError(s.T(), err)
	require.Equal(s.T(), expectedCreatedAt, updates[0].CreatedAt)
	createdAt, err := time.Parse(time.RFC3339Nano, updates[0].CreatedAt)
	require.NoError(s.T(), err)
	require.NotEqual(s.T(), createdAt.UnixMilli(), updates[0].UpdatedAt.AsTime().UnixMilli())
	require.NotEqual(s.T(), updates[0].AnInt32, oldInt32)
	require.Equal(s.T(), updates[0].AnInt32, newInt32)
	// test list
	listedUsers := UserProtos{}
	err = listedUsers.List(context.Background(), cockroachdbDb, 100, 0, nil)
	require.NoError(s.T(), err)
	assertProtosEquality(s.T(), updates, listedUsers)
	// test get by ids
	ids := []string{*listedUsers[0].Id}
	fetchedUsers := UserProtos{}
	err = fetchedUsers.GetByIds(context.Background(), cockroachdbDb, ids)
	require.NoError(s.T(), err)
	assertProtosEquality(s.T(), listedUsers, fetchedUsers)
	// test delete
	err = DeleteUserGormModels(context.Background(), cockroachdbDb, ids)
	require.NoError(s.T(), err)
	err = listedUsers.List(context.Background(), cockroachdbDb, 100, 0, nil)
	require.NoError(s.T(), err)
	require.Len(s.T(), listedUsers, 0)
}

func (s *PluginSuite) TestSliceTransformers() {
	user, err := getPopulatedUser(s.T())
	require.NoError(s.T(), err)
	users := UserProtos{user}
	models, err := users.ToModels()
	require.NoError(s.T(), err)
	transformedThings, err := models.ToProtos()
	require.NoError(s.T(), err)
	assertProtosEquality(s.T(), users, transformedThings)
}

func (s *PluginSuite) SetupSuite() {
	s.T().Parallel()
	preset := cockroachdb.Preset()
	var err error
	portOpt := gnomock.WithCustomNamedPorts(gnomock.NamedPorts{"default": gnomock.Port{
		Protocol: "tcp",
		Port:     26257,
		HostPort: 26257,
	}})
	cockroachdbContainer, err = gnomock.Start(preset, portOpt)
	require.NoError(s.T(), err)
	dsn := fmt.Sprintf("host=%s port=%d user=root dbname=%s sslmode=disable", cockroachdbContainer.Host, cockroachdbContainer.DefaultPort(), "postgres")
	logger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,        // Disable color
		},
	)
	cockroachdbDb, err = gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger})
	require.NoError(s.T(), err)
	err = cockroachdbDb.AutoMigrate(&UserGormModel{}, &AddressGormModel{}, &CommentGormModel{})
	require.NoError(s.T(), err)
}

func (s *PluginSuite) TearDownSuite() {
	require.NoError(s.T(), gnomock.Stop(cockroachdbContainer))
}

func (s *PluginSuite) SetupTest() {
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

func getPopulatedUser(t *testing.T) (thing *User, err error) {
	thing = &User{}
	company, company2, company3 := &Company{}, &Company{}, &Company{}
	address := &Address{}
	comment1, comment2, comment3 := &Comment{}, &Comment{}, &Comment{}
	profile1, profile2, profile3 := &Profile{}, &Profile{}, &Profile{}
	err = gofakeit.Struct(&thing)
	require.NoError(t, err)
	err = gofakeit.Struct(&company)
	require.NoError(t, err)
	err = gofakeit.Struct(&company2)
	require.NoError(t, err)
	err = gofakeit.Struct(&company3)
	require.NoError(t, err)
	err = gofakeit.Struct(&address)
	require.NoError(t, err)
	err = gofakeit.Struct(&comment1)
	require.NoError(t, err)
	err = gofakeit.Struct(&comment2)
	require.NoError(t, err)
	err = gofakeit.Struct(&comment3)
	require.NoError(t, err)
	err = gofakeit.Struct(&profile1)
	require.NoError(t, err)
	err = gofakeit.Struct(&profile2)
	require.NoError(t, err)
	err = gofakeit.Struct(&profile3)
	require.NoError(t, err)
	thing.Company = company
	thing.CompanyTwo = company2
	thing.CompanyThree = company3
	thing.Address = address
	thing.Comments = []*Comment{comment1, comment2, comment3}
	thing.Profiles = []*Profile{profile1, profile2, profile3}
	theMap := gofakeit.Map()
	bytes, err := json.Marshal(theMap)
	err = json.Unmarshal(bytes, &thing.AStructpb)
	require.NoError(t, err)
	return
}
