package rfid

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/reeceappling/goUtils/v2/utils"
	"github.com/reeceappling/mushDb/rfid/pics"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"reflect"
	slices2 "slices"
	"strings"
)

type GrainJar struct {
	MainCollectionIdField             `bson:"inline"`
	SizeCups                          int `bson:"sizeCups"` // 1==1cup, 2 == pint, 4==quart, 16==gal // TODO: new! use!
	JarRecipeField                    `bson:"inline"`
	WetnessField                      `bson:"inline"` // TODO: HANDLE IN JAVASCRIPT
	BurstGrainsField                  `bson:"inline"`
	PcRunOptionalField                `bson:"inline"`
	CreationDateField                 `bson:"inline"`
	SpeciesOptionalField              `bson:"inline"`
	SubspeciesOptionalField           `bson:"inline"`
	InnocField                        `bson:"inline"` // TODO: multiple? What if first innoc does not work?
	GenerationsFields                 `bson:"inline"`
	TransfersOutField                 `bson:"inline"`
	ParentTypeField                   `bson:"inline"` // nil == mainCollectionType, can also be MSS or clone! // TODO: INDEX???? // TODO: multiple?
	MainCollectionOptionalParentField `bson:"inline"`
	PicsField                         `bson:"inline"`
	ContaminationsField               `bson:"inline"`
	KnownFruitableField               `bson:"inline"`
	SaleField                         `bson:"inline"`
	DisposedField                     `bson:"inline"`
	MostRecentImageField              `bson:"inline"`
	NotesField                        `bson:"inline"`
	LastUpdatedField                  `bson:"inline"`
	AclField                          `bson:"inline"`
}

type BurstGrainsField struct {
	BurstGrains *int `bson:"burstGrains,omitempty" json:"burstGrains,omitempty"` // TODO: HANDLE IN JAVASCRIPT
}

func (j GrainJar) CanTransferTo(dst geneticSource) error {
	if j.Innoc == nil {
		return errors.New("source not innoculated. Cannot transfer nothing")
	}
	if slices2.Contains([]string{FruitingChamberSourceType, FruitSourceType, LcSyringeSourceType, MssSourceType, SporePrintSourceType, SporeSwabSourceType}, dst.SourceType()) {
		return errors.New("jar cannot transfer to " + dst.SourceType())
	}
	return nil
}

func (j GrainJar) GeneticInfoAsParent() (GeneticParentInfo, error) {
	return GeneticParentInfo{
		SpeciesOptionalField:    SpeciesOptionalField{j.Species},
		SubspeciesOptionalField: j.SubspeciesOptionalField,
		KnownFruitableField:     j.KnownFruitableField,
		GenerationsFields:       j.GenerationsFields,
	}, nil
}

func (j GrainJar) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return j.GenSinceSpore, j.GenSinceFruitOrSpore
}

func (j GrainJar) setTransferParent(ctx context.Context, xfer Transfer) (error, func() error) {
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(GrainJarCollectionName)
	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
	if err != nil {
		return err, nil
	}
	res, err := coll.UpdateByID(ctx, j.Id, upd)
	if err != nil {
		return err, nil
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer, nil
	}
	return nil, func() error {
		return coll.FindOneAndReplace(ctx, bson.D{{"_id", j.Id}}, j).Err()
	}
}

func (j GrainJar) setTransferChild(ctx context.Context, xfer Transfer, from geneticSource) error {
	parentInfo, genSpore, genFruitSpore, err := childGensForParent(from)
	if err != nil {
		return err
	}
	upd, err := xfer.
		PicsModsForChild().
		withInnoc(xfer).
		withParentType(&xfer.FromType).
		withParent(utils.Pointer(from.DbId())).
		withGens(genSpore, genFruitSpore).
		withSpecies(parentInfo.Species).
		withKnownFruitable(parentInfo.KnownFruitable).
		withSubspecies(parentInfo.SubSpecies).
		withLastUpdated(xfer.LastUpdated).
		Finalized()
	if err != nil {
		return ErrFailedToFinalizeMods
	}
	res, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(GrainJarCollectionName).UpdateByID(ctx, j.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer
	}
	return nil
}

