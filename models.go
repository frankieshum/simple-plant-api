package main

import (
	"fmt"
	"strings"
)

// --------------- Plant models ---------------

type PlantRequest struct {
	Name       string   `json:"name"`
	OtherNames []string `json:"otherNames"`
	Light      string   `json:"light"`
	Humidity   string   `json:"humidity"`
	Water      string   `json:"water"`
}

func (plant *PlantRequest) Validate() []string {
	results := make([]string, 0)
	if len(plant.Name) == 0 {
		results = append(results, "The name value is required")
	}
	if len(plant.Light) == 0 {
		results = append(results, "The light value is required")
	}
	if len(plant.Humidity) == 0 {
		results = append(results, "The humidity value is required")
	}
	if len(plant.Water) == 0 {
		results = append(results, "The water value is required")
	}
	return results
}

type Plant struct {
	Id         int      `json:"id"`
	Name       string   `json:"name"`
	OtherNames []string `json:"otherNames"`
	Light      string   `json:"light"`
	Humidity   string   `json:"humidity"`
	Water      string   `json:"water"`
}

func (plant *Plant) PrettyString() string {
	return fmt.Sprintf("Id: %v, Name: %v, OtherNames: [%v], Light: %v, Humidity: %v, Water: %v",
		plant.Id, plant.Name, strings.Join(plant.OtherNames, ", "), plant.Light, plant.Humidity, plant.Water)
}

// --------------- Database ---------------

type Database interface {
	GetAllPlants() ([]Plant, error)
	GetPlantById(id int) (Plant, error)
	CreatePlant(plant Plant) error
	UpsertPlant(id int, plant Plant) error
	DeletePlant(id int) error
	Connect() error
	Disconnect() error
}

// --------------- Errors ---------------

type NotFoundError struct{}

func (err *NotFoundError) Error() string {
	return "specified record was not found"
}

type ConflictError struct {
	ConflictingKey   string
	ConflictingValue string
}

func (err *ConflictError) Error() string {
	return fmt.Sprintf("the request conflicts with the target resource (conflicting key: %v, conflicting value: %v)",
		err.ConflictingKey, err.ConflictingValue)
}
