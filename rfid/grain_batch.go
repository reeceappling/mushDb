package rfid

// TODO: new

// TODO: DO THIS WHOLE THING???
// TODO: what about mixed-grain batches????

//const grainBatchesCollectionName = "grainBatches"
//
//type GrainBatch struct { // TODO: use this
//	AlternateCollectionIdField
//	//WetnessField
//	SoakTimeHours *int `bson:"soakTimeHrs,omitempty" json:"soakTimeHrs,omitempty"`
//	BoilTimeMins  *int `bson:"boilTimeMins,omitempty" json:"boilTimeMins,omitempty"`
//	DryTimeHours  *int `bson:"dryTimeHours,omitempty" json:"dryTimeHours,omitempty"`
//	//BurstKernels      *int `bson:"burstKernels,omitempty" json:"burstKernels,omitempty"` // 0 == perfect/none, avg == 1-3, 5 == a very noticeable amount, 10 == a ton (over 50%)
//	CreationDateField // Date of first hydration
//	JarRecipeRequiredField
//	NotesField
//	LastUpdatedField
//  AclField
//}
//
//func (batch GrainBatch) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
//	out := batch
//	err := decodeItem(&out, encoded)
//	return out, err
//}
//
//func (batch GrainBatch) EntryTypeField() *string {
//	return nil
//}
//
//func (batch GrainBatch) CollectionName() string {
//	return grainBatchesCollectionName
//}
//
//func initializeGrainBatches(ctx context.Context) error {
//	// Indices
//	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(grainBatchesCollectionName)
//	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
//		//newSimpleIndex("wetness", "wetness", false, true, false),
//		newSimpleIndex("creationDate", "creationDate", true, false, false),
//		newSimpleIndex("recipe", "recipe", false, false, false),
//		//Notes (no index unless tags)
//		lastUpdatedIndexModel,
//	})
//	if err != nil {
//		return err
//	}
//
//	// If test jar recipe does not exist, then create it
//	existingEntry := GrainBatch{}
//	testItem := GrainBatch{
//		AlternateCollectionIdField: AlternateCollectionIdField{exAltId},
//		SoakTimeHours:              utils.Pointer(8),
//		BoilTimeMins:               utils.Pointer(30),
//		DryTimeHours:               utils.Pointer(4),
//		CreationDateField:          CreationDateField{},
//		JarRecipeRequiredField:     JarRecipeRequiredField{},
//		NotesField:                 NotesField{exampleNotes()},
//		LastUpdatedField:           LastUpdatedField{exampleTime},
//	}
//	err = coll.FindOne(ctx, bson.D{{"_id", exAltId}}).Decode(&existingEntry)
//	if err == nil {
//		if reflect.DeepEqual(existingEntry, testItem) {
//			return nil
//		}
//	}
//	return testExistingEntry(ctx, coll, exAltId, testItem, existingEntry)
//}
//
//type createGrainBatchRequest struct {
//	JarRecipeRequiredField
//	NotesField
//}
//
//func createGrainBatchHandler(w http.ResponseWriter, r *http.Request) {
//	body, err := io.ReadAll(r.Body)
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusBadRequest)
//		return
//	}
//	req := createGrainBatchRequest{}
//	err = json.Unmarshal(body, &req)
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusBadRequest)
//		return
//	}
//	_, err = req.Get(r.Context())
//	if err != nil {
//		http.Error(w, "invalid request: "+err.Error(), http.StatusBadRequest)
//		return
//	}
//	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
//		coll := ctx.Client().Database(dbName).Collection(grainBatchesCollectionName)
//		toInsert := GrainBatch{
//			AlternateCollectionIdField: AlternateCollectionIdField{newAlternateCollectionId()},
//			CreationDateField:          CreationDateField{unixTimeForNow()},
//			JarRecipeRequiredField:     req.JarRecipeRequiredField,
//			NotesField:                 NotesField{req.Notes},
//			LastUpdatedField:           LastUpdatedField{unixTimeForNow()},
//		}
//		_, err = coll.InsertOne(r.Context(), toInsert)
//		if err != nil {
//			return dbErr(w, err.Error(), http.StatusInternalServerError)
//		}
//		bs, err := json.Marshal(toInsert)
//		if err != nil {
//			return dbErr(w, err.Error(), http.StatusInternalServerError)
//		}
//		return w.Write(bs)
//	})
//	if err != nil {
//		handleWriteErr(err, w)
//	}
//}
//
//type updateGrainBatchRequest struct {
//	SoakTimeHours *int             `bson:"soakTimeHrs,omitempty" json:"soakTimeHrs,omitempty"`
//	BoilTimeMins  *int             `bson:"boilTimeMins,omitempty" json:"boilTimeMins,omitempty"`
//	DryTimeHours  *int             `bson:"dryTimeHours,omitempty" json:"dryTimeHours,omitempty"`
//	Notes         AllEntries[Note] `json:"notes"`
//}
//
//func updateGrainBatchHandler(w http.ResponseWriter, r *http.Request) { // TODO: txn?
//	b58Id := Base58Str(r.PathValue("id"))
//	defer r.Body.Close()
//	bs, err := io.ReadAll(r.Body)
//	if err != nil {
//		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
//		return
//	}
//	req := updateGrainBatchRequest{}
//	err = json.Unmarshal(bs, &req)
//	if err != nil {
//		http.Error(w, "failed to unmarshal body: "+err.Error(), http.StatusBadRequest)
//		return
//	}
//	id, err := b58Id.toAltCollectionId()
//	if err != nil {
//		http.Error(w, "Invalid id! "+err.Error(), http.StatusBadRequest)
//		return
//	}
//	existing, err := GetAltCollectionItem(r.Context(), id, GrainBatch{})
//	if err != nil {
//		stat := http.StatusInternalServerError
//		if err == mongo.ErrNoDocuments {
//			stat = http.StatusNotFound
//		}
//		http.Error(w, err.Error(), stat)
//		return
//	}
//	// TODO: GET GRAIN BATCH
//	// TODO: make and/or validate grain changes?
//	mods := NewMods()
//	mods = updatePointerIfNeeded(mods, "soakTimeHours", req.SoakTimeHours, existing.SoakTimeHours)
//	mods = updatePointerIfNeeded(mods, "boilTimeMins", req.BoilTimeMins, existing.BoilTimeMins)
//	mods = updatePointerIfNeeded(mods, "dryTimeHours", req.DryTimeHours, existing.DryTimeHours)
//	upd, err := mods.
//		updateNotesIfNeeded(req.Notes, existing.Notes).
//		Finalized()
//	if err != nil {
//		http.Error(w, "error creating txn:"+err.Error(), http.StatusInternalServerError)
//		return
//	}
//	if len(upd) == 0 {
//		http.Error(w, "no changes made", http.StatusBadRequest)
//		return
//	}
//
//	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
//		coll := ctx.Client().Database(dbName).Collection(grainBatchesCollectionName)
//		bsonId := bson.D{{"_id", existing.Email}}
//		err = coll.FindOneAndUpdate(ctx, bsonId, upd).Err()
//		if err != nil {
//			return dbErr(w, err.Error(), http.StatusInternalServerError)
//		}
//		err = coll.FindOne(ctx, bsonId).Decode(&existing)
//		if err != nil {
//			return dbErr(w, err.Error(), http.StatusInternalServerError)
//		}
//		bs, err = json.Marshal(existing)
//		if err != nil {
//			return dbErr(w, err.Error(), http.StatusInternalServerError)
//		}
//		return w.Write(bs)
//	})
//	if err != nil {
//		HandleHttpWriteError(err)
//	}
//}
