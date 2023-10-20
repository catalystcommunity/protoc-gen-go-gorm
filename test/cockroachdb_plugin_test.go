package test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	. "github.com/catalystsquad/protoc-gen-go-gorm/example/cockroachdb"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/orlangure/gnomock"
	cockroachdb_preset "github.com/orlangure/gnomock/preset/cockroachdb"
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
)

var cockroachdbContainer *gnomock.Container
var cockroachdbDb *gorm.DB

type CockroachdbPluginSuite struct {
	suite.Suite
}

func TestCockroachdbPluginSuite(t *testing.T) {
	suite.Run(t, new(CockroachdbPluginSuite))
}

func (s *CockroachdbPluginSuite) TestPlugin() {
	user := getPopulatedCockroachdbUser(s.T())
	expectedUser := proto.Clone(user)
	users := UserProtos{user}
	err := users.Upsert(context.Background(), cockroachdbDb, nil, nil, false)
	require.NoError(s.T(), err)
	upsertedUser := users[0]
	// assert all objects have the appropriate ids
	require.Equal(s.T(), upsertedUser.Id, upsertedUser.Address.UserId)
	require.Equal(s.T(), upsertedUser.CompanyTwoId, *upsertedUser.CompanyTwo.Id)
	require.Equal(s.T(), upsertedUser.AnUnexpectedId, *upsertedUser.CompanyThree.Id)
	for _, comment := range upsertedUser.Comments {
		require.Equal(s.T(), upsertedUser.Id, comment.UserId)
	}
	// assert equality ignoring generated ids and timestamps
	assertCockroachdbProtosEquality(s.T(), expectedUser, upsertedUser,
		protocmp.IgnoreFields(&Address{}, "id", "created_at", "updated_at", "user_id"),
		protocmp.IgnoreFields(&Company{}, "id", "created_at", "updated_at"),
		protocmp.IgnoreFields(&Comment{}, "id", "created_at", "updated_at", "user_id"),
		protocmp.IgnoreFields(&Profile{}, "id", "created_at", "updated_at"),
		protocmp.IgnoreFields(&User{}, "id", "created_at", "updated_at", "an_unexpected_id", "company_two_id"),
	)
	var firstUserModel *UserGormModel
	var firstUser *User
	err = cockroachdbDb.Preload(clause.Associations).First(&firstUserModel).Error
	require.NoError(s.T(), err)
	firstUser, err = firstUserModel.ToProto()
	require.NoError(s.T(), err)
	assertCockroachdbProtosEquality(s.T(), users[0], firstUser)
	// test update
	expectedUpdatedUser := getPopulatedCockroachdbUser(s.T())
	expectedUpdatedUser.Id = users[0].Id
	toUpdate := proto.Clone(expectedUpdatedUser)
	updatedUsers := UserProtos{toUpdate.(*User)}
	err = updatedUsers.Upsert(context.Background(), cockroachdbDb, nil, nil, false)
	require.NoError(s.T(), err)
	updatedUser := updatedUsers[0]
	// assert all objects have the appropriate ids
	require.Equal(s.T(), updatedUser.Id, updatedUser.Address.UserId)
	require.Equal(s.T(), updatedUser.CompanyTwoId, *updatedUser.CompanyTwo.Id)
	require.Equal(s.T(), updatedUser.AnUnexpectedId, *updatedUser.CompanyThree.Id)
	for _, comment := range updatedUser.Comments {
		require.Equal(s.T(), updatedUser.Id, comment.UserId)
	}
	// assert equality ignoring generated ids and timestamps
	assertCockroachdbProtosEquality(s.T(), expectedUpdatedUser, updatedUsers[0],
		protocmp.IgnoreFields(&Address{}, "id", "created_at", "updated_at", "user_id"),
		protocmp.IgnoreFields(&Company{}, "id", "created_at", "updated_at"),
		protocmp.IgnoreFields(&Comment{}, "id", "created_at", "updated_at", "user_id"),
		protocmp.IgnoreFields(&Profile{}, "id", "created_at", "updated_at"),
		protocmp.IgnoreFields(&User{}, "id", "created_at", "updated_at", "an_unexpected_id", "company_two_id"),
	)
	// test list
	listedUsers := UserProtos{}
	err = listedUsers.List(context.Background(), cockroachdbDb, 100, 0, nil)
	require.NoError(s.T(), err)
	assertCockroachdbProtosEquality(s.T(), updatedUsers, listedUsers)
	// test get by ids
	ids := []string{*listedUsers[0].Id}
	fetchedUsers := UserProtos{}
	err = fetchedUsers.GetByIds(context.Background(), cockroachdbDb, ids)
	require.NoError(s.T(), err)
	assertCockroachdbProtosEquality(s.T(), listedUsers, fetchedUsers)
	// test delete
	err = DeleteUserGormModels(context.Background(), cockroachdbDb, ids)
	require.NoError(s.T(), err)
	err = listedUsers.List(context.Background(), cockroachdbDb, 100, 0, nil)
	require.NoError(s.T(), err)
	require.Len(s.T(), listedUsers, 0)
}

