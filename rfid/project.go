package rfid

import (
	"context"
	"encoding/json"
	"github.com/reeceappling/goUtils/v2/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
)

type projectName string

type Project struct {
	Name              projectName `bson:"_id" json:"_id"`
	CreationDateField `bson:"inline"`
	Completed         *unixTime `bson:"completed,omitempty" json:"completed,omitempty"` // TODO: index?
	NotesField        `bson:"inline"`
	LastUpdatedField  `bson:"inline"`
	Perms             ProjectPerms `bson:"perms" json:"perms"`
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
}

func (p Project) IdValue() any {
	return string(p.Name)
}

func (p Project) EntryTypeField() *string {
	return nil
}

func initializeProjects(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(ProjectsCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		newSimpleIndex("creationDate", "creationDate", true, false, false),
		//newSimpleIndex("completed", "creationDate", true, true, false),
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	for _, testItem := range []Project{
		{
			Name:              testProj,
			CreationDateField: CreationDateField{exampleTime},
			Completed:         &exampleTime,
			NotesField:        NotesField{exampleNotes()},
			LastUpdatedField:  LastUpdatedField{exampleTime},
			Perms:             exProjPerms,
		},
		{
			Name:              exProjWrite,
			CreationDateField: CreationDateField{exampleTime},
			Completed:         nil,
			NotesField:        NotesField{exampleNotes()},
			LastUpdatedField:  LastUpdatedField{exampleTime},
			Perms:             exProjPerms,
		}, {
			Name:              exProjRead,
			CreationDateField: CreationDateField{exampleTime},
			Completed:         nil,
			NotesField:        NotesField{exampleNotes()},
			LastUpdatedField:  LastUpdatedField{exampleTime},
			Perms:             exProjPerms,
		},
	} {
		// If test item does not exist or does not match, then create/update it
		err = coll.FindOneAndReplace(ctx, bson.D{{"_id", testItem.Name}}, testItem).Err()
		if err == nil {
			return err
		}
	}
	return nil
}

type createProjectRequest struct {
	NameField
	NotesField
}

func createProjectHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: ADD CREATING USER TO THE PROJECT AS THE INITIAL USER. HANDLE ALL PERMS
	// TODO: update sessions with perm updates?
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
	ctx, db := Db(r)
	coll := db.Collection(ProjectsCollectionName)
	user, err := GetAuthInfo(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// dont allow guests to create projects
	if user.isGuest() {
		http.Error(w, "guests cannot create new projects", http.StatusForbidden)
		return
	}

	now := unixTimeForNow()
	toInsert := Project{
		Name:              projectName(req.Name),
		CreationDateField: CreationDateField{now},
		NotesField:        req.NotesField,
		LastUpdatedField:  LastUpdatedField{now},
		Perms: ProjectPerms(map[string]*bool{
			user.Email: utils.Pointer(true),
		}),
	}
	finishCreateAlternateEntry(ctx, coll, toInsert, w)
	// TODO: add this project to the user session
}

type updateProjectRequest struct {
	Completed *unixTime        `json:"completed,omitempty"`
	Notes     AllEntries[Note] `json:"notes"`
	Perms     ProjectPerms     `json:"perms"`
	// TODO: update perms should update users too!
}

func (mods updateProjectRequest) modsFor(existing *Project) (bson.D, error) {
	return NewMods().
		updateProjectCompletedIfNeeded(mods.Completed, existing.Completed).
		updateNotesIfNeeded(mods.Notes, existing.Notes).
		updateProjectPermsIfNeeded(mods.Perms, existing.Perms).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateProjectHandler(w http.ResponseWriter, r *http.Request) {
	urlEncodedProjectName := r.PathValue("id") // TODO: USE!
	projNameStr, err := urlDecodeString(urlEncodedProjectName)
	if err != nil {
		http.Error(w, "bad project name in url", http.StatusBadRequest)
		return
	}
	projName := projectName(projNameStr)
	ctx := r.Context()
	user, err := GetAuthInfo(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Validate user perms for this project
	// TODO: validate current user can edit
	if projUserPerm := user.PermsForProject(projName); projUserPerm == nil || bool(*projUserPerm) != true {
		http.Error(w, "user is not project admin", http.StatusForbidden)
		return
	}
	// TODO: HANDLE CHANGES TO PERMS
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}
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
	for u, _ := range req.Perms {
		if _, exists := existing.Perms[u]; !exists {
			// validate new user exists
			result := db.Collection(UserCollName).FindOne(ctx, bson.D{{"_id", u}})
			if err = result.Err(); err != nil {
				dbErr(w, "user "+u+" not found", http.StatusNotFound)
				return
			}
		}
	}
	// TODO: add/remove users (on users)
	// Validate user is admin of project
	existingUserPerm := existing.Perms[user.Email]
	if existingUserPerm == nil || !*existingUserPerm {
		dbErr(w, "unauthorized to edit", http.StatusForbidden)
		return
	}

	// Create and write modifications
	upd, err := req.modsFor(&existing)
	handleUpdateMods(ctx, w, coll, existing, existing.DbId(), upd, err)
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
