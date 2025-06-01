package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_SearchStations(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/stations/search" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			// Проверяем query param name
			name := r.URL.Query().Get("name")
			if name != "rock" {
				t.Errorf("unexpected query param name: %s", name)
			}

			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `[{"name":"Rock FM","url":"http://rockfm.example","country":"US","codec":"mp3","bitrate":128}]`)
		}))
		defer server.Close()

		client := NewClient(server.URL, 5*time.Second)
		stations, err := client.SearchStations(context.Background(), "rock")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(stations) != 1 {
			t.Fatalf("expected 1 station, got %d", len(stations))
		}

		if stations[0].Name != "Rock FM" {
			t.Errorf("unexpected station name: %s", stations[0].Name)
		}
	})

	// Пустой query
	t.Run("empty query", func(t *testing.T) {
		client := NewClient("http://example.com", 5*time.Second)
		_, err := client.SearchStations(context.Background(), "")
		if err == nil {
			t.Fatal("expected error for empty query, got nil")
		}
	})

	t.Run("http status not ok", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "bad request", http.StatusBadRequest)
		}))
		defer server.Close()

		client := NewClient(server.URL, 5*time.Second)
		_, err := client.SearchStations(context.Background(), "rock")
		if err == nil {
			t.Fatal("expected error for bad status, got nil")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `invalid json`)
		}))
		defer server.Close()

		client := NewClient(server.URL, 5*time.Second)
		_, err := client.SearchStations(context.Background(), "rock")
		if err == nil {
			t.Fatal("expected error for invalid json, got nil")
		}
	})

	t.Run("request error", func(t *testing.T) {
		client := NewClient("http://invalid-host", 1*time.Second)
		_, err := client.SearchStations(context.Background(), "rock")
		if err == nil {
			t.Fatal("expected error for request failure, got nil")
		}
	})
}
