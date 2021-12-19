package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Api struct {
	Router *mux.Router // TODO interface this?
	DB     Database
}

func (api *Api) Initialise() {
	api.Router = mux.NewRouter()
	api.addRoutes()
	api.connectToDb()
}

func (api *Api) Run() {
	defer api.DB.Disconnect() // Is this necessary?
	log.Fatal(http.ListenAndServe(":8081", api.Router))
}

func (api *Api) addRoutes() {
	api.Router.HandleFunc("/plants", api.listPlants).Methods("GET")
	api.Router.HandleFunc("/plants/{name}", api.getPlant).Methods("GET")
	api.Router.HandleFunc("/plants", api.postPlant).Methods("POST")
	api.Router.HandleFunc("/plants/{name}", api.putPlant).Methods("PUT")
	api.Router.HandleFunc("/plants/{name}", api.deletePlant).Methods("DELETE")
}

func (api *Api) connectToDb() {
	fmt.Println("Connecting to MongoDB...")
	db := MongoDb{DbName: "plantsdb", CollectionName: "plants"}                                                                       // TODO rename DB and env var this
	dbClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://127.0.0.1:27017/?maxPoolSize=20&w=majority")) // TODO parameterise
	if err != nil {
		log.Fatal(fmt.Sprintf("An error occurred while connecting to MongoDB: %v", err))
	}

	if err := dbClient.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(fmt.Sprintf("Error: failed to ping MongoDB: %v", err))
	}
	fmt.Println("Connected to MongoDB.")

	db.Driver = dbClient
	api.DB = &db
}
