package rfid

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type User struct {
	Email string    `bson:"_id" json:"_id"`                         // TODO: INDEX? MUST BE UNIQUE
	Perms UserPerms `bson:"perms,omitempty" json:"perms,omitempty"` // TODO: PROJECTS COME FROM PERMS
	// All can view?
}

func (u User) DbId() string {
	return u.Email
}

func (u User) IdValue() any {
	return u.Email
}

func initializeUsers(ctx context.Context, usern, unhashedPass string) error {
	//Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(UserCollName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		//newSimpleIndex("username", "username", false, false, true), // TODO: is true ok?
		//newSimpleIndex("email", "email", false, false, true),       // TODO: is true ok?
		//newSimpleIndex("googleId", "googleIde", false, true, true), // TODO: is true ok?
	})
	if err != nil {
		return err
	}
	return nil
}

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
		Find(ctx, bson.M{"_id": bson.M{"$in": u.Perms.Projects}})
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
			// TODO: ERROR
		} else {
			userProjPerms[project.Name] = perm
		}
	}
	if err = cursor.Err(); err != nil {
		return out, errors.Join(errors.New("mongo cursor error after UserPerms project iteration"), err)
	}
	out.projects = userProjPerms
	return out, nil
}
