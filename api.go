package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Api struct {
	Router *mux.Router
	DB     Database
}

func (api *Api) Initialise() {
	api.Router = mux.NewRouter()
	api.addRoutes()

	api.DB = &MongoDb{DbName: "plantsdb", CollectionName: "plants"} // TODO env var?
	api.DB.Connect()
}

func (api *Api) Run() {
	defer api.DB.Disconnect()
	log.Fatal(http.ListenAndServe(":8081", api.Router))
}

func (api *Api) addRoutes() {
	api.Router.HandleFunc("/plants", api.listPlants).Methods("GET")
	api.Router.HandleFunc("/plants/{id}", api.getPlant).Methods("GET")
	api.Router.HandleFunc("/plants", api.postPlant).Methods("POST")
	api.Router.HandleFunc("/plants/{id}", api.putPlant).Methods("PUT")
	api.Router.HandleFunc("/plants/{id}", api.deletePlant).Methods("DELETE")
}
