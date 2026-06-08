package api

import (
	"context"
	"encoding/json"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
)

// used in substrateBatch, species, (bag, fruiting chamber, get it either provided or from the substrateBatch)

var _ CollectionItem = &SubstrateRecipe{}

type SubstrateRecipeField struct {
	Substrate AlternateCollectionId `bson:"recipe" json:"recipe"`
}

func (field SubstrateRecipeField) Get(ctx context.Context) (out SubstrateRecipe, err error) {
	err = ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(PcRunCollectionName).FindOne(ctx, bson.M{
		"_id": field.Substrate,
	}).Decode(&out)
	return out, err
}

type SubstrateRecipe struct {
	AlternateCollectionIdField `bson:"inline"`
	NameField                  `bson:"inline"`
	StandardField              `bson:"inline"`
	AliasesField               `bson:"inline"` // must be unique everywhere
	NotesField                 `bson:"inline"` // ingredients in notes
	LastUpdatedField           `bson:"inline"`
	AclField                   `bson:"inline"`
}

func initializeSubstrates(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(SubstrateRecipesCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		newSimpleIndex("name", "name", false, false, true),
		standardIndexModel,
		aliasesIndexModel,
		//Notes (no index unless tags)
		//LastUpdated
		projectsIndexModel,
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	basicEntries := []*SubstrateRecipe{
		// Coir
		{
			AlternateCollectionIdField: altCollIdFieldForint(idCoir),
			NameField:                  NameField{"Coir"},
			AliasesField:               AliasesField{[]string{}},
			StandardField:              StandardField{true},
			NotesField: NotesField{[]Note{
				newNote(ogTime, "roughly 40g dry coir, 1 cup H20 per quart"),
			}},
			AclField: allCanReadAcl(nil),
		},
		// Coir and Vermiculite
		{
			AlternateCollectionIdField: altCollIdFieldForint(idCoirVerm),
			NameField:                  NameField{"CVG"},
			AliasesField:               AliasesField{[]string{"Coir with Vermiculite"}},
			StandardField:              StandardField{true},
			NotesField: NotesField{[]Note{
				newNote(ogTime, "Recipe: roughly 40g dry coir, up to 1/2 cup vermiculite, 1 cup H20 per quart"),
				newNote(ogTime, "Vermiculite helps to keep more moisture in the substrate over time"),
			}},
			AclField: allCanReadAcl(nil),
		},
		{
			AlternateCollectionIdField: altCollIdFieldForint(idWoodPellets),
			NameField:                  NameField{"HWFP"},
			AliasesField:               AliasesField{[]string{"Hardwood Fuel Pellets"}},
			StandardField:              StandardField{true},
			NotesField: NotesField{[]Note{
				newNote(ogTime, "Roughly equal parts wood pellets and water (maybe less water. Do less at first to ensure field capacity)"),
			}},
			AclField: allCanReadAcl(nil),
		},
	}
	err = addBasicAltEntries(ctx, basicEntries...)
	if err != nil {
		return err
	}
	// Add test entry
	testItem := &SubstrateRecipe{
		AlternateCollectionIdField: altCollIdFieldForint(idTestingOnly),
		NameField:                  NameField{testEntryStringId},
		StandardField:              StandardField{false},
		AliasesField:               AliasesField{[]string{"testSubstrate", "example substrate"}},
		NotesField:                 NotesField{exampleNotes()},
		LastUpdatedField:           LastUpdatedField{exampleTime},
		AclField:                   allCanWriteAcl(),
	}
	return addTestAltEntries(ctx, testItem)
}

type PermsOnRequest struct {
	UserPerms    map[string]bool      `json:"users,omitempty"` // Bool is canEdit
	ProjectPerms map[projectName]bool `json:"projects,omitempty"`
	BlanketPerm  *bool                `json:"blanketPerm,omitempty"` // If true then these entries are publicly writeable, if false then publicly readable
}

func (requestPerms PermsOnRequest) GetPermsOnRequest() PermsOnRequest {
	return requestPerms
}

func (requestPerms PermsOnRequest) DefaultAcl() ACL {
	return ACL{
		Users:       requestPerms.UserPerms,
		Projects:    requestPerms.ProjectPerms,
		BlanketPerm: requestPerms.BlanketPerm,
	}
}

// TODO: use this everywhere needed?
func (requestPerms PermsOnRequest) AclForUser(ctx context.Context, perms ResolvedUserPerms) (AclField, error) {
	client := ctx.Value(mongoClientContextKey).(*mongo.Client)

	// validate Projects
	// TODO: count instead?
	projColl := client.Database(dbName).Collection(ProjectsCollectionName)
	for projName, _ := range requestPerms.ProjectPerms {
		err := projColl.FindOne(ctx, bsonFindFilter("_id", projName)).Err()
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return AclField{}, errors.New(string("could not find project " + projName))
			}
			return AclField{}, err
		}
	}
	// validate users
	// TODO: count instead?
	userColl := client.Database(dbName).Collection(UserCollName)
	for userEmail, _ := range requestPerms.UserPerms {
		err := userColl.FindOne(ctx, bsonFindFilter("_id", userEmail)).Err()
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return AclField{}, errors.New(string("could not find email " + userEmail))
			}
			return AclField{}, err
		}
	}

	// Resolve acl
	acl := ACL{
		Users:       requestPerms.UserPerms,
		Projects:    requestPerms.ProjectPerms,
		BlanketPerm: requestPerms.BlanketPerm,
	}
	if acl.Users == nil {
		acl.Users = map[string]bool{}
	}
	if acl.Projects == nil {
		acl.Projects = map[projectName]bool{}
	}
	acl.Users[perms.Email] = true
	return AclField{ACL: acl}, nil
}

