package test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	. "github.com/catalystsquad/protoc-gen-go-gorm/example/postgres"
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/orlangure/gnomock"
	postgres_preset "github.com/orlangure/gnomock/preset/postgres"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/testing/protocmp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"testing"
)

var postgresContainer *gnomock.Container
var postgresDb *gorm.DB

type PostgresPluginSuite struct {
	suite.Suite
}

func TestPostgresPluginSuite(t *testing.T) {
	suite.Run(t, new(PostgresPluginSuite))
}

// TestList tests that the list function works as expected
func (s *PostgresPluginSuite) TestList() {
	// create profiles
	profiles := getPostgresProfiles(s.T(), 3)
	profileProtos := ProfileProtos(profiles)
	_, err := profileProtos.Upsert(context.Background(), postgresDb)
	require.NoError(s.T(), err)

	// list profiles
	fetchedProfiles := ProfileProtos{}
	err = fetchedProfiles.List(context.Background(), postgresDb, 10, 0, nil)
	require.NoError(s.T(), err)

	// assert equality, tests are run in parallel so filter down to the ids we know about
	idsSet := hashset.New()
	for _, profile := range profileProtos {
		idsSet.Add(*profile.Id)
	}
	actualProfiles := ProfileProtos{}
	for _, profile := range fetchedProfiles {
		if idsSet.Contains(*profile.Id) {
			actualProfiles = append(actualProfiles, profile)
		}
	}
	assertPostgresProtosEquality(s.T(), profileProtos, actualProfiles,
		protocmp.IgnoreFields(&Profile{}, "created_at", "updated_at"),
	)
}

// TestGetByIds tests that the getByIds function works as expected
func (s *PostgresPluginSuite) TestGetByIds() {
	// create profiles
	profiles := getPostgresProfiles(s.T(), 3)
	profileProtos := ProfileProtos(profiles)
	_, err := profileProtos.Upsert(context.Background(), postgresDb)
	require.NoError(s.T(), err)

	// get profiles
	ids := lo.Map(profileProtos, func(item *Profile, index int) string {
		return *item.Id
	})
	fetchedProfiles := ProfileProtos{}
	err = fetchedProfiles.GetByIds(context.Background(), postgresDb, ids)
	require.NoError(s.T(), err)

	// assert equality
	assertPostgresProtosEquality(s.T(), profileProtos, fetchedProfiles,
		protocmp.IgnoreFields(&Profile{}, "created_at", "updated_at"),
	)
}

// TestBase tests that scalar fields are persisted as we expect them to be
func (s *PostgresPluginSuite) TestBase() {
	// create the user
	user := getPostgresUser(s.T())
	userProtos := UserProtos{user}
	_, err := userProtos.Upsert(context.Background(), postgresDb)
	require.NoError(s.T(), err)

	// fetch the user
	fetchedUserModel, err := getPostgresUserById(*user.Id)
	require.NoError(s.T(), err)
	fetchedUserProto, err := fetchedUserModel.ToProto()
	require.NoError(s.T(), err)

	// assert equality
	assertPostgresProtosEquality(s.T(), userProtos[0], fetchedUserProto,
		protocmp.IgnoreFields(&User{}, "created_at", "updated_at"),
	)
}

// TestHasOneByObject tests that fields related with a has one relationship are persisted as we expect them to be when saved as an object
func (s *PostgresPluginSuite) TestHasOneByObject() {
	// create the user
	user := getPostgresUser(s.T())
	userProtos := UserProtos{user}
	_, err := userProtos.Upsert(context.Background(), postgresDb)
	require.NoError(s.T(), err)
	expectedUser := userProtos[0]

	// create the address
	address := getPostgresAddress(s.T())
	address.User = user
	addressProtos := AddressProtos{address}
	_, err = addressProtos.Upsert(context.Background(), postgresDb)
	require.NoError(s.T(), err)

	// set the address on the expected proto for comparison
	expectedUser.Address = addressProtos[0]
	expectedUser.Address.User = nil
	expectedUser.Address.UserId = expectedUser.Id

	// fetch the user
	fetchedUserModel, err := getPostgresUserById(*user.Id)
	require.NoError(s.T(), err)
	fetchedUserProto, err := fetchedUserModel.ToProto()
	require.NoError(s.T(), err)

	// assert equality
	assertPostgresProtosEquality(s.T(), userProtos[0], fetchedUserProto,
		protocmp.IgnoreFields(&User{}, "created_at", "updated_at"),
		protocmp.IgnoreFields(&Address{}, "created_at", "updated_at"),
	)
}

