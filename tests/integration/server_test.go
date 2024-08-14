//go:build integration
// +build integration

package integration

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/olad5/caution-companion/config"
	"github.com/olad5/caution-companion/config/data"
	"github.com/olad5/caution-companion/internal/infra/cloudinary"
	"github.com/olad5/caution-companion/internal/infra/postgres"
	"github.com/olad5/caution-companion/internal/infra/redis"
	"github.com/olad5/caution-companion/internal/infra/smtpexpress"
	"github.com/olad5/caution-companion/pkg/api"
	"github.com/olad5/caution-companion/pkg/utils/logger"
	"github.com/olad5/caution-companion/tests"
)

var (
	appRouter      http.Handler
	configurations *config.Configurations
)

var (
	userEmail     = "will@gmail.com"
	userPassword  = "some-random-password"
	adminEmail    = "admin@app.com"
	adminPassword = "some-random-password"
)

func TestMain(m *testing.M) {
	// TODO:TODO: I should be able to allow it to wait in the make file and not here
	time.Sleep(8 * time.Second) // Wait for docker containers to start
	configurations = config.GetConfig("../config/.test.env")
	ctx := context.Background()
	l := logger.Get(configurations)

	postgresConnection := data.StartPostgres(configurations.DatabaseUrl, l)
	if err := postgres.Migrate(ctx, postgresConnection); err != nil {
		log.Fatal("Error Migrating postgres", err)
	}

	defer postgresConnection.Close()

	userRepo, err := postgres.NewPostgresUserRepo(ctx, postgresConnection)
	if err != nil {
		log.Fatal("Error Initializing User Repo", err)
	}
	reportsRepo, err := postgres.NewPostgresReportRepo(ctx, postgresConnection)
	if err != nil {
		log.Fatal("Error Initializing Reports Repo", err)
	}

	redisCache, err := redis.New(ctx, configurations)
	if err != nil {
		log.Fatal("Error Initializing redisCache", err)
	}
	fileStore, err := cloudinary.NewCloudinaryFileStore(ctx, configurations)
	if err != nil {
		log.Fatal("Error Initializing fileStore", err)
	}

	mailService, err := smtpexpress.New(ctx, configurations)
	if err != nil {
		log.Fatal("Error Initializing smtpexpress mailservice", err)
	}
	appRouter = api.NewHttpRouter(
		ctx,
		userRepo,
		reportsRepo,
		fileStore,
		redisCache,
		mailService,
		configurations,
		l)

	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestRegister(t *testing.T) {
	route := "/users"
	t.Run("test for invalid json request body",
		func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPost, route, nil)
			response := tests.ExecuteRequest(req, appRouter)
			tests.AssertStatusCode(t, http.StatusBadRequest, response.Code)
		},
	)
	t.Run(`Given a valid user registration request, when the user submits the request, 
    then the server should respond with a success status code, and the user's account 
    should be created in the database.`,
		func(t *testing.T) {
			email := "will" + fmt.Sprint(tests.GenerateUniqueId()) + "@gmail.com"
			requestBody := []byte(fmt.Sprintf(`{
      "email": "%s",
      "first_name": "will",
      "last_name": "hansen",
      "password": "some-random-password"
      }`, email))
			req, _ := http.NewRequest(http.MethodPost, route, bytes.NewBuffer(requestBody))
			response := tests.ExecuteRequest(req, appRouter)
			tests.AssertStatusCode(t, http.StatusOK, response.Code)
			data := tests.ParseResponse(t, response)["data"].(map[string]interface{})
			tests.AssertResponseMessage(t, data["email"].(string), email)
		},
	)

	t.Run(`Given a user registration request with an email address that already exists,
    when the user submits the request, then the server should respond with an error
    status code, and the response should indicate that the email address is already
    taken. `,
		func(t *testing.T) {
			email := "will@gmail.com"
			requestBody := []byte(fmt.Sprintf(`{
      "email": "%s",
      "first_name": "will",
      "last_name": "hansen",
      "password": "some-random-password"
      }`, email))
			req, _ := http.NewRequest(http.MethodPost, route, bytes.NewBuffer(requestBody))
			_ = tests.ExecuteRequest(req, appRouter)

			secondRequestBody := []byte(fmt.Sprintf(`{
      "email": "%s",
      "first_name": "will",
      "last_name": "hansen",
      "password": "passcode-password"
      }`, email))
			req, _ = http.NewRequest(http.MethodPost, route, bytes.NewBuffer(secondRequestBody))

			response := tests.ExecuteRequest(req, appRouter)
			tests.AssertStatusCode(t, http.StatusBadRequest, response.Code)
			message := tests.ParseResponse(t, response)["message"].(string)
			tests.AssertResponseMessage(t, message, "email already exist")
		},
	)
}

