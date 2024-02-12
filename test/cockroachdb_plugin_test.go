package test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	. "github.com/catalystsquad/protoc-gen-go-gorm/example/cockroachdb"
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/orlangure/gnomock"
	cockroachdb_preset "github.com/orlangure/gnomock/preset/cockroachdb"
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

var cockroachdbContainer *gnomock.Container
var cockroachdbDb *gorm.DB

type CockroachdbPluginSuite struct {
	suite.Suite
}

func TestCockroachdbPluginSuite(t *testing.T) {
	suite.Run(t, new(CockroachdbPluginSuite))
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

func BenchmarkConvertProtosToProtosMSingle(b *testing.B) {
	b.StopTimer()
	profiles, err := generateCockroachdbProfiles(1)
	if err != nil {
		panic(err)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ConvertProtosToProtosM[*Profile, *ProfileGormModel](profiles)
	}
}

func BenchmarkConvertProtosToProtosMTen(b *testing.B) {
	b.StopTimer()
	profiles, err := generateCockroachdbProfiles(10)
	if err != nil {
		panic(err)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ConvertProtosToProtosM[*Profile, *ProfileGormModel](profiles)
	}
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

func getCockroachdbUser(t *testing.T) (thing *User) {
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
	profiles, err := generateCockroachdbProfiles(num)
	require.NoError(t, err)
	return profiles
}

func generateCockroachdbProfiles(num int) ([]*Profile, error) {
	profiles := []*Profile{}
	for i := 0; i < num; i++ {
		var profile *Profile
		err := gofakeit.Struct(&profile)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}
	return profiles, nil
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

func getUserById(id string) (*UserGormModel, error) {
	session := cockroachdbDb.Session(&gorm.Session{})
	var user *UserGormModel
	err := session.Preload(clause.Associations).First(&user, "id = ?", id).Error
	return user, err
}

// TestList tests that the list function works as expected
func (s *CockroachdbPluginSuite) TestList() {
	// create profiles
	numProfiles := gofakeit.Number(2, 5)
	profiles := getCockroachdbProfiles(s.T(), numProfiles)
	_, err := Upsert[*Profile, *ProfileGormModel](context.Background(), cockroachdbDb, profiles)
	require.NoError(s.T(), err)
	// list profiles
	models, err := List[*ProfileGormModel](context.Background(), cockroachdbDb, 100, 0, "", nil)
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
	assertCockroachdbProtosEquality(s.T(), profiles, fetchedProfiles,
		protocmp.IgnoreFields(&Profile{}, "created_at", "updated_at"),
	)
}

// TestPreloadBelongsTo tests preloading belongs to relationship
func (s *CockroachdbPluginSuite) TestPreloadBelongsTo() {
	// create a user and a company
	company := getCockroachdbCompany(s.T())
	_, err := Upsert[*Company, *CompanyGormModel](context.Background(), cockroachdbDb, []*Company{company})
	require.NoError(s.T(), err)
	user := getCockroachdbUser(s.T())
	user.CompanyId = company.Id
	_, err = Upsert[*User, *UserGormModel](context.Background(), cockroachdbDb, []*User{user})
	require.NoError(s.T(), err)
	// get with preload
	fetchedUsers, err := GetByIds[*UserGormModel](context.Background(), cockroachdbDb, []string{*user.Id}, map[string][]interface{}{"Company": nil})
	require.NoError(s.T(), err)
	// assert
	expectedUserModel := fetchedUsers[0]
	expectedUser, err := expectedUserModel.ToProto()
	require.NoError(s.T(), err)
	assertCockroachdbProtosEquality(s.T(), company, expectedUser.Company,
		protocmp.IgnoreFields(&Company{}, "created_at", "updated_at"),
	)
}

// TestPreloadHasOne tests preloading has one relationship
func (s *CockroachdbPluginSuite) TestPreloadHasOne() {
	// create a user and a address
	user := getCockroachdbUser(s.T())
	_, err := Upsert[*User, *UserGormModel](context.Background(), cockroachdbDb, []*User{user})
	require.NoError(s.T(), err)
	address := getCockroachdbAddress(s.T())
	address.UserId = user.Id
	_, err = Upsert[*Address, *AddressGormModel](context.Background(), cockroachdbDb, []*Address{address})
	require.NoError(s.T(), err)
	// get with preload
	fetchedUsers, err := GetByIds[*UserGormModel](context.Background(), cockroachdbDb, []string{*user.Id}, map[string][]interface{}{"Address": nil})
	require.NoError(s.T(), err)
	// assert
	expectedUserModel := fetchedUsers[0]
	expectedUser, err := expectedUserModel.ToProto()
	require.NoError(s.T(), err)
	assertCockroachdbProtosEquality(s.T(), address, expectedUser.Address,
		protocmp.IgnoreFields(&Address{}, "created_at", "updated_at"),
	)
}

// TestPreloadHasMany tests preloading has many relationship
func (s *CockroachdbPluginSuite) TestPreloadHasMany() {
	// create a user and a address
	user := getCockroachdbUser(s.T())
	_, err := Upsert[*User, *UserGormModel](context.Background(), cockroachdbDb, []*User{user})
	require.NoError(s.T(), err)
	numComments := gofakeit.Number(2, 5)
	comments := getCockroachdbComments(s.T(), numComments)
	for _, comment := range comments {
		comment.UserId = user.Id
	}
	_, err = Upsert[*Comment, *CommentGormModel](context.Background(), cockroachdbDb, comments)
	require.NoError(s.T(), err)
	// get with preload
	fetchedUsers, err := GetByIds[*UserGormModel](context.Background(), cockroachdbDb, []string{*user.Id}, map[string][]interface{}{"Comments": nil})
	require.NoError(s.T(), err)
	// assert
	expectedUserModel := fetchedUsers[0]
	expectedUser, err := expectedUserModel.ToProto()
	require.NoError(s.T(), err)
	assertCockroachdbProtosEquality(s.T(), comments, expectedUser.Comments,
		protocmp.IgnoreFields(&Comment{}, "created_at", "updated_at"),
	)
}

// TestPreloadManyToMany tests preloading many to many relationship
func (s *CockroachdbPluginSuite) TestPreloadManyToMany() {
	// create a user and profiles
	user := getCockroachdbUser(s.T())
	userModels, err := Upsert[*User, *UserGormModel](context.Background(), cockroachdbDb, []*User{user})
	require.NoError(s.T(), err)
	numProfiles := gofakeit.Number(2, 5)
	profiles := getCockroachdbProfiles(s.T(), numProfiles)
	profileModels, err := Upsert[*Profile, *ProfileGormModel](context.Background(), cockroachdbDb, profiles)
	require.NoError(s.T(), err)
	expectedUser := userModels[0]
	// associate the users and profiles
	associations := &ManyToManyAssociations{}
	for _, profile := range profileModels {
		associations.AddAssociation(*expectedUser.Id, *profile.Id)
	}
	err = AssociateManyToMany[*UserGormModel, *ProfileGormModel](context.Background(), cockroachdbDb, associations, "Profiles")
	require.NoError(s.T(), err)
	// get with preload
	fetchedUsers, err := GetByIds[*UserGormModel](context.Background(), cockroachdbDb, []string{*user.Id}, map[string][]interface{}{"Profiles": nil})
	require.NoError(s.T(), err)
	// assert
	expectedUserModel := fetchedUsers[0]
	fetchedUser, err := expectedUserModel.ToProto()
	require.NoError(s.T(), err)
	assertCockroachdbProtosEquality(s.T(), profiles, fetchedUser.Profiles,
		protocmp.IgnoreFields(&Profile{}, "created_at", "updated_at"),
	)
}

// TestDissociateManyToMany tests that dissociateManyToMany works as expected
func (s *CockroachdbPluginSuite) TestDissociateManyToMany() {
	// create a user and profiles
	user := getCockroachdbUser(s.T())
	userModels, err := Upsert[*User, *UserGormModel](context.Background(), cockroachdbDb, []*User{user})
	require.NoError(s.T(), err)
	numProfiles := gofakeit.Number(5, 10)
	profiles := getCockroachdbProfiles(s.T(), numProfiles)
	profileModels, err := Upsert[*Profile, *ProfileGormModel](context.Background(), cockroachdbDb, profiles)
	require.NoError(s.T(), err)
	// associate the users and profiles
	associations := &ManyToManyAssociations{}
	for _, profile := range profileModels {
		associations.AddAssociation(*userModels[0].Id, *profile.Id)
	}
	err = AssociateManyToMany[*UserGormModel, *ProfileGormModel](context.Background(), cockroachdbDb, associations, "Profiles")
	require.NoError(s.T(), err)
	// get with preload
	fetchedUsers, err := GetByIds[*UserGormModel](context.Background(), cockroachdbDb, []string{*user.Id}, map[string][]interface{}{"Profiles": nil})
	require.NoError(s.T(), err)
	// assert
	expectedUserModel := fetchedUsers[0]
	expectedUser, err := expectedUserModel.ToProto()
	require.NoError(s.T(), err)
	assertCockroachdbProtosEquality(s.T(), profiles, expectedUser.Profiles,
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
	err = DissociateManyToMany[*UserGormModel, *ProfileGormModel](context.Background(), cockroachdbDb, dissociations, "Profiles")
	require.NoError(s.T(), err)
	// get with preload
	fetchedUsersAfterDissociate, err := GetByIds[*UserGormModel](context.Background(), cockroachdbDb, []string{*user.Id}, map[string][]interface{}{"Profiles": nil})
	require.NoError(s.T(), err)
	fetchedUserAfterDissociate := fetchedUsersAfterDissociate[0]
	// assert no longer associated
	require.Len(s.T(), fetchedUserAfterDissociate.Profiles, len(profiles)-len(profilesToDissociate))
	for _, profile := range fetchedUserAfterDissociate.Profiles {
		require.False(s.T(), dissociatedIds.Contains(*profile.Id))
	}
}

// TestDissociateManyToMany tests that dissociateManyToMany works as expected
func (s *CockroachdbPluginSuite) TestReplaceManyToMany() {
	// create a user and profiles
	user := getCockroachdbUser(s.T())
	userModels, err := Upsert[*User, *UserGormModel](context.Background(), cockroachdbDb, []*User{user})
	require.NoError(s.T(), err)
	numProfiles := gofakeit.Number(5, 10)
	profiles := getCockroachdbProfiles(s.T(), numProfiles)
	profileModels, err := Upsert[*Profile, *ProfileGormModel](context.Background(), cockroachdbDb, profiles)
	require.NoError(s.T(), err)
	// associate the users and profiles
	associations := &ManyToManyAssociations{}
	for _, profile := range profileModels {
		associations.AddAssociation(*userModels[0].Id, *profile.Id)
	}
	err = AssociateManyToMany[*UserGormModel, *ProfileGormModel](context.Background(), cockroachdbDb, associations, "Profiles")
	require.NoError(s.T(), err)
	// get with preload
	fetchedUsers, err := GetByIds[*UserGormModel](context.Background(), cockroachdbDb, []string{*user.Id}, map[string][]interface{}{"Profiles": nil})
	require.NoError(s.T(), err)
	// assert
	expectedUserModel := fetchedUsers[0]
	expectedUser, err := expectedUserModel.ToProto()
	require.NoError(s.T(), err)
	assertCockroachdbProtosEquality(s.T(), profiles, expectedUser.Profiles,
		protocmp.IgnoreFields(&Profile{}, "created_at", "updated_at"),
	)
	// replace
	replacementProfiles := getCockroachdbProfiles(s.T(), numProfiles)
	replacementProfileModels, err := Upsert[*Profile, *ProfileGormModel](context.Background(), cockroachdbDb, replacementProfiles)
	require.NoError(s.T(), err)
	// associate the users and profiles
	replacementAssociations := &ManyToManyAssociations{}
	for _, profile := range replacementProfileModels {
		replacementAssociations.AddAssociation(*userModels[0].Id, *profile.Id)
	}
	err = ReplaceManyToMany[*UserGormModel, *ProfileGormModel](context.Background(), cockroachdbDb, replacementAssociations, "Profiles")
	require.NoError(s.T(), err)
	// get with preload
	fetchedUsers, err = GetByIds[*UserGormModel](context.Background(), cockroachdbDb, []string{*user.Id}, map[string][]interface{}{"Profiles": nil})
	require.NoError(s.T(), err)
	// assert
	expectedUserModel = fetchedUsers[0]
	expectedUser, err = expectedUserModel.ToProto()
	require.NoError(s.T(), err)
	assertCockroachdbProtosEquality(s.T(), replacementProfiles, expectedUser.Profiles,
		protocmp.IgnoreFields(&Profile{}, "created_at", "updated_at"),
	)
}

// TestListWithWhere tests that the list function works with a where clause set on the tx
func (s *CockroachdbPluginSuite) TestListWithWhere() {
	// create profiles
	numProfiles := gofakeit.Number(2, 5)
	profiles := getCockroachdbProfiles(s.T(), numProfiles)
	_, err := Upsert[*Profile, *ProfileGormModel](context.Background(), cockroachdbDb, profiles)
	require.NoError(s.T(), err)
	// list profiles using session with a where clause
	expected := profiles[0]
	session := cockroachdbDb.Session(&gorm.Session{})
	session = session.Where("name = ?", expected.Name)
	models, err := List[*ProfileGormModel](context.Background(), session, 100, 0, "", nil)
	require.NoError(s.T(), err)
	require.Len(s.T(), models, 1)
	// assert equality
	fetchedProfiles, err := ToProtos[*Profile, *ProfileGormModel](models)
	require.NoError(s.T(), err)
	assertCockroachdbProtosEquality(s.T(), expected, fetchedProfiles[0],
		protocmp.IgnoreFields(&Profile{}, "created_at", "updated_at"),
	)
}

// TestGetByIds tests that the getByIds function works as expected
func (s *CockroachdbPluginSuite) TestGetByIds() {
	// create profiles
	numProfiles := gofakeit.Number(5, 10)
	profiles := getCockroachdbProfiles(s.T(), numProfiles)
	upsertedProfiles, err := Upsert[*Profile, *ProfileGormModel](context.Background(), cockroachdbDb, profiles)
	require.NoError(s.T(), err)
	// get by id
	ids := lo.Map(upsertedProfiles[:2], func(item *ProfileGormModel, index int) string { return *item.Id })
	fetchedModels, err := GetByIds[*ProfileGormModel](context.Background(), cockroachdbDb, ids, nil)
	require.NoError(s.T(), err)
	require.Len(s.T(), fetchedModels, len(ids))
	// assert equality
	fetchedProfiles, err := ToProtos[*Profile, *ProfileGormModel](fetchedModels)
	require.NoError(s.T(), err)
	assertCockroachdbProtosEquality(s.T(), profiles[:2], fetchedProfiles,
		protocmp.IgnoreFields(&Profile{}, "created_at", "updated_at"),
	)
}

// TestDelete tests that the delete function works as expected
func (s *CockroachdbPluginSuite) TestDelete() {
	// create profiles
	numProfiles := gofakeit.Number(2, 5)
	profiles := getCockroachdbProfiles(s.T(), numProfiles)
	_, err := Upsert[*Profile, *ProfileGormModel](context.Background(), cockroachdbDb, profiles)
	require.NoError(s.T(), err)
	// list profiles
	models, err := List[*ProfileGormModel](context.Background(), cockroachdbDb, 100, 0, "", nil)
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
	assertCockroachdbProtosEquality(s.T(), profiles, fetchedProfiles,
		protocmp.IgnoreFields(&Profile{}, "created_at", "updated_at"),
	)
	// delete
	session := cockroachdbDb.Session(&gorm.Session{})
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
	fetchedModels, err := GetByIds[*ProfileGormModel](context.Background(), cockroachdbDb, idsToDelete, nil)
	require.NoError(s.T(), err)
	require.Len(s.T(), fetchedModels, 0)
}
