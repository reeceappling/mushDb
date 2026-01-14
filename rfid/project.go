package rfid

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
)

const projectsCollectionName = "Projects"

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

func (p Project) AddUser(u User, perm ReadWritePerm) string {
	// TODO: this!!!!
	// TODO: update email entry
	// TODO: update email session?
	panic("implement me")
}

func (p Project) StringId() string {
	return string(p.Name)
}

func (p Project) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := p
	err := decodeItem(&out, encoded)
	return out, err
}

func (p Project) CollectionName() string {
	return projectsCollectionName
}

func (p Project) EntryTypeField() *string {
	return nil
}

func initializeProjects(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(projectsCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		newSimpleIndex("creationDate", "creationDate", true, false, false),
		newSimpleIndex("completed", "creationDate", true, true, false),
		lastUpdatedIndexModel,
		// TODO: Perms?
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
	// TODO: PERMS!
	//Perms ProjectPerms `json:"perms"` // TODO: validate
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
	//authinfo, err := GetAuthInfo(r.Context())
	//if err != nil {
	//	http.Error(w, "failed to get auth info: "+err.Error(), http.StatusUnauthorized)
	//	return
	//}
	//if req.Perms.Blanket != perms.Write {
	//	authorExistsInPerms := false
	//	for _, email := range req.Perms.Users.Ids {
	//		// TODO: validate that each userId exists in the db
	//		if email.Email.String() == authinfo.Email.String() {
	//			authorExistsInPerms = true
	//			break
	//		}
	//	}
	//	if !authorExistsInPerms {
	//		req.Perms.Users = req.Perms.Users.WithAuthor(ProjectPermUserId{
	//			Email:  authinfo.Email,
	//			Val: "FIXME!", // TODO: this
	//		})
	//	}
	//}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(projectsCollectionName)
		now := unixTimeForNow()
		// TODO: validate all users exist
		toInsert := Project{
			Name:              projectName(req.Name),
			CreationDateField: CreationDateField{now},
			NotesField:        req.NotesField,
			LastUpdatedField:  LastUpdatedField{now},
		}
		_, err = coll.InsertOne(r.Context(), toInsert)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		bsOut, err := json.Marshal(toInsert)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		// TODO: add this project to all email sessions that need it (only if non-blanket-write)
		//
		return w.Write(bsOut)
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}

type updateProjectRequest struct {
	NameField
	Completed *unixTime        `json:"completed,omitempty"`
	Notes     AllEntries[Note] `json:"notes"`
	//Perms     ProjectPerms     `json:"perms"`
	// TODO: PERMS!
	// TODO: update perms should update users too!
}

func updateProjectHandler(w http.ResponseWriter, r *http.Request) {
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
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(projectsCollectionName)
		existing := Project{}
		err = coll.FindOne(ctx, bson.M{"_id": req.Name}).Decode(&existing)
		if err != nil {
			stat := http.StatusInternalServerError
			if err == mongo.ErrNoDocuments {
				stat = http.StatusNotFound
			}
			return DbTxnStdErr(w, err.Error(), stat)
		}
		mods := NewMods().
			// UpdatePointerIfNeeded("completed", req.Completed, existing.Completed). // TODO: FIX
			updateNotesIfNeeded(req.Notes, existing.Notes)
		// TODO: ENSURE THIS USER CAN WRITE TO THIS PROJECT

		//newUsers := map[ProjectPermUserId]bool{} // TODO brand new email to this project. Handle
		//removedUsers := []ProjectPermUserId{}    // TODO: this email has had their perms revoked. Handle
		//promotedPerms := []ProjectPermUserId{}   // TODO email can now write. Handle
		//demotedPerms := []ProjectPermUserId{}    // TODO: email can no longer write, handle
		//
		//// TODO: simplify req perms if needed
		//if req.Perms.Blanket != existing.Perms.Blanket {
		//	if req.Perms.Blanket == perms.Write {
		//		mods.Unset("perms")
		//		// TODO: remove project from users and email sessions as needed
		//	} else {
		//		mods = mods.Set("perms", req.Perms)
		//	}
		//} else {
		//	// TODO: check all users in projectPerms for changes (and make sure they exist)
		//	tempCanWrite := map[Base58Str]bool{}
		//	tempName := map[Base58Str]string{}
		//	var existsAlready = false
		//	for i, userIds := range req.Perms.Users.Ids {
		//		userIdStr := userIds.Email.asBase58()
		//		_, existsAlready = tempCanWrite[userIdStr]
		//		if existsAlready {
		//			return DbTxnStdErr(w, fmt.Sprintf(`userId %d already exists in request`, i), http.StatusBadRequest)
		//		}
		//		tempCanWrite[userIdStr] = req.Perms.Users.CanWrite[i]
		//		tempName[userIdStr] = userIds.Val
		//
		//	}
		//
		//	for i, current := range existing.Perms.Users.Ids {
		//		id := current.Email.asBase58()
		//		newPerm, exists := tempCanWrite[]
		//		if !exists {
		//			removedUsers = append(removedUsers, current)
		//		} else {
		//			if newPerm != existing.Perms.Users.CanWrite[i] {
		//				if newPerm {
		//					promotedPerms = append(promotedPerms, current)
		//				} else {
		//					demotedPerms = append(demotedPerms, current)
		//				}
		//			}
		//		}
		//		delete(tempCanWrite, id)
		//	}
		//	for uidBase58, canWrite := range tempCanWrite {
		//		id, err := uidBase58.toAltCollectionId()
		//		if err != nil {
		//			panic("SHOULD NEVER HAPPEN: " + err.Error()) // TODO: this
		//		}
		//		newUsers[ProjectPermUserId{
		//			Email:  id,
		//			Val: tempName[uidBase58],
		//		}] = canWrite
		//	}
		//	if len(newUsers) > 0 {
		//		// TODO: validate users exist
		//	}
		//
		//	if len(newUsers) > 0 || len(removedUsers) > 0 || len(promotedPerms) > 0 || len(demotedPerms) > 0 {
		//		mods = mods.Set("perms", req.Perms)
		//	}
		//	// TODO: modify users and email sessions if the perms changed
		//
		//	// TODO: MODIFY DB,
		//}
		upd, err := mods.Finalized()
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError) // TODO: ok?
		}
		if len(upd) == 0 {
			return DbTxnStdErr(w, "no changes made", http.StatusBadRequest)
		}
		bsonId := bson.D{{"_id", existing.Name}}
		err = coll.FindOneAndUpdate(ctx, bsonId, upd).Err()
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		err = coll.FindOne(ctx, bsonId).Decode(&existing)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		bsOut, err := json.Marshal(existing)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		//for email, canWrite := range newUsers { //:= map[ProjectPermUserId]bool{} // TODO brand new email to this project. Handle
		//	// TODO: add project to email in db
		//	// TODO: add to email session perms (if email has a session)
		//}
		//for _, email := range removedUsers { //:= []ProjectPermUserId{}    // TODO: this email has had their perms revoked. Handle
		//	// TODO: remove project from email in db if can no-longer read
		//	// TODO: change email session perms
		//}
		//for _, email := range promotedPerms { //:= []ProjectPermUserId{}   // TODO email can now write. Handle
		//	// TODO: change email session perms
		//}
		//for _, email := range demotedPerms { //:= []ProjectPermUserId{}    // TODO: email can no longer write, handle
		//	// TODO: change email session perms
		//}
		return w.Write(bsOut)
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}

//func allUnfinishedProjectsForUser(ctx context.Context, auth AuthInfo) ([]ProjectWithPerm, error) {
//	out := []ProjectWithPerm{}
//	if auth.Opts != nil {
//		filter := bson.M{} // TODO: make sure this works
//		if !auth.isAdmin() {
//			filter = bson.M{"_id": bson.M{"$in": maps.Keys(auth.Opts.Projects)}}
//		}
//		cursor, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(projectsCollectionName).
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
//		cursor, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(projectsCollectionName).
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
