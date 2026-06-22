package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/reeceappling/goUtils/v2/utils"
	sliceutils "github.com/reeceappling/goUtils/v2/utils/slices"
	"github.com/reeceappling/mushDb/api/request"
	"github.com/reeceappling/mushDb/api/request/unix"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"golang.org/x/exp/maps"
	"io"
	"net/http"
)

// TODO: needed for: ???????????????

type projectName string

type Project struct {
	Name              projectName `bson:"_id" json:"_id"`
	CreationDateField `bson:"inline"`
	Completed         *unix.Time `bson:"completed,omitempty" json:"completed,omitempty"`
	NotesField        `bson:"inline"`
	LastUpdatedField  `bson:"inline"`
	Perms             ProjectPerms `bson:"perms" json:"perms"` // Map of email of user to permission on project
	// TODO: make it so we can add/remove users from Projects (maybe one at a time?)
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
	coll := DbFrom(ctx).Collection(ProjectsCollectionName)
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
		_, errRep := coll.ReplaceOne(ctx, BsonFindFilter("_id", testItem.Name), testItem, options.Replace().SetUpsert(true))
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
		var keepFilter func(*Project) bool
		if *complete {
			keepFilter = func(pr *Project) bool {
				return pr.Completed != nil
			}
		} else {
			keepFilter = func(pr *Project) bool {
				return pr.Completed == nil
			}
		}
		projs = sliceutils.FilterInPlace(projs, keepFilter)
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
			newNote(exampleTime, "test user should be admin. Admin user is admin"),
		}},
		LastUpdatedField: LastUpdatedField{exampleTime},
		Perms: map[string]ProjectPerm{
			testUserEmail:     ProjectAdmin,
			testUserEmailSelf: ProjectAdmin,
		},
	}, {
		Name:              "testProjectWrite",
		CreationDateField: CreationDateField{exampleTime},
		Completed:         &exampleTime,
		NotesField: NotesField{Notes: []Note{
			newNote(exampleTime, "test user should be able to write but not admin. Admin user is admin"),
		}},
		LastUpdatedField: LastUpdatedField{exampleTime},
		Perms: map[string]ProjectPerm{
			testUserEmail:     ProjectWrite,
			testUserEmailSelf: ProjectAdmin,
		},
	}, {
		Name:              "testProjectRead",
		CreationDateField: CreationDateField{exampleTime},
		Completed:         nil,
		NotesField: NotesField{Notes: []Note{
			newNote(exampleTime, "test user should be able to read. Admin user is admin"),
		}},
		LastUpdatedField: LastUpdatedField{exampleTime},
		Perms: map[string]ProjectPerm{
			testUserEmail:     ProjectRead, // TODO: ensure to remove dots and plusses from (pre-@) emails? Maybe keep them because that's nicer as a service provider :)
			testUserEmailSelf: ProjectAdmin,
		},
	}, {
		Name:              "testProjectNone",
		CreationDateField: CreationDateField{exampleTime},
		Completed:         nil,
		NotesField: NotesField{Notes: []Note{
			newNote(exampleTime, "test user should not be able to do anything. Admin user is admin"),
		}},
		LastUpdatedField: LastUpdatedField{exampleTime},
		Perms: map[string]ProjectPerm{
			testUserEmailSelf: ProjectAdmin, // This is self and not test user because I want only my main email to be admin
		},
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

	ctx, now := request.UnixTime(r.Context()) // TODO: no more r.Context below
	toInsert := Project{
		Name:              projectName(req.Name),
		CreationDateField: CreationDateField{now},
		NotesField:        req.NotesField,
		LastUpdatedField:  LastUpdatedField{now},
		Perms: ProjectPerms(map[string]ProjectPerm{
			GetUserEmail(ctx): ProjectAdmin,
		}),
	}
	// TODO: try to add project to user!
	finishCreateProject(ctx, toInsert, w, func() error {
		// TODO: add project to the user session, both in db and mem
		return nil
	}) // TODO: handle create project that is different
}

