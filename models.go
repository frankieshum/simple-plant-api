package main

import (
	"fmt"
	"strings"
)

// TODO - request and response models?

type PostPlantRequest struct {
	Name       string   `json:"name"`
	OtherNames []string `json:"otherNames"`
	Light      string   `json:"light"`
	Humidity   string   `json:"humidity"`
	Water      string   `json:"water"`
}

func (plant *PostPlantRequest) Validate() []string {
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

type PutPlantRequest struct {
	OtherNames []string `json:"otherNames"`
	Light      string   `json:"light"`
	Humidity   string   `json:"humidity"`
	Water      string   `json:"water"`
}

func (plant *PutPlantRequest) Validate() []string {
	results := make([]string, 0)
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
	Name       string   `json:"name"`
	OtherNames []string `json:"otherNames"`
	Light      string   `json:"light"`
	Humidity   string   `json:"humidity"`
	Water      string   `json:"water"`
}

func (plant *Plant) PrettyString() string {
	return fmt.Sprintf("Name: %v, OtherNames: [%v], Light: %v, Humidity: %v, Water: %v",
		plant.Name, strings.Join(plant.OtherNames, ", "), plant.Light, plant.Humidity, plant.Water)
}

// --------------------------------- Errors ---------------------------------

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