// TestHasOneByObject tests that fields related with a has one relationship are persisted as we expect them to be when saved as an id
func (s *PostgresPluginSuite) TestHasOneById() {
	// create the user
	user := getPostgresUser(s.T())
	userProtos := UserProtos{user}
	_, err := userProtos.Upsert(context.Background(), postgresDb)
	require.NoError(s.T(), err)
	expectedUser := userProtos[0]

	// create the address
	address := getPostgresAddress(s.T())
	address.UserId = user.Id
	addressProtos := AddressProtos{address}
	_, err = addressProtos.Upsert(context.Background(), postgresDb)
	require.NoError(s.T(), err)

	// set the address on the expected proto for comparison
	expectedUser.Address = addressProtos[0]
	expectedUser.Address.User = nil
	expectedUser.Address.UserId = expectedUser.Id

	// fetch the user
	fetchedUserModel, err := getPostgresUserById(*user.Id)
	require.NoError(s.T(), err)
	fetchedUserProto, err := fetchedUserModel.ToProto()
	require.NoError(s.T(), err)

	// assert equality
	assertPostgresProtosEquality(s.T(), userProtos[0], fetchedUserProto,
		protocmp.IgnoreFields(&User{}, "created_at", "updated_at"),
		protocmp.IgnoreFields(&Address{}, "created_at", "updated_at"),
	)
}

// TestHasMany tests that fields related with a has many relationship are persisted as we expect them to be
func (s *PostgresPluginSuite) TestHasMany() {
	// create the user
	user := getPostgresUser(s.T())
	userProtos := UserProtos{user}
	_, err := userProtos.Upsert(context.Background(), postgresDb)
	require.NoError(s.T(), err)
	expectedUser := userProtos[0]

	// create comments
	comments := getPostgresComments(s.T(), 3)
	for _, comment := range comments {
		comment.User = user
	}
	commentProtos := CommentProtos(comments)
	_, err = commentProtos.Upsert(context.Background(), postgresDb)
	require.NoError(s.T(), err)

	// set the comments on the expected proto for comparison
	expectedUser.Comments = commentProtos
	for _, comment := range expectedUser.Comments {
		// nil user to avoid stack overflow
		comment.User = nil
	}

	// fetch the user
	fetchedUserModel, err := getPostgresUserById(*user.Id)
	require.NoError(s.T(), err)
	fetchedUserProto, err := fetchedUserModel.ToProto()
	require.NoError(s.T(), err)

	// assert equality
	assertPostgresProtosEquality(s.T(), userProtos[0], fetchedUserProto,
		protocmp.IgnoreFields(&User{}, "created_at", "updated_at"),
		protocmp.IgnoreFields(&Comment{}, "created_at", "updated_at"),
	)
}

// TestManyToMany tests that fields related with a many-to-many relationship are persisted as we expect them to be
func (s *PostgresPluginSuite) TestManyToMany() {
	// create the user
	user := getPostgresUser(s.T())
	userProtos := UserProtos{user}
	_, err := userProtos.Upsert(context.Background(), postgresDb)
	require.NoError(s.T(), err)
	expectedUser := userProtos[0]

	// create profiles
	profiles := getPostgresProfiles(s.T(), 3)
	profileProtos := ProfileProtos(profiles)
	_, err = profileProtos.Upsert(context.Background(), postgresDb)
	require.NoError(s.T(), err)

	// associate profiles
	session := postgresDb.Session(&gorm.Session{})
	userModel, err := user.ToModel()
	require.NoError(s.T(), err)
	profileModels, err := profileProtos.ToModels()
	require.NoError(s.T(), err)

	err = session.Model(userModel).Association("Profiles").Replace(profileModels)
	require.NoError(s.T(), err)

	// set the profiles on the expected proto for comparison
	expectedUser.Profiles = profiles

	// fetch the user
	fetchedUserModel, err := getPostgresUserById(*user.Id)
	require.NoError(s.T(), err)
	fetchedUserProto, err := fetchedUserModel.ToProto()
	require.NoError(s.T(), err)

	// assert equality
	assertPostgresProtosEquality(s.T(), userProtos[0], fetchedUserProto,
		protocmp.IgnoreFields(&User{}, "created_at", "updated_at"),
		protocmp.IgnoreFields(&Profile{}, "created_at", "updated_at"),
	)
}

