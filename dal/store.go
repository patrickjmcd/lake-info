package dal

import (
	"context"
	"fmt"
	lakeinfov1 "github.com/patrickjmcd/lake-info/gen/lakeinfo/v1"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (m *Client) StoreLakeInfo(ctx context.Context, measurements []*lakeinfov1.LakeInfoMeasurement) error {
	coll := m.client.Database(dbName).Collection(collectionName)

	mm := bson.A{}
	skipped := 0
	for _, v := range measurements {
		measuredAt := primitive.NewDateTimeFromTime(v.MeasuredAt.AsTime())

		// Skip a record that is already stored for this lake and timestamp.
		// The collection is a MongoDB time-series collection, which does not
		// support unique indexes or upserts, so duplicate inserts (e.g. from a
		// re-triggered or retried job) are guarded against here. This is a
		// best-effort check-then-insert and is not atomic under true
		// concurrency; prevent concurrent runs at the orchestration layer.
		count, err := coll.CountDocuments(ctx, bson.M{
			"lakeName":   v.LakeName,
			"measuredAt": measuredAt,
		})
		if err != nil {
			return fmt.Errorf("checking for existing measurement: %w", err)
		}
		if count > 0 {
			skipped++
			continue
		}

		record := bson.M{
			"lakeName":            v.LakeName,
			"level":               v.Level,
			"temperature":         v.Temperature,
			"generation":          v.Generation,
			"totalReleaseRate":    v.TotalReleaseRate,
			"spillwayReleaseRate": v.SpillwayReleaseRate,
			"turbineReleaseRate":  v.TurbineReleaseRate,
			"measuredAt":          measuredAt,
			"createdAt":           primitive.NewDateTimeFromTime(v.CreatedAt.AsTime()),
		}
		mm = append(mm, record)
	}

	if len(mm) == 0 {
		fmt.Printf("No new documents to insert into lake-info.level (%d already present)\n", skipped)
		return nil
	}

	res, err := coll.InsertMany(ctx, mm)
	if err != nil {
		return err
	}
	fmt.Printf("Inserted %v documents into lake-info.level collection (skipped %d already present)!\n", len(res.InsertedIDs), skipped)
	return nil
}
