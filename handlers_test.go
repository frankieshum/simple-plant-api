package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type MockDB struct {
	DbResponse interface{}
	DbError    error
}

func (db *MockDB) Connect() error {
	return nil
}

func (db *MockDB) Disconnect() error {
	return nil
}

func (db *MockDB) GetAllPlants() ([]Plant, error) {
	return db.DbResponse.([]Plant), db.DbError
}

func (db *MockDB) GetPlantById(id int) (Plant, error) {
	return Plant{}, nil
}

func (db *MockDB) CreatePlant(plant Plant) error {
	return nil
}

func (db *MockDB) UpsertPlant(id int, plant Plant) error {
	return nil
}

func (db *MockDB) DeletePlant(id int) error {
	return nil
}

func TestListPlants(t *testing.T) {
	cases := []struct {
		testName             string
		dbResponse           []Plant
		dbError              error
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			testName: "valid_db_response_returns_200_and_plant_array",
			dbResponse: []Plant{
				{
					Id:         99,
					Name:       "Plant A",
					OtherNames: []string{"Other name A"},
					Light:      "low",
					Humidity:   "high",
					Water:      "low",
				},
			},
			dbError:              nil,
			expectedStatusCode:   200,
			expectedResponseBody: "[{\"id\":99,\"name\":\"Plant A\",\"otherNames\":[\"Other name A\"],\"light\":\"low\",\"humidity\":\"high\",\"water\":\"low\"}]",
		},
		{
			testName:             "error_db_response_returns_500_and_error",
			dbResponse:           []Plant{},
			dbError:              errors.New("something went wrong!"),
			expectedStatusCode:   500,
			expectedResponseBody: "{\"error\":\"An error occurred while processing the request\"}",
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(t *testing.T) {
			// Arrange
			db := &MockDB{DbResponse: tc.dbResponse, DbError: tc.dbError}
			req := http.Request{RequestURI: "some/uri"}
			w := httptest.NewRecorder()
			api := Api{DB: db}

			// Act
			api.listPlants(w, &req)

			// Assert
			actualResponseBody := strings.TrimSpace(w.Body.String())
			if actualResponseBody != tc.expectedResponseBody {
				t.Errorf("handler returned unexpected body: got %v, want %v",
					actualResponseBody, tc.expectedResponseBody)
			}
			actualStatusCode := w.Result().StatusCode
			if actualStatusCode != tc.expectedStatusCode {
				t.Errorf("handler returned unexpected status code: got %v, want %v",
					actualStatusCode, tc.expectedStatusCode)
			}
		})
	}
}

func TestGetPlant(t *testing.T) {
	// 200
	// 400
	// 404
	// 500
}

func TestPostPlant(t *testing.T) {
	// 201
	// 400
	// 409
	// 500
}

func TestPutPlant(t *testing.T) {
	// 200
	// 400
	// 409
	// 500
}

func TestDeletePlant(t *testing.T) {
	// 204
	// 400
	// 500
}
