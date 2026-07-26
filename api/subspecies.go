package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/reeceappling/mushDb/api/env"
	"github.com/reeceappling/mushDb/api/request"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"net/url"
	"slices"
)

type SubspeciesOptionalField struct {
	Subspecies *string `bson:"subspecies,omitempty" json:"subspecies,omitempty"`
}

func (s SubspeciesOptionalField) RequireNoSubspecies() error {
	if s.Subspecies != nil {
		return errors.New("subspecies field should not be populated")
	}
	return nil
}

type Subspecies struct {
	NameIdField      `bson:"inline"`
	SpeciesField     `bson:"inline"`
	AliasesField     `bson:"inline"`
	NotesField       `bson:"inline"`
	LastUpdatedField `bson:"inline"`
	AclField         `bson:"inline"`
	DefaultAcl       ACL `bson:"defaultAcl" json:"defaultAcl"` // Only used when importing mainCollectionItems
}

const TestSubspeciesName = "TestSubspecies"

func initializeSubspecies(ctx context.Context) error {
	// Indices
	coll := DbFrom(ctx).Collection(SubspeciesCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		newSimpleIndex("species", "species", false, false, false),
		aliasesIndexModel,
		//Notes (no index) (maybe later with tags?)
		//projectsIndexModel,
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}

	return env.IfNotProd(ctx, func() error {
		basicEntries := []Subspecies{
			// White Beech
			{
				NameIdField:  NameIdField{"White Beech"},
				SpeciesField: SpeciesField{"Beech"},
				AliasesField: AliasesField{[]string{"Bunapi-shimeji"}},
				NotesField: NotesField{Notes: []Note{
					newNote(ogTime, "something to do with light, fixme"),
				}},
				AclField:   allCanWriteAcl(),
				DefaultAcl: allCanWriteAcl().ACL,
			},
			// Brown Beech
			{
				NameIdField:  NameIdField{"Brown Beech"},
				SpeciesField: SpeciesField{"Beech"},
				AliasesField: AliasesField{[]string{"Buna-shimeji"}},
				NotesField: NotesField{Notes: []Note{
					newNote(ogTime, "something to do with light, fixme"),
				}},
				AclField:   allCanWriteAcl(),
				DefaultAcl: allCanWriteAcl().ACL,
			},
		}
		err = addBasicAltEntries(ctx, basicEntries...)
		if err != nil {
			return err
		}

		// Add test entry
		testItem := &Subspecies{
			NameIdField:      NameIdField{TestSubspeciesName},
			SpeciesField:     SpeciesField{TestSpeciesName},
			AliasesField:     AliasesField{[]string{"testSubSpecies", "example subspecies"}},
			NotesField:       NotesField{exampleNotes()},
			LastUpdatedField: LastUpdatedField{exampleTime},
			AclField:         allCanReadAcl(nil),
			DefaultAcl:       allCanWriteAcl().ACL,
		}
		return addTestAltEntries(ctx, testItem)
	})
}

