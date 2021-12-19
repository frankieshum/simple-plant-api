package main

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database interface {
	GetAllPlants() ([]Plant, error)
	GetPlantByName(name string) (Plant, error)
	CreatePlant(plant Plant) error
	UpsertPlant(name string, plant Plant) error
	DeletePlant(name string) error
	Disconnect()
}

type MongoDb struct {
	Driver         *mongo.Client
	DbName         string
	CollectionName string
}

func (db *MongoDb) Disconnect() { // TODO - nece4ssary?
	fmt.Println("Disconnecting from MongoDB...")
	if err := db.Driver.Disconnect(context.TODO()); err != nil {
		fmt.Printf("Error while disconnecting from MongoDB: %v\n", err)
	}
	fmt.Println("Disconnected from MongoDB.")
}

func (db *MongoDb) GetAllPlants() ([]Plant, error) {
	fmt.Println("Retrieving all plants from MongoDB...")

	// Get plants from DB
	filter := bson.D{}
	collection := *db.Driver.Database(db.DbName).Collection(db.CollectionName)
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("An error occurred during MongoDB find: ", err)
		return []Plant{}, err
	}

	// Decode all documents
	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		fmt.Println("An error occurred while decoding find results: ", err)
		return []Plant{}, err
	}

	// Parse docs into Plant objects
	plants := make([]Plant, 0)
	for _, result := range results {
		var plant Plant
		if err := bsonToPlant(result, &plant); err != nil {
			fmt.Println("An error occurred while parsing the DB result: ", err)
			return []Plant{}, err
		}
		plants = append(plants, plant)
	}

	fmt.Println("Retrieved all plants from MongoDB. Item count: ", len(plants))
	return plants, nil
}

func (db *MongoDb) GetPlantByName(name string) (Plant, error) {
	fmt.Printf("Retrieving plant from MongoDB with name '%v'...\n", name)

	// Get plant from DB
	filter := bson.D{{Key: "name", Value: name}}
	collection := *db.Driver.Database(db.DbName).Collection(db.CollectionName)
	var result bson.D
	err := collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		if errors.As(err, &mongo.ErrNoDocuments) {
			fmt.Println("No Plant found with name: ", name)
			return Plant{}, &NotFoundError{}
		}
		fmt.Println("An error occurred during MongoDB find: ", err)
		return Plant{}, err
	}

	// Parse DB results into Plant objects
	var plant Plant
	if err := bsonToPlant(result, &plant); err != nil {
		fmt.Println("An error occurred while parsing the DB result: ", err)
		return Plant{}, err
	}

	fmt.Printf("Retrieved plant from MongoDB with name '%v'. Result: %v\n", name, plant.PrettyString())
	return plant, nil
}

func (db *MongoDb) CreatePlant(plant Plant) error {
	fmt.Printf("Inserting plant into MongoDB: %v\n", plant.PrettyString())

	// Convert Plant object into BSON doc
	_, doc, err := bson.MarshalValue(plant)
	if err != nil {
		fmt.Println("An error occurred while converting the plant into BSON: ", err)
		return err
	}

	// Insert plant into DB
	collection := *db.Driver.Database(db.DbName).Collection(db.CollectionName)
	result, err := collection.InsertOne(context.TODO(), doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			fmt.Printf("A Plant with the name '%v' already exists in MongoDB\n", plant.Name)
			return &ConflictError{ConflictingKey: "name", ConflictingValue: plant.Name}
		}
		fmt.Println("An error occurred while inserting the plant into MongoDB: ", err)
		return err
	}

	fmt.Println("Inserted plant into MongoDB. _id: ", result.InsertedID)
	return nil
}

func (db *MongoDb) UpsertPlant(name string, plant Plant) error {
	fmt.Printf("Upserting plant with name '%v' into MongoDB: %v\n", name, plant.PrettyString())

	// Convert Plant object into BSON doc
	_, doc, err := bson.MarshalValue(plant)
	if err != nil {
		fmt.Println("An error occurred while converting the plant into BSON: ", err)
		return err
	}

	// Upsert plant into DB
	collection := *db.Driver.Database(db.DbName).Collection(db.CollectionName)
	filter := bson.D{{Key: "name", Value: name}}
	options := options.Replace().SetUpsert(true)
	result, err := collection.ReplaceOne(context.TODO(), filter, doc, options)
	if err != nil {
		fmt.Println("An error occurred while upserting the plant into MongoDB: ", err)
		return err
	}

	fmt.Printf("Upserted plant into MongoDB. ModifiedCount: %v, UpsertedCount: %v\n", result.ModifiedCount, result.UpsertedCount)
	return nil
}

func (db *MongoDb) DeletePlant(name string) error {
	fmt.Printf("Deleting plant with name '%v' in MongoDB\n", name)

	collection := *db.Driver.Database(db.DbName).Collection(db.CollectionName)
	filter := bson.D{{Key: "name", Value: name}}
	opts := options.Delete().SetHint(bson.D{{Key: "name", Value: 1}})
	result, err := collection.DeleteOne(context.TODO(), filter, opts)
	if err != nil {
		fmt.Println("An error occurred while deleting the plant in MongoDB: ", err)
		return err
	}

	fmt.Printf("Deleted plant in MongoDB. DeletedCount: %v\n", result.DeletedCount)
	return nil
}

func bsonToPlant(result interface{}, plant *Plant) error {
	doc, err := bson.Marshal(result)
	if err != nil {
		fmt.Println("Error while marshalling BSON into []byte: ", err)
		return err
	}
	if err := bson.Unmarshal(doc, &plant); err != nil {
		fmt.Println("Error while unmarshalling document into Plant: ", err)
		return err
	}
	return nil
}