func TestLogin(t *testing.T) {
	route := "/users/login"
	t.Run("test for invalid json request body",
		func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPost, route, nil)
			response := tests.ExecuteRequest(req, appRouter)
			tests.AssertStatusCode(t, http.StatusBadRequest, response.Code)
		},
	)
	t.Run("test for undefined email address in request body",
		func(t *testing.T) {
			requestBody := []byte(fmt.Sprintf(`{
      "email": "%v",
      "password": "%v"
      }`, nil, nil))
			req, _ := http.NewRequest(http.MethodPost, route, bytes.NewBuffer(requestBody))
			response := tests.ExecuteRequest(req, appRouter)
			tests.AssertStatusCode(t, http.StatusBadRequest, response.Code)
		},
	)
	t.Run(`Given a user attempts to log in with valid credentials,
    when they make a POST request to the login endpoint with their username and password,
    then they should receive a 200 OK response,
    and the response should contain a JSON web token (JWT) in the 'token' field,
    and the token should be valid and properly signed.`,
		func(t *testing.T) {
			requestBody := []byte(fmt.Sprintf(`{
      "email": "%s",
      "password": "%s"
      }`, userEmail, userPassword))
			req, _ := http.NewRequest(http.MethodPost, route, bytes.NewBuffer(requestBody))
			response := tests.ExecuteRequest(req, appRouter)

			tests.AssertStatusCode(t, http.StatusOK, response.Code)
			responseBody := tests.ParseResponse(t, response)
			message := responseBody["message"].(string)
			tests.AssertResponseMessage(t, message, "user logged in successfully")

			data := responseBody["data"].(map[string]interface{})

			accessToken, exists := data["access_token"]
			if !exists {
				t.Error("Missing 'access_token' key in the JSON response")
			}

			_, isString := accessToken.(string)
			if !isString {
				t.Error("'access_token' value is not a string")
			}
		},
	)

	t.Run(`Given a user attempts to log in with invalid password,
    when they make a POST request to the login endpoint with their username and password,
    then they should receive a 401 unauthorized response,
    and the response should contain an invalidd credentials message.`,
		func(t *testing.T) {
			email := "will@gmail.com"
			requestBody := []byte(fmt.Sprintf(`{
      "email": "%s",
      "password": "invalid-password"
      }`, email))
			req, _ := http.NewRequest(http.MethodPost, route, bytes.NewBuffer(requestBody))
			response := tests.ExecuteRequest(req, appRouter)

			tests.AssertStatusCode(t, http.StatusUnauthorized, response.Code)
			responseBody := tests.ParseResponse(t, response)
			message := responseBody["message"].(string)
			tests.AssertResponseMessage(t, message, "invalid credentials")
		},
	)

	t.Run(`Given a user tries to log in with an account that does not exist,
    when they make a POST request to the login endpoint with a non-existent email,
    then they should receive a 404 Not Found response,
    and the response should contain an error message indicating that the account 
    does not exist.`,
		func(t *testing.T) {
			email := "emailnoexist@gmail.com"
			requestBody := []byte(fmt.Sprintf(`{
      "email": "%s",
      "first_name": "mike",
      "last_name": "wilson",
      "password": "some-random-password"
      }`, email))
			req, _ := http.NewRequest(http.MethodPost, route, bytes.NewBuffer(requestBody))
			response := tests.ExecuteRequest(req, appRouter)

			tests.AssertStatusCode(t, http.StatusNotFound, response.Code)
			message := tests.ParseResponse(t, response)["message"].(string)
			tests.AssertResponseMessage(t, message, "user does not exist")
		},
	)
}

