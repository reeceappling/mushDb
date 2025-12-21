package rfid

type WetnessField struct {
	Wetness *int // nil==unknown, 0== very dry, 10==veryWet, 5==perfect fieldCapacity, normal range 4-6
}

type GrainBatch struct { // TODO: use this
	AlternateCollectionIdField
	WetnessField
	SoakTimeHours     *int `bson:"soakTimeHrs,omitempty" json:"soakTimeHrs,omitempty"`
	BoilTimeMins      *int `bson:"boilTimeMins,omitempty" json:"boilTimeMins,omitempty"`
	DryTimeHours      *int `bson:"dryTimeHours,omitempty" json:"dryTimeHours,omitempty"`
	BurstKernels      *int `bson:"burstKernels,omitempty" json:"burstKernels,omitempty"` // 0 == perfect/none, avg == 1-3, 5 == a very noticeable amount, 10 == a ton (over 50%)
	CreationDateField      // Date of first hydration
	JarRecipeField         // TODO: change to grain recipe field?
	NotesField
	LastUpdatedField
	PermsField // TODO: delete
}
