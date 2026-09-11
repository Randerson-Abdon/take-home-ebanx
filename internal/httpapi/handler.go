// Package httpapi provides the HTTP transport for the application.
package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/Randerson-Abdon/take-home-ebanx/internal/account"
)

// NewHandler creates the application HTTP handler.
func NewHandler(service *account.Service) http.Handler {
	handler := &handler{service: service}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("POST /reset", handler.reset)
	mux.HandleFunc("GET /balance", handler.balance)
	mux.HandleFunc("POST /event", handler.event)

	return allowCrossOriginRequests(mux)
}

type handler struct {
	service *account.Service
}

type eventRequest struct {
	Type        string `json:"type"`
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Amount      int64  `json:"amount"`
}

type accountResponse struct {
	ID      string `json:"id"`
	Balance int64  `json:"balance"`
}

type depositResponse struct {
	Destination accountResponse `json:"destination"`
}

type withdrawResponse struct {
	Origin accountResponse `json:"origin"`
}

type transferResponse struct {
	Origin      accountResponse `json:"origin"`
	Destination accountResponse `json:"destination"`
}

func allowCrossOriginRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		if request.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, request)
	})
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, "ok")
}

func (h *handler) reset(w http.ResponseWriter, _ *http.Request) {
	h.service.Reset()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, "OK")
}

func (h *handler) balance(w http.ResponseWriter, request *http.Request) {
	balance, err := h.service.Balance(request.URL.Query().Get("account_id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeNumber(w, http.StatusOK, balance)
}

func (h *handler) event(w http.ResponseWriter, request *http.Request) {
	event, err := decodeEvent(request.Body)
	if err != nil {
		writeNumber(w, http.StatusBadRequest, 0)
		return
	}

	switch event.Type {
	case "deposit":
		h.deposit(w, event)
	case "withdraw":
		h.withdraw(w, event)
	case "transfer":
		h.transfer(w, event)
	default:
		writeNumber(w, http.StatusBadRequest, 0)
	}
}

func (h *handler) deposit(w http.ResponseWriter, event eventRequest) {
	if event.Destination == "" {
		writeNumber(w, http.StatusBadRequest, 0)
		return
	}

	destination, err := h.service.Deposit(event.Destination, event.Amount)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, depositResponse{
		Destination: newAccountResponse(destination),
	})
}

func (h *handler) withdraw(w http.ResponseWriter, event eventRequest) {
	if event.Origin == "" {
		writeNumber(w, http.StatusBadRequest, 0)
		return
	}

	origin, err := h.service.Withdraw(event.Origin, event.Amount)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, withdrawResponse{
		Origin: newAccountResponse(origin),
	})
}

func (h *handler) transfer(w http.ResponseWriter, event eventRequest) {
	if event.Origin == "" || event.Destination == "" {
		writeNumber(w, http.StatusBadRequest, 0)
		return
	}

	origin, destination, err := h.service.Transfer(event.Origin, event.Destination, event.Amount)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, transferResponse{
		Origin:      newAccountResponse(origin),
		Destination: newAccountResponse(destination),
	})
}

func decodeEvent(body io.Reader) (eventRequest, error) {
	var event eventRequest
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&event); err != nil {
		return eventRequest{}, err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return eventRequest{}, errors.New("request body must contain one JSON object")
	}

	return event, nil
}

func newAccountResponse(domainAccount account.Account) accountResponse {
	return accountResponse{
		ID:      domainAccount.ID,
		Balance: domainAccount.Balance,
	}
}

func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, account.ErrAccountNotFound):
		writeNumber(w, http.StatusNotFound, 0)
	case errors.Is(err, account.ErrInvalidAmount):
		writeNumber(w, http.StatusBadRequest, 0)
	case errors.Is(err, account.ErrInsufficientFunds):
		writeNumber(w, http.StatusUnprocessableEntity, 0)
	default:
		writeNumber(w, http.StatusInternalServerError, 0)
	}
}

func writeNumber(w http.ResponseWriter, status int, value int64) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = fmt.Fprint(w, strconv.FormatInt(value, 10))
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
