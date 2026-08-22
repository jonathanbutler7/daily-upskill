package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	cmd "ledger-db/cmd"
	"ledger-db/internal/ledgerstore"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	defaultAddr = ":8080"
	defaultDSN  = "postgresql://ledger_db:password@localhost:5432/ledger_db"
)

type server struct {
	db *sql.DB
}

type postTransferRequest struct {
	FromAccountID  int64  `json:"from_account_id"`
	ToAccountID    int64  `json:"to_account_id"`
	Amount         int64  `json:"amount"`
	IdempotencyKey string `json:"idempotency_key"`
}

type postExternalTransferRequest struct {
	UserAccountID             int64  `json:"user_account_id"`
	TransferAmount            int64  `json:"transfer_amount"`
	Rail                      string `json:"rail"`
	ExternalReference         string `json:"external_reference"`
	IdempotencyKey            string `json:"idempotency_key"`
	ExternalTransferDirection string `json:"direction"`
}

type reversalRequest struct {
	TransactionID  int64  `json:"transaction_id"`
	IdempotencyKey string `json:"idempotency_key"`
	Reason         string `json:"reason"`
}

type commandResponse struct {
	TransactionID int64 `json:"transaction_id"`
}

type errorResponse struct {
	Error ledgerstore.LedgerErrorInfo `json:"error"`
}

func main() {
	ctx := context.Background()
	db, err := sql.Open("pgx", envOrDefault("LEDGER_DB_DSN", defaultDSN))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		log.Fatal(err)
	}

	srv := &server{db: db}
	httpServer := &http.Server{
		Addr:         serviceAddr(),
		Handler:      srv.routes(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("ledger-db service listening on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	<-shutdown

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc("/healthz", requireMethod(http.MethodGet, s.handleHealth))
	mux.HandleFunc("/routes", requireMethod(http.MethodGet, s.handleRoutes))
	mux.HandleFunc("/transfers", requireMethod(http.MethodPost, s.handlePostTransfer))
	mux.HandleFunc("/external-transfers", requireMethod(http.MethodPost, s.handlePostExternalTransfer))
	mux.HandleFunc("/reversals", requireMethod(http.MethodPost, s.handleReversal))
	return mux
}

func (s *server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeNotFound(w)
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}
	s.handleRoutes(w, r)
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := s.db.PingContext(ctx); err != nil {
		writeLedgerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) handleRoutes(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"service": "ledger-db",
		"routes": []map[string]string{
			{"method": "GET", "path": "/healthz", "command": "health check"},
			{"method": "POST", "path": "/transfers", "command": "PostTransfer"},
			{"method": "POST", "path": "/external-transfers", "command": "PostExternalTransfer"},
			{"method": "POST", "path": "/reversals", "command": "Reversal"},
		},
	})
}

func (s *server) handlePostTransfer(w http.ResponseWriter, r *http.Request) {
	var req postTransferRequest
	if !decodeRequest(w, r, &req) {
		return
	}

	transactionID, err := cmd.PostTransfer(r.Context(), s.db, ledgerstore.TransferCommand{
		FromAccountID:  ledgerstore.AccountID(req.FromAccountID),
		ToAccountID:    ledgerstore.AccountID(req.ToAccountID),
		Amount:         ledgerstore.Amount(req.Amount),
		IdempotencyKey: ledgerstore.IdempotencyKey(req.IdempotencyKey),
	})
	if err != nil {
		writeLedgerError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, commandResponse{TransactionID: transactionID})
}

func (s *server) handlePostExternalTransfer(w http.ResponseWriter, r *http.Request) {
	var req postExternalTransferRequest
	if !decodeRequest(w, r, &req) {
		return
	}

	transactionID, err := cmd.PostExternalTransfer(r.Context(), s.db, ledgerstore.PostExternalTransferCommand{
		UserAccountID:             ledgerstore.AccountID(req.UserAccountID),
		TransferAmount:            ledgerstore.Amount(req.TransferAmount),
		Rail:                      ledgerstore.PaymentRail(req.Rail),
		ExternalReference:         ledgerstore.ExternalReference(req.ExternalReference),
		IdempotencyKey:            ledgerstore.IdempotencyKey(req.IdempotencyKey),
		ExternalTransferDirection: ledgerstore.ExternalTransferDirection(req.ExternalTransferDirection),
	})
	if err != nil {
		writeLedgerError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, commandResponse{TransactionID: int64(transactionID)})
}

func (s *server) handleReversal(w http.ResponseWriter, r *http.Request) {
	var req reversalRequest
	if !decodeRequest(w, r, &req) {
		return
	}

	transactionID, err := cmd.Reversal(r.Context(), s.db, ledgerstore.ReversalCommand{
		TransactionID:  ledgerstore.TransactionID(req.TransactionID),
		IdempotencyKey: ledgerstore.IdempotencyKey(req.IdempotencyKey),
		Reason:         ledgerstore.Reason(req.Reason),
	})
	if err != nil {
		writeLedgerError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, commandResponse{TransactionID: transactionID})
}

func decodeRequest(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		writeBadRequest(w, err)
		return false
	}

	if err := decoder.Decode(&struct{}{}); err == nil {
		writeBadRequest(w, errors.New("request body must contain a single JSON object"))
		return false
	} else if err != io.EOF {
		writeBadRequest(w, err)
		return false
	}

	return true
}

func writeBadRequest(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusBadRequest, errorResponse{
		Error: ledgerstore.LedgerErrorInfo{
			Code:     ledgerstore.LedgerErrorCode("validation.invalid_request"),
			Category: ledgerstore.LedgerErrorCategoryValidation,
			Expected: true,
			Message:  err.Error(),
		},
	})
}

func writeNotFound(w http.ResponseWriter) {
	writeJSON(w, http.StatusNotFound, errorResponse{
		Error: ledgerstore.LedgerErrorInfo{
			Code:     ledgerstore.LedgerErrorCode("not_found.route"),
			Category: ledgerstore.LedgerErrorCategoryNotFound,
			Expected: true,
			Message:  "route not found",
		},
	})
}

func writeMethodNotAllowed(w http.ResponseWriter, allowedMethod string) {
	w.Header().Set("Allow", allowedMethod)
	writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
		Error: ledgerstore.LedgerErrorInfo{
			Code:     ledgerstore.LedgerErrorCode("validation.method_not_allowed"),
			Category: ledgerstore.LedgerErrorCategoryValidation,
			Expected: true,
			Message:  "method not allowed",
		},
	})
}

func writeLedgerError(w http.ResponseWriter, err error) {
	info := ledgerstore.ClassifyError(err)
	writeJSON(w, statusForLedgerError(info), errorResponse{Error: info})
}

func statusForLedgerError(info ledgerstore.LedgerErrorInfo) int {
	switch info.Category {
	case ledgerstore.LedgerErrorCategoryValidation:
		return http.StatusBadRequest
	case ledgerstore.LedgerErrorCategoryNotFound:
		return http.StatusNotFound
	case ledgerstore.LedgerErrorCategoryBusiness:
		return http.StatusConflict
	case ledgerstore.LedgerErrorCategoryDB:
		if info.Retryable {
			return http.StatusServiceUnavailable
		}
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("write response: %v", err)
	}
}

func requireMethod(method string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			writeMethodNotAllowed(w, method)
			return
		}
		handler(w, r)
	}
}

func serviceAddr() string {
	if addr := os.Getenv("LEDGER_DB_ADDR"); addr != "" {
		return addr
	}
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}
	return defaultAddr
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
