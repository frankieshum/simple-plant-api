package main

import (
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
)

type Plant struct {
	Name       string   `json:"name"`
	OtherNames []string `json:"otherNames"`
	Light      string   `json:"light"`
	Humidity   string   `json:"humidity"`
	Water      string   `json:"water"`
}

type PlantsRepository interface {
	GetAllPlants() ([]Plant, error)
	GetPlantByName(name string) (Plant, error)
	CreatePlant(plant Plant) error
	UpdatePlant(plant Plant) error
	DeletePlant(name string) error
}

type PlantsMongoDb struct {
	Driver *mongo.Client
}

func (db *PlantsMongoDb) GetAllPlants() ([]Plant, error) {
	// TODO - implement this
	plants := []Plant{
		{
			Name:       "Monstera Deliciosa",
			OtherNames: []string{"Swiss Cheese Plant"},
			Light:      "Bright Indirect",
			Humidity:   "High",
			Water:      "Moderate",
		},
		{
			Name:       "Ficus Elastica",
			OtherNames: []string{"Rubber Plant", "Rubber Tree", "Rubber Fig"},
			Light:      "Bright Indirect",
			Humidity:   "Moderate",
			Water:      "Moderate",
		},
		{
			Name:     "Aloe Vera",
			Light:    "Bright Indirect",
			Humidity: "Low",
			Water:    "Low",
		},
	}
	return plants, nil
}

func (db *PlantsMongoDb) GetPlantByName(name string) (Plant, error) {
	// TODO - implement this
	if name == "unknown plant" {
		return Plant{}, nil
	}
	if name == "error plant" {
		return Plant{}, errors.New("plant db blew up")
	}
	plant := Plant{
		Name:     name,
		Light:    "Bright Indirect",
		Humidity: "High",
		Water:    "Moderate",
	}
	return plant, nil
}

func (db *PlantsMongoDb) CreatePlant(plant Plant) error {
	// TODO - implement this
	if plant.Name == "error plant" {
		return errors.New("plant db blew up")
	}
	return nil
}

func (db *PlantsMongoDb) UpdatePlant(plant Plant) error {
	// TODO - implement this
	if plant.Name == "error plant" {
		return errors.New("plant db blew up")
	}
	return nil
}

func (db *PlantsMongoDb) DeletePlant(name string) error {
	// TODO - implement this
	if name == "error plant" {
		return errors.New("plant db blew up")
	}
	return nil
}
