package dal

import (
	"context"
	"github.com/labstack/gommon/log"
	lakeinfov1 "github.com/patrickjmcd/lake-info/gen/lakeinfo/v1"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

var (
	ErrIncompleteTimeRange = status.Error(codes.InvalidArgument, "invalid time range, if specifying start or end time, you must specify both")
	ErrInvalidTimeRange    = status.Error(codes.InvalidArgument, "invalid time range, start time must be before end time")
	ErrLakeNameEmpty       = status.Error(codes.InvalidArgument, "lake name cannot be empty")
)

func timeEmpty(t time.Time) bool {
	return t.IsZero() || t.Equal(time.Unix(0, 0))
}

func (m *Client) GetLakeInfo(ctx context.Context, lakeName string, startTime, endTime time.Time, latest bool) ([]*lakeinfov1.LakeInfoMeasurement, error) {
	if lakeName == "" {
		return nil, ErrLakeNameEmpty
	}
	find := bson.M{
		"lakeName": lakeName,
	}

	if timeEmpty(startTime) && !timeEmpty(endTime) || !timeEmpty(startTime) && timeEmpty(endTime) {
		return nil, ErrIncompleteTimeRange
	}

	if !timeEmpty(startTime) && !timeEmpty(endTime) {
		if startTime.After(endTime) {
			return nil, ErrInvalidTimeRange
		}

		find["measuredAt"] = bson.M{
			"$gte": startTime,
			"$lte": endTime,
		}
	}

	log.Printf("Query: %+v", find)
	o := options.Find()
	o.SetSort(bson.D{{Key: "measuredAt", Value: -1}})
	if latest {
		o.SetLimit(1)
	}

	cur, err := m.client.Database(dbName).Collection(collectionName).Find(ctx, find, o)
	if err != nil {
		return nil, err
	}
	var docs []levelDocument
	if err := cur.All(ctx, &docs); err != nil {
		return nil, err
	}
	measurements := make([]*lakeinfov1.LakeInfoMeasurement, len(docs))
	for i, d := range docs {
		measurements[i] = d.toProto()
	}
	return measurements, nil
}
