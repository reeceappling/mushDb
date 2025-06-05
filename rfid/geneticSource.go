package rfid

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	// Alts
	// SporePrint
	_ geneticSource = Fruit{} // TODO: only clones via transfer
)

type GeneticParentInfo struct {
	Species               *string     `json:"species"`
	Subspecies            *string     `json:"subspecies"`
	KnownFruitable        *bool       `json:"knownFruitable"`
	GensSinceSpore        *Generation `json:"gensSinceSpore"`
	GensSinceFruitOrSpore *Generation `json:"gensSinceFruitOrSpore"`
}

type geneticSource interface {
	SourceType() string
	GeneticInfoAsParent() (GeneticParentInfo, error)
	DbId() string
	projects() []string
	setTransferParent(ctx mongo.SessionContext, xfer Transfer) error
	setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error
	children(ctx context.Context) ([]geneticSource, error) // TODO: DELETE ALL?
	generation() (sinceSpore *Generation, sinceSporeOrClone *Generation)
}

type hasImages interface {
	// TODO: THIS!!!! ----------------------
}