func (j GrainJar) EntryTypeField() *string {
	return utils.Pointer(GrainJarSourceType)
}

func (j GrainJar) Collection(ctx mongo.SessionContext) *mongo.Collection { // TODO: DO THIS ON EVERYTHING!
	return ctx.Client().Database(dbName).Collection(GrainJarCollectionName) // TODO: switch all references to jarCollectionName
}

func (j *GrainJar) Refresh(ctx mongo.SessionContext) error { // TODO; DO THIS ON EVERYTHING!
	return j.Collection(ctx).FindOne(ctx, bson.D{{"_id", j.Id}}).Decode(j)
}

func LookupGrainJar(ctx context.Context, id MainCollectionId) (j *GrainJar, err error) {
	j = &GrainJar{}
	err = ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(GrainJarCollectionName).FindOne(ctx, bson.D{{"_id", id}}).Decode(j)
	return j, err
}

type innoculateJarFromRequest struct { // TODO: this
	parent MainCollectionId // TODO: can this be alt?

}

func initializeJars(ctx context.Context) error {
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(GrainJarCollectionName)
	err := createIndexes(ctx, coll,
		[]mongo.IndexModel{
			creationDateIndexModel,
			//newSimpleIndex("sizeCups", "sizeCups", true, false, false),
			//newSimpleIndex("recipe", "recipe", false, true, false),
			//newSimpleIndex("wetness", "wetness", false, true, false),
			//newSimpleIndex("burstGrains", "burstGrains", false, true, false),
			newSimpleIndex("pcRun", "pcRun", false, true, false),
			creationDateIndexModel,
			newSimpleIndex("species", "species", false, true, false),
			newSimpleIndex("subspecies", "subspecies", false, true, false),
			//newSimpleIndex("innoc", "innoc", false, true, false),
			//newSimpleIndex("genSinceSpore", "genSpore", true, true, false),
			//newSimpleIndex("genSinceFruitOrSpore", "genFruitOrSpore", true, true, false),
			//transfersOutIndexModel,
			//newSimpleIndex("parent", "parent", false, true, false),         // TODO: nil is store or outside?
			//newSimpleIndex("parentType", "parentType", false, true, false), // TODO: nil is store or outside?
			//Pics (no index)
			//TODO: Contams
			//newSimpleIndex("knownFruitable", "knownFruitable", false, true, false),
			//saleIndexModel,
			//newSimpleIndex("disposed", "disposed", false, true, false),
			// MostRecentImage
			//Notes (no index) (maybe later with tags?)
			projectsIndexModel,
			lastUpdatedIndexModel,
		})
	if err != nil {
		// TODO: I dont like the second one here!
		if !strings.Contains(err.Error(), "Identical index already exists:") && !strings.Contains(err.Error(), "Cannot build two identical indexes") {
			println("failed to create indices", err.Error())
			return err
		}
	}
	// If test agar batch does not exist, then create it
	testId := mainCollIdForint(idTestJar)
	testItem := &GrainJar{
		MainCollectionIdField:   MainCollectionIdField{testId},
		JarRecipeField:          JarRecipeField{&exAltId},
		PcRunOptionalField:      PcRunOptionalField{&exAltId},
		CreationDateField:       CreationDateField{exampleTime},
		SpeciesOptionalField:    SpeciesOptionalField{&exampleSpecies},
		SubspeciesOptionalField: SubspeciesOptionalField{exampleSubspecies},
		InnocField:              InnocField{&exAltId},
		GenerationsFields: GenerationsFields{
			GenSporeField:        GenSporeField{&exGenSinceSpore},
			GenSinceFruitOrSpore: &exGenSinceFruitSpore,
		},
		TransfersOutField:                 TransfersOutField{exAlts},
		ParentTypeField:                   ParentTypeField{&exParentType},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&exPlate},
		PicsField:                         PicsField{exPics},
		ContaminationsField:               ContaminationsField{exContams},
		KnownFruitableField:               KnownFruitableField{exBool},
		SaleField:                         SaleField{&exAltId},
		DisposedField:                     DisposedField{&exampleTime},
		MostRecentImageField:              MostRecentImageField{&exPics[0]},
		NotesField:                        NotesField{exampleNotes()},
		LastUpdatedField:                  LastUpdatedField{exampleTime},
	}
	return addTestMainEntries(ctx, testItem)
}

