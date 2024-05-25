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

	"github.com/olad5/go-hackathon-starter-template/config"
	"github.com/olad5/go-hackathon-starter-template/config/data"
	"github.com/olad5/go-hackathon-starter-template/internal/infra/postgres"
	"github.com/olad5/go-hackathon-starter-template/internal/infra/redis"
	"github.com/olad5/go-hackathon-starter-template/pkg/api"
	"github.com/olad5/go-hackathon-starter-template/pkg/utils/logger"
	"github.com/olad5/go-hackathon-starter-template/tests"
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

	redisCache, err := redis.New(ctx, configurations)
	if err != nil {
		log.Fatal("Error Initializing redisCache", err)
	}
	appRouter = api.NewHttpRouter(ctx, userRepo, redisCache, configurations, l)

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
				t.Error("Missing 'accesstoken' key in the JSON response")
			}

			_, isString := accessToken.(string)
			if !isString {
				t.Error("'accesstoken' value is not a string")
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

func logUserIn(t testing.TB, email, password string) string {
	requestBody := []byte(fmt.Sprintf(`{
      "email": "%s",
      "password": "%s"
      }`, email, password))
	loginReq, _ := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(requestBody))
	loginResponse := tests.ExecuteRequest(loginReq, appRouter)
	loginResponseBody := tests.ParseResponse(t, loginResponse)
	loginData := loginResponseBody["data"].(map[string]interface{})
	accessToken := loginData["access_token"]
	token := accessToken.(string)
	return token
}
