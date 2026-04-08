package currency_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/trancee/DealScout/internal/currency"
	"github.com/trancee/DealScout/internal/storage"
)

func mustOpenDB(t *testing.T) *storage.Database {
	t.Helper()
	db, err := storage.Open(":memory:")
	if err != nil {
		t.Fatalf("Open(:memory:): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func mockRateServer(t *testing.T, rates map[string]float64) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"base":  "CHF",
			"rates": rates,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

func TestConvertEURtoCHF(t *testing.T) {
	db := mustOpenDB(t)
	server := mockRateServer(t, map[string]float64{"EUR": 1.0753})
	defer server.Close()

	conv := currency.New(db, server.URL, "CHF", 24)

	if err := conv.RefreshRates(); err != nil {
		t.Fatalf("RefreshRates: %v", err)
	}

	got, err := conv.Convert(100.0, "EUR")
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}

	// 100 EUR × (1/1.0753) = ~93.00 CHF — but we store rate as EUR→CHF
	// The API gives base=CHF, rates={EUR: 1.0753} meaning 1 CHF = 1.0753 EUR
	// So EUR→CHF = 1/1.0753
	want := 100.0 / 1.0753
	if diff := got - want; diff > 0.01 || diff < -0.01 {
		t.Errorf("Convert(100, EUR) = %f, want ~%f", got, want)
	}
}

func TestConvertBaseCurrencyPassthrough(t *testing.T) {
	db := mustOpenDB(t)
	conv := currency.New(db, "", "CHF", 24)

	got, err := conv.Convert(250.0, "CHF")
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if got != 250.0 {
		t.Errorf("Convert(250, CHF) = %f, want 250.0", got)
	}
}

func TestConvertUsesCache(t *testing.T) {
	db := mustOpenDB(t)

	// Pre-seed the cache.
	_ = db.UpsertExchangeRate("EUR", 1.08)

	// Server that should NOT be called.
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	conv := currency.New(db, server.URL, "CHF", 24)

	// RefreshRates should use cache, not call server.
	_ = conv.RefreshRates()

	got, err := conv.Convert(100.0, "EUR")
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}

	want := 100.0 / 1.08
	if diff := got - want; diff > 0.01 || diff < -0.01 {
		t.Errorf("Convert(100, EUR) = %f, want ~%f", got, want)
	}

	if calls > 0 {
		t.Error("server was called despite fresh cache")
	}
}

func TestConvertUnknownCurrencyFails(t *testing.T) {
	db := mustOpenDB(t)
	conv := currency.New(db, "", "CHF", 24)

	_, err := conv.Convert(100.0, "XYZ")
	if err == nil {
		t.Fatal("expected error for unknown currency")
	}
}