//const AliasesCollectionName = "aliasesEntries"
//
//type AliasItem struct {
//	AlternateCollectionIdField `bson:"inline"`
//	Alias                      string `bson:"alias" json:"alias"`
//	Type                       string `bson:"entryType" json:"entryType"`
//	Primary                    bool   `bson:"primary" json:"primary"`
//}
//
//func initializeAliasesCollection(ctx context.Context) error {
//	coll := DbFrom(ctx).Collection(AliasesCollectionName)
//	indexModel := mongo.IndexModel{
//		Keys: bson.D{
//			{Key: "alias", Value: 1},     // Index alias in ascending order
//			{Key: "entryType", Value: 1}, // Index entryType in ascending order
//		},
//		Options: options.Index().SetUnique(true),
//	}
//
//	if _, err := coll.Indexes().CreateOne(ctx, indexModel); err != nil {
//		return err
//	}
//	// Do for all existing species, subspecies, substrateRecipes
//	loadAllItems := true // TODO: DONT DO THIS EVERYTIME! MAKE FALSE NORMALLY!
//	if !loadAllItems {
//		return nil
//	}
//	_, er := newTxn(ctx, func(sessCtx mongo.SessionContext) (any, error) {
//		subCurs, err := DbFrom(sessCtx).Collection(SubspeciesCollectionName).Find(sessCtx, bson.D{})
//		if err != nil {
//			return nil, err
//		}
//		for subs, err := range cursorIterator[*Subspecies](sessCtx, subCurs) {
//			if err != nil {
//				return nil, err
//			}
//			err := coll.FindOne(sessCtx, bson.D{{Key: "alias", Value: subs.Name}, {Key: "entryType", Value: "subspecies"}}).Err()
//			if err != nil {
//				if errors.Is(err, mongo.ErrNoDocuments) {
//					if _, err = coll.InsertOne(sessCtx, AliasItem{
//						AlternateCollectionIdField: newAlternateCollectionId().asIdField(),
//						Alias:                      subs.Name,
//						Type:                       "subspecies",
//						Primary:                    true,
//					}); err != nil {
//						return nil, err
//					}
//				}
//			}
//			for _, alias := range subs.Aliases {
//				err := coll.FindOne(sessCtx, bson.D{{Key: "alias", Value: alias}, {Key: "entryType", Value: "subspecies"}}).Err()
//				if err != nil {
//					if errors.Is(err, mongo.ErrNoDocuments) {
//						if _, err = coll.InsertOne(sessCtx, AliasItem{
//							AlternateCollectionIdField: newAlternateCollectionId().asIdField(),
//							Alias:                      alias,
//							Type:                       "subspecies",
//							Primary:                    false,
//						}); err != nil {
//							return nil, err
//						}
//					}
//				}
//			}
//		}
//		specCurs, err := DbFrom(ctx).Collection(SpeciesCollectionName).Find(ctx, bson.D{})
//		if err != nil {
//			return nil, err
//		}
//		for spec, err := range cursorIterator[*Species](ctx, specCurs) {
//			if err != nil {
//				return nil, err
//			}
//			err := coll.FindOne(sessCtx, bson.D{{Key: "alias", Value: spec.Name}, {Key: "entryType", Value: "species"}}).Err()
//			if err != nil {
//				if errors.Is(err, mongo.ErrNoDocuments) {
//					if _, err = coll.InsertOne(sessCtx, AliasItem{
//						AlternateCollectionIdField: newAlternateCollectionId().asIdField(),
//						Alias:                      spec.Name,
//						Type:                       "species",
//						Primary:                    true,
//					}); err != nil {
//						return nil, err
//					}
//				}
//			}
//			for _, alias := range spec.Aliases {
//				err := coll.FindOne(sessCtx, bson.D{{Key: "alias", Value: alias}, {Key: "entryType", Value: "species"}}).Err()
//				if err != nil {
//					if errors.Is(err, mongo.ErrNoDocuments) {
//						if _, err = coll.InsertOne(sessCtx, AliasItem{
//							AlternateCollectionIdField: newAlternateCollectionId().asIdField(),
//							Alias:                      alias,
//							Type:                       "species",
//							Primary:                    false,
//						}); err != nil {
//							return nil, err
//						}
//					}
//				}
//			}
//		}
//		recCurs, err := DbFrom(ctx).Collection(SubstrateRecipesCollectionName).Find(ctx, bson.D{})
//		if err != nil {
//			return nil, err
//		}
//		for subRec, err := range cursorIterator[*SubstrateRecipe](ctx, recCurs) {
//			if err != nil {
//				return nil, err
//			}
//			err := coll.FindOne(sessCtx, bson.D{{Key: "alias", Value: subRec.Name}, {Key: "entryType", Value: "substrateRecipe"}}).Err()
//			if err != nil {
//				if errors.Is(err, mongo.ErrNoDocuments) {
//					if _, err = coll.InsertOne(sessCtx, AliasItem{
//						AlternateCollectionIdField: newAlternateCollectionId().asIdField(),
//						Alias:                      subRec.Name,
//						Type:                       "substrateRecipe",
//						Primary:                    true,
//					}); err != nil {
//						return nil, err
//					}
//				}
//			}
//			for _, alias := range subRec.Aliases {
//				err := coll.FindOne(sessCtx, bson.D{{Key: "alias", Value: alias}, {Key: "entryType", Value: "substrateRecipe"}}).Err()
//				if err != nil {
//					if errors.Is(err, mongo.ErrNoDocuments) {
//						if _, err = coll.InsertOne(sessCtx, AliasItem{
//							AlternateCollectionIdField: newAlternateCollectionId().asIdField(),
//							Alias:                      alias,
//							Type:                       "substrateRecipe",
//							Primary:                    false,
//						}); err != nil {
//							return nil, err
//						}
//					}
//				}
//			}
//		}
//		return nil, nil
//	})
//	return er
//}

