package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type MockDB struct {
	DbResponse interface{}
	DbError    error
}

type TestCase struct {
	testName             string
	requestPathId        string
	requestBody          string
	dbResponse           interface{}
	dbError              error
	expectedStatusCode   int
	expectedResponseBody string
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
	return db.DbResponse.(Plant), db.DbError
}

func (db *MockDB) CreatePlant(plant Plant) error {
	return db.DbError
}

func (db *MockDB) UpsertPlant(id int, plant Plant) error {
	return db.DbError
}

func (db *MockDB) DeletePlant(id int) error {
	return db.DbError
}

func TestListPlants(t *testing.T) {
	cases := []TestCase{
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
			req := http.Request{RequestURI: "api/plants"}
			w := httptest.NewRecorder()
			api := Api{DB: db}

			// Act
			api.listPlants(w, &req)

			// Assert
			responseBody := strings.TrimSpace(w.Body.String())
			if responseBody != tc.expectedResponseBody {
				t.Errorf("handler returned unexpected body: got %v, want %v",
					responseBody, tc.expectedResponseBody)
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
	cases := []TestCase{
		{
			testName:      "valid_db_response_returns_200_and_plant",
			requestPathId: "99",
			dbResponse: Plant{
				Id:         99,
				Name:       "Plant A",
				OtherNames: []string{"Other name A"},
				Light:      "low",
				Humidity:   "high",
				Water:      "low",
			},
			dbError:              nil,
			expectedStatusCode:   200,
			expectedResponseBody: "{\"id\":99,\"name\":\"Plant A\",\"otherNames\":[\"Other name A\"],\"light\":\"low\",\"humidity\":\"high\",\"water\":\"low\"}",
		},
		{
			testName:             "error_db_response_returns_500_and_error",
			requestPathId:        "99",
			dbResponse:           Plant{},
			dbError:              errors.New("something went wrong!"),
			expectedStatusCode:   500,
			expectedResponseBody: "{\"error\":\"An error occurred while processing the request\"}",
		},
		{
			testName:             "notfound_db_response_returns_404_and_error",
			requestPathId:        "99",
			dbResponse:           Plant{},
			dbError:              &NotFoundError{},
			expectedStatusCode:   404,
			expectedResponseBody: "{\"error\":\"The specified Plant was not found\"}",
		},
		{
			testName:             "invalid_id_returns_400_and_error",
			requestPathId:        "abc",
			dbResponse:           []Plant{},
			dbError:              nil,
			expectedStatusCode:   400,
			expectedResponseBody: "{\"error\":\"The Plant id must be an integer\"}",
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(t *testing.T) {
			// Arrange
			db := &MockDB{DbResponse: tc.dbResponse, DbError: tc.dbError}
			req, _ := http.NewRequest("GET", "api/plants", nil)
			query := url.Values{}
			query.Add("id", tc.requestPathId)
			req.URL.RawQuery = query.Encode()
			w := httptest.NewRecorder()
			api := Api{DB: db}

			// Act
			api.getPlant(w, req)

			// Assert
			responseBody := strings.TrimSpace(w.Body.String())
			if responseBody != tc.expectedResponseBody {
				t.Errorf("handler returned unexpected body: got %v, want %v",
					responseBody, tc.expectedResponseBody)
			}
			actualStatusCode := w.Result().StatusCode
			if actualStatusCode != tc.expectedStatusCode {
				t.Errorf("handler returned unexpected status code: got %v, want %v",
					actualStatusCode, tc.expectedStatusCode)
			}
		})
	}
}

func TestPostPlant(t *testing.T) {
	cases := []TestCase{
		{
			testName:             "valid_db_response_returns_201",
			requestBody:          "{\"name\":\"plant A\",\"light\":\"low\",\"humidity\":\"low\",\"water\":\"low\",\"otherNames\":[]}",
			dbError:              nil,
			expectedStatusCode:   201,
			expectedResponseBody: "{}",
		},
		{
			testName:             "error_db_response_returns_500_and_error",
			requestBody:          "{\"name\":\"plant A\",\"light\":\"low\",\"humidity\":\"low\",\"water\":\"low\",\"otherNames\":[]}",
			dbError:              errors.New("something went wrong!"),
			expectedStatusCode:   500,
			expectedResponseBody: "{\"error\":\"An error occurred while processing the request\"}",
		},
		{
			testName:             "conflict_db_response_returns_409_and_error",
			requestBody:          "{\"name\":\"plant A\",\"light\":\"low\",\"humidity\":\"low\",\"water\":\"low\",\"otherNames\":[]}",
			dbError:              &ConflictError{ConflictingKey: "name", ConflictingValue: "plant X"},
			expectedStatusCode:   409,
			expectedResponseBody: "{\"error\":\"Plant with name 'plant X' already exists\"}",
		},
		{
			testName:             "invalid_payload_returns_400_and_error",
			requestBody:          "{\"name\":123,\"invalid\":\"plant\"}",
			expectedStatusCode:   400,
			expectedResponseBody: "{\"error\":\"The request payload could not be parsed into a Plant\"}",
		},
		{
			testName:             "failed_validation_returns_400_and_error",
			requestBody:          "{\"light\":\"low\",\"humidity\":\"low\",\"water\":\"low\",\"otherNames\":[]}",
			expectedStatusCode:   400,
			expectedResponseBody: "{\"error\":\"The name value is required\"}",
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(t *testing.T) {
			// Arrange
			db := &MockDB{DbResponse: tc.dbResponse, DbError: tc.dbError}
			req, _ := http.NewRequest("POST", "api/plants", strings.NewReader(tc.requestBody))
			w := httptest.NewRecorder()
			api := Api{DB: db}

			// Act
			api.postPlant(w, req)

			// Assert
			responseBody := strings.TrimSpace(w.Body.String())
			if responseBody != tc.expectedResponseBody {
				t.Errorf("handler returned unexpected body: got %v, want %v",
					responseBody, tc.expectedResponseBody)
			}
			actualStatusCode := w.Result().StatusCode
			if actualStatusCode != tc.expectedStatusCode {
				t.Errorf("handler returned unexpected status code: got %v, want %v",
					actualStatusCode, tc.expectedStatusCode)
			}
		})
	}
}

func TestPutPlant(t *testing.T) {
	cases := []TestCase{
		{
			testName:             "valid_db_response_returns_200",
			requestPathId:        "99",
			requestBody:          "{\"name\":\"plant A\",\"light\":\"low\",\"humidity\":\"low\",\"water\":\"low\",\"otherNames\":[]}",
			dbError:              nil,
			expectedStatusCode:   200,
			expectedResponseBody: "{}",
		},
		{
			testName:             "error_db_response_returns_500_and_error",
			requestPathId:        "99",
			requestBody:          "{\"name\":\"plant A\",\"light\":\"low\",\"humidity\":\"low\",\"water\":\"low\",\"otherNames\":[]}",
			dbError:              errors.New("something went wrong!"),
			expectedStatusCode:   500,
			expectedResponseBody: "{\"error\":\"An error occurred while processing the request\"}",
		},
		{
			testName:             "conflict_db_response_returns_409_and_error",
			requestPathId:        "99",
			requestBody:          "{\"name\":\"plant A\",\"light\":\"low\",\"humidity\":\"low\",\"water\":\"low\",\"otherNames\":[]}",
			dbError:              &ConflictError{ConflictingKey: "name", ConflictingValue: "plant X"},
			expectedStatusCode:   409,
			expectedResponseBody: "{\"error\":\"Plant with name 'plant X' already exists\"}",
		},
		{
			testName:             "invalid_payload_returns_400_and_error",
			requestPathId:        "99",
			requestBody:          "{\"name\":123,\"invalid\":\"plant\"}",
			expectedStatusCode:   400,
			expectedResponseBody: "{\"error\":\"The request payload could not be parsed into a Plant\"}",
		},
		{
			testName:             "invalid_id_returns_400_and_error",
			requestPathId:        "abc",
			requestBody:          "{\"light\":\"low\",\"humidity\":\"low\",\"water\":\"low\",\"otherNames\":[]}",
			expectedStatusCode:   400,
			expectedResponseBody: "{\"error\":\"The Plant id must be an integer\"}",
		},
		{
			testName:             "failed_validation_returns_400_and_error",
			requestPathId:        "99",
			requestBody:          "{\"light\":\"low\",\"humidity\":\"low\",\"water\":\"low\",\"otherNames\":[]}",
			expectedStatusCode:   400,
			expectedResponseBody: "{\"error\":\"The name value is required\"}",
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(t *testing.T) {
			// Arrange
			db := &MockDB{DbResponse: tc.dbResponse, DbError: tc.dbError}
			req, _ := http.NewRequest("PUT", "api/plants", strings.NewReader(tc.requestBody))
			query := url.Values{}
			query.Add("id", tc.requestPathId)
			req.URL.RawQuery = query.Encode()
			w := httptest.NewRecorder()
			api := Api{DB: db}

			// Act
			api.putPlant(w, req)

			// Assert
			responseBody := strings.TrimSpace(w.Body.String())
			if responseBody != tc.expectedResponseBody {
				t.Errorf("handler returned unexpected body: got %v, want %v",
					responseBody, tc.expectedResponseBody)
			}
			actualStatusCode := w.Result().StatusCode
			if actualStatusCode != tc.expectedStatusCode {
				t.Errorf("handler returned unexpected status code: got %v, want %v",
					actualStatusCode, tc.expectedStatusCode)
			}
		})
	}
}

func TestDeletePlant(t *testing.T) {
	cases := []TestCase{
		{
			testName:             "valid_db_response_returns_204",
			requestPathId:        "99",
			dbError:              nil,
			expectedStatusCode:   204,
			expectedResponseBody: "{}",
		},
		{
			testName:             "error_db_response_returns_500_and_error",
			requestPathId:        "99",
			dbError:              errors.New("something went wrong!"),
			expectedStatusCode:   500,
			expectedResponseBody: "{\"error\":\"An error occurred while processing the request\"}",
		},
		{
			testName:             "invalid_id_returns_400_and_error",
			requestPathId:        "abc",
			expectedStatusCode:   400,
			expectedResponseBody: "{\"error\":\"The Plant id must be an integer\"}",
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(t *testing.T) {
			// Arrange
			db := &MockDB{DbResponse: tc.dbResponse, DbError: tc.dbError}
			req, _ := http.NewRequest("PUT", "api/plants", nil)
			query := url.Values{}
			query.Add("id", tc.requestPathId)
			req.URL.RawQuery = query.Encode()
			w := httptest.NewRecorder()
			api := Api{DB: db}

			// Act
			api.deletePlant(w, req)

			// Assert
			responseBody := strings.TrimSpace(w.Body.String())
			if responseBody != tc.expectedResponseBody {
				t.Errorf("handler returned unexpected body: got %v, want %v",
					responseBody, tc.expectedResponseBody)
			}
			actualStatusCode := w.Result().StatusCode
			if actualStatusCode != tc.expectedStatusCode {
				t.Errorf("handler returned unexpected status code: got %v, want %v",
					actualStatusCode, tc.expectedStatusCode)
			}
		})
	}
}
