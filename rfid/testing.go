package rfid

import "github.com/reeceappling/goUtils/v2/utils"

// TODO: probably get rid of this whole file

var okPicture = imageLocation("testOk.jpg")   // TODO: this!
var badPicture = imageLocation("testBad.jpg") // TODO: this!

var testNotes = []Note{
	{Time: exampleTime,
		Note: "toKeep"},
	{Time: exampleTime,
		Note: "should be changed"},
	{Time: exampleTime,
		Note: "should be deleted"},
}

var testFinalNotes = []Note{
	testNotes[0],
	{Time: exampleTime,
		Note: "changed"},
	{exampleTime, "new note"},
}

// TODO: USE THESE!
var testNotesChange = AllEntries[Note]{
	Existing: []Data[Note]{
		{Data: testNotes[0], Disabled: false},
		{Data: testFinalNotes[1], Disabled: false},
		{Data: testNotes[2], Disabled: true},
	},
	New: []Data[Note]{
		{Data: testFinalNotes[2], Disabled: false},
	},
}

var testContaminations = []Contamination{
	{
		Time:       exampleTime,
		Confirmed:  false,
		Bacteria:   false,
		Mold:       true,
		Location:   &badPicture,           // TODO: fixme!
		NotesField: NotesField{testNotes}, // TODO: ?????
	},
	{
		Time:       exampleTime,
		Confirmed:  false,
		Bacteria:   true,
		Mold:       false,
		Location:   &badPicture,           // TODO: fixme!
		NotesField: NotesField{testNotes}, // TODO: ?????
	},
}

var testContamsChange = SplitEntries[contamForm, Contamination]{
	Existing: []Data[contamForm]{
		{Data: contamForm{
			Time:      exampleTime,
			Confirmed: true,
			Bacteria:  false,
			Mold:      true,
			Location:  utils.Pointer(string(okPicture)),
			Notes:     testNotesChange,
		}, Disabled: false},
		{Data: contamForm{}, Disabled: true},
	},
	New: []Contamination{
		{
			Time:       exampleTime,
			Confirmed:  true,
			Bacteria:   true,
			Mold:       true,
			Location:   &okPicture, // TODO: fixme
			NotesField: NotesField{testFinalNotes},
		},
		// TODO: fixme
	},
}

var testImages = []PicWithNotes{
	{
		Time:       exampleTime,
		Location:   badPicture,
		NotesField: NotesField{testNotes},
	},
	{
		Time:       exampleTime,
		Location:   badPicture,
		NotesField: NotesField{testNotes},
	},
}

var testFinalImages = []PicWithNotes{
	{
		Time:       exampleTime,
		Location:   okPicture,
		NotesField: NotesField{testFinalNotes},
	},
	{
		Time:       exampleTime,
		Location:   okPicture,
		NotesField: NotesField{testFinalNotes},
	},
}

var testImagesChange = SplitEntries[picWithNotesForm, PicWithNotes]{
	Existing: []Data[picWithNotesForm]{
		{Data: picWithNotesForm{
			Time:  exampleTime,
			Img:   string(okPicture),
			Notes: testNotesChange,
		}, Disabled: false},
		{Data: picWithNotesForm{}, Disabled: true},
	},
	New: []PicWithNotes{
		testFinalImages[1],
	},
}