func TestRefreshToken(t *testing.T) {
	route := "/users/token/refresh"
	t.Run(`Given a user tries to refresh their JWT token and provides a valid
    refresh token, when the request is processed, they receive a new JWT token
    and a response indicating success. `,
		func(t *testing.T) {
			token, existingRefreshToken := logUserIn(t, userEmail, userPassword)
			existingUserId := getCurrentUser(t, token)["id"].(string)

			requestBody := []byte(fmt.Sprintf(`{
      "refresh_token": "%s"
      }`, existingRefreshToken))
			req, _ := http.NewRequest(http.MethodPost, route, bytes.NewBuffer(requestBody))
			response := tests.ExecuteRequest(req, appRouter)

			tests.AssertStatusCode(t, http.StatusOK, response.Code)
			responseBody := tests.ParseResponse(t, response)
			message := responseBody["message"].(string)
			tests.AssertResponseMessage(t, message, "access token refreshed successfully")

			data := responseBody["data"].(map[string]interface{})

			accessTokenField, exists := data["access_token"]
			if !exists {
				t.Error("Missing 'access_token' key in the JSON response")
			}

			newAccessToken, isString := accessTokenField.(string)
			if !isString {
				t.Error("'access_token' value is not a string")
			}
			refreshTokenField, exists := data["refresh_token"]
			if !exists {
				t.Error("Missing 'refresh_token' key in the JSON response")
			}

			_, isString = refreshTokenField.(string)
			if !isString {
				t.Error("'refresh_token' value is not a string")
			}

			userId := getCurrentUser(t, newAccessToken)["id"].(string)
			if existingUserId != userId {
				t.Error("user ids do not match")
			}
		},
	)
}

func TestEditUserProfile(t *testing.T) {
	route := "/users"
	t.Run(`Given a user tries to update their profile with valid data, when they 
    submit the changes, the profile is updated successfully and they receive a 
    confirmation message. `,
		func(t *testing.T) {
			firstName := "mike"
			lastName := "fisher"
			email := lastName + fmt.Sprint(tests.GenerateUniqueId()) + "@gmail.com"
			password := "some-random-password"
			createUser(t, firstName, lastName, email, password)
			ac, _ := logUserIn(t, email, password)

			newEmail := lastName + fmt.Sprint(tests.GenerateUniqueId()) + "@gmail.com"
			newFirstName := "michael"
			newLocation := ""
			someUniqueId := (fmt.Sprint(tests.GenerateUniqueId()))
			newUserName := "iamfisher" + someUniqueId[len(someUniqueId)-3:]
			newAvatar := "https://res.cloudinary.com/deda4nfxl/image/upload/v1721583338/caution-companion/caution-companion/avatars/4608bc1b98c84a06838fafb5e38fb552.jpg"
			phone := "08093487904"
			requestBody := []byte(fmt.Sprintf(`{
      "email": "%s",
      "first_name": "%s",
      "last_name": "%s",
      "avatar": "%s",
      "user_name": "%s",
      "location": "%s",
      "phone": "%s"
      }`, newEmail, newFirstName, lastName, newAvatar, newUserName, newLocation, phone))
			req, _ := http.NewRequest(http.MethodPut, route, bytes.NewBuffer(requestBody))
			req.Header.Set("Authorization", "Bearer "+ac)
			response := tests.ExecuteRequest(req, appRouter)
			tests.AssertStatusCode(t, http.StatusOK, response.Code)
			message := tests.ParseResponse(t, response)["message"].(string)
			tests.AssertResponseMessage(t, message, "user profile updated successfully")

			ac, _ = logUserIn(t, newEmail, password)
			user := getCurrentUser(t, ac)
			tests.AssertResponseMessage(t, user["email"].(string), newEmail)
			tests.AssertResponseMessage(t, user["first_name"].(string), newFirstName)
			tests.AssertResponseMessage(t, user["last_name"].(string), lastName)
			tests.AssertResponseMessage(t, user["location"].(string), newLocation)
			tests.AssertResponseMessage(t, user["phone"].(string), phone)
		},
	)
}