type updateProjectRequest struct {
	Completed *unix.Time `json:"completed,omitempty"`
	NotesUpdateField
	Perms ProjectPerms `json:"perms"`
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
	if !user.IsAdmin() && (existingUserPerm != "admin") {
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
			result := db.Collection(UserCollName).FindOne(ctx, BsonFindFilter("_id", u))
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
		authSvc := GetAuthService(sessCtx)
		authSvc.Lock()
		defer authSvc.Unlock()
		for u, _ := range usersWithProjectRemoved {
			// remove project from each user that no longer has the project // TODO: VALIDATE WORKING PROPERLY
			_, e := txColl.UpdateByID(sessCtx, u, bson.D{{"$unset", bson.D{{permKey, ""}}}}) // TODO: ensure ok
			if e != nil {
				return nil, e
			}
			// remove the project from the user in the session stuff!
			if sessId, exists := authSvc.UserSessionMap[u]; exists {
				delete(authSvc.sessMap[sessId].Data.Projects, projName)
			}
		}
		// Add project with perm to user, or change the project perm  // TODO: VALIDATE WORKING PROPERLY
		for _, subset := range []map[string]ProjectPerm{usersWithProjectChanged, usersWithProjectAdded} {
			for u, userPerm := range subset {
				_, e := txColl.UpdateByID(sessCtx, u, bson.D{{"$set", bson.D{{permKey, userPerm}}}}) // TODO: ensure ok
				if e != nil {
					return nil, e
				}
				// TODO: add the project to the user in the session stuff!
				if sessId, exists := authSvc.UserSessionMap[u]; exists {
					sess := authSvc.sessMap[sessId]
					tempProjects := sess.Data.Projects
					if tempProjects == nil {
						tempProjects = map[projectName]*UserProjectPerm{}
					}
					tempProjects[projName] = userPerm.UserProjectPerm()
					sess.Data.Projects = tempProjects
					authSvc.sessMap[sessId] = sess
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
//		cursor, err := DbFrom(ctx).Collection(ProjectsCollectionName).
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
//		cursor, err := DbFrom(ctx).Collection(ProjectsCollectionName).
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

// TODO: MOVE!
func handleUpdateProject(ctx context.Context, w http.ResponseWriter, existing Project, upd bson.D, err error, updateUsers func(mongo.SessionContext) (any, error)) {
	if err != nil {
		println("mod creation failure: " + err.Error())
		dbErr(w, "error creating txn:"+err.Error(), http.StatusInternalServerError)
		return
	}
	if len(upd) == 0 {
		dbErr(w, "no changes made", http.StatusBadRequest)
		return
	}

	sessionOptions := options.Session() // TODO: change?
	sess, err := GetMongoClient(ctx).StartSession(sessionOptions)
	if err != nil {
		http.Error(w, "failed to start mongo session: "+err.Error(), http.StatusInternalServerError)
		return
	}
	wc := writeconcern.Majority() // TODO: ok?
	txnOptions := options.Transaction().SetWriteConcern(wc)
	// Defers ending the session after the transaction is committed or ended
	_, err = sess.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		defer sess.EndSession(ctx)
		// update the users (if needed)
		if _, e := updateUsers(sessCtx); e != nil {
			errTxn := errors.Join(e, sess.AbortTransaction(ctx))
			http.Error(w, "failed to update users: "+errTxn.Error(), http.StatusInternalServerError)
			return nil, errTxn
		}
		// Update the project
		coll := mongo.SessionFromContext(sessCtx).Client().Database(dbName).Collection(ProjectsCollectionName)
		bsonId := BsonFindFilter("_id", existing.DbId())
		err = coll.FindOneAndUpdate(ctx, bsonId, upd).Err()
		if err != nil {
			dbErr(w, "failed to write update to db: "+err.Error(), http.StatusInternalServerError)
			return nil, err
		}
		var updated Project
		err = coll.FindOne(ctx, bsonId).Decode(&updated)
		if err != nil {
			dbErr(w, "failed to write update to db: "+err.Error(), http.StatusInternalServerError)
			return nil, err
		}
		bsOut, err := json.Marshal(updated)
		if err != nil {
			dbErr(w, err.Error(), http.StatusInternalServerError)
			return nil, err
		}

		// Try to commit the txn
		if e := sess.CommitTransaction(ctx); e != nil {
			errTxn := errors.Join(e, sess.AbortTransaction(ctx))
			http.Error(w, "failed to commit: "+errTxn.Error(), http.StatusInternalServerError)
			return nil, errTxn
		}
		// TODO: move the write!
		bsOut2, err2 := json.MarshalIndent(updated, "", " ") // TODO: delete later
		if err2 != nil {
			dbErr(w, err2.Error(), http.StatusInternalServerError)
			return nil, err
		}
		println("Writing update:", string(bsOut2)) // TODO: del
		_, err = w.Write(bsOut)
		handleWriteErr(err, w)

		return nil, nil
	}, txnOptions)
	return
}

func finishCreateProject(ctx context.Context, toInsert CollectionItem, w http.ResponseWriter, inTxn func() error) {
	sessionOptions := options.Session() // TODO: change?
	sess, err := GetMongoClient(ctx).StartSession(sessionOptions)
	if err != nil {
		http.Error(w, "failed to start mongo session: "+err.Error(), http.StatusInternalServerError)
		return
	}
	wc := writeconcern.Majority() // TODO: ok?
	txnOptions := options.Transaction().SetWriteConcern(wc)
	// Defers ending the session after the transaction is committed or ended
	_, err = sess.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		defer sess.EndSession(ctx)
		// do the insert
		_, err := mongo.SessionFromContext(sessCtx).Client().
			Database(dbName).Collection(toInsert.CollectionName()).InsertOne(ctx, toInsert)
		if err != nil {
			http.Error(w, "failed to insert one: "+err.Error(), http.StatusInternalServerError)
			return nil, errors.Join(err, ErrTxnWriteFail) // TODO: ok?
		}
		// do the thing needed to be successful // TODO: UPDATE THE USER!
		if e := inTxn(); e != nil {
			errTxn := errors.Join(e, sess.AbortTransaction(ctx))
			http.Error(w, "failed to do post-insert call: "+errTxn.Error(), http.StatusInternalServerError)
			return nil, errTxn
		}
		if e := sess.CommitTransaction(ctx); e != nil {
			errTxn := errors.Join(e, sess.AbortTransaction(ctx))
			http.Error(w, "failed to commit: "+errTxn.Error(), http.StatusInternalServerError)
			return nil, errTxn
		}
		return nil, nil
	}, txnOptions)
	if err != nil {
		return
	}
	bsOut, err := json.Marshal(toInsert)
	if err != nil {
		http.Error(w, "failed to marshal: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(bsOut)
	if err != nil {
		handleWriteErr(err, w)
		return
	}
}
