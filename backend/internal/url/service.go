package url

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"net/url"
	"time"

	"url-shortener/internal/click"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/skip2/go-qrcode"
)

type Service interface {
	CreateShortURL(userID int64, originalURL string, expriesAt time.Time) (string, string, error)
	GetOriginalURL(shortCode string) (string, error)
	ListURLs(userID int64) ([]*URL, error)
	GetUserStats(userID int64) ([]*URLStats, error)
	DeleteURL(id int64) error
}

type URLStats struct {
	ID          int64  `json:"id"`
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
	QRURL       string `json:"qr_url"`
	Clicks      int    `json:"clicks"`
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

func (s *service) CreateShortURL(userID int64, originalURL string, expiresAt time.Time) (string, string, error) {
	if !isValidURL(originalURL) {
		return "", "", errors.New("invalid URL")
	}
	if (expiresAt.Before(time.Now())) {
		return "", "", errors.New("expiration date must be in the future")
	}
	shortCode := generateShortCode(6)
	shortURL := "http://localhost:8080/" + shortCode

	// Tạo QR code dưới dạng PNG trong memory
	qrBytes, err := qrcode.Encode(shortURL, qrcode.Medium, 256)
	if err != nil {
		return "", "", err
	}

	// Upload QR code lên Cloudinary
	uploadResp, err := s.cld.Upload.Upload(context.Background(), bytes.NewReader(qrBytes), uploader.UploadParams{
		PublicID: "qr_" + shortCode,
		Folder:   "qr_codes",
	})
	if err != nil {
		return "", "", err
	}

	// Lưu vào DB
	if err := s.repo.Create(userID, originalURL, shortCode, uploadResp.SecureURL, expiresAt); err != nil {
		return "", "", err
	}

	return shortCode, uploadResp.SecureURL, nil
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
			ID:          u.ID,
			OriginalURL: u.OriginalURL,
			ShortURL:    "http://localhost:8080/" + u.ShortCode,
			QRURL:       u.QRURL,
			CreatedAt:   u.CreatedAt,
			ExpiresAt:   u.ExpiresAt,
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
func (s *service) DeleteURL(id int64) error {
	return s.repo.DeleteByID(id)
}
