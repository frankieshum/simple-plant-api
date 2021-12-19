package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func (api *Api) listPlants(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("GET %v\n", r.RequestURI)

	plants, err := api.DB.GetAllPlants()
	if err != nil {
		writeError(w, 500, "An error occurred while processing the request")
		return
	}
	writeResponse(w, 200, plants)
}

func (api *Api) getPlant(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("GET %v\n", r.RequestURI)

	name := mux.Vars(r)["name"]
	plant, err := api.DB.GetPlantByName(name)
	if err != nil {
		if errors.Is(err, &NotFoundError{}) {
			writeError(w, 404, "The specified plant was not found")
			return
		}
		writeError(w, 500, "An error occurred while processing the request")
		return
	}
	writeResponse(w, 200, plant)
}

func (api *Api) postPlant(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("POST %v\n", r.RequestURI)

	// Read payload and parse body into Plant
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("An error occurred while reading the request body: %v\n", err)
		writeError(w, 500, "An error occurred while processing the request")
		return
	}
	newPlantRequest := PostPlantRequest{}
	if err = json.Unmarshal(body, &newPlantRequest); err != nil {
		fmt.Printf("The request body could not be parsed into a Plant: %v", err)
		writeError(w, 400, "The request payload could not be parsed into a Plant")
		return
	}

	// Validate the request
	if validationResults := newPlantRequest.Validate(); len(validationResults) > 0 {
		fmt.Println("The Plant request is invalid: ", strings.Join(validationResults, "; "))
		writeError(w, 400, strings.Join(validationResults, "; "))
		return
	}

	// Add the plant
	if err = api.DB.CreatePlant(Plant(newPlantRequest)); err != nil {
		var conflictErr *ConflictError
		if errors.As(err, &conflictErr) {
			writeError(w, 409, fmt.Sprintf("Plant with %v '%v' already exists", conflictErr.ConflictingKey, conflictErr.ConflictingValue))
			return
		}
		writeError(w, 500, "An error occurred while processing the request")
		return
	}
	writeResponse(w, 201, map[string]string{})
}

func (api *Api) putPlant(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("PUT %v\n", r.RequestURI)

	// Read payload and parse body into Plant
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("An error occurred while reading the request body: %v\n", err)
		writeError(w, 500, "An error occurred while processing the request")
		return
	}
	newPlantRequest := PutPlantRequest{}
	if err = json.Unmarshal(body, &newPlantRequest); err != nil {
		fmt.Printf("The request body could not be parsed into a Plant: %v", err)
		writeError(w, 400, "The request payload could not be parsed into a Plant")
		return
	}

	// Update the plant
	name := mux.Vars(r)["name"]
	newPlant := Plant{Name: name, OtherNames: newPlantRequest.OtherNames, Humidity: newPlantRequest.Humidity, Light: newPlantRequest.Light, Water: newPlantRequest.Water} // TODO use a mapper?
	if err = api.DB.UpsertPlant(name, newPlant); err != nil {
		writeError(w, 500, "An error occurred while processing the request")
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
	writeResponse(w, 204, map[string]string{})
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
