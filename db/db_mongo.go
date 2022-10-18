package db

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/weebank/pucci"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoDB Database Service
type MongoDatabaseService struct {
	client *mongo.Client
}

// Instance new MongoDB service
func NewMongoDatabaseService() DatabaseService {
	return &MongoDatabaseService{}
}

// Establish a connection to MongoDB Cluster
func (s *MongoDatabaseService) Connect() (context.Context, context.CancelFunc) {
	// Get URI from .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	// Create new client
	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		log.Fatal(err)
	}

	// Create context
	ctx, cancel := context.WithCancel(context.Background())

	// Connect
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Test connection
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
	log.Println("database initialized successfully")

	// Set client on service
	s.client = client

	return ctx, cancel
}

// Establish a connection to MongoDB Cluster
func (s *MongoDatabaseService) Disconnect(ctx context.Context, cancel context.CancelFunc) {
	// Cancel context
	s.client.Disconnect(ctx)
	cancel()
}

// Create a document on database
func (s *MongoDatabaseService) Create(ctx context.Context, database, table, id string, doc any) (string, error) {
	// Get database and collection
	db := s.client.Database(database)
	col := db.Collection(table)

	// Convert document to map
	b, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}
	m := make(map[string]interface{})
	json.Unmarshal(b, &m)

	// Insert ID
	if objectID, err := primitive.ObjectIDFromHex(id); err == nil {
		m["_id"] = objectID
	}

	// Insert into database
	res, err := col.InsertOne(ctx, m)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return "", pucci.ErrorDuplicatedID
		}
		return "", err
	}

	// Return ID
	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

// Read a document from database
func (s *MongoDatabaseService) Read(ctx context.Context, database, table string, filter map[string]interface{}, to any) (string, error) {
	// Get database and collection
	db := s.client.Database(database)
	col := db.Collection(table)

	// Find document
	res := col.FindOne(ctx, filter)
	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			return primitive.NilObjectID.Hex(), pucci.ErrorItemDoesNotExist
		}
		return primitive.NilObjectID.Hex(), res.Err()
	}

	// Obtain ID
	type ID struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	var id ID
	err := res.Decode(&id)
	if err != nil {
		return primitive.NilObjectID.Hex(), err
	}

	// Decode document
	return id.ID.Hex(), res.Decode(to)
}

// Read a document from database by its ID
func (s *MongoDatabaseService) ReadByID(ctx context.Context, database, table, id string, to any) error {
	// Get database and collection
	db := s.client.Database(database)
	col := db.Collection(table)

	// Validate ID
	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// Find document
	res := col.FindOne(ctx, bson.M{"_id": docID})
	if res.Err() != nil {
		return res.Err()
	}

	// Decode document
	return res.Decode(to)
}

// Update a document on database
func (s *MongoDatabaseService) Update(ctx context.Context, database, table string, filter map[string]interface{}, doc any) error {
	// Get database and collection
	db := s.client.Database(database)
	col := db.Collection(table)

	// Convert document to map
	b, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	m := make(map[string]interface{})
	json.Unmarshal(b, &m)

	// Find document and update
	res := col.FindOneAndReplace(ctx, filter, doc)
	if res.Err() == mongo.ErrNoDocuments {
		return pucci.ErrorItemDoesNotExist
	}

	// Return error
	return res.Err()
}

// Update a document on database
func (s *MongoDatabaseService) Delete(ctx context.Context, database, table, id string) error {
	// Get database and collection
	db := s.client.Database(database)
	col := db.Collection(table)

	// Delete document
	_, err := col.DeleteOne(ctx, bson.M{"_id": id})
	if err == mongo.ErrNoDocuments {
		return pucci.ErrorItemDoesNotExist
	}

	// Return error
	return err
}
