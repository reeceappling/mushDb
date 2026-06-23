package api

import (
	"github.com/reeceappling/mushDb/api/request/unix"
	"time"
)

const ( // MainCollection test ints
	idTestPlate = iota // Blanket write
	idTestBag
	idTestFC
	idTestJar
	idTestMSS
	idTestSlant
	idTestStasis
	idTestLC
	idTestSp
	idTestFruit
	idTestSwab
	idTestPlug
	idTestPlateBlanketRead
	idTestPlateProjectWrite
	idTestPlateProjectRead
	idTestPlateUserWriteProjRead
	idTestPlateUserOutsideProject
	idTestLCS
	idTestWaterJar
	idTestPlateEmpty
	idTestLC2
	idTestPlateBlanketWrite
	idTestPlateAdminOnly // No blanket permission, no users, no projects
)

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

	idExampleTransfer

	idTestBottle
	idTestBatch

	idTestGrainBatch
)

var ogTime = unix.TimeFor(time.Date(2024, 12, 13, 20, 14, 0, 0, time.Local))

func builtInNote(note string) Note {
	return Note{
		RequiredTimeField: newRequiredTimeField(ogTime),
		Note:              note,
	}
}
func altCollIdForint(i int) AlternateCollectionId { // TODO: FIX FOR uint8 overflow
	return [12]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, uint8(i)}
}

func altCollIdFieldForint(i int) AlternateCollectionIdField {
	return AlternateCollectionIdField{altCollIdForint(i)}
}

func mainCollIdForint(i int) MainCollectionId { // TODO: FIX FOR uint8 overflow
	return [RfidByteSize]byte{0, 0, 0, 0, 0, 0, 0, uint8(i)}
}

func mainCollIdFieldForint(i int) MainCollectionIdField {
	return MainCollectionIdField{mainCollIdForint(i)}
}
