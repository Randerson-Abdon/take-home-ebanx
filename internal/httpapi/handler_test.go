package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Randerson-Abdon/take-home-ebanx/internal/account"
	"github.com/Randerson-Abdon/take-home-ebanx/internal/store"
)

func TestHealth(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	newTestHandler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	if recorder.Body.String() != "ok" {
		t.Fatalf("expected body %q, got %q", "ok", recorder.Body.String())
	}
}

func TestCORSPreflight(t *testing.T) {
	request := httptest.NewRequest(http.MethodOptions, "/event", nil)
	recorder := httptest.NewRecorder()

	newTestHandler().ServeHTTP(recorder, request)

	assertResponse(t, recorder, http.StatusNoContent, "")
	if recorder.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("expected cross-origin requests to be allowed")
	}
	if recorder.Header().Get("Access-Control-Allow-Methods") != "GET, POST, OPTIONS" {
		t.Fatalf("expected allowed methods to include API and preflight methods")
	}
}

func TestResetClearsAccountState(t *testing.T) {
	handler := newTestHandler()
	performRequest(handler, http.MethodPost, "/event", `{"type":"deposit","destination":"100","amount":10}`)

	resetResponse := performRequest(handler, http.MethodPost, "/reset", "")
	balanceResponse := performRequest(handler, http.MethodGet, "/balance?account_id=100", "")

	assertResponse(t, resetResponse, http.StatusOK, "")
	assertResponse(t, balanceResponse, http.StatusNotFound, "0")
}

func TestBalanceForExistingAccount(t *testing.T) {
	handler := newTestHandler()
	performRequest(handler, http.MethodPost, "/event", `{"type":"deposit","destination":"100","amount":20}`)

	response := performRequest(handler, http.MethodGet, "/balance?account_id=100", "")

	assertResponse(t, response, http.StatusOK, "20")
}

func TestBalanceForMissingAccount(t *testing.T) {
	response := performRequest(newTestHandler(), http.MethodGet, "/balance?account_id=1234", "")

	assertResponse(t, response, http.StatusNotFound, "0")
	if response.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("expected API response to allow cross-origin requests")
	}
}

func TestDepositCreatesAndCreditsDestination(t *testing.T) {
	handler := newTestHandler()

	firstResponse := performRequest(handler, http.MethodPost, "/event", `{"type":"deposit","destination":"100","amount":10}`)
	secondResponse := performRequest(handler, http.MethodPost, "/event", `{"type":"deposit","destination":"100","amount":10}`)

	assertResponse(t, firstResponse, http.StatusCreated, "{\"destination\":{\"id\":\"100\",\"balance\":10}}\n")
	assertResponse(t, secondResponse, http.StatusCreated, "{\"destination\":{\"id\":\"100\",\"balance\":20}}\n")
}

func TestWithdrawFromExistingAccount(t *testing.T) {
	handler := newTestHandler()
	performRequest(handler, http.MethodPost, "/event", `{"type":"deposit","destination":"100","amount":20}`)

	response := performRequest(handler, http.MethodPost, "/event", `{"type":"withdraw","origin":"100","amount":5}`)

	assertResponse(t, response, http.StatusCreated, "{\"origin\":{\"id\":\"100\",\"balance\":15}}\n")
}

func TestWithdrawFromMissingAccount(t *testing.T) {
	response := performRequest(newTestHandler(), http.MethodPost, "/event", `{"type":"withdraw","origin":"200","amount":10}`)

	assertResponse(t, response, http.StatusNotFound, "0")
}

func TestTransferFromExistingAccount(t *testing.T) {
	handler := newTestHandler()
	performRequest(handler, http.MethodPost, "/event", `{"type":"deposit","destination":"100","amount":15}`)

	response := performRequest(handler, http.MethodPost, "/event", `{"type":"transfer","origin":"100","amount":15,"destination":"300"}`)

	assertResponse(t, response, http.StatusCreated, "{\"origin\":{\"id\":\"100\",\"balance\":0},\"destination\":{\"id\":\"300\",\"balance\":15}}\n")
}

func TestTransferFromMissingAccount(t *testing.T) {
	response := performRequest(newTestHandler(), http.MethodPost, "/event", `{"type":"transfer","origin":"200","amount":15,"destination":"300"}`)

	assertResponse(t, response, http.StatusNotFound, "0")
}

func TestEventRejectsInvalidRequests(t *testing.T) {
	testCases := []struct {
		name string
		body string
	}{
		{name: "malformed JSON", body: `{`},
		{name: "unknown event", body: `{"type":"unknown","amount":10}`},
		{name: "missing destination", body: `{"type":"deposit","amount":10}`},
		{name: "invalid amount", body: `{"type":"deposit","destination":"100","amount":0}`},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			response := performRequest(newTestHandler(), http.MethodPost, "/event", testCase.body)

			assertResponse(t, response, http.StatusBadRequest, "0")
		})
	}
}

func TestEventReturnsUnprocessableEntityForInsufficientFunds(t *testing.T) {
	handler := newTestHandler()
	performRequest(handler, http.MethodPost, "/event", `{"type":"deposit","destination":"100","amount":10}`)

	response := performRequest(handler, http.MethodPost, "/event", `{"type":"withdraw","origin":"100","amount":15}`)

	assertResponse(t, response, http.StatusUnprocessableEntity, "0")
}

func newTestHandler() http.Handler {
	memoryStore := store.NewMemoryStore()
	service := account.NewService(memoryStore)
	return NewHandler(service)
}

func performRequest(handler http.Handler, method, target, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func assertResponse(t *testing.T, response *httptest.ResponseRecorder, wantStatus int, wantBody string) {
	t.Helper()

	if response.Code != wantStatus {
		t.Fatalf("expected status %d, got %d", wantStatus, response.Code)
	}
	if response.Body.String() != wantBody {
		t.Fatalf("expected body %q, got %q", wantBody, response.Body.String())
	}
}