func TestCreatReport(t *testing.T) {
	route := "/reports"
	t.Run(` Given a user wants to create an emergency report, when they submit all 
    the required information correctly, then the report is created successfully 
    and a confirmation message is sent back to the user.
    `,
		func(t *testing.T) {
			incidentType := "fire"
			longitude := "11.11"
			latitude := "23.991818118"
			description := "some-description"
			requestBody := []byte(fmt.Sprintf(`{
      "incident_type": "%s",
      "location": {
        "longitude": "%s",
        "latitude": "%s"
        },
      "description": "%s"
      }`, incidentType, longitude, latitude, description))
			req, _ := http.NewRequest(http.MethodPost, route, bytes.NewBuffer(requestBody))
			token, _ := logUserIn(t, userEmail, userPassword)
			req.Header.Set("Authorization", "Bearer "+token)
			response := tests.ExecuteRequest(req, appRouter)

			tests.AssertStatusCode(t, http.StatusOK, response.Code)
			responseBody := tests.ParseResponse(t, response)
			message := responseBody["message"].(string)
			tests.AssertResponseMessage(t, message, "report created successfully")

			data := responseBody["data"].(map[string]interface{})
			tests.AssertResponseMessage(t, data["incident_type"].(string), incidentType)
			location := data["location"].(map[string]interface{})
			tests.AssertResponseMessage(t, location["longitude"].(string), longitude)
			tests.AssertResponseMessage(t, location["latitude"].(string), latitude)
			tests.AssertResponseMessage(t, data["description"].(string), description)
		},
	)
}

