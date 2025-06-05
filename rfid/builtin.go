package rfid

import "time"

const (
	idLmea = iota
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
)

var ogTime = unixTime(time.Date(2024, 12, 13, 20, 14, 0, 0, time.Local).UnixMilli())

func builtInNote(note string) Note {
	return Note{
		Time: ogTime,
		Note: note,
	}
}
func altCollIdForint(i int) alternateCollectionId {
	z := byte(0)
	return [12]byte{z, z, z, z, z, z, z, z, z, z, z, uint8(i)}
}

func mainCollIdForint(i int) MainCollectionId {
	z := byte(0)
	return [8]byte{z, z, z, z, z, z, z, uint8(i)}
}