func (s *PostgresPluginSuite) TestSliceTransformers() {
	user := getPostgresUser(s.T())
	users := UserProtos{user}
	models, err := users.ToModels()
	require.NoError(s.T(), err)
	transformedThings, err := models.ToProtos()
	require.NoError(s.T(), err)
	assertPostgresProtosEquality(s.T(), users, transformedThings)
}

func (s *PostgresPluginSuite) SetupSuite() {
	s.T().Parallel()
	preset := postgres_preset.Preset(
		postgres_preset.WithUser("test", "test"),
		postgres_preset.WithDatabase("test"),
		postgres_preset.WithQueriesFile("postgres_queries.sql"))
	var err error
	portOpt := gnomock.WithCustomNamedPorts(gnomock.NamedPorts{"default": gnomock.Port{
		Protocol: "tcp",
		Port:     5432,
		HostPort: 5432,
	}})
	postgresContainer, err = gnomock.Start(preset, portOpt)
	require.NoError(s.T(), err)
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", postgresContainer.Host, postgresContainer.DefaultPort(), "test", "test", "test", "disable")
	logger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,        // Disable color
		},
	)
	postgresDb, err = gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger})
	require.NoError(s.T(), err)
	err = postgresDb.AutoMigrate(&UserGormModel{}, &AddressGormModel{}, &CommentGormModel{})
	require.NoError(s.T(), err)
}

func (s *PostgresPluginSuite) TearDownSuite() {
	require.NoError(s.T(), gnomock.Stop(postgresContainer))
}

func (s *PostgresPluginSuite) SetupTest() {
}

func assertPostgresProtosEquality(t *testing.T, expected, actual interface{}, opts ...cmp.Option) {
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

func getPostgresUser(t *testing.T) (thing *User) {
	thing = &User{}
	err := gofakeit.Struct(&thing)
	require.NoError(t, err)
	theMap := gofakeit.Map()
	bytes, err := json.Marshal(theMap)
	require.NoError(t, err)
	err = json.Unmarshal(bytes, &thing.AStructpb)
	require.NoError(t, err)
	return
}

func getRandomNumPostgresComments(t *testing.T) []*Comment {
	return getPostgresComments(t, gofakeit.Number(2, 10))
}

func getPostgresComments(t *testing.T, num int) []*Comment {
	comments := []*Comment{}
	for i := 0; i < num; i++ {
		var comment *Comment
		err := gofakeit.Struct(&comment)
		require.NoError(t, err)
		comments = append(comments, comment)
	}
	return comments
}

func getRandomNumPostgresProfiles(t *testing.T) []*Profile {
	return getPostgresProfiles(t, gofakeit.Number(2, 10))
}

func getPostgresProfiles(t *testing.T, num int) []*Profile {
	profiles := []*Profile{}
	for i := 0; i < num; i++ {
		var profile *Profile
		err := gofakeit.Struct(&profile)
		require.NoError(t, err)
		profiles = append(profiles, profile)
	}
	return profiles
}

func getRandomNumPostgresCompanys(t *testing.T) []*Company {
	return getPostgresCompanys(t, gofakeit.Number(2, 10))
}

func getPostgresCompanys(t *testing.T, num int) []*Company {
	companys := []*Company{}
	for i := 0; i < num; i++ {
		companys = append(companys, getPostgresCompany(t))
	}
	return companys
}

func getPostgresCompany(t *testing.T) *Company {
	var company *Company
	err := gofakeit.Struct(&company)
	require.NoError(t, err)
	return company
}

func getPostgresAddress(t *testing.T) *Address {
	var address *Address
	err := gofakeit.Struct(&address)
	require.NoError(t, err)
	address.CompanyBlob = getPostgresCompany(t)
	return address
}

func getPostgresUserById(id string) (*UserGormModel, error) {
	session := postgresDb.Session(&gorm.Session{})
	var user *UserGormModel
	err := session.Preload(clause.Associations).First(&user, "id = ?", id).Error
	return user, err
}
