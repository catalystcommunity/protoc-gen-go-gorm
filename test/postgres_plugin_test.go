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

// TestList tests that the list function works as expected
func (s *PostgresPluginSuite) TestList() {
	// create profiles
	numProfiles := gofakeit.Number(2, 5)
	profiles := getPostgresProfiles(s.T(), numProfiles)
	_, err := Upsert[*Profile, *ProfileGormModel](context.Background(), postgresDb, profiles)
	require.NoError(s.T(), err)
	// list profiles
	models, err := List[*ProfileGormModel](context.Background(), postgresDb, 100, 0, "", nil)
	require.NoError(s.T(), err)
	// assert equality
	idsSet := hashset.New()
	for _, profile := range profiles {
		idsSet.Add(*profile.Id)
	}
	fetchedProfiles, err := ToProtos[*Profile, *ProfileGormModel](models)
	require.NoError(s.T(), err)
	// filter down to the ids we created
	fetchedProfiles = lo.Filter(fetchedProfiles, func(item *Profile, index int) bool { return idsSet.Contains(*item.Id) })
	require.Len(s.T(), fetchedProfiles, numProfiles)
	assertPostgresProtosEquality(s.T(), profiles, fetchedProfiles,
		protocmp.IgnoreFields(&Profile{}, "created_at", "updated_at"),
	)
}

// TestPreloadBelongsTo tests preloading belongs to relationship
func (s *PostgresPluginSuite) TestPreloadBelongsTo() {
	// create a user and a company
	company := getPostgresCompany(s.T())
	_, err := Upsert[*Company, *CompanyGormModel](context.Background(), postgresDb, []*Company{company})
	require.NoError(s.T(), err)
	user := getPostgresUser(s.T())
	user.CompanyId = company.Id
	_, err = Upsert[*User, *UserGormModel](context.Background(), postgresDb, []*User{user})
	require.NoError(s.T(), err)
	// get with preload
	fetchedUsers, err := GetByIds[*UserGormModel](context.Background(), postgresDb, []string{*user.Id}, []string{"Company"})
	require.NoError(s.T(), err)
	// assert
	expectedUserModel := fetchedUsers[0]
	expectedUser, err := expectedUserModel.ToProto()
	require.NoError(s.T(), err)
	assertPostgresProtosEquality(s.T(), company, expectedUser.Company,
		protocmp.IgnoreFields(&Company{}, "created_at", "updated_at"),
	)
}

// TestPreloadHasOne tests preloading has one relationship
func (s *PostgresPluginSuite) TestPreloadHasOne() {
	// create a user and a address
	user := getPostgresUser(s.T())
	_, err := Upsert[*User, *UserGormModel](context.Background(), postgresDb, []*User{user})
	require.NoError(s.T(), err)
	address := getPostgresAddress(s.T())
	address.UserId = user.Id
	_, err = Upsert[*Address, *AddressGormModel](context.Background(), postgresDb, []*Address{address})
	require.NoError(s.T(), err)
	// get with preload
	fetchedUsers, err := GetByIds[*UserGormModel](context.Background(), postgresDb, []string{*user.Id}, []string{"Address"})
	require.NoError(s.T(), err)
	// assert
	expectedUserModel := fetchedUsers[0]
	expectedUser, err := expectedUserModel.ToProto()
	require.NoError(s.T(), err)
	assertPostgresProtosEquality(s.T(), address, expectedUser.Address,
		protocmp.IgnoreFields(&Address{}, "created_at", "updated_at"),
	)
}

// TestPreloadHasMany tests preloading has many relationship
func (s *PostgresPluginSuite) TestPreloadHasMany() {
	// create a user and a address
	user := getPostgresUser(s.T())
	_, err := Upsert[*User, *UserGormModel](context.Background(), postgresDb, []*User{user})
	require.NoError(s.T(), err)
	numComments := gofakeit.Number(2, 5)
	comments := getPostgresComments(s.T(), numComments)
	for _, comment := range comments {
		comment.UserId = user.Id
	}
	_, err = Upsert[*Comment, *CommentGormModel](context.Background(), postgresDb, comments)
	require.NoError(s.T(), err)
	// get with preload
	fetchedUsers, err := GetByIds[*UserGormModel](context.Background(), postgresDb, []string{*user.Id}, []string{"Comments"})
	require.NoError(s.T(), err)
	// assert
	expectedUserModel := fetchedUsers[0]
	expectedUser, err := expectedUserModel.ToProto()
	require.NoError(s.T(), err)
	assertPostgresProtosEquality(s.T(), comments, expectedUser.Comments,
		protocmp.IgnoreFields(&Comment{}, "created_at", "updated_at"),
	)
}

