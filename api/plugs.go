package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/reeceappling/goUtils/v2/utils"
	"github.com/reeceappling/mushDb/api/env"
	"github.com/reeceappling/mushDb/api/pics"
	"github.com/reeceappling/mushDb/api/request"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"slices"
)

// sometimes needed for transfers

type PlugsJar struct {
	MainCollectionIdField             `bson:"inline"`
	ParentTypeField                   `bson:"inline"` // empty==bought
	MainCollectionOptionalParentField `bson:"inline"` // empty=bought. From plugs, jar, LC, plate/slant
	CreationDateField                 `bson:"inline"`
	DowelTypes                        []Dowel `bson:"dowelTypes" json:"dowelTypes"`
	GenerationsFields                 `bson:"inline"`
	SpeciesOptionalField              `bson:"inline"`
	TransfersOutField                 `bson:"inline"`
	SubspeciesOptionalField           `bson:"inline"`
	InnocField                        `bson:"inline"`
	PicsField                         `bson:"inline"` // TODO: ADDED 7/6/26
	ContaminationsField               `bson:"inline"` // TODO: ADDED 7/6/26
	MostRecentImageField              `bson:"inline"`
	KnownFruitableField               `bson:"inline"`
	PcRunOptionalField                `bson:"inline"` // defaults on import, but can be created without a run!
	SalesField                        `bson:"inline"` // TODO: MULTIPLE!
	DisposedField                     `bson:"inline"` // Also changed once all pegs are sold/used?
	NotesField                        `bson:"inline"`
	LastUpdatedField                  `bson:"inline"`
	AclField                          `bson:"inline"`
}

func (pl PlugsJar) CanTransferTo(dst geneticSource) error {
	if !slices.Contains([]string{BagSourceType, GrainJarSourceType, LcSourceType, PlateSourceType, PlugSourceType, SlantSourceType, StasisTubeSourceType}, dst.SourceType()) {
		return errors.New("plugs cannot transfer to " + dst.SourceType())
	}
	return nil
}

func (pl PlugsJar) GeneticInfoAsParent() (GeneticParentInfo, error) {
	return GeneticParentInfo{
		SpeciesOptionalField:    SpeciesOptionalField{pl.Species},
		SubspeciesOptionalField: pl.SubspeciesOptionalField,
		KnownFruitableField:     pl.KnownFruitableField,
		GenerationsFields:       pl.GenerationsFields,
	}, nil
}

func (pl PlugsJar) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
	parentInfo, genSpore, genFruitSpore, err := childGensForParent(from)
	if err != nil {
		return err
	}
	upd, err := xfer.PicsModsForChild().
		withInnoc(xfer).
		withParentType(&xfer.FromType).
		withParent(utils.Pointer(from.DbId())).
		withGens(genSpore, genFruitSpore).
		withSpecies(parentInfo.Species).
		withSubspecies(parentInfo.Subspecies).
		withPerms(from.Permissions()).
		withLastUpdated(xfer.LastUpdated).
		Finalized()
	if err != nil {
		return errors.Join(err, ErrFailedToFinalizeMods)
	}
	res, err := mongo.SessionFromContext(ctx).Client().Database(dbName).Collection(PlugsCollectionName).UpdateByID(ctx, pl.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer
	}
	return nil
}

func (pl PlugsJar) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return pl.GenSinceSpore, pl.GenSinceFruitOrSpore
}

func (pl PlugsJar) Innoculatable() error {
	var soldErr error = nil
	if pl.Sales != nil || len(pl.Sales) != 0 {
		soldErr = errors.New("cannot innoculate sold plugs")
	}
	return errors.Join(
		pl.RequireNoSpecies(),
		pl.RequireNoSubspecies(),
		pl.RequireNotDisposed(),
		soldErr,
		pl.RequireUnknownFruitable(),
		pl.RequireNoInnoculation(),
		pl.HasPcRun())
}

type Dowel struct {
	Radius `bson:"inline"`
	Wood   Wood `bson:"wood" json:"wood"`
}

func (d Dowel) validate() error {
	if !slices.Contains(woods, d.Wood) {
		return errors.New("invalid wood type")
	}
	if d.Size <= 0.0 {
		return errors.New("invalid dowel radius magnitude")
	}
	if !slices.Contains(dowelRadiusUnits, d.Units) {
		return errors.New("invalid dowel radius units")
	}
	return nil
}

type Radius struct {
	Size  float64    `bson:"size" json:"size"`
	Units LengthUnit `bson:"units" json:"units"`
}

type Wood string

var woods = []Wood{oak, poplar, bamboo}

const (
	poplar Wood = "Poplar"
	oak    Wood = "Oak"
	bamboo Wood = "Bamboo"
)

type LengthUnit string

