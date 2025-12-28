package url

import (
	"crypto/rand"
	"errors"
	"math/big"
	"net/url"

	"url-shortener/internal/click"
)

type Service interface {
	CreateShortURL(userID int64, originalURL string) (string, error)
	GetOriginalURL(shortCode string) (string, error)
	ListURLs(userID int64) ([]*URL, error)
	GetUserStats(userID int64) ([]*URLStats, error)
}

type URLStats struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
	Clicks      int    `json:"clicks"`
}

type service struct {
	repo         Repository
	clickService click.Service
}

func NewService(repo Repository, clickService click.Service) Service {
	return &service{repo: repo, clickService: clickService}
}

const base62 = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func (s *service) CreateShortURL(userID int64, originalURL string) (string, error) {
	if !isValidURL(originalURL) {
		return "", errors.New("invalid URL")
	}
	shortCode := generateShortCode(6)
	if err := s.repo.Create(userID, originalURL, shortCode); err != nil {
		return "", err
	}
	return shortCode, nil
}

func (s *service) GetOriginalURL(shortCode string) (string, error) {
	u, err := s.repo.GetByShortCode(shortCode)
	if err != nil || u == nil {
		return "", errors.New("URL not found")
	}
	s.clickService.AddClick(u.ID)
	return u.OriginalURL, nil
}

func (s *service) ListURLs(userID int64) ([]*URL, error) {
	return s.repo.List(userID)
}

func (s *service) GetUserStats(userID int64) ([]*URLStats, error) {
	urls, err := s.repo.List(userID)
	if err != nil {
		return nil, err
	}
	var stats []*URLStats
	for _, u := range urls {
		count, _ := s.clickService.GetClicks(u.ID)
		stats = append(stats, &URLStats{
			OriginalURL: u.OriginalURL,
			ShortURL:    "http://localhost:8080/" + u.ShortCode,
			Clicks:      count,
		})
	}
	return stats, nil
}

func generateShortCode(length int) string {
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(base62))))
		result[i] = base62[n.Int64()]
	}
	return string(result)
}

func isValidURL(raw string) bool {
	u, err := url.ParseRequestURI(raw)
	return err == nil && u.Scheme != "" && u.Host != ""
}
