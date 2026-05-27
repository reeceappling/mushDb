package rfid

import (
	"context"
	"errors"
	"github.com/reeceappling/goUtils/v2/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	Email string    `bson:"_id" json:"_id"`
	Perms UserPerms `bson:"perms,omitempty" json:"perms,omitempty"` // TODO: PROJECTS COME FROM PERMS
	// All can view?
}

func (u User) DbId() string {
	return u.Email
}

func (u User) IdValue() any {
	return u.Email
}

func initializeUsers(ctx context.Context) error {
	//Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(UserCollName)
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
			Admin:    utils.Pointer(false),
			Projects: []projectName{testProjects[0].Name, testProjects[1].Name, testProjects[2].Name},
		},
	}
	_, err := coll.ReplaceOne(ctx, bsonFindFilter("_id", testUser.Email), testUser, options.Replace().SetUpsert(true))
	if err != nil {
		return err
	}
	// TODO: DELETE THIS AFTER TESTING!!!!
	testUserSelf := User{
		Email: testUserEmailSelf,
		Perms: UserPerms{
			Admin:    utils.Pointer(false),
			Projects: []projectName{testProjects[0].Name, testProjects[1].Name, testProjects[2].Name},
		},
	}
	_, err = coll.ReplaceOne(ctx, bsonFindFilter("_id", testUserSelf.Email), testUserSelf, options.Replace().SetUpsert(true))
	return err
}

const testUserEmailSelf = "reeceappling@gmail.com" // TODO: or dot?
const testUserEmail = "nessapatch2408@gmail.com"

func (u User) ResolvePerms(ctx context.Context) (ResolvedUserPerms, error) {
	userIsAdmin := u.Perms.Admin
	out := ResolvedUserPerms{
		Email: u.Email,
		admin: userIsAdmin,
	}
	// If guest or Admin, return early
	if u.Perms.Admin == nil {
		println("user is guest") // TODO; del
		return out, nil
	}
	if *u.Perms.Admin {
		println("user is Admin") // TODO; del
		return out, nil
	}

	// TODO: ensure all Projects are looped through

	// Resolve project perms // TODO: MAKE SURE THIS WORKS
	cursor, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(ProjectsCollectionName).
		Find(ctx, bson.M{"_id": bson.M{"$in": u.Perms.Projects}}) // TODO: not sure I like this. Means that more projects will need to be on more users??
	if err != nil {
		return out, errors.Join(errors.New("failed to get cursor for UserPerms Projects"), err)
	}
	userProjPerms := map[projectName]*bool{}

	for cursor.Next(context.TODO()) {
		var project Project
		if err = cursor.Decode(&project); err != nil {
			return out, errors.Join(errors.New("cursor decode error for UserPerms project"), err)
		}
		perm, exists := project.Perms[u.Email]
		if !exists {
			return out, errors.New("user not on project")
		} else {
			switch perm {
			case "admin":
				userProjPerms[project.Name] = utils.Pointer(true)
			case "write":
				userProjPerms[project.Name] = utils.Pointer(false)
			case "read":
				userProjPerms[project.Name] = nil
			default:
				return out, errors.New("invalid user perm on project: " + perm)
			}
		}
	}
	if err = cursor.Err(); err != nil {
		return out, errors.Join(errors.New("mongo cursor error after UserPerms project iteration"), err)
	}
	out.projects = userProjPerms // TODO: ensure checking public projects as well???
	return out, nil
}