func (s *CockroachdbPluginSuite) TestSliceTransformers() {
	user := getPopulatedCockroachdbUser(s.T())
	users := UserProtos{user}
	models, err := users.ToModels()
	require.NoError(s.T(), err)
	transformedThings, err := models.ToProtos()
	require.NoError(s.T(), err)
	assertCockroachdbProtosEquality(s.T(), users, transformedThings)
}

func (s *CockroachdbPluginSuite) SetupSuite() {
	s.T().Parallel()
	preset := cockroachdb_preset.Preset()
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

func (s *CockroachdbPluginSuite) TearDownSuite() {
	require.NoError(s.T(), gnomock.Stop(cockroachdbContainer))
}

func (s *CockroachdbPluginSuite) SetupTest() {
}

func assertCockroachdbProtosEquality(t *testing.T, expected, actual interface{}, opts ...cmp.Option) {
	// ignoring id, created_at, updated_at, user_id because the original proto doesn't have those, but the
	// one converted from the created model does
	defaultOpts := []cmp.Option{
		cmpopts.SortSlices(func(x, y *Comment) bool {
			return x.Name < y.Name
		}),
		cmpopts.SortSlices(func(x, y *Profile) bool {
			return x.Name < y.Name
		}),
		protocmp.Transform(),
		protocmp.SortRepeated(func(x, y *Comment) bool {
			return x.Name < y.Name
		}),
		protocmp.SortRepeated(func(x, y *Profile) bool {
			return x.Name < y.Name
		}),
	}
	defaultOpts = append(defaultOpts, opts...)
	diff := cmp.Diff(
		expected,
		actual,
		defaultOpts...,
	)
	require.Empty(t,
		diff,
		diff,
	)
}

func getPopulatedCockroachdbUser(t *testing.T) (thing *User) {
	thing = &User{}
	companies := getCockroachdbCompanys(t, 3)
	err := gofakeit.Struct(&thing)
	require.NoError(t, err)
	thing.Company = companies[0]
	thing.CompanyTwo = companies[1]
	thing.CompanyThree = companies[2]
	thing.Address = getCockroachdbAddress(t)
	thing.Comments = getRandomNumCockroachdbComments(t)
	thing.Profiles = getRandomNumCockroachdbProfiles(t)
	theMap := gofakeit.Map()
	bytes, err := json.Marshal(theMap)
	require.NoError(t, err)
	err = json.Unmarshal(bytes, &thing.AStructpb)
	require.NoError(t, err)
	return
}

func getRandomNumCockroachdbComments(t *testing.T) []*Comment {
	return getCockroachdbComments(t, gofakeit.Number(2, 10))
}

func getCockroachdbComments(t *testing.T, num int) []*Comment {
	comments := []*Comment{}
	for i := 0; i < num; i++ {
		var comment *Comment
		err := gofakeit.Struct(&comment)
		require.NoError(t, err)
		comments = append(comments, comment)
	}
	return comments
}

func getRandomNumCockroachdbProfiles(t *testing.T) []*Profile {
	return getCockroachdbProfiles(t, gofakeit.Number(2, 10))
}

func getCockroachdbProfiles(t *testing.T, num int) []*Profile {
	profiles := []*Profile{}
	for i := 0; i < num; i++ {
		var profile *Profile
		err := gofakeit.Struct(&profile)
		require.NoError(t, err)
		profiles = append(profiles, profile)
	}
	return profiles
}

func getRandomNumCockroachdbCompanys(t *testing.T) []*Company {
	return getCockroachdbCompanys(t, gofakeit.Number(2, 10))
}

func getCockroachdbCompanys(t *testing.T, num int) []*Company {
	companys := []*Company{}
	for i := 0; i < num; i++ {
		companys = append(companys, getCockroachdbCompany(t))
	}
	return companys
}

func getCockroachdbCompany(t *testing.T) *Company {
	var company *Company
	err := gofakeit.Struct(&company)
	require.NoError(t, err)
	return company
}

func getCockroachdbAddress(t *testing.T) *Address {
	var address *Address
	err := gofakeit.Struct(&address)
	require.NoError(t, err)
	address.CompanyBlob = getCockroachdbCompany(t)
	return address
}