var filterSizeUnits = []LengthUnit{um} // TODO: use? can be 0.2 micron (not pc-able), 0.5 micron (pc-able), or 5 micron (pc-able)
var dowelRadiusUnits = []LengthUnit{mm, cm, in}

// var dowelLengthUnits = []LengthUnit{mm, cm, in, ft} // TODO: use?
// TODO: use?
var lengthUnits = []LengthUnit{um, mm, cm, in, ft, meter} // TODO: use?
const (
	um    = "um"
	mm    = "mm"
	cm    = "cm"
	in    = "in"
	ft    = "ft"
	meter = "m"
)

//func (pl PlugsJar) setTransferParent(ctx context.Context, xfer Transfer) error {
//	// TODO: can this even occur?
//	coll := DbFrom(ctx).Collection(pl.CollectionName())
//	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
//	if err != nil {
//		return err
//	}
//	res, err := coll.UpdateByID(ctx, pl.Id, upd)
//	if err != nil {
//		return err
//	}
//	if res.ModifiedCount == 0 {
//		return ErrNoParentModifiedForTransfer
//	}
//	return nil
//}

func initializePlugs(ctx context.Context) error {
	// Indices
	coll := DbFrom(ctx).Collection(PlugsCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		creationDateIndexModel,
		//newSimpleIndex("parentType", "parentType", false, true, false), // TODO: nil is store or outside?
		//newSimpleIndex("parent", "parent", false, true, false),         // TODO: nil is store or outside?
		// TODO: combo dowel types index?
		//newSimpleIndex("dowelTypes", "dowelTypes.wood", false, false, false),
		//newSimpleIndex("dowelSizes", "dowelTypes.radius", false, false, false),
		newSimpleIndex("species", "species", false, true, false),
		newSimpleIndex("subspecies", "subspecies", false, true, false),
		//newSimpleIndex("innoc", "innoc", false, true, false),
		newSimpleIndex("pcRun", "pcRun", false, false, false),
		// TODO: ensure sales index is for each item!
		//newSimpleIndex("sales", "sales", false, true, false),
		//disposedIndexModel,
		//Notes (no index unless tags)
		projectsIndexModel,
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// Create test plugs
	return env.IfNotProd(ctx, func() error {
		testItem := &PlugsJar{
			MainCollectionIdField: MainCollectionIdField{exPlugId},
			ParentTypeField: ParentTypeField{
				utils.Pointer("plate"),
			},
			MainCollectionOptionalParentField: MainCollectionOptionalParentField{&exPlate},
			CreationDateField:                 CreationDateField{exampleTime},
			DowelTypes: []Dowel{
				{
					Radius: Radius{
						Size:  2,
						Units: "cm",
					},
					Wood: "oak",
				},
			},
			SpeciesOptionalField:    SpeciesOptionalField{&testEntryStringId},
			SubspeciesOptionalField: SubspeciesOptionalField{&testEntryStringId},
			InnocField:              InnocField{&exAltId},
			PcRunOptionalField:      PcRunOptionalField{&exAltId},
			SalesField:              SalesField{[]AlternateCollectionId{exAltId}},
			DisposedField:           DisposedField{&exampleTime},
			NotesField:              NotesField{exampleNotes()},
			LastUpdatedField:        LastUpdatedField{exampleTime},
			AclField:                allCanReadAcl(nil),
		}
		return addTestMainEntries(ctx, testItem)
	})
}

type createPlugsRequest struct {
	DowelTypes         []Dowel `json:"dowelTypes"`
	PcRunOptionalField         // OPTIONAL! // TODO: Can be created before pc!
	NotesField
	WriteTagToField
}

func createPlugsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := createPlugsRequest{}
	id := NextMainCollectionId()
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
	for i, d := range data.DowelTypes {
		if err = d.validate(); err != nil {
			errTxt := fmt.Sprintf("failed to validate dowel type for entry #%d", i)
			http.Error(w, errTxt+": "+err.Error(), http.StatusBadRequest)
			return
		}
	}
	if _, err = data.PcRunOptionalField.Get(ctx); err != nil {
		if !errors.Is(err, ErrMissingOptionalField) {
			http.Error(w, "invalid pc run field: "+err.Error(), http.StatusBadRequest)
			return
		}
	}
	// TODO: validate dowels

	ctx, now := request.UnixTime(r.Context())
	toInsert := PlugsJar{
		MainCollectionIdField: MainCollectionIdField{id},
		CreationDateField:     CreationDateField{now},
		DowelTypes:            data.DowelTypes,
		PcRunOptionalField:    PcRunOptionalField{data.PcRun},
		NotesField:            NotesField{data.Notes},
		LastUpdatedField:      LastUpdatedField{now},
		// No Perms here for basic plugs
		AclField: allCanWriteAcl(),
	}

	err = writeRfidTagIfNecessary(ctx, data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}

	finishCreateMainCollectionEntry(ctx, &toInsert, w)
}

