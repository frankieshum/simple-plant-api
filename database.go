package main

import (
	"context"
	"log"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoDb struct {
	Driver         *mongo.Client
	DbName         string
	CollectionName string
}

func (db *MongoDb) Connect() error {
	log.Println("Connecting to MongoDB...")
	dbClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://127.0.0.1:27017/?maxPoolSize=20&w=majority")) // TODO config file
	if err != nil {
		return errors.Wrap(err, "MongoDB connect failed")
	}

	if err := dbClient.Ping(context.TODO(), readpref.Primary()); err != nil {
		return errors.Wrap(err, "MongoDB ping failed")
	}

	db.Driver = dbClient
	log.Println("Connected to MongoDB.")
	return nil
}

func (db *MongoDb) Disconnect() error {
	log.Println("Disconnecting from MongoDB...")
	if err := db.Driver.Disconnect(context.TODO()); err != nil {
		return err
	}
	log.Println("Disconnected from MongoDB.")
	return nil
}

func (db *MongoDb) GetAllPlants() ([]Plant, error) {
	// Get plants from DB
	log.Println("Finding all Plants in MongoDB")
	filter := bson.D{}
	collection := *db.Driver.Database(db.DbName).Collection(db.CollectionName)
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		if errors.As(err, &mongo.ErrNoDocuments) {
			log.Println("No Plants in database")
			return []Plant{}, nil
		}
		return []Plant{}, errors.Wrap(err, "MongoDB find failed")
	}

	// Decode all documents
	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		return []Plant{}, errors.Wrap(err, "MongoDB decode failed")
	}

	// Parse docs into Plant objects
	plants := make([]Plant, 0)
	for _, result := range results {
		var plant Plant
		if err := bsonToPlant(result, &plant); err != nil {
			return []Plant{}, errors.Wrap(err, "BSON to Plant conversion failed")
		}
		plants = append(plants, plant)
	}

	log.Println("Retrieved all Plants from MongoDB. Item count: ", len(plants))
	return plants, nil
}

func (db *MongoDb) GetPlantById(id int) (Plant, error) {
	// Get plant from DB
	log.Printf("Finding Plant in MongoDB with id %v...\n", id)
	filter := bson.D{{Key: "id", Value: id}}
	collection := *db.Driver.Database(db.DbName).Collection(db.CollectionName)
	var result bson.D
	err := collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		if errors.As(err, &mongo.ErrNoDocuments) {
			return Plant{}, &NotFoundError{}
		}
		return Plant{}, errors.Wrap(err, "MongoDB findOne failed")
	}

	// Parse DB results into Plant objects
	var plant Plant
	if err := bsonToPlant(result, &plant); err != nil {
		return Plant{}, errors.Wrap(err, "BSON to Plant conversion failed")
	}

	log.Printf("Retrieved Plant from MongoDB with id %v. Result: %v\n", id, plant.PrettyString())
	return plant, nil
}

func (db *MongoDb) CreatePlant(plant Plant) error {
	log.Printf("Inserting new Plant into MongoDB: %v\n", plant.PrettyString())

	newId, err := db.generateNewId()
	if err != nil {
		return err
	}
	plant.Id = newId

	// Convert Plant object into BSON doc
	_, doc, err := bson.MarshalValue(plant)
	if err != nil {
		return errors.Wrap(err, "Plant to BSON conversion failed")
	}

	// Insert plant into DB
	collection := *db.Driver.Database(db.DbName).Collection(db.CollectionName)
	result, err := collection.InsertOne(context.TODO(), doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return &ConflictError{ConflictingKey: "name", ConflictingValue: plant.Name}
		}
		return errors.Wrap(err, "MongoDB insertOne failed")
	}

	log.Println("Inserted Plant into MongoDB. _id: ", result.InsertedID)
	return nil
}

func (db *MongoDb) UpsertPlant(id int, plant Plant) error {
	log.Printf("Upserting Plant with id %v into MongoDB: %v\n", id, plant.PrettyString())

	// Convert Plant object into BSON doc
	_, doc, err := bson.MarshalValue(plant)
	if err != nil {
		return errors.Wrap(err, "Plant to BSON conversion failed")
	}

	// Upsert plant into DB
	collection := *db.Driver.Database(db.DbName).Collection(db.CollectionName)
	filter := bson.D{{Key: "id", Value: id}}
	options := options.Replace().SetUpsert(true)
	result, err := collection.ReplaceOne(context.TODO(), filter, doc, options)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return &ConflictError{ConflictingKey: "name", ConflictingValue: plant.Name}
		}
		return errors.Wrap(err, "MongoDB replaceOne failed")
	}

	log.Printf("Upserted Plant into MongoDB. ModifiedCount: %v, UpsertedCount: %v\n", result.ModifiedCount, result.UpsertedCount)
	return nil
}

func (db *MongoDb) DeletePlant(id int) error {
	log.Printf("Deleting Plant with id %v in MongoDB\n", id)

	collection := *db.Driver.Database(db.DbName).Collection(db.CollectionName)
	filter := bson.D{{Key: "id", Value: id}}
	opts := options.Delete().SetHint(bson.D{{Key: "id", Value: 1}})
	result, err := collection.DeleteOne(context.TODO(), filter, opts)
	if err != nil {
		return errors.Wrap(err, "MongoDB deleteOne failed")
	}

	log.Printf("Deleted Plant in MongoDB. DeletedCount: %v\n", result.DeletedCount)
	return nil
}

func bsonToPlant(result interface{}, plant *Plant) error {
	doc, err := bson.Marshal(result)
	if err != nil {
		return errors.Wrap(err, "BSON marshall failed")
	}
	if err := bson.Unmarshal(doc, &plant); err != nil {
		return errors.Wrap(err, "BSON unmarshall failed")
	}
	return nil
}

func (db *MongoDb) generateNewId() (int, error) {
	log.Println("Getting max ID from MongoDB")
	collection := *db.Driver.Database(db.DbName).Collection(db.CollectionName)
	filter := bson.D{}
	options := options.FindOne().SetSort(bson.D{{Key: "id", Value: -1}})
	var result bson.D
	err := collection.FindOne(context.TODO(), filter, options).Decode(&result)
	if err != nil {
		if errors.As(err, &mongo.ErrNoDocuments) {
			log.Println("No Plants in database. New ID is 1.")
			return 1, nil
		}
		return -1, errors.Wrap(err, "MongoDB findOne failed")
	}

	var plant Plant
	if err = bsonToPlant(result, &plant); err != nil {
		return -1, errors.Wrap(err, "BSON to Plant conversion failed")
	}

	newId := plant.Id + 1
	log.Printf("New plant ID: %v\n", newId)
	return newId, nil
}
