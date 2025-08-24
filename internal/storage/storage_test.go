package storage

import (
	"os"
	"path/filepath"
	"radio/internal/client"
	"testing"
)

func tempFilePath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "test_storage.json")
}

func testStation() client.Station {
	return client.Station{
		URL:     "http://example.com/stream",
		Name:    "Test Station",
		Bitrate: 128,
		Country: "Testland",
		Tags:    "pop,rock",
	}
}

func TestAddAndListFavorites(t *testing.T) {
	path := tempFilePath(t)

	s, err := NewStorage(path)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	station := testStation()

	err = s.AddFavorite(station)
	if err != nil {
		t.Fatalf("AddFavorite failed: %v", err)
	}

	favs := s.ListFavorites()
	if len(favs) != 1 {
		t.Fatalf("expected 1 favorite, got %d", len(favs))
	}

	if favs[0].URL != station.URL {
		t.Errorf("expected URL %s, got %s", station.URL, favs[0].URL)
	}
}

func TestIsFavorite(t *testing.T) {
	path := tempFilePath(t)
	s, _ := NewStorage(path)
	station := testStation()

	if s.IsFavorite(station.URL) {
		t.Error("expected IsFavorite to be false")
	}

	s.AddFavorite(station)

	if !s.IsFavorite(station.URL) {
		t.Error("expected IsFavorite to be true after AddFavorite")
	}
}

func TestRemoveFavorite(t *testing.T) {
	path := tempFilePath(t)
	s, _ := NewStorage(path)
	station := testStation()

	s.AddFavorite(station)
	if err := s.RemoveFavorite(station.URL); err != nil {
		t.Fatalf("RemoveFavorite failed: %v", err)
	}

	if s.IsFavorite(station.URL) {
		t.Error("expected station to be removed from favorites")
	}
}

func TestPersistence(t *testing.T) {
	path := tempFilePath(t)
	station := testStation()

	s1, _ := NewStorage(path)
	s1.AddFavorite(station)

	s2, err := NewStorage(path)
	if err != nil {
		t.Fatalf("failed to reopen storage: %v", err)
	}

	if !s2.IsFavorite(station.URL) {
		t.Error("expected station to persist after reload")
	}
}

func TestEmptyStorageFile(t *testing.T) {
	path := tempFilePath(t)

	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write empty file: %v", err)
	}

	s, err := NewStorage(path)
	if err != nil {
		t.Fatalf("failed to load empty storage file: %v", err)
	}

	if len(s.ListFavorites()) != 0 {
		t.Error("expected empty favorites for empty file")
	}
}