type importPlugsRequest struct {
	DowelTypes []Dowel     `json:"dowelTypes"`
	Generation *Generation `json:"generation,omitempty"` // required when innoculated
	SpeciesOptionalField
	SubspeciesOptionalField
	KnownFruitableField
	NotesField
	WriteTagToField
}

func importPlugsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctx, now := request.UnixTime(r.Context())
	data := importPlugsRequest{}
	id := NextMainCollectionId()
	b58id := id.AsBase58()
	reader, err := multipartReaderForRequest(r.WithContext(ctx), w, &data)
	if err != nil {
		// Already written
		return
	}
	err = writeRfidTagIfNecessary(ctx, data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Try to get pic if exists
	picsSaved := []string{}
	defer func() {
		if err != nil {
			err = pics.DeleteFiles(ctx, picsSaved...)
			if err != nil {
				handleFileDeleteErr(err)
			}
		}
	}()
	// Go to next part, if exists to get image
	var importedPic *PicWithNotes = nil
	p, errr := reader.NextPart()
	if errr != nil {
		err = errr
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
		newFileNameWithPrefixPath, errr := pics.SaveFile(ctx, fieldBytes, "plugs", string(b58id), "img") // TODO: plugs ok?
		if errr != nil {
			err = errr
			http.Error(w, "failed to save file: "+err.Error(), http.StatusBadRequest)
			return
		}
		picsSaved = append(picsSaved, newFileNameWithPrefixPath)
		importedPic = utils.Pointer(newPicWithNotes(now, []Note{}, ImageLocation(newFileNameWithPrefixPath)))
	}
	// Validation
	for i, d := range data.DowelTypes {
		if err = d.validate(); err != nil {
			errTxt := fmt.Sprintf("failed to validate dowel type for entry #%d", i)
			http.Error(w, errTxt+": "+err.Error(), http.StatusBadRequest)
			return
		}
	}
	var gen *Generation = nil
	if data.Species != nil {
		if data.Generation == nil {
			http.Error(w, "innoculated must have generation: "+err.Error(), http.StatusBadRequest)
			return
		}
		if *data.Generation < 1 {
			http.Error(w, "gen must be positive", http.StatusBadRequest)
			return
		}
		gen = data.Generation
	} else {
		data.KnownFruitable = nil
		data.Subspecies = nil
	}
	pix := []PicWithNotes{}
	if importedPic != nil {
		pix = []PicWithNotes{*importedPic}
	}

	var finalPerms ACL
	if data.Species == nil { // Not innoculated
		finalPerms = allCanWriteAcl().ACL
	} else {
		finalPerms, err = ImportFinalPerms(ctx, *data.Species, data.Subspecies)
		if err != nil {
			http.Error(w, "failed to get species and/or subspecies: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	toInsert := PlugsJar{
		MainCollectionIdField: MainCollectionIdField{id},
		PcRunOptionalField:    PcRunOptionalField{&impPcRun}, // default for imports
		CreationDateField:     CreationDateField{now},
		DowelTypes:            data.DowelTypes,
		GenerationsFields: GenerationsFields{
			GenSporeField:        GenSporeField{gen},
			GenSinceFruitOrSpore: gen,
		},
		SpeciesOptionalField:    SpeciesOptionalField{data.Species},
		SubspeciesOptionalField: SubspeciesOptionalField{data.Subspecies},
		KnownFruitableField:     data.KnownFruitableField,
		PicsField:               PicsField{pix},
		ContaminationsField:     ContaminationsField{Contaminations: []Contamination{}},
		NotesField:              NotesField{data.Notes},
		LastUpdatedField:        LastUpdatedField{now},
		// No Perms here for basic plugs
		AclField: finalPerms.AsField(),
	}
	err = writeRfidTagIfNecessary(ctx, data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	finishImportMainCollectionEntry(ctx, &toInsert, w)
}

// TODO: new sale?

type updatePlugsRequest struct {
	PcRunOptionalField // Can only be set once!
	KnownFruitableField
	NotesUpdateField
	ImagesUpdateField
	ContamsUpdateField
	DisposedField
	PermsOnRequest `json:"acl"`
}

func (upr updatePlugsRequest) reform() resolvedUpdatePlugsRequest {
	return resolvedUpdatePlugsRequest{
		PcRunOptionalField:  upr.PcRunOptionalField,
		KnownFruitableField: upr.KnownFruitableField,
		DisposedField:       upr.DisposedField,
		NotesUpdateField:    upr.NotesUpdateField,
		Images:              imageUpdates(upr.Images),
		Contams:             contamUpdates(upr.Contams),
		PermsOnRequest:      upr.PermsOnRequest,
	}
}

type resolvedUpdatePlugsRequest struct {
	PcRunOptionalField // Can only be set once!
	KnownFruitableField
	NotesUpdateField
	Images  SplitEntries[picWithNotesForm, PicWithNotes]
	Contams SplitEntries[contamForm, Contamination]
	DisposedField
	PermsOnRequest `json:"acl"`
}

func (req resolvedUpdatePlugsRequest) modsFor(existing *PlugsJar, acl AclField) (bson.D, error) {
	mods := NewMods()
	return mods.
		updatePcRunIfNeeded(req, existing).
		updateNotesIfNeeded(req, existing).
		updateDisposedIfNeeded(req, existing).
		updatePicsIfNeeded(req.Images, existing.Pics).
		updateContamsIfNeeded(req.Contams, existing.Contaminations).
		updateMostRecentImageIfNeeded(existing.MostRecentImage, loadMriPics(&req.Images, &req.Contams, nil)).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updatePlugsHandler(w http.ResponseWriter, r *http.Request) { // TODO: overhauled for images and contams! thoroughly test!
	data := &updatePlugsRequest{}
	idStr, err := UrlDecodeString(r.PathValue("id"))
	if err != nil {
		http.Error(w, "failed to url decode string: "+err.Error(), http.StatusInternalServerError)
		return
	}
	mainCollId, err := StandardizeMainCollectionId(idStr)
	if err != nil {
		http.Error(w, "failed to standardize main collection id: "+err.Error(), http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	newPics, newContams, _, err := fullMultipartWithNoBreaks(w, r, &data, Base58Str(idStr))
	if err != nil {
		// Already wrote
		return
	}

	// CHECK THAT ALL NEW PICS EXIST
	// PROCESS ALL NEW PICS AND CONTAMS
	//// TODO: PANICKING WHEN SENDING EMPTY THINGS!!!!
	//for i, picNote := range data.Images.Existing[0].Data.Notes.asEntries() {
	//	println("note", i, picNote.Note)
	//}
	out := data.reform()
	for i, _ := range data.Images.New {
		loc, exists := newPics[i]
		if !exists {
			http.Error(w, fmt.Sprintf("error, location for new picture index %d not found (should never happen)", i), http.StatusInternalServerError)
			return
		}
		out.Images.New[i].Location = ImageLocation(loc)
	}
	for i, _ := range data.Contams.New {
		if loc, exists := newContams[i]; exists {
			finalLoc := ImageLocation(loc)
			out.Contams.New[i].Location = &finalLoc
		} else {
			println("no contam location for", i)
		}
	}
	//finalReqBs, err := json.MarshalIndent(out, "", " ")
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusBadRequest)
	//	return
	//}
	//println("REQUEST BYTES: ", string(finalReqBs)) // TODO: del

	existing := &PlugsJar{}
	db := DbFrom(ctx)
	err = db.Collection(PlugsCollectionName).FindOne(ctx, BsonFindFilter(IDfld, *mainCollId)).Decode(existing)
	if err != nil {
		http.Error(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	if data.PcRun != nil { // TODO: ONLY CHECK THIS IF THE RUN CHANGED!!!!
		_, err = data.PcRunOptionalField.Get(ctx)
		if err != nil {
			http.Error(w, "invalid pc run: "+err.Error(), http.StatusBadRequest)
			return
		}
	}
	finishMainCollItemUpdate(ctx, w, out.modsFor, existing, out.PermsOnRequest)

}

func deletePlugsHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Empty id for delete request", http.StatusBadRequest)
		return
	}
	id, err := Base58Str(idStr).ToMainCollectionId()
	if err != nil {
		http.Error(w, "Invalid ID to delete: "+err.Error(), http.StatusBadRequest)
		return
	}
	// Validate not used in other places...
	ctx := r.Context()
	// ensure item does not have any transfers in or out
	item, err := GetMainCollectionItemSpecific[*PlugsJar](ctx, id, &PlugsJar{})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			http.Error(w, "Item to be deleted not found: "+err.Error(), http.StatusNotFound) // TODO: ok?
		} else {
			http.Error(w, "Failed to retrieve item to be deleted: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if item.Parent != nil {
		// TODO: what if we want to remove it from the parent as well?
		http.Error(w, "Cannot delete innoculated items!", http.StatusConflict)
		return
	}
	if item.TransfersOut != nil && len(item.TransfersOut) > 0 {
		http.Error(w, "Cannot delete items with transfers out", http.StatusConflict)
		return
	}

	// Delete if not found elsewhere!
	DeleteCollectionItem(ctx, item.CollectionName(), id, w)
}