// TestPreloadManyToMany tests preloading many to many relationship
func (s *PostgresPluginSuite) TestPreloadManyToMany() {
	// create a user and profiles
	user := getPostgresUser(s.T())
	userModels, err := Upsert[*User, *UserGormModel](context.Background(), postgresDb, []*User{user})
	require.NoError(s.T(), err)
	numProfiles := gofakeit.Number(2, 5)
	profiles := getPostgresProfiles(s.T(), numProfiles)
	profileModels, err := Upsert[*Profile, *ProfileGormModel](context.Background(), postgresDb, profiles)
	require.NoError(s.T(), err)
	expectedUser := userModels[0]
	// associate the users and profiles
	associations := &ManyToManyAssociations{}
	for _, profile := range profileModels {
		associations.AddAssociation(*expectedUser.Id, *profile.Id)
	}
	err = AssociateManyToMany[*UserGormModel, *ProfileGormModel](context.Background(), postgresDb, associations, "Profiles")
	require.NoError(s.T(), err)
	// get with preload
	fetchedUsers, err := GetByIds[*UserGormModel](context.Background(), postgresDb, []string{*user.Id}, []string{"Profiles"})
	require.NoError(s.T(), err)
	// assert
	expectedUserModel := fetchedUsers[0]
	fetchedUser, err := expectedUserModel.ToProto()
	require.NoError(s.T(), err)
	assertPostgresProtosEquality(s.T(), profiles, fetchedUser.Profiles,
		protocmp.IgnoreFields(&Profile{}, "created_at", "updated_at"),
	)
}

// TestDissociateManyToMany tests that dissociateManyToMany works as expected
func (s *PostgresPluginSuite) TestDissociateManyToMany() {
	// create a user and profiles
	user := getPostgresUser(s.T())
	userModels, err := Upsert[*User, *UserGormModel](context.Background(), postgresDb, []*User{user})
	require.NoError(s.T(), err)
	numProfiles := gofakeit.Number(5, 10)
	profiles := getPostgresProfiles(s.T(), numProfiles)
	profileModels, err := Upsert[*Profile, *ProfileGormModel](context.Background(), postgresDb, profiles)
	require.NoError(s.T(), err)
	// associate the users and profiles
	associations := &ManyToManyAssociations{}
	for _, profile := range profileModels {
		associations.AddAssociation(*userModels[0].Id, *profile.Id)
	}
	err = AssociateManyToMany[*UserGormModel, *ProfileGormModel](context.Background(), postgresDb, associations, "Profiles")
	require.NoError(s.T(), err)
	// get with preload
	fetchedUsers, err := GetByIds[*UserGormModel](context.Background(), postgresDb, []string{*user.Id}, []string{"Profiles"})
	require.NoError(s.T(), err)
	// assert
	expectedUserModel := fetchedUsers[0]
	expectedUser, err := expectedUserModel.ToProto()
	require.NoError(s.T(), err)
	assertPostgresProtosEquality(s.T(), profiles, expectedUser.Profiles,
		protocmp.IgnoreFields(&Profile{}, "created_at", "updated_at"),
	)
	// dissociate
	profilesToDissociate := profileModels[:3]
	dissociatedIds := hashset.New()
	dissociations := &ManyToManyAssociations{}
	for _, profile := range profilesToDissociate {
		dissociatedIds.Add(*profile.Id)
		dissociations.AddAssociation(*userModels[0].Id, *profile.Id)
	}
	err = DissociateManyToMany[*UserGormModel, *ProfileGormModel](context.Background(), postgresDb, dissociations, "Profiles")
	require.NoError(s.T(), err)
	// get with preload
	fetchedUsersAfterDissociate, err := GetByIds[*UserGormModel](context.Background(), postgresDb, []string{*user.Id}, []string{"Profiles"})
	require.NoError(s.T(), err)
	fetchedUserAfterDissociate := fetchedUsersAfterDissociate[0]
	// assert no longer associated
	require.Len(s.T(), fetchedUserAfterDissociate.Profiles, len(profiles)-len(profilesToDissociate))
	for _, profile := range fetchedUserAfterDissociate.Profiles {
		require.False(s.T(), dissociatedIds.Contains(*profile.Id))
	}
}

