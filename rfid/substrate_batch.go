package rfid

type SubstrateBatch struct { // TODO: use this
	AlternateCollectionIdField
	// Initial wetness is quantified on bags/boxes
	CreationDateField // Date of first hydration
	SubstrateRecipeField
	NotesField
	LastUpdatedField
	PermsField // TODO: delete?
}
