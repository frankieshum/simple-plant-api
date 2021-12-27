package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

type Api struct {
	Router *mux.Router
	DB     Database
}

func (api *Api) Initialise() {
	loadConfig()
	api.initialiseRouter()
	api.initialiseDatabase()
}

func (api *Api) Run() {
	exitCode := 0
	defer func() { os.Exit(exitCode) }()
	defer api.DB.Disconnect()

	if err := http.ListenAndServe(":8081", api.Router); err != nil {
		log.Println("Error while running API: ", err)
		exitCode = 1
		return
	}
}

func loadConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath("./config/")
	err := viper.ReadInConfig()
	if err != nil {
		log.Println("Error while loading config from file: ", err)
		os.Exit(1)
	}
}

func (api *Api) initialiseRouter() {
	api.Router = mux.NewRouter()

	api.Router.HandleFunc("/plants", api.listPlants).Methods("GET")
	api.Router.HandleFunc("/plants/{id}", api.getPlant).Methods("GET")
	api.Router.HandleFunc("/plants", api.postPlant).Methods("POST")
	api.Router.HandleFunc("/plants/{id}", api.putPlant).Methods("PUT")
	api.Router.HandleFunc("/plants/{id}", api.deletePlant).Methods("DELETE")
}

func (api *Api) initialiseDatabase() {
	dbName := viper.GetString("MongoDb.DbName")
	collectionName := viper.GetString("MongoDb.CollectionName")
	api.DB = &MongoDb{DbName: dbName, CollectionName: collectionName}
	if err := api.DB.Connect(); err != nil {
		log.Println("Error while connecting to MongoDB: ", err)
		os.Exit(1)
	}
}
