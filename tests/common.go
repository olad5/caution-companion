package tests

import (
	"crypto/rand"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
)

func ExecuteRequest(req *http.Request, s http.Handler) *httptest.ResponseRecorder {
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.ServeHTTP(rr, req)

	return rr
}

func AssertStatusCode(t *testing.T, expected, actual int) {
	t.Helper()
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func AssertResponseMessage(t *testing.T, got, expected string) {
	t.Helper()
	if got != expected {
		t.Errorf("got message: %q expected: %q", got, expected)
	}
}

func ParseResponse(t testing.TB, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	res := make(map[string]interface{})
	body := w.Body
	err := json.NewDecoder(body).Decode(&res)
	if err != nil {
		t.Fatalf("Unable to parse response from body %q '%v'", body, err)
	}
	return res
}

func GenerateUniqueId() int {
	MAX_INT := 7935425686241
	b := new(big.Int).SetInt64(int64(MAX_INT))
	randomBigInt, _ := rand.Int(rand.Reader, b)
	randomeNewInt := int(randomBigInt.Int64())
	return randomeNewInt
}
