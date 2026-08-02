package api

import (
	"context"
	"errors"
	"github.com/reeceappling/goUtils/v2/utils"
	sliceutils "github.com/reeceappling/goUtils/v2/utils/slices"
	"github.com/reeceappling/mushDb/api/env"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//var _ webauthn.User = &User{}

type User struct {
	Email string    `bson:"_id" json:"_id"`
	Perms UserPerms `bson:"perms,omitempty" json:"perms,omitempty"` // PROJECTS COME FROM PERMS! So even projects with only read perms can be associated with an item!
	// All can view?
}

func (u User) DbId() string {
	return u.Email
}

func (u User) IdValue() any {
	return u.Email
}

func (u *User) Reload(ctx context.Context) error {
	if u == nil {
		return errors.New("User is nil")
	}
	return DbFrom(ctx).
		Collection(UserCollName).
		FindOne(ctx, BsonFindFilter(IDfld, u.Email)).
		Decode(u) // TODO: ensure works!
}

//func (u User) WebAuthnID() []byte {
//	panic("not implemented") // TODO: see webauthn.User
//}
//func (u User) WebAuthnName() string {
//	panic("not implemented") // TODO: see webauthn.User
//}
//func (u User) WebAuthnDisplayName() string {
//	panic("not implemented") // TODO: see webauthn.User
//}
//func (u User) WebAuthnCredentials() []webauthn.Credential {
//	panic("not implemented") // TODO: see webauthn.User
//}
//func (u User) AddCredential(cred *webauthn.Credential) {
//	panic("not implemented") // TODO: see PasskeyUser
//}
//func (u User) UpdateCredential(cred *webauthn.Credential) {
//	panic("not implemented") // TODO: see PasskeyUser
//}

func initializeUsers(ctx context.Context) error {
	//Indices
	coll := DbFrom(ctx).Collection(UserCollName)
	//err := createIndexes(ctx, coll, []mongo.IndexModel{ // TODO: FIX!
	//	//newSimpleIndex("username", "username", false, false, true), // TODO: is true ok?
	//	//newSimpleIndex("email", "email", false, false, true),       // TODO: is true ok?
	//	//newSimpleIndex("googleId", "googleIde", false, true, true), // TODO: is true ok?
	//})
	//if err != nil {
	//	return err
	//}
	// TODO: DELETE THIS AFTER TESTING!!!!
	println("Adding test users that are not completely reset after every boot") // TODO: del!
	return env.IfNotProd(ctx, func() error {
		// resolve final test user projects list
		testUserProjectsBase := []projectName{TestProjectNamePublic}
		testUser := &User{
			Email: testUserEmailGoogleNormal,
		}
		if err := testUser.Reload(ctx); err != nil {
			if !errors.Is(err, mongo.ErrNoDocuments) {
				return errors.Join(errors.New("failed to find possibly existing user"), err)
			}
			// User does not exist, create base built-in perms
			testUser.Perms = UserPerms{
				Admin:    AcctTypeNormal(),
				Projects: testUserProjectsBase,
			}
		} else {
			allProjects := make([]projectName, 0, len(testUserProjectsBase)+len(testUser.Perms.Projects))
			for _, projGroup := range [][]projectName{testUserProjectsBase, testUser.Perms.Projects} {
				allProjects = append(allProjects, projGroup...)
			}
			testUser.Perms.Projects = utils.SetFrom(allProjects...).ToSlice() // TODO: ensure ok that we do not overwrite the Admin field
		}
		//err := coll.FindOne(ctx, BsonFindFilter(IDfld, testUserEmailGoogleNormal)).Decode(&testUser)
		//if err != nil {
		//	if !errors.Is(err, mongo.ErrNoDocuments) {
		//		return errors.Join(errors.New("failed to find possibly existing user"), err)
		//	}
		//	testUser.Perms = UserPerms{
		//		Admin:    AcctTypeNormal(),
		//		Projects: testUserProjectsFinal.ToSlice(),
		//	}
		//} else {
		//	testUserProjectsFinal.Add(testUser.Perms.Projects...)
		//	testUser.Perms.Projects = testUserProjectsFinal.ToSlice() // TODO: ensure ok that we do not overwrite the Admin field
		//}
		// Create user
		_, err := coll.ReplaceOne(ctx, BsonFindFilter(IDfld, testUser.Email), testUser, options.Replace().SetUpsert(true))
		if err != nil {
			return err
		}
		// TODO: DELETE THIS AFTER TESTING!!!!
		var testUsers []User

		println("Adding test users that are reset on every boot")

		for _, info := range []struct { // TODO: validate working properly!
			Projects []projectName
			Emails   []string
		}{
			{
				Projects: []projectName{TestProjectNamePublic},
				Emails: []string{
					testUserEmailPAA, testUserEmailPAB, testUserEmailPAC,
					testUserEmailPWA, testUserEmailPWB, testUserEmailPWC,
					testUserEmailPRA, testUserEmailPRB, testUserEmailPRC,
				},
			},
			{
				Projects: []projectName{},
				Emails: []string{
					testUserEmailPNA, testUserEmailPNB, testUserEmailPNC,
				},
			},
		} {
			for _, email := range info.Emails {
				testUsers = append(testUsers, User{
					Email: email,
					Perms: UserPerms{
						Admin:    AcctTypeNormal(),
						Projects: info.Projects,
					},
				})
			}
		}
		models := sliceutils.Map(testUsers, func(u User) mongo.WriteModel {
			return mongo.NewReplaceOneModel().
				SetFilter(bson.M{IDfld: u.Email}).
				SetReplacement(u).
				SetUpsert(true)
		})
		opts := options.BulkWrite().SetOrdered(false)
		_, err = coll.BulkWrite(ctx, models, opts)
		return err
	})
}

const (
	//testUserEmailAdmin = "reece.appling@gmail.com" // TODO: remove
	testUserEmailGoogleNormal = "nessapatch2408@gmail.com" // TODO: remove after testing!
	testUserEmailPAA          = "testProjAdminA@appli.ng"  // Admin on project
	testUserEmailPWA          = "testProjWriteA@appli.ng"  // Can write on project
	testUserEmailPRA          = "testProjReadA@appli.ng"   // Can read on project
	testUserEmailPNA          = "testProjNoneA@appli.ng"   // No project access
	testUserEmailPAB          = "testProjAdminB@appli.ng"  // Admin on project
	testUserEmailPWB          = "testProjWriteB@appli.ng"  // Can write on project
	testUserEmailPRB          = "testProjReadB@appli.ng"   // Can read on project
	testUserEmailPNB          = "testProjNoneB@appli.ng"   // No project access
	testUserEmailPAC          = "testProjAdminC@appli.ng"  // Admin on project
	testUserEmailPWC          = "testProjWriteC@appli.ng"  // Can write on project
	testUserEmailPRC          = "testProjReadC@appli.ng"   // Can read on project
	testUserEmailPNC          = "testProjNoneC@appli.ng"   // No project access
)

// TODO: update user!!!

type UserProjectPerm bool // Always referenced as a pointer, where true===admin, false===write, and nil===read
func UserProjectAdmin() *UserProjectPerm {
	return utils.Pointer(UserProjectPerm(true))
}
func UserProjectWrite() *UserProjectPerm {
	return utils.Pointer(UserProjectPerm(false))
}
func UserProjectRead() *UserProjectPerm {
	return nil
}
func (upp *UserProjectPerm) ProjectPerm() ProjectPerm {
	if upp == nil {
		return ProjectRead
	}
	if *upp {
		return ProjectAdmin
	}
	return ProjectWrite
}
func (upp *UserProjectPerm) RWPerm() *ReadWritePerm {
	if upp == nil {
		return nil // User can only read
	}
	out := ReadWritePerm(*upp) // User can write or admin
	return &out
}

func (u User) ResolvePerms(ctx context.Context) (ResolvedUserPerms, error) {
	acctType := u.Perms.Admin
	out := ResolvedUserPerms{
		Email:       u.Email,
		AccountType: acctType,
	}
	// If not regular user (is guest or admin), return early
	if !acctType.IsRegular() {
		return out, nil
	}

	// TODO: ensure all Projects are looped through? Or just the user's projects?

	// Resolve project perms // TODO: MAKE SURE THIS WORKS
	cursor, err := DbFrom(ctx).Collection(ProjectsCollectionName).
		Find(ctx, bson.M{IDfld: bson.M{"$in": u.Perms.Projects}}) // TODO: not sure I like this. Means that more projects will need to be on more users??
	if err != nil {
		return out, errors.Join(errors.New("failed to get cursor for UserPerms Projects"), err)
	}
	userProjPerms := map[projectName]*UserProjectPerm{}

	for cursor.Next(ctx) {
		var project Project
		if err = cursor.Decode(&project); err != nil {
			return out, errors.Join(errors.New("cursor decode error for UserPerms project"), err)
		}
		perm, exists := project.Perms[u.Email]
		if !exists {
			return out, errors.New("user not on project when they should have been")
		} else {
			userProjPerms[project.Name] = perm.UserProjectPerm()
		}
	}
	if err = cursor.Err(); err != nil {
		return out, errors.Join(errors.New("mongo cursor error after UserPerms project iteration"), err)
	}
	out.Projects = userProjPerms // TODO: ensure checking public projects as well???
	return out, nil
}
