package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type DBHandler interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Client() *mongo.Client
	Database() *mongo.Database
	Collection(name string) *mongo.Collection
	InsertOne(ctx context.Context, coll string, doc interface{}) (*mongo.InsertOneResult, error)
	FindOne(ctx context.Context, coll string, filter interface{}, result interface{}) error
	FindMany(ctx context.Context, coll string, filter interface{}) (*mongo.Cursor, error)
	UpdateOne(ctx context.Context, coll string, filter interface{}, update interface{}) (*mongo.UpdateResult, error)
	DeleteOne(ctx context.Context, coll string, filter interface{}) (*mongo.DeleteResult, error)
}

type MongoHandler struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewMongoHandler() *MongoHandler {
	return &MongoHandler{}
}

func (m *MongoHandler) Connect(ctx context.Context) error {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = os.Getenv("MONGODB_URI")
	}
	if uri == "" {
		return errors.New("MONGO_URI or MONGODB_URI not set in environment or .env")
	}

	clientOpts := options.Client().ApplyURI(uri)
	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(clientOpts)
	if err != nil {
		return fmt.Errorf("mongo connect: %w", err)
	}

	if err := client.Ping(cctx, nil); err != nil {
		return fmt.Errorf("mongo ping: %w", err)
	}

	dbName := os.Getenv("MONGO_DB")
	if dbName == "" {
		dbName = "tuubaa"
	}

	m.client = client
	m.db = client.Database(dbName)
	return nil
}

func (m *MongoHandler) Disconnect(ctx context.Context) error {
	if m.client == nil {
		return nil
	}
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return m.client.Disconnect(cctx)
}

func (m *MongoHandler) Client() *mongo.Client                    { return m.client }
func (m *MongoHandler) Database() *mongo.Database                { return m.db }
func (m *MongoHandler) Collection(name string) *mongo.Collection { return m.db.Collection(name) }

func (m *MongoHandler) InsertOne(ctx context.Context, coll string, doc interface{}) (*mongo.InsertOneResult, error) {
	return m.Collection(coll).InsertOne(ctx, doc)
}

func (m *MongoHandler) FindOne(ctx context.Context, coll string, filter interface{}, result interface{}) error {
	return m.Collection(coll).FindOne(ctx, filter).Decode(result)
}

func (m *MongoHandler) FindMany(ctx context.Context, coll string, filter interface{}) (*mongo.Cursor, error) {
	return m.Collection(coll).Find(ctx, filter)
}

func (m *MongoHandler) UpdateOne(ctx context.Context, coll string, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	return m.Collection(coll).UpdateOne(ctx, filter, update)
}

func (m *MongoHandler) DeleteOne(ctx context.Context, coll string, filter interface{}) (*mongo.DeleteResult, error) {
	return m.Collection(coll).DeleteOne(ctx, filter)
}

var globalDB *mongo.Database

func InitDB() error {
	h := NewMongoHandler()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := h.Connect(ctx); err != nil {
		return err
	}
	globalDB = h.db
	return nil
}

func DB() *mongo.Database {
	return globalDB
}

func (m *MongoHandler) EnsureConnected() error {
	if m.client == nil {
		return errors.New("mongo client not connected; call Connect")
	}
	return nil
}
