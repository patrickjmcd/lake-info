package dal

import (
	"time"

	lakeinfov1 "github.com/patrickjmcd/lake-info/gen/lakeinfo/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// levelDocument is a measurement as stored in lake-info.level. Queries decode
// into it rather than straight into the protobuf type: the driver has no codec
// for *timestamppb.Timestamp, so the dates come back as time.Time and are
// converted here.
type levelDocument struct {
	LakeName            string    `bson:"lakeName"`
	Level               float64   `bson:"level"`
	Temperature         float64   `bson:"temperature"`
	Generation          float64   `bson:"generation"`
	TurbineReleaseRate  float64   `bson:"turbineReleaseRate"`
	SpillwayReleaseRate float64   `bson:"spillwayReleaseRate"`
	TotalReleaseRate    float64   `bson:"totalReleaseRate"`
	MeasuredAt          time.Time `bson:"measuredAt"`
	CreatedAt           time.Time `bson:"createdAt"`
}

func (d levelDocument) toProto() *lakeinfov1.LakeInfoMeasurement {
	return &lakeinfov1.LakeInfoMeasurement{
		LakeName:            d.LakeName,
		Level:               d.Level,
		Temperature:         d.Temperature,
		Generation:          d.Generation,
		TurbineReleaseRate:  d.TurbineReleaseRate,
		SpillwayReleaseRate: d.SpillwayReleaseRate,
		TotalReleaseRate:    d.TotalReleaseRate,
		MeasuredAt:          timestamp(d.MeasuredAt),
		CreatedAt:           timestamp(d.CreatedAt),
	}
}

// timestamp keeps a missing date nil, as the old protobuf codec did, instead
// of turning it into 0001-01-01.
func timestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}
