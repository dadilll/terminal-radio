package storage

import (
	"encoding/json"
	"os"
	"radio/pkg/logger"
	"sync"
)

type Storage struct {
	Favorites map[string]bool `json:"favorites"`
	path      string
	mu        sync.Mutex
}

func NewStorage(path string) (*Storage, error) {
	s := &Storage{
		Favorites: make(map[string]bool),
		path:      path,
	}

	if err := s.load(); err != nil {
		logger.Log.Error().Err(err).Msgf("Failed to load storage from file %s: %v", path, err)
		return nil, err
	}

	logger.Log.Info().Msgf("Storage loaded from %s successfully", path)
	return s, nil
}

func (s *Storage) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := os.Stat(s.path); os.IsNotExist(err) {
		logger.Log.Warn().Msgf("Storage file %s does not exist, starting with empty favorites", s.path)
		return nil
	}

	data, err := os.ReadFile(s.path)
	if err != nil {
		logger.Log.Error().Err(err).Msgf("Error reading storage file %s", s.path)
		return err
	}

	if len(data) == 0 {
		logger.Log.Warn().Msgf("Storage file %s is empty, initializing with empty favorites", s.path)
		s.Favorites = make(map[string]bool)
		return nil
	}

	var tmp struct {
		Favorites map[string]bool `json:"favorites"`
	}

	if err := json.Unmarshal(data, &tmp); err != nil {
		logger.Log.Error().Err(err).Msgf("Error unmarshaling JSON from %s", s.path)
		return err
	}

	s.Favorites = tmp.Favorites

	logger.Log.Info().Msgf("Storage loaded from %s successfully", s.path)
	return nil
}

func (s *Storage) save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(struct {
		Favorites map[string]bool `json:"favorites"`
	}{
		Favorites: s.Favorites,
	}, "", "  ")
	if err != nil {
		logger.Log.Error().Err(err).Msg("Error marshaling storage data")
		return err
	}

	if err := os.WriteFile(s.path, data, 0644); err != nil {
		logger.Log.Error().Err(err).Msgf("Error writing storage file %s", s.path)
		return err
	}

	logger.Log.Info().Msgf("Storage saved successfully with %d favorites", len(s.Favorites))
	return nil
}

func (s *Storage) AddFavorite(stationID string) error {
	s.Favorites[stationID] = true
	return s.save()
}

func (s *Storage) RemoveFavorite(stationID string) error {
	delete(s.Favorites, stationID)
	return s.save()
}

func (s *Storage) IsFavorite(stationID string) bool {
	return s.Favorites[stationID]
}

func (s *Storage) ListFavorites() []string {
	favs := make([]string, 0, len(s.Favorites))
	for id := range s.Favorites {
		favs = append(favs, id)
	}
	return favs
}