// TestListWithWhere tests that the list function works with a where clause set on the tx
func (s *PostgresPluginSuite) TestListWithWhere() {
	// create profiles
	numProfiles := gofakeit.Number(2, 5)
	profiles := getPostgresProfiles(s.T(), numProfiles)
	_, err := Upsert[*Profile, *ProfileGormModel](context.Background(), postgresDb, profiles)
	require.NoError(s.T(), err)
	// list profiles using session with a where clause
	expected := profiles[0]
	session := postgresDb.Session(&gorm.Session{})
	session = session.Where("name = ?", expected.Name)
	models, err := List[*ProfileGormModel](context.Background(), session, 100, 0, "", nil)
	require.NoError(s.T(), err)
	require.Len(s.T(), models, 1)
	// assert equality
	fetchedProfiles, err := ToProtos[*Profile, *ProfileGormModel](models)
	require.NoError(s.T(), err)
	assertPostgresProtosEquality(s.T(), expected, fetchedProfiles[0],
		protocmp.IgnoreFields(&Profile{}, "created_at", "updated_at"),
	)
}

// TestGetByIds tests that the getByIds function works as expected
func (s *PostgresPluginSuite) TestGetByIds() {
	// create profiles
	numProfiles := gofakeit.Number(5, 10)
	profiles := getPostgresProfiles(s.T(), numProfiles)
	upsertedProfiles, err := Upsert[*Profile, *ProfileGormModel](context.Background(), postgresDb, profiles)
	require.NoError(s.T(), err)
	// get by id
	ids := lo.Map(upsertedProfiles[:2], func(item *ProfileGormModel, index int) string { return *item.Id })
	fetchedModels, err := GetByIds[*ProfileGormModel](context.Background(), postgresDb, ids, nil)
	require.NoError(s.T(), err)
	require.Len(s.T(), fetchedModels, len(ids))
	// assert equality
	fetchedProfiles, err := ToProtos[*Profile, *ProfileGormModel](fetchedModels)
	require.NoError(s.T(), err)
	assertPostgresProtosEquality(s.T(), profiles[:2], fetchedProfiles,
		protocmp.IgnoreFields(&Profile{}, "created_at", "updated_at"),
	)
}

// TestDelete tests that the delete function works as expected
func (s *PostgresPluginSuite) TestDelete() {
	// create profiles
	numProfiles := gofakeit.Number(2, 5)
	profiles := getPostgresProfiles(s.T(), numProfiles)
	_, err := Upsert[*Profile, *ProfileGormModel](context.Background(), postgresDb, profiles)
	require.NoError(s.T(), err)
	// list profiles
	models, err := List[*ProfileGormModel](context.Background(), postgresDb, 100, 0, "", nil)
	require.NoError(s.T(), err)
	// assert equality
	idsSet := hashset.New()
	for _, profile := range profiles {
		idsSet.Add(*profile.Id)
	}
	fetchedProfiles, err := ToProtos[*Profile, *ProfileGormModel](models)
	require.NoError(s.T(), err)
	// filter down to the ids we created
	fetchedProfiles = lo.Filter(fetchedProfiles, func(item *Profile, index int) bool { return idsSet.Contains(*item.Id) })
	require.Len(s.T(), fetchedProfiles, numProfiles)
	assertPostgresProtosEquality(s.T(), profiles, fetchedProfiles,
		protocmp.IgnoreFields(&Profile{}, "created_at", "updated_at"),
	)
	// delete
	session := postgresDb.Session(&gorm.Session{})
	// add returning id clause to test returned models
	session = session.Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}})
	idsToDelete := []string{}
	for _, id := range idsSet.Values() {
		idsToDelete = append(idsToDelete, id.(string))
	}
	deletedModels, err := Delete[*ProfileGormModel](context.Background(), session, idsToDelete)
	require.NoError(s.T(), err)
	for _, model := range deletedModels {
		require.NotNil(s.T(), model.Id)
		require.True(s.T(), idsSet.Contains(*model.Id))
	}
	fetchedModels, err := GetByIds[*ProfileGormModel](context.Background(), postgresDb, idsToDelete, nil)
	require.NoError(s.T(), err)
	require.Len(s.T(), fetchedModels, 0)
}
