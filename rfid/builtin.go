package rfid

import "time"

const (
	idTestingOnly = iota
	idLmea
	idPda
	idWaterAgar
	idMeaLC
	idMeaSugLC
	idJarPop
	idJarOat
	idCoir
	idCoirVerm
	idWoodPellets
	idGrainWaterAgar
	idJarOatWithVermGypsum
	idAntibioticAgar
	idTestBag
	idExampleTransfer
	idTestPlate
	idTestFC
	idTestJar
	idTestMSS
	idTestSlant
	idTestStasis
	idTestLC
	idTestBottle
	idTestBatch
)

var ogTime = unixTimeFor(time.Date(2024, 12, 13, 20, 14, 0, 0, time.Local))

func builtInNote(note string) Note {
	return Note{
		Time: ogTime,
		Note: note,
	}
}
func altCollIdForint(i int) AlternateCollectionId { // TODO: FIX FOR uint8 overflow
	return [12]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, uint8(i)}
}

func altCollIdFieldForint(i int) AlternateCollectionIdField {
	return AlternateCollectionIdField{altCollIdForint(i)}
}

func mainCollIdForint(i int) MainCollectionId { // TODO: FIX FOR uint8 overflow
	return [8]byte{0, 0, 0, 0, 0, 0, 0, uint8(i)}
}

func mainCollIdFieldForint(i int) MainCollectionIdField {
	return MainCollectionIdField{mainCollIdForint(i)}
}
