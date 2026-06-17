package api

import (
	"context"
	"errors"
	"github.com/reeceappling/goUtils/v2/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//var _ webauthn.User = &User{}

type User struct {
	Email string `bson:"_id" json:"_id"`
	// TODO: can we make UserPerms.Admin not a pointer?
	Perms UserPerms `bson:"perms,omitempty" json:"perms,omitempty"` // TODO: PROJECTS COME FROM PERMS! So even projects with only read perms can be associated with an item!
	// All can view?
}

func (u User) DbId() string {
	return u.Email
}

func (u User) IdValue() any {
	return u.Email
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
	//err := createIndexes(ctx, coll, []mongo.IndexModel{
	//	//newSimpleIndex("username", "username", false, false, true), // TODO: is true ok?
	//	//newSimpleIndex("email", "email", false, false, true),       // TODO: is true ok?
	//	//newSimpleIndex("googleId", "googleIde", false, true, true), // TODO: is true ok?
	//})
	//if err != nil {
	//	return err
	//}
	// TODO: DELETE THIS AFTER TESTING!!!!
	testUser := User{
		Email: testUserEmail,
		Perms: UserPerms{
			Admin:    AcctTypeNormal(),
			Projects: []projectName{testProjects[0].Name, testProjects[1].Name, testProjects[2].Name},
		},
	}
	_, err := coll.ReplaceOne(ctx, BsonFindFilter("_id", testUser.Email), testUser, options.Replace().SetUpsert(true))
	if err != nil {
		return err
	}
	// TODO: DELETE THIS AFTER TESTING!!!!
	testUserSelf := User{
		Email: testUserEmailSelf,
		Perms: UserPerms{
			Admin:    AcctTypeNormal(),
			Projects: []projectName{testProjects[0].Name, testProjects[1].Name, testProjects[2].Name},
		},
	}
	_, err = coll.ReplaceOne(ctx, BsonFindFilter("_id", testUserSelf.Email), testUserSelf, options.Replace().SetUpsert(true))
	return err
}

const testUserEmailSelf = "reeceappling@gmail.com" // TODO: or dot?
const testUserEmail = "nessapatch2408@gmail.com"

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
		accountType: acctType,
	}
	// If not regular user (is guest or admin), return early
	if !acctType.IsRegular() {
		println("user is guest or admin") // TODO; del
		return out, nil
	}

	// TODO: ensure all Projects are looped through? Or just the user's projects?

	// Resolve project perms // TODO: MAKE SURE THIS WORKS
	cursor, err := DbFrom(ctx).Collection(ProjectsCollectionName).
		Find(ctx, bson.M{"_id": bson.M{"$in": u.Perms.Projects}}) // TODO: not sure I like this. Means that more projects will need to be on more users??
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
			return out, errors.New("user not on project when they should have been!")
		} else {
			userProjPerms[project.Name] = perm.UserProjectPerm()
		}
	}
	if err = cursor.Err(); err != nil {
		return out, errors.Join(errors.New("mongo cursor error after UserPerms project iteration"), err)
	}
	out.projects = userProjPerms // TODO: ensure checking public projects as well???
	return out, nil
}
