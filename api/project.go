package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/reeceappling/goUtils/v2/utils"
	sliceutils "github.com/reeceappling/goUtils/v2/utils/slices"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/exp/maps"
	"io"
	"net/http"
)

// TODO: needed for: ???????????????

type projectName string

type Project struct {
	Name              projectName `bson:"_id" json:"_id"`
	CreationDateField `bson:"inline"`
	Completed         *UnixTime `bson:"completed,omitempty" json:"completed,omitempty"` // TODO: index?
	NotesField        `bson:"inline"`
	LastUpdatedField  `bson:"inline"`
	Perms             ProjectPerms `bson:"perms" json:"perms"` // Map of email of user to permission on project
	// TODO: make it so we can add/remove users from Projects
}

func (p Project) DbId() string {
	return string(p.Name)
}

func (p Project) AddUser(u User, perm *ReadWritePerm) string {
	// TODO: this!!!!
	// TODO: update email entry
	// TODO: update email session?
	panic("implement me")
} // TODO: impl!?

func (p Project) IdValue() any {
	return string(p.Name)
}

func initializeProjects(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(ProjectsCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		newSimpleIndex("creationDate", "creationDate", true, false, false),
		//newSimpleIndex("completed", "creationDate", true, true, false), // TODO: ???
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	for _, testItem := range testProjects {
		// If test item does not exist or does not match, then create/update it
		_, errRep := coll.ReplaceOne(ctx, bsonFindFilter("_id", testItem.Name), testItem, options.Replace().SetUpsert(true))
		err = errors.Join(errRep, err)
		if err != nil {
		}
	}
	return err
}

func GetAllProjects(ctx context.Context, complete *bool) ([]Project, error) {
	projs, err := getAllEntries[*Project](ctx, &Project{})
	if err != nil {
		return nil, err
	}
	if complete != nil {
		if *complete {
			projs = sliceutils.FilterInPlace(projs, func(pr *Project) bool {
				return pr.Completed != nil
			})
		} else {
			projs = sliceutils.FilterInPlace(projs, func(pr *Project) bool {
				return pr.Completed == nil
			})
		}
	}

	return sliceutils.Map(projs, func(pr *Project) Project {
		return *pr
	}), nil
}

var testProjects = []Project{
	{
		Name:              "testProjectAdmin",
		CreationDateField: CreationDateField{exampleTime},
		Completed:         nil,
		NotesField: NotesField{Notes: []Note{
			newNote(exampleTime, "test user should be admin"),
		}},
		LastUpdatedField: LastUpdatedField{exampleTime},
		Perms: map[string]ProjectPerm{
			testUserEmail: ProjectAdmin,
		},
	}, {
		Name:              "testProjectWrite",
		CreationDateField: CreationDateField{exampleTime},
		Completed:         &exampleTime,
		NotesField: NotesField{Notes: []Note{
			newNote(exampleTime, "test user should be able to write but not admin"),
		}},
		LastUpdatedField: LastUpdatedField{exampleTime},
		Perms: map[string]ProjectPerm{
			testUserEmail: ProjectWrite,
		},
	}, {
		Name:              "testProjectRead",
		CreationDateField: CreationDateField{exampleTime},
		Completed:         nil,
		NotesField: NotesField{Notes: []Note{
			newNote(exampleTime, "test user should be able to read"),
		}},
		LastUpdatedField: LastUpdatedField{exampleTime},
		Perms: map[string]ProjectPerm{
			testUserEmail: ProjectRead,
		},
	}, {
		Name:              "testProjectNone",
		CreationDateField: CreationDateField{exampleTime},
		Completed:         nil,
		NotesField: NotesField{Notes: []Note{
			newNote(exampleTime, "test user should not be able to do anything"),
		}},
		LastUpdatedField: LastUpdatedField{exampleTime},
		Perms:            nil,
	},
}

type createProjectRequest struct {
	NameField
	NotesField
}

func createProjectHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: update sessions with perm updates? // TODO: THIS!
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req := createProjectRequest{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx, _ := Db(r)
	user, err := GetAuthInfo(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// dont allow guests to create projects
	if user.isGuest() { // TODO: is this needed?
		http.Error(w, "guests cannot create new projects", http.StatusForbidden)
		return
	}

	now := unixTimeForNow()
	toInsert := Project{
		Name:              projectName(req.Name),
		CreationDateField: CreationDateField{now},
		NotesField:        req.NotesField,
		LastUpdatedField:  LastUpdatedField{now},
		Perms: ProjectPerms(map[string]ProjectPerm{
			user.Email: ProjectAdmin,
		}),
	}
	// TODO: try to add project to user!
	finishCreateProject(ctx, toInsert, w, func() error {
		// TODO: add project to the user session...
		return nil
	})
}

type updateProjectRequest struct {
	Completed *UnixTime `json:"completed,omitempty"`
	NotesUpdateField
	Perms ProjectPerms `json:"perms"`
	// TODO: update perms should update users too!
}

func (req updateProjectRequest) modsFor(existing *Project) (bson.D, error) {
	return NewMods().
		updateProjectCompletedIfNeeded(req.Completed, existing.Completed).
		updateNotesIfNeeded(req, existing).
		updateProjectPermsIfNeeded(req.Perms, existing.Perms).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateProjectHandler(w http.ResponseWriter, r *http.Request) {
	urlEncodedProjectName := r.PathValue("id") // TODO: NOT FINDING PROJECT!
	println("project update url used: " + r.URL.String())
	projNameStr, err := UrlDecodeString(urlEncodedProjectName)
	if err != nil {
		http.Error(w, "bad project name in url", http.StatusBadRequest)
		return
	}

	println("decoded project name", projNameStr)
	projName := projectName(projNameStr)
	ctx := r.Context()
	user, err := GetAuthInfo(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// validate current user can edit (is admin of service or admin on project)
	if projUserPerm := user.PermsForProject(projName); projUserPerm == nil || bool(*projUserPerm) != true {
		http.Error(w, "user is not project admin", http.StatusForbidden)
		return
	}
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	println("RECEIVED: ", string(bs))
	req := updateProjectRequest{}
	err = json.Unmarshal(bs, &req)
	if err != nil {
		http.Error(w, "failed to unmarshal body: "+err.Error(), http.StatusBadRequest)
		return
	}
	ctx, db := Db(r)
	coll := db.Collection(ProjectsCollectionName)
	existing := Project{}
	err = coll.FindOne(ctx, bson.M{"_id": projName}).Decode(&existing)
	if err != nil {
		stat := http.StatusInternalServerError
		if err == mongo.ErrNoDocuments {
			stat = http.StatusNotFound
		}
		dbErr(w, err.Error(), stat)
		return
	}
	// Validate user perms for this project
	// Validate user is admin of project
	existingUserPerm := existing.Perms[user.Email]
	if !user.isAdmin() && (existingUserPerm != "admin") {
		dbErr(w, "unauthorized to edit", http.StatusForbidden)
		return
	}
	// Validate any changes, ensure new users exist, and sort perms changes into groups
	usersWithProjectRemoved := utils.SetOf(maps.Keys(existing.Perms))
	usersWithProjectChanged := map[string]ProjectPerm{}
	usersWithProjectAdded := map[string]ProjectPerm{}
	for u, futurePerm := range req.Perms {
		usersWithProjectRemoved.Remove(u)
		existingPerm, exists := existing.Perms[u]
		if !exists {
			usersWithProjectAdded[u] = futurePerm
			// validate new user exists
			result := db.Collection(UserCollName).FindOne(ctx, bsonFindFilter("_id", u))
			if err = result.Err(); err != nil {
				dbErr(w, "user "+u+" does not exist. Invalid request", http.StatusBadRequest)
				return
			}
		} else {
			if futurePerm != existingPerm {
				usersWithProjectChanged[u] = futurePerm
			}
		}
	}

	updateUsers := func(sessCtx mongo.SessionContext) (any, error) {
		permKey := "perms." + string(projName)
		txColl := mongo.SessionFromContext(sessCtx).
			Client().Database(dbName).
			Collection(UserCollName)
		for u, _ := range usersWithProjectRemoved {
			// remove project from each user that no longer has the project // TODO: VALIDATE WORKING PROPERLY
			_, e := txColl.UpdateByID(sessCtx, u, bson.D{{"$unset", bson.D{{permKey, ""}}}}) // TODO: ensure ok
			if e != nil {
				return nil, e
			}
		}
		// Add project with perm to user, or change the project perm  // TODO: VALIDATE WORKING PROPERLY
		for _, subset := range []map[string]ProjectPerm{usersWithProjectChanged, usersWithProjectAdded} {
			for u, userPerm := range subset {
				_, e := txColl.UpdateByID(sessCtx, u, bson.D{{"$set", bson.D{{permKey, userPerm}}}}) // TODO: ensure ok
				if e != nil {
					return nil, e
				}
			}
		}
		return nil, nil
	}

	// Create and write modifications
	upd, err := req.modsFor(&existing)
	handleUpdateProject(ctx, w, existing, upd, err, updateUsers)
}

//func allUnfinishedProjectsForUser(ctx context.Context, auth AuthInfo) ([]ProjectWithPerm, error) {
//	out := []ProjectWithPerm{}
//	if auth.Opts != nil {
//		filter := bson.M{} // TODO: make sure this works
//		if !auth.isAdmin() {
//			filter = bson.M{"_id": bson.M{"$in": maps.Keys(auth.Opts.Projects)}}
//		}
//		cursor, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(ProjectsCollectionName).
//			Find(ctx, filter) // TODO: ok?
//		if err != nil {
//			return nil, errors.Join(errors.New("failed to get cursor for UserPerms Projects"), err)
//		}
//		for cursor.Next(ctx) {
//			var project Project
//			if err = cursor.Decode(&project); err != nil {
//				return nil, errors.Join(errors.New("cursor decode error for UserPerms project"), err)
//			}
//			if project.Completed != nil {
//				continue
//			}
//			out = append(out, ProjectWithPerm{
//				ProjectName: project.Name,
//				CanWrite:    utils.Pointer((*auth.Opts).Projects[project.Name]),
//			})
//		}
//		if err = cursor.Err(); err != nil {
//			return nil, errors.Join(errors.New("mongo cursor error after UserPerms project iteration"), err)
//		}
//	}
//	return out, nil
//}
//
//// TODO: USE THIS
//func listProjectsForUser(ctx context.Context, unfinishedOnly bool, writeOnly bool, maxProjects *int) ([]ProjectWithPerm, error) {
//	auth, err := GetAuthInfo(ctx)
//	if err != nil {
//		return nil, err
//	}
//	var out []ProjectWithPerm
//	if maxProjects == nil {
//		out = []ProjectWithPerm{}
//	} else {
//		out = make([]ProjectWithPerm, 0, *maxProjects)
//	}
//	if auth.Opts != nil {
//		filter := bson.M{} // TODO: make sure this works, will bson.D be better?
//		if unfinishedOnly {
//			filter["completed"] = bson.M{"$exists": false}
//		}
//		cursor, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(ProjectsCollectionName).
//			Find(ctx, filter) // TODO: ok?
//		if err != nil {
//			return nil, errors.Join(errors.New("failed to get cursor for UserPerms Projects"), err)
//		}
//		for cursor.Next(ctx) {
//			if maxProjects != nil && len(out) == *maxProjects {
//				break
//			}
//			var project Project
//			if err = cursor.Decode(&project); err != nil {
//				return nil, errors.Join(errors.New("cursor decode error for UserPerms project"), err)
//			}
//			if project.Perms.Blanket == perms.Write || auth.isAdmin() {
//				out = append(out, ProjectWithPerm{
//					ProjectName: project.Name,
//					CanWrite:    utils.Pointer(true),
//				})
//				continue
//			}
//			userIndex := slices.IndexFunc(project.Perms.Users.Ids, func(idPair ProjectPermUserId) bool {
//				return idPair.Email.String() == auth.Email.String()
//			})
//			if userIndex == -1 {
//				if project.Perms.Blanket == perms.Read && !writeOnly {
//					out = append(out, ProjectWithPerm{
//						ProjectName: project.Name,
//						CanWrite:    utils.Pointer(false),
//					})
//				}
//				continue
//			}
//			canWrite := project.Perms.Users.CanWrite[userIndex]
//			if !canWrite && writeOnly {
//				continue
//			}
//			out = append(out, ProjectWithPerm{
//				ProjectName: project.Name,
//				CanWrite:    utils.Pointer(project.Perms.Users.CanWrite[userIndex]),
//			})
//		}
//		if err = cursor.Err(); err != nil {
//			return nil, errors.Join(errors.New("mongo cursor error after UserPerms project iteration"), err)
//		}
//	}
//	return out, nil
//}