func TestGetReportByReportId(t *testing.T) {
	route := "/reports"
	t.Run("test for invalid json request body",
		func(t *testing.T) {
			t.Skip()
			// TODO:TODO: fix this bit later
			req, _ := http.NewRequest(http.MethodPost, route, nil)
			response := tests.ExecuteRequest(req, appRouter)
			tests.AssertStatusCode(t, http.StatusBadRequest, response.Code)
		},
	)
	// TODO:TODO: i might need to test this
	// t.Run("test for undefined email address in request body",
	// 	func(t *testing.T) {
	// 		requestBody := []byte(fmt.Sprintf(`{
	// "email": "%v",
	// "password": "%v"
	// }`, nil, nil))
	// 		req, _ := http.NewRequest(http.MethodPost, route, bytes.NewBuffer(requestBody))
	// 		response := tests.ExecuteRequest(req, appRouter)
	// 		tests.AssertStatusCode(t, http.StatusBadRequest, response.Code)
	// 	},
	// )
	t.Run(` Given a user tries to get an emergency report by ID and the report 
    exists, when they provide the correct ID, they receive the report details.
    `,
		func(t *testing.T) {
			incidentType := "fire"
			longitude := "11.11"
			latitude := "23.991818118"
			description := "some-description"
			token, _ := logUserIn(t, userEmail, userPassword)
			reportId := createReport(t, token, incidentType, longitude, latitude, description)
			req, _ := http.NewRequest(http.MethodGet, route+"/"+reportId, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			response := tests.ExecuteRequest(req, appRouter)

			tests.AssertStatusCode(t, http.StatusOK, response.Code)
			responseBody := tests.ParseResponse(t, response)
			message := responseBody["message"].(string)
			tests.AssertResponseMessage(t, message, "report retrieved successfully")

			data := responseBody["data"].(map[string]interface{})
			tests.AssertResponseMessage(t, data["incident_type"].(string), incidentType)
			location := data["location"].(map[string]interface{})
			tests.AssertResponseMessage(t, location["longitude"].(string), longitude)
			tests.AssertResponseMessage(t, location["latitude"].(string), latitude)
			tests.AssertResponseMessage(t, data["description"].(string), description)
		},
	)
}

func TestGetLatestReports(t *testing.T) {
	route := "/reports"
	t.Run(`Given a user tries to get the latest emergency reports and there are 
    recent reports available, when they make the request, they receive a list 
    of the most recent emergency reports.
    `,
		func(t *testing.T) {
			// TODO:TODO: this test case title is wrong
			t.Skip()
			token, _ := logUserIn(t, userEmail, userPassword)

			req, _ := http.NewRequest(http.MethodGet, route+"/latest"+"?page=1&rows=20", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			response := tests.ExecuteRequest(req, appRouter)

			tests.AssertStatusCode(t, http.StatusOK, response.Code)
			responseBody := tests.ParseResponse(t, response)
			message := responseBody["message"].(string)
			tests.AssertResponseMessage(t, message, "latest reports retrieved successfully")

			data := responseBody["data"].(map[string]interface{})
			reports := data["items"].([]interface{})
			const numberOfReports = 3
			if len(reports) != numberOfReports {
				t.Errorf("got files length: %d expected: %d", len(reports), numberOfReports)
			}
		},
	)
	t.Run(`Given a user tries to get the latest emergency reports and there are 
    recent reports available, when they make the request, they receive a list 
    of the most recent emergency reports.
    `,
		func(t *testing.T) {
			token, _ := logUserIn(t, userEmail, userPassword)

			const numberOfReports = 2
			const pageToRetrieve = 1
			req, _ := http.NewRequest(http.MethodGet, route+"/latest"+"?page="+fmt.Sprintf("%d", pageToRetrieve)+"&rows="+fmt.Sprintf("%d", numberOfReports), nil)
			req.Header.Set("Authorization", "Bearer "+token)
			response := tests.ExecuteRequest(req, appRouter)

			tests.AssertStatusCode(t, http.StatusOK, response.Code)
			responseBody := tests.ParseResponse(t, response)
			message := responseBody["message"].(string)
			tests.AssertResponseMessage(t, message, "latest reports retrieved successfully")

			data := responseBody["data"].(map[string]interface{})
			reports := data["items"].([]interface{})
			page := data["page"].(float64)
			rows := data["rows"].(float64)
			if len(reports) != numberOfReports {
				t.Errorf("got items length: %d expected: %d", len(reports), numberOfReports)
			}
			if rows != numberOfReports {
				t.Errorf("got rows number retrieved: %v expected: %d", rows, numberOfReports)
			}
			if page != pageToRetrieve {
				t.Errorf("got rows number retrieved: %v expected: %d", page, pageToRetrieve)
			}
		},
	)
}

func createReport(t testing.TB, token, incidentType, longitude, latitude, description string) string {
	t.Helper()
	route := "/reports"
	requestBody := []byte(fmt.Sprintf(`{
      "incident_type": "%s",
      "location": {
        "longitude": "%s",
        "latitude": "%s"
        },
      "description": "%s"
      }`, incidentType, longitude, latitude, description))
	req, _ := http.NewRequest(http.MethodPost, route, bytes.NewBuffer(requestBody))
	req.Header.Set("Authorization", "Bearer "+token)

	response := tests.ExecuteRequest(req, appRouter)
	responseBody := tests.ParseResponse(t, response)
	data := responseBody["data"].(map[string]interface{})
	reportId := data["id"].(string)

	return reportId
}

func createUser(t testing.TB, firstName, lastName, email, password string) string {
	t.Helper()
	route := "/users"
	requestBody := []byte(fmt.Sprintf(`{
      "first_name": "%s",
      "last_name": "%s",
      "email": "%s",
      "password": "%s"
      }`, firstName, lastName, email, password))
	req, _ := http.NewRequest(http.MethodPost, route, bytes.NewBuffer(requestBody))
	response := tests.ExecuteRequest(req, appRouter)
	data := tests.ParseResponse(t, response)["data"].(map[string]interface{})
	userId := data["id"].(string)

	return userId
}

func getCurrentUser(t testing.TB, token string) map[string]interface{} {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, "/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	response := tests.ExecuteRequest(req, appRouter)
	responseBody := tests.ParseResponse(t, response)
	data := responseBody["data"].(map[string]interface{})
	return data
}

func logUserIn(t testing.TB, email, password string) (string, string) {
	t.Helper()
	requestBody := []byte(fmt.Sprintf(`{
      "email": "%s",
      "password": "%s"
      }`, email, password))
	loginReq, _ := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(requestBody))
	loginResponse := tests.ExecuteRequest(loginReq, appRouter)
	loginResponseBody := tests.ParseResponse(t, loginResponse)
	loginData := loginResponseBody["data"].(map[string]interface{})
	accessTokenField := loginData["access_token"]
	accessToken := accessTokenField.(string)
	refreshTokenField := loginData["refresh_token"]
	refreshToken := refreshTokenField.(string)
	return accessToken, refreshToken
}
