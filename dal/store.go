package dal

import (
	"context"
	"fmt"
	lakeinfov1 "github.com/patrickjmcd/lake-info/gen/lakeinfo/v1"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (m *Client) StoreLakeInfo(ctx context.Context, measurements []*lakeinfov1.LakeInfoMeasurement) error {

	mm := bson.A{}
	for _, v := range measurements {
		record := bson.M{
			"lakeName":            v.LakeName,
			"level":               v.Level,
			"temperature":         v.Temperature,
			"totalReleaseRate":    v.TotalReleaseRate,
			"spillwayReleaseRate": v.SpillwayReleaseRate,
			"turbineReleaseRate":  v.TurbineReleaseRate,
			"measuredAt":          primitive.NewDateTimeFromTime(v.MeasuredAt.AsTime()),
			"createdAt":           primitive.NewDateTimeFromTime(v.CreatedAt.AsTime()),
		}
		mm = append(mm, record)
	}
	res, err := m.client.Database(dbName).Collection(collectionName).InsertMany(ctx, mm)
	if err != nil {
		return err
	}
	fmt.Printf("Inserted %v documents into lake-info.level collection!\n", len(res.InsertedIDs))
	return nil
}
