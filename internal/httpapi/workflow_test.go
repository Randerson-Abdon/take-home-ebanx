package httpapi

import (
	"net/http"
	"testing"
)

func TestOfficialAPIWorkflow(t *testing.T) {
	handler := newTestHandler()
	steps := []struct {
		name       string
		method     string
		target     string
		body       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "reset state before starting tests",
			method:     http.MethodPost,
			target:     "/reset",
			wantStatus: http.StatusOK,
		},
		{
			name:       "get balance for non-existing account",
			method:     http.MethodGet,
			target:     "/balance?account_id=1234",
			wantStatus: http.StatusNotFound,
			wantBody:   "0",
		},
		{
			name:       "create account with initial balance",
			method:     http.MethodPost,
			target:     "/event",
			body:       `{"type":"deposit","destination":"100","amount":10}`,
			wantStatus: http.StatusCreated,
			wantBody:   "{\"destination\":{\"id\":\"100\",\"balance\":10}}\n",
		},
		{
			name:       "deposit into existing account",
			method:     http.MethodPost,
			target:     "/event",
			body:       `{"type":"deposit","destination":"100","amount":10}`,
			wantStatus: http.StatusCreated,
			wantBody:   "{\"destination\":{\"id\":\"100\",\"balance\":20}}\n",
		},
		{
			name:       "get balance for existing account",
			method:     http.MethodGet,
			target:     "/balance?account_id=100",
			wantStatus: http.StatusOK,
			wantBody:   "20",
		},
		{
			name:       "withdraw from non-existing account",
			method:     http.MethodPost,
			target:     "/event",
			body:       `{"type":"withdraw","origin":"200","amount":10}`,
			wantStatus: http.StatusNotFound,
			wantBody:   "0",
		},
		{
			name:       "withdraw from existing account",
			method:     http.MethodPost,
			target:     "/event",
			body:       `{"type":"withdraw","origin":"100","amount":5}`,
			wantStatus: http.StatusCreated,
			wantBody:   "{\"origin\":{\"id\":\"100\",\"balance\":15}}\n",
		},
		{
			name:       "transfer from existing account",
			method:     http.MethodPost,
			target:     "/event",
			body:       `{"type":"transfer","origin":"100","amount":15,"destination":"300"}`,
			wantStatus: http.StatusCreated,
			wantBody:   "{\"origin\":{\"id\":\"100\",\"balance\":0},\"destination\":{\"id\":\"300\",\"balance\":15}}\n",
		},
		{
			name:       "transfer from non-existing account",
			method:     http.MethodPost,
			target:     "/event",
			body:       `{"type":"transfer","origin":"200","amount":15,"destination":"300"}`,
			wantStatus: http.StatusNotFound,
			wantBody:   "0",
		},
		{
			name:       "failed transfer preserves origin",
			method:     http.MethodGet,
			target:     "/balance?account_id=100",
			wantStatus: http.StatusOK,
			wantBody:   "0",
		},
		{
			name:       "failed transfer preserves destination",
			method:     http.MethodGet,
			target:     "/balance?account_id=300",
			wantStatus: http.StatusOK,
			wantBody:   "15",
		},
	}

	for _, step := range steps {
		passed := t.Run(step.name, func(t *testing.T) {
			response := performRequest(handler, step.method, step.target, step.body)
			assertResponse(t, response, step.wantStatus, step.wantBody)
		})
		if !passed {
			return
		}
	}
}
