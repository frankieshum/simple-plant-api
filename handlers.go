package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

func (api *Api) listPlants(w http.ResponseWriter, r *http.Request) {
	log.Printf("GET %v\n", r.RequestURI)

	plants, err := api.DB.GetAllPlants()
	if err != nil {
		log.Printf("Error: %v\n", err)
		writeErrorResponse(w, 500, "An error occurred while processing the request")
		return
	}
	writeResponse(w, 200, plants)
}

func (api *Api) getPlant(w http.ResponseWriter, r *http.Request) {
	log.Printf("GET %v\n", r.RequestURI)

	// Retrieve plant ID
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		log.Printf("Plant Id '%v' is not an integer", id)
		writeErrorResponse(w, 400, "The Plant id must be an integer")
		return
	}

	plant, err := api.DB.GetPlantById(id)
	if err != nil {
		if errors.Is(err, &NotFoundError{}) {
			log.Println("The specified Plant was not found")
			writeErrorResponse(w, 404, "The specified Plant was not found")
			return
		}
		log.Printf("Error: %v\n", err)
		writeErrorResponse(w, 500, "An error occurred while processing the request")
		return
	}
	writeResponse(w, 200, plant)
}

func (api *Api) postPlant(w http.ResponseWriter, r *http.Request) {
	log.Printf("POST %v\n", r.RequestURI)

	// Read body and parse into Plant
	plantRequest := PlantRequest{}
	if err := json.NewDecoder(r.Body).Decode(&plantRequest); err != nil {
		log.Printf("The request body could not be parsed into a Plant: %v", err)
		writeErrorResponse(w, 400, "The request payload could not be parsed into a Plant")
		return
	}

	// Validate the request
	if validationResults := plantRequest.Validate(); len(validationResults) > 0 {
		log.Println("The Plant request is invalid: ", strings.Join(validationResults, "; "))
		writeErrorResponse(w, 400, strings.Join(validationResults, "; "))
		return
	}

	newPlant := Plant{
		Name:       plantRequest.Name,
		OtherNames: plantRequest.OtherNames,
		Humidity:   plantRequest.Humidity,
		Light:      plantRequest.Light,
		Water:      plantRequest.Water,
	}
	if err := api.DB.CreatePlant(newPlant); err != nil {
		var conflictErr *ConflictError
		if errors.As(err, &conflictErr) {
			errMsg := fmt.Sprintf("Plant with %v '%v' already exists", conflictErr.ConflictingKey, conflictErr.ConflictingValue)
			log.Println(errMsg)
			writeErrorResponse(w, 409, errMsg)
			return
		}
		log.Printf("Error: %v\n", err)
		writeErrorResponse(w, 500, "An error occurred while processing the request")
		return
	}
	writeResponse(w, 201, map[string]string{})
}

func (api *Api) putPlant(w http.ResponseWriter, r *http.Request) {
	log.Printf("PUT %v\n", r.RequestURI)

	// Read body and parse into Plant
	plantRequest := PlantRequest{}
	if err := json.NewDecoder(r.Body).Decode(&plantRequest); err != nil {
		log.Printf("The request body could not be parsed into a Plant: %v", err)
		writeErrorResponse(w, 400, "The request payload could not be parsed into a Plant")
		return
	}

	// Get ID and validate the request
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		log.Printf("Plant Id '%v' is not an integer", id)
		writeErrorResponse(w, 400, "The Plant id must be an integer")
		return
	}
	if validationResults := plantRequest.Validate(); len(validationResults) > 0 {
		log.Println("The Plant request is invalid: ", strings.Join(validationResults, "; "))
		writeErrorResponse(w, 400, strings.Join(validationResults, "; "))
		return
	}

	newPlant := Plant{
		Id:         id,
		Name:       plantRequest.Name,
		OtherNames: plantRequest.OtherNames,
		Humidity:   plantRequest.Humidity,
		Light:      plantRequest.Light,
		Water:      plantRequest.Water,
	}
	if err = api.DB.UpsertPlant(id, newPlant); err != nil {
		var conflictErr *ConflictError
		if errors.As(err, &conflictErr) {
			errMsg := fmt.Sprintf("Plant with %v '%v' already exists", conflictErr.ConflictingKey, conflictErr.ConflictingValue)
			log.Println(errMsg)
			writeErrorResponse(w, 409, errMsg)
			return
		}
		log.Printf("Error: %v\n", err)
		writeErrorResponse(w, 500, "An error occurred while processing the request")
		return
	}
	writeResponse(w, 200, map[string]string{})
}

func (api *Api) deletePlant(w http.ResponseWriter, r *http.Request) {
	log.Printf("DELETE %v\n", r.RequestURI)

	// Retrieve plant ID
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		log.Printf("Plant Id '%v' is not an integer", id)
		writeErrorResponse(w, 400, "The Plant id must be an integer")
		return
	}

	if err := api.DB.DeletePlant(id); err != nil {
		log.Printf("Error: %v\n", err)
		writeErrorResponse(w, 500, "An error occurred while deleting the plant")
		return
	}
	writeResponse(w, 204, map[string]string{})
}

func writeResponse(w http.ResponseWriter, httpStatusCode int, responseBody interface{}) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(httpStatusCode)
	json.NewEncoder(w).Encode(responseBody)
}

func writeErrorResponse(w http.ResponseWriter, httpStatusCode int, errorMessage string) {
	resp := ErrorResponse{
		Error: errorMessage,
	}
	writeResponse(w, httpStatusCode, resp)
}