type createSubstrateRecipeRequest struct {
	NameField
	AliasesField
	StandardField // If this is a standard recipe
	NotesField
	// Initial perms are read by all and write only by owner
}

func createSubstrateRecipeHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	req := createSubstrateRecipeRequest{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	id := newAlternateCollectionId()
	ctx := r.Context()
	toInsert := SubstrateRecipe{
		AlternateCollectionIdField: AlternateCollectionIdField{id},
		NameField:                  req.NameField,
		AliasesField:               req.AliasesField,
		StandardField:              req.StandardField,
		NotesField:                 req.NotesField,
		LastUpdatedField:           LastUpdatedFieldForNow(),
		AclField:                   allCanReadAcl(GetUserEmailPtr(ctx)),
	}
	finishCreateAlternateEntry(ctx, toInsert, w)
}

type updateSubstrateRecipeRequest struct {
	NameField
	AliasesField
	StandardField
	NotesUpdateField
	PermsOnRequest `json:"acl"`
}

func (req updateSubstrateRecipeRequest) modsFor(existing *SubstrateRecipe, aclField AclField) (bson.D, error) {
	return NewMods().
		updateNameIfNeeded(req.Name, existing.Name).
		updateAliasesIfNeeded(req.Aliases, existing.Aliases). // TODO: make sure no duplicates
		updateStandardIfNeeded(req.Standard, existing.Standard).
		updateNotesIfNeeded(req, existing).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateSubstrateRecipeHandler(w http.ResponseWriter, r *http.Request) {
	b58Id := Base58Str(r.PathValue("id"))
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req := updateSubstrateRecipeRequest{}
	err = json.Unmarshal(bs, &req)
	if err != nil {
		http.Error(w, "failed to unmarshal body: "+err.Error(), http.StatusBadRequest)
		return
	}
	id, err := b58Id.toAltCollectionId()
	if err != nil {
		http.Error(w, "Invalid id! "+err.Error(), http.StatusBadRequest)
		return
	}
	ctx, db := Db(r)
	coll := db.Collection(SubstrateRecipesCollectionName)
	existing, err := GetAltCollectionItemOutsideTxn(ctx, id, SubstrateRecipe{})
	if err != nil {
		stat := http.StatusInternalServerError
		if err == mongo.ErrNoDocuments {
			stat = http.StatusNotFound
		}
		dbErr(w, err.Error(), stat)
		return
	}
	finishAltCollItemUpdate(ctx, w, coll, req.modsFor, &existing, req.PermsOnRequest)
}
