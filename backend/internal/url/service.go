package url

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"time"
	"os"
	"url-shortener/internal/click"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/skip2/go-qrcode"
	"github.com/lib/pq" // PostgreSQL error codes
)

type Service interface {
	CreateShortURL(userID int64, originalURL string, expiresAt time.Time) (string, string, error)
	GetOriginalURL(shortCode string) (string, error)
	ListURLs(userID int64) ([]*URL, error)
	GetUserStats(userID int64) ([]*URLStats, error)
	DeleteURL(id int64) error
	GetURLByID(id int64) (*URL, error)
}

type URLStats struct {
	ID          int64     `json:"id"`
	OriginalURL string    `json:"original_url"`
	ShortURL    string    `json:"short_url"`
	QRURL       string    `json:"qr_url"`
	Clicks      int       `json:"clicks"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type service struct {
	repo         Repository
	clickService click.Service
	cld          *cloudinary.Cloudinary
}

func NewService(repo Repository, clickService click.Service, cld *cloudinary.Cloudinary) Service {
	return &service{repo: repo, clickService: clickService, cld: cld}
}

const base62 = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var blacklistedDomains = []string{
	"localhost",
	"127.0.0.1",
	"0.0.0.0",
	"[::1]",
	"internal",
	"local",
}


func (s *service) CreateShortURL(userID int64, originalURL string, expiresAt time.Time) (string, string, error) {

	if err := validateURL(originalURL); err != nil {
		return "", "", err
	}


	if expiresAt.Before(time.Now()) {
		return "", "", errors.New("expiration date must be in the future")
	}

	count, err := s.repo.CountURLsCreatedToday(userID)
	if err != nil {
		return "", "", fmt.Errorf("failed to check rate limit")
	}
	if count >= 100 {
		return "", "", errors.New("daily limit exceeded (100 URLs per day)")
	}

	existingURL, err := s.repo.FindExistingURL(userID, originalURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to check existing URL")
	}


	if existingURL != nil {

		if existingURL.ExpiresAt.After(time.Now()) {
			return existingURL.ShortCode, existingURL.QRURL, nil
		}
	
	}

	return s.createNewShortURL(userID, originalURL, expiresAt)
}


func (s *service) createNewShortURL(userID int64, originalURL string, expiresAt time.Time) (string, string, error) {
	maxRetries := 5
	baseURL := os.Getenv("BACKEND_URL"+"/")
	if baseURL == "" {
		baseURL = "http://localhost:8080/"
	}

	for attempt := 0; attempt < maxRetries; attempt++ {

		shortCode := generateShortCode(6)
		shortURL := baseURL + shortCode


		qrBytes, err := qrcode.Encode(shortURL, qrcode.Medium, 256)
		if err != nil {
			return "", "", fmt.Errorf("failed to generate QR code")
		}

		uploadResp, err := s.cld.Upload.Upload(context.Background(), bytes.NewReader(qrBytes), uploader.UploadParams{
			PublicID: "qr_" + shortCode,
			Folder:   "qr_codes",
		})
		if err != nil {
			return "", "", fmt.Errorf("failed to upload QR code")
		}

		err = s.repo.Create(userID, originalURL, shortCode, uploadResp.SecureURL, expiresAt)
		if err == nil {
			return shortCode, uploadResp.SecureURL, nil
		}
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": 
				if strings.Contains(pqErr.Message, "short_code") {

					continue
				}
				if strings.Contains(pqErr.Message, "user_original_url") {

					existingURL, fetchErr := s.repo.FindExistingURL(userID, originalURL)
					if fetchErr == nil && existingURL != nil {
						return existingURL.ShortCode, existingURL.QRURL, nil
					}
				
					continue
				}
			}
		}

		
		return "", "", fmt.Errorf("database error")
	}

	return "", "", errors.New("failed to generate unique short code after multiple attempts")
}

func (s *service) GetOriginalURL(shortCode string) (string, error) {
	u, err := s.repo.GetByShortCode(shortCode)
	if err != nil || u == nil {
		return "", errors.New("URL not found")
	}

	if u.ExpiresAt.Before(time.Now()) {
		return "", errors.New("URL has expired")
	}

	go s.clickService.AddClick(u.ID)

	return u.OriginalURL, nil
}

func (s *service) ListURLs(userID int64) ([]*URL, error) {
	return s.repo.List(userID)
}

func (s *service) GetUserStats(userID int64) ([]*URLStats, error) {
	return s.repo.GetUserStatsOptimized(userID)
}

func (s *service) GetURLByID(id int64) (*URL, error) {
	return s.repo.GetByID(id)
}

func (s *service) DeleteURL(id int64) error {
	return s.repo.DeleteByID(id)
}


func generateShortCode(length int) string {
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(base62))))
		result[i] = base62[n.Int64()]
	}
	return string(result)
}


func validateURL(rawURL string) error {
	if len(rawURL) > 2048 {
		return errors.New("URL too long (max 2048 characters)")
	}

	if len(rawURL) < 10 {
		return errors.New("URL too short")
	}

	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format")
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("only HTTP and HTTPS protocols are allowed")
	}

	if u.Host == "" {
		return errors.New("URL must contain a valid host")
	}

	
	hostLower := strings.ToLower(u.Host)
	for _, blocked := range blacklistedDomains {
		if strings.Contains(hostLower, blocked) {
			return fmt.Errorf("domain '%s' is not allowed", u.Host)
		}
	}

	return nil
}