type createSubspeciesRequest struct {
	NameField
	SpeciesField
	AliasesField
	NotesField
	// ACL/DefaultACL are initially inherited from parent species
}

func createSubspeciesHandler(w http.ResponseWriter, r *http.Request) {
	req := createSubspeciesRequest{}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	spec, _, err := getSpeciesAndSubspecies(r.Context(), req.Species, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx, now := request.UnixTime(r.Context())
	toInsert := Subspecies{
		NameIdField:      NameIdField{req.Name},
		SpeciesField:     req.SpeciesField,
		AliasesField:     req.AliasesField,
		NotesField:       req.NotesField,
		LastUpdatedField: LastUpdatedField{now},
		AclField:         spec.AclField,   // Use parent perms
		DefaultAcl:       spec.DefaultAcl, // use parent default acl
	}
	// Create species update modifications
	var tempSpecUpd *Mods
	var speciesUpdate bson.D
	if spec.Subspecies == nil {
		tempSpecUpd = NewMods().Set("subspecies", []string{req.Name})
	} else {
		if slices.Contains(spec.Subspecies, req.Name) {
			// Name already exists, do nothing
			http.Error(w, "subspecies already exists on species", http.StatusBadRequest)
			return
		}
		tempSpecUpd = NewMods().Push("subspecies", req.Name)
	}
	speciesUpdate, err = tempSpecUpd.withLastUpdated(now).Finalized()
	if err != nil {
		http.Error(w, "failed to create species update: "+err.Error(), http.StatusBadRequest)
		return
	}
	if len(speciesUpdate) == 0 {
		http.Error(w, "no changes made to species, should be impossible!", http.StatusInternalServerError)
		return
	}
	// Validate new aliases
	ctx, db := Db(r)
	coll := db.Collection(SubspeciesCollectionName)
	if err = validateAliasesNameUnused(ctx, coll, req.Name, req.Aliases); err != nil { // TODO: validate working!
		http.Error(w, "aliases or name already in use: "+err.Error(), http.StatusBadRequest)
		return
	}
	// Do DB stuff
	_, err = newTxn(ctx, func(ctx mongo.SessionContext) (any, error) {
		db := mongo.SessionFromContext(ctx).Client().Database(dbName)
		// Update species with subspecies

		bsonId := BsonFindFilter(IDfld, spec.Name)
		err = db.Collection(SpeciesCollectionName).FindOneAndUpdate(ctx, bsonId, speciesUpdate).Err()
		if err != nil {
			dbErr(w, "failed to write update to db: "+err.Error(), http.StatusInternalServerError)
			return nil, err
		}
		// Insert subspecies
		if _, e := db.Collection(SubspeciesCollectionName).InsertOne(ctx, toInsert); e != nil {
			dbErr(w, e.Error(), http.StatusInternalServerError)
			return nil, e
		}
		return nil, nil
	})
	if err != nil {
		// Already wrote in all cases except success
		println("failed in txn: " + err.Error())
		return
	}
	bsOut, err := json.Marshal(toInsert)
	if err != nil {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(bsOut)
	if err != nil {
		handleWriteErr(err, w)
	}
}

type updateSubspeciesRequest struct {
	NotesUpdateField
	AliasesField
	PermsOnRequest `json:"acl"`
	DefaultAcl     PermsOnRequest
}

func (req updateSubspeciesRequest) modsFor(existing *Subspecies, aclField AclField) (bson.D, error) {
	return NewMods().
		updateAliasesIfNeeded(req.Aliases, existing.Aliases).
		updateNotesIfNeeded(req, existing).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateDefaultAclIfNeeded(req.DefaultAcl, existing.DefaultAcl).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateSubspeciesHandler(w http.ResponseWriter, r *http.Request) {
	urlEncodedSubspeciesName := r.PathValue("id")
	subspeciesName, err := url.QueryUnescape(urlEncodedSubspeciesName)
	if err != nil {
		http.Error(w, "failed to decode subspecies name from url: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req := updateSubspeciesRequest{}
	err = json.Unmarshal(bs, &req)
	if err != nil {
		http.Error(w, "failed to unmarshal body: "+err.Error(), http.StatusBadRequest)
		return
	}
	// TODO: ensure ok! updating perms to allow for creating user as well...
	user, _ := GetAuthInfo(r.Context())
	finalDefaultAcl, err := req.DefaultAcl.AclForUser(r.Context(), user)
	if err != nil {
		http.Error(w, "failed to create default acl: "+err.Error(), http.StatusInternalServerError)
		return
	}
	req.DefaultAcl = finalDefaultAcl.ACL.AsPermsOnRequest()

	ctx, db := Db(r)
	coll := db.Collection(SubspeciesCollectionName)
	existing, err := GetSubspeciesByNameInTxn(ctx, subspeciesName)
	if err != nil {
		stat := http.StatusInternalServerError
		if err == mongo.ErrNoDocuments {
			stat = http.StatusNotFound
		}
		dbErr(w, err.Error(), stat)
		return
	}

	err = validateAliasesUnused(ctx, coll, existing.Name, existing.Aliases, req.Aliases)
	if err != nil {
		http.Error(w, "At least one new alias already exists as an alias or name on another entry, or there was an error querying: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishStringIdAltCollItemUpdate(ctx, w, coll, req.modsFor, &existing, req.PermsOnRequest) // TODO: use on species, project, user(?)
}

func deleteSubspeciesHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "subspecies deletion not implemented yet", http.StatusNotImplemented)
	return
	//// TODO: DELETE SUBSPECIES FROM SPECIES!!!!!
	//sub := r.PathValue("id") // TODO: recipe by name?
	//if idStr == "" {
	//	http.Error(w, "Empty id for delete request", http.StatusBadRequest)
	//	return
	//}
	//id, err := Base58Str(idStr).toAltCollectionId()
	//if err != nil {
	//	http.Error(w, "Invalid ID to delete: "+err.Error(), http.StatusBadRequest)
	//	return
	//}
	//// Validate not used in other places...
	//ctx := r.Context()
	//db := DbFrom(ctx)
	//// ensure recipe not used by any batches first
	//err = db.Collection(AgarBatchCollectionName).FindOne(ctx, bson.M{"agarRecipe": id}).Err()
	//if err != nil {
	//	if !errors.Is(err, mongo.ErrNoDocuments) {
	//		http.Error(w, "failed to check for agarRecipe usage in agarBatch collection. "+err.Error(), http.StatusInternalServerError)
	//		return
	//	}
	//} else {
	//	// At least one item exists, fail
	//	http.Error(w, "at least one agarBatch utilizes the item you are attempting to delete.", http.StatusExpectationFailed)
	//	return
	//}
	//
	//
	//// Delete if not found elsewhere!
	//deleteResult, err := db.Collection(AgarRecipesCollectionName).DeleteOne(ctx, bson.M{IDfld: id})
	//if err != nil {
	//	http.Error(w, "failed to delete: "+err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//if deleteResult.DeletedCount == 0 {
	//	http.Error(w, "failed to delete id "+idStr+" from agar recipes. Id not found", http.StatusNotFound)
	//	return
	//}
	//_, err = w.Write([]byte(idStr))
	//handleWriteErr(err, w)
}