// TODO: RENAME AND MOVE!
func testExistingEntry[T any](ctx context.Context, coll *mongo.Collection, testId any, testItem T, existingEntry T) error {
	res, err := coll.InsertOne(ctx, testItem)
	if err != nil {
		return err
	}
	if res == nil {
		return errors.New("result should not be nil")
	}
	err = coll.FindOne(ctx, bson.D{{"_id", testId}}).Decode(&existingEntry)
	if err != nil {
		return errors.New("not found at specified id. " + err.Error())
	}
	if !reflect.DeepEqual(existingEntry, testItem) {
		ee, err := json.Marshal(existingEntry)
		if err != nil {
			println("bad existing json")
		}
		te, err := json.Marshal(testItem)
		if err != nil {
			println("bad test json")
		}
		println("-------------------")
		println(string(te))
		println(string(ee))
		return errors.New("entries (as updated) were not equal")
	}
	return nil
}

// TODO: no innoculateJarHandler. Comes from create transfer handler

type createJarRequest struct {
	Recipe AlternateCollectionId // grain recipe
	WetnessField
	BurstGrainsField
	CreationDateField
	PcRunField
	NotesField
	WriteTagToField
}

func createJarHandler(w http.ResponseWriter, r *http.Request) {
	data := createJarRequest{}
	id, err := newCollectionId(r.Context(), GrainJarCollectionName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(bs, &data)
	if err != nil {
		http.Error(w, "failed to unmarshal request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	ctx, db := Db(r)
	coll := db.Collection(GrainJarCollectionName)
	now := unixTimeForNow()
	pcrun := PcRunField{data.PcRun}
	toInsert := GrainJar{
		MainCollectionIdField: MainCollectionIdField{id},
		JarRecipeField:        JarRecipeField{&data.Recipe},
		PcRunOptionalField:    pcrun.asOptional(),
		BurstGrainsField:      data.BurstGrainsField,
		WetnessField:          data.WetnessField,
		CreationDateField:     CreationDateField{data.CreationDate},
		NotesField:            NotesField{data.Notes},
		LastUpdatedField:      LastUpdatedField{now},
		AclField:              allCanWriteAcl(),
	}
	_, err = pcrun.Get(ctx)
	if err != nil {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = toInsert.JarRecipeField.Get(ctx)
	if err != nil && !errors.Is(err, ErrMissingOptionalField) {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		dbErr(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	finishCreateMainCollectionEntry(ctx, coll, &toInsert, w)
}

type importJarRequest struct {
	Recipe AlternateCollectionId // Jar Recipe
	CreationDateField
	SpeciesField
	SubspeciesOptionalField
	Generation *int
	KnownFruitableField
	WriteTagToField
	PermsOnRequest
	// image as "img"
}

func importJarHandler(w http.ResponseWriter, r *http.Request) {
	data := importJarRequest{}
	id, err := newCollectionId(r.Context(), GrainJarCollectionName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	b58id := id.asBase58()
	r.Body = http.MaxBytesReader(w, r.Body, maxMultipartRequestSize) // TODO: do multipart streamlined way
	defer r.Body.Close()
	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "unable to open multipart reader: "+err.Error(), http.StatusBadRequest)
		return
	}
	p1, err := reader.NextPart()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer p1.Close()
	// Process text (or object)
	bs, errr := io.ReadAll(p1)
	if errr != nil {
		err = errr
		http.Error(w, "unable to read data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	// PARSE INTO CORRECT DATA FORMAT
	err = json.Unmarshal(bs, &data)
	if err != nil {
		http.Error(w, "unable to unmarshal json form data: "+err.Error(), http.StatusBadRequest)
		return
	}
	user, err := GetAuthInfo(r.Context()) // TODO: fix
	if err != nil {
		http.Error(w, "failed to get auth info: "+err.Error(), http.StatusUnauthorized)
		return
	}
	//sp, subsp, err := getSpeciesAndSubspecies(r.Context(), data.Species, data.SubSpecies)
	//if err != nil {
	//	http.Error(w, "failed to get species and subspecies: "+err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//finalPerms := minimalPermsBetween(data.Perms, sp, subsp)
	//finalPerms.Users = finalPerms.Users.WithAuthor(authinfo.Email)
	//err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	//if err != nil {
	//	http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
	//	return
	//}
	// Try to get pic if exists
	picsSaved := []string{}
	defer func() {
		if err != nil {
			err = pics.DeleteFiles(r.Context(), picsSaved...)
			if err != nil {
				handleFileDeleteErr(err)
			}
		}
	}()
	// Go to next part, if exists to get image
	var importedPic *PicWithNotes = nil
	p, err := reader.NextPart()
	if err != nil {
		if err != io.EOF {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		fileName := p.FileName()
		defer p.Close()
		if fileName != "img" {
			http.Error(w, "invalid image name", http.StatusBadRequest)
			return
		}
		// Process file
		fieldBytes, err := multipartToImageBytes(p, w)
		if err != nil {
			// Already wrote
			return
		}
		newFileNameWithPrefixPath, errr := pics.SaveFile(r.Context(), fieldBytes, "bag", string(b58id), "img")
		if errr != nil {
			err = errr
			http.Error(w, "failed to save file: "+err.Error(), http.StatusBadRequest)
			return
		}
		picsSaved = append(picsSaved, newFileNameWithPrefixPath)
		now := unixTimeForNow()
		importedPic = &PicWithNotes{
			Time:       now,
			Location:   imageLocation(newFileNameWithPrefixPath),
			NotesField: NotesField{[]Note{}},
		}
	}
	var gen *Generation = nil
	if data.Generation != nil {
		gen = (*Generation)(data.Generation)
	}
	pix := []PicWithNotes{}
	if importedPic != nil {
		pix = []PicWithNotes{*importedPic}
	}
	ctx, db := Db(r)
	acl, err := data.PermsOnRequest.AclForUser(ctx, user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	toInsert := GrainJar{
		MainCollectionIdField:   MainCollectionIdField{id},
		JarRecipeField:          JarRecipeField{&data.Recipe},
		PcRunOptionalField:      PcRunOptionalField{}, // No pc runs on imports
		CreationDateField:       CreationDateField{data.CreationDate},
		SpeciesOptionalField:    SpeciesOptionalField{&data.Species},
		SubspeciesOptionalField: data.SubspeciesOptionalField,
		GenerationsFields: GenerationsFields{
			GenSporeField:        GenSporeField{gen},
			GenSinceFruitOrSpore: gen,
		},
		PicsField:            PicsField{pix},
		KnownFruitableField:  data.KnownFruitableField,
		MostRecentImageField: MostRecentImageField{importedPic},
		LastUpdatedField:     LastUpdatedField{unixTimeForNow()},
		AclField:             acl,
	}

	coll := db.Collection(GrainJarCollectionName)
	_, err = toInsert.PcRunOptionalField.Get(ctx)
	if err != nil && !errors.Is(err, ErrMissingOptionalField) {
		dbErr(w, "invalid jar recipe: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = toInsert.JarRecipeField.Get(ctx)
	if err != nil && !errors.Is(err, ErrMissingOptionalField) {
		dbErr(w, "invalid jar recipe: "+err.Error(), http.StatusInternalServerError)
		return
	}
	finishCreateMainCollectionEntry(ctx, coll, &toInsert, w)
}

type updateJarRequest struct {
	Notes AllEntries[Note]
	KnownFruitableField
	DisposedField
	SaleField
	Images  SplitEntries[picWithNotesForm, PicWithNotesLessLocation] //"newPic-1"
	Contams SplitEntries[contamForm, ContaminationLessLocation]      //"newContam-1"
	WriteTagToField
	PermsOnRequest
}

func (upr updateJarRequest) reform() resolvedUpdateJarRequest {
	return resolvedUpdateJarRequest{
		KnownFruitableField: upr.KnownFruitableField,
		SaleField:           upr.SaleField,
		DisposedField:       upr.DisposedField,
		Notes:               upr.Notes,
		Images:              imageUpdates(upr.Images),
		Contams:             contamUpdates(upr.Contams),
		PermsOnRequest:      upr.PermsOnRequest,
	}
}

type resolvedUpdateJarRequest struct {
	KnownFruitableField
	SaleField
	DisposedField
	Notes   AllEntries[Note]
	Images  SplitEntries[picWithNotesForm, PicWithNotes]
	Contams SplitEntries[contamForm, Contamination]
	PermsOnRequest
}

func (out resolvedUpdateJarRequest) modsFor(existing *GrainJar, aclField AclField) (bson.D, error) {
	return NewMods().
		updateKnownFruitableIfNeeded(out.KnownFruitable, existing.KnownFruitable).
		updateSaleIfNeeded(out.Sale, existing.Sale).
		updateDisposedIfNeeded(out.Disposed, existing.Disposed).
		updateNotesIfNeeded(out.Notes, existing.Notes).
		updatePicsIfNeeded(out.Images, existing.Pics).
		updateContamsIfNeeded(out.Contams, existing.Contaminations).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateJarHandler(w http.ResponseWriter, r *http.Request) {
	data := updateJarRequest{}
	defer r.Body.Close()
	b58Id := Base58Str(r.PathValue("id"))
	id, err := b58Id.toMainCollectionId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	reader, err := multipartReaderForRequest(r, w, &data)
	if err != nil {
		// Already written
		return
	}
	//if err = data.Perms.ValidateUserCanWrite(r.Context()); err != nil {
	//	http.Error(w, "cannot write to new perms: "+err.Error(), http.StatusBadRequest)
	//	return
	//}
	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	newPics, newContams, _, err := getMultipartImages(r.Context(), "jar", w, reader, b58Id)
	if err != nil {
		// Already wrotw
		return
	}

	// CHECK THAT ALL NEW PICS EXIST
	// PROCESS ALL NEW PICS AND CONTAMS
	out := data.reform()
	for i, _ := range data.Images.New {
		loc, exists := newPics[i]
		if !exists {
			http.Error(w, fmt.Sprintf("error, location for new picture index %d not found (should never happen)", i), http.StatusInternalServerError)
			return
		}
		out.Images.New[i].Location = imageLocation(loc)
	}
	for i, _ := range data.Contams.New {
		if loc, exists := newContams[i]; exists {
			finalLoc := imageLocation(loc)
			out.Contams.New[i].Location = &finalLoc
		}
	}
	ctx, db := Db(r)
	coll := db.Collection(GrainJarCollectionName)
	// go get current
	existing := &GrainJar{}
	err = coll.FindOne(ctx, bson.D{{"_id", id}}).Decode(existing)
	if err != nil {
		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishMainCollItemUpdate(ctx, w, coll, out.modsFor, existing, out.PermsOnRequest)
}

func Db(r *http.Request) (context.Context, *mongo.Database) {
	ctx := r.Context()
	return ctx, ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
}
