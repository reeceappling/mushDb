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

func init() { // TODO: remove this block!!!!!
	if idTestPlateAdminOnly >= 255 {
		panic("test values 1 go too high")
	}
	if idTestGrainBatch >= 255 {
		panic("test values 2 go too high")
	}
}

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
	//temp := i // TODO: switch and check if we start going above 255!
	//out := [12]byte{0, 0, 0, 0, 0, 0, 0, 0}
	//for j := 0; j<12; j++ {
	//	out[12-j] = byte(temp % 8)
	//	temp = temp/8
	//	if temp == 0 {
	//		return out
	//	}
	//}
	//return out
	return [12]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, uint8(i)}
}

func altCollIdFieldForint(i int) AlternateCollectionIdField {
	return AlternateCollectionIdField{altCollIdForint(i)}
}

func mainCollIdForint(i int) MainCollectionId { // TODO: FIX FOR uint8 overflow
	//temp := i // TODO: switch and check if we start going above 255!
	//out := [RfidByteSize]byte{0, 0, 0, 0, 0, 0, 0, 0}
	//for j := 0; j<RfidByteSize; j++ {
	//	out[8-j] = byte(temp % 8)
	//	temp = temp/8
	//	if temp == 0 {
	//		return out
	//	}
	//}
	//return out
	return [RfidByteSize]byte{0, 0, 0, 0, 0, 0, 0, uint8(i)} // TODO: reenable if the above does not work! TODO: ENSURE TO REENABLE IF NEEDED!
}
