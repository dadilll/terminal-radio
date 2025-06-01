package storage

import (
	"encoding/json"
	"os"
	"radio/internal/api"
	"radio/pkg/logger"
	"sync"
)

type Storage struct {
	Favorites map[string]FavoriteStation `json:"favorites"`
	path      string
	mu        sync.Mutex
}

type FavoriteStation struct {
	URL     string `json:"url"`
	Name    string `json:"name"`
	Bitrate int    `json:"bitrate"`
	Country string `json:"country"`
	Tags    string `json:"tags"`
}

func NewStorage(path string) (*Storage, error) {
	s := &Storage{
		Favorites: make(map[string]FavoriteStation),
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
		s.Favorites = make(map[string]FavoriteStation)
		return nil
	}

	var tmp struct {
		Favorites map[string]FavoriteStation `json:"favorites"`
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
	// Предполагается, что вызывающий уже захватил s.mu
	data, err := json.MarshalIndent(struct {
		Favorites map[string]FavoriteStation `json:"favorites"`
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

func (s *Storage) AddFavorite(station api.Station) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Favorites[station.URL] = FavoriteStation{
		URL:     station.URL,
		Name:    station.Name,
		Bitrate: station.Bitrate,
		Country: station.Country,
		Tags:    station.Tags,
	}
	return s.save()
}

func (s *Storage) RemoveFavorite(url string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.Favorites, url)
	return s.save()
}

func (s *Storage) IsFavorite(url string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.Favorites[url]
	return exists
}

func (s *Storage) ListFavorites() []api.Station {
	s.mu.Lock()
	defer s.mu.Unlock()

	stations := make([]api.Station, 0, len(s.Favorites))
	for _, fav := range s.Favorites {
		stations = append(stations, api.Station{
			URL:     fav.URL,
			Name:    fav.Name,
			Bitrate: fav.Bitrate,
			Country: fav.Country,
			Tags:    fav.Tags,
		})
	}
	return stations
}
