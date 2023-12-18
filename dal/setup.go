package dal

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (m *Client) Setup(ctx context.Context) error {
	createCollectionOptions := options.CreateCollection()
	createCollectionOptions.SetTimeSeriesOptions(options.TimeSeries().SetTimeField("measuredAt").SetMetaField("lakeName").SetGranularity("hours"))
	err := m.client.Database(dbName).CreateCollection(ctx, collectionName, createCollectionOptions)
	if err != nil {
		return err
	}

	return nil
}
