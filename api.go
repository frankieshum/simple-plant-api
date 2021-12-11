package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Api struct {
	Router *mux.Router // TODO interface this?
	DB     PlantsRepository
}

func (api *Api) Initialise() {
	api.Router = mux.NewRouter()
	api.addRoutes()
	api.setupDb()
}

func (api *Api) Run() {
	log.Fatal(http.ListenAndServe(":8081", api.Router))
}

func (api *Api) addRoutes() {
	api.Router.HandleFunc("/plants", api.getAllPlants).Methods("GET")
	api.Router.HandleFunc("/plants/{name}", api.getPlant).Methods("GET")
	api.Router.HandleFunc("/plants", api.postPlant).Methods("POST")
	api.Router.HandleFunc("/plants/{name}", api.putPlant).Methods("PUT")
	api.Router.HandleFunc("/plants/{name}", api.deletePlant).Methods("DELETE")
}

func (api *Api) setupDb() {
	db := PlantsMongoDb{}
	dbClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://127.0.0.1:27017/?maxPoolSize=20&w=majority")) // TODO parameterise
	if err != nil {
		panic(err) // TODO
	}

	defer func() {
		dbClient.Disconnect(context.TODO())
	}()

	if err := dbClient.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err) // TODO
	}

	db.Driver = dbClient
	api.DB = &db
}

// ---------------- ENDPOINTS ----------------

func (api *Api) getAllPlants(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("GET %v\n", r.RequestURI)

	plants, err := api.DB.GetAllPlants()
	if err != nil {
		writeError(w, 500, "An error occurred while getting the plants")
		return
	}
	writeResponse(w, 200, plants)
}

func (api *Api) getPlant(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("GET %v\n", r.RequestURI)

	name := mux.Vars(r)["name"]
	plant, err := api.DB.GetPlantByName(name)
	if err != nil {
		writeError(w, 500, "An error occurred while getting the plant")
		return
	}
	if plant.Name == "" {
		writeError(w, 404, "The specified plant was not found")
		return
	}
	writeResponse(w, 200, plant)
}

func (api *Api) postPlant(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("POST %v\n", r.RequestURI)

	// Read payload
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("An error occurred while reading the request body: %v\n", err)
		writeError(w, 500, "An error occurred while adding the plant")
		return
	}

	// Parse body into Plant
	newPlant := Plant{}
	if err = json.Unmarshal(body, &newPlant); err != nil {
		fmt.Printf("The request body could not be parsed into a Plant: %v", err)
		writeError(w, 400, "The plant payload is in invalid format")
		return
	}

	// Add the plant
	if err = api.DB.CreatePlant(newPlant); err != nil {
		writeError(w, 500, "An error occurred while adding the plant")
		return
	}
	writeResponse(w, 200, map[string]string{})
}

func (api *Api) putPlant(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("PUT %v\n", r.RequestURI)

	// Read payload
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("An error occurred while reading the request body: %v\n", err)
		writeError(w, 500, "An error occurred while updating the plant")
		return
	}

	// Parse body into Plant
	newPlant := Plant{}
	if err = json.Unmarshal(body, &newPlant); err != nil {
		fmt.Printf("The request body could not be parsed into a Plant: %v", err)
		writeError(w, 400, "The plant payload is in invalid format")
		return
	}

	// Update the plant
	if err = api.DB.UpdatePlant(newPlant); err != nil {
		writeError(w, 500, "An error occurred while updating the plant")
		return
	}
	writeResponse(w, 200, map[string]string{})
}

func (api *Api) deletePlant(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("DELETE %v\n", r.RequestURI)

	name := mux.Vars(r)["name"]
	if err := api.DB.DeletePlant(name); err != nil {
		writeError(w, 500, "An error occurred while deleting the plant")
		return
	}
	writeResponse(w, 200, map[string]string{})
}

func writeResponse(w http.ResponseWriter, httpStatusCode int, responseBody interface{}) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(httpStatusCode)
	json.NewEncoder(w).Encode(responseBody)
}

func writeError(w http.ResponseWriter, httpStatusCode int, errorMessage string) {
	errorResponse := map[string]string{
		"error": errorMessage,
	}
	writeResponse(w, httpStatusCode, errorResponse)
}
