package dal

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Client struct {
	client *mongo.Client
	cancel context.CancelFunc
	uri    string
}

const dbName = "lake-info"
const collectionName = "level"

type ClientOption func(*Client)

func NewConnectionAsync(ctx context.Context, cancel context.CancelFunc, endpoint string, ch chan *Client) error {
	o := options.Client().
		ApplyURI(endpoint).
		SetServerSelectionTimeout(10 * time.Second)

	client, err := mongo.Connect(o)
	if err != nil {
		cancel()
		return fmt.Errorf("error connecting to mongodb: %w", err)
	}

	fmt.Println("Ping to mongodb")
	if err := client.Ping(ctx, nil); err != nil {
		cancel()
		return fmt.Errorf("error pinging mongodb: %w", err)
	}
	_, err = client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("error listing databases: %w", err)
	}

	cli := &Client{
		client: client,
		cancel: cancel,
		uri:    endpoint,
	}

	ch <- cli

	return nil
}

func (m *Client) Ping() error {
	return m.client.Ping(context.Background(), nil)
}

func (m *Client) Close() error {
	m.cancel()
	return m.client.Disconnect(context.Background())
}

func New(opts ...ClientOption) (*Client, error) {
	mongoURI := "mongodb://localhost:27017"
	ctx := context.Background()
	timeCtx, cancel := context.WithTimeout(ctx, 15*time.Second)

	if newURI, ok := os.LookupEnv("ATLAS_CONNECTION_URI"); ok {
		mongoURI = newURI
	}

	ch := make(chan *Client, 1)
	err := NewConnectionAsync(timeCtx, cancel, mongoURI, ch)
	if err != nil {
		fmt.Println("failed to connect to mongo")
		return nil, fmt.Errorf("failed to connect to mongo: %w", err)
	}

	select {
	case <-timeCtx.Done():
		fmt.Println("timeout connecting to mongo")
		return nil, fmt.Errorf("timeout connecting to mongo")
	case result := <-ch:
		return result, nil
	}
}
