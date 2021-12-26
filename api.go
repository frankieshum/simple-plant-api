package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

type Api struct {
	Router *mux.Router
	DB     Database
}

func (api *Api) Initialise() {
	api.Router = mux.NewRouter()
	api.addRoutes()

	api.DB = &MongoDb{DbName: "plantsdb", CollectionName: "plants"} // TODO add to config file?
	if err := api.DB.Connect(); err != nil {
		log.Println("Error while connecting to MongoDB: ", err)
		os.Exit(1)
	}
}

func (api *Api) Run() {
	defer api.DB.Disconnect() // TODO exit gracefully to allow deferred func to run
	log.Fatal(http.ListenAndServe(":8081", api.Router))
}

func (api *Api) addRoutes() {
	api.Router.HandleFunc("/plants", api.listPlants).Methods("GET")
	api.Router.HandleFunc("/plants/{id}", api.getPlant).Methods("GET")
	api.Router.HandleFunc("/plants", api.postPlant).Methods("POST")
	api.Router.HandleFunc("/plants/{id}", api.putPlant).Methods("PUT")
	api.Router.HandleFunc("/plants/{id}", api.deletePlant).Methods("DELETE")
}
