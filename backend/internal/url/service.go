package url

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"strings"
	"time"
	"url-shortener/internal/click"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"

	"github.com/skip2/go-qrcode"
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
	baseURL := os.Getenv("FRONTEND_URL")
	if baseURL == "" {
		baseURL = "https://shorty-black.vercel.app"
	}
	baseURL += "/l/"

	id, err := s.repo.Create(userID, originalURL, "", "", expiresAt)
	if err != nil {
		return "", "", fmt.Errorf("failed to create URL record: %w", err)
	}

	shortCode := encodeBase62(id)
	shortURL := baseURL + shortCode

	qrBytes, err := qrcode.Encode(shortURL, qrcode.Medium, 256)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate QR code: %w", err)
	}

	uploadResp, err := s.cld.Upload.Upload(context.Background(), bytes.NewReader(qrBytes), uploader.UploadParams{
		PublicID: "qr_" + shortCode,
		Folder:   "qr_codes",
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to upload QR code: %w", err)
	}

	err = s.repo.UpdateShortCodeAndQR(id, shortCode, uploadResp.SecureURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to update short code and QR URL: %w", err)
	}

	return shortCode, uploadResp.SecureURL, nil
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
	return s.repo.GetUserStats(userID)
}

func (s *service) GetURLByID(id int64) (*URL, error) {
	return s.repo.GetByID(id)
}

func (s *service) DeleteURL(id int64) error {
	return s.repo.DeleteByID(id)
}
func encodeBase62(num int64) string {
	if num == 0 {
		return string(base62[0])
	}
	var result []byte
	for num > 0 {
		result = append([]byte{base62[num%62]}, result...)
		num /= 62
	}
	return string(result)
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

	if len(rawURL) < 4 {
		return errors.New("URL too short (minimum 4 characters)")
	}

	normalizedURL := rawURL
	if !strings.HasPrefix(strings.ToLower(rawURL), "http://") && 
	   !strings.HasPrefix(strings.ToLower(rawURL), "https://") {
		normalizedURL = "https://" + rawURL
	}

	u, err := url.ParseRequestURI(normalizedURL)
	if err != nil {
		return errors.New("invalid URL format")
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("only HTTP and HTTPS protocols are allowed")
	}

	if u.Host == "" {
		return errors.New("URL must contain a valid host")
	}

	hostParts := strings.Split(u.Host, ":")
	hostname := hostParts[0]

	if hostname == "" {
		return errors.New("URL must contain a valid hostname")
	}

	if err := validateHostname(hostname); err != nil {
		return err
	}

	hostnameLower := strings.ToLower(hostname)
	for _, blocked := range blacklistedDomains {
		if hostnameLower == blocked || strings.HasSuffix(hostnameLower, "."+blocked) {
			return fmt.Errorf("domain '%s' is not allowed", hostname)
		}
	}

	return nil
}

func validateHostname(hostname string) error {
	if len(hostname) > 253 {
		return errors.New("hostname too long (max 253 characters)")
	}

	if strings.Contains(hostname, "..") {
		return errors.New("hostname cannot contain consecutive dots")
	}

	if strings.HasPrefix(hostname, ".") || strings.HasSuffix(hostname, ".") {
		return errors.New("hostname cannot start or end with a dot")
	}

	if strings.HasPrefix(hostname, "-") || strings.HasSuffix(hostname, "-") {
		return errors.New("hostname cannot start or end with a hyphen")
	}

	parts := strings.Split(strings.ToLower(hostname), ".")
	
	if len(parts) < 2 {
		isIP := isValidIPv4(hostname)
		if !isIP {
			return errors.New("hostname must be a valid domain or IP address")
		}
		return nil
	}

	for _, part := range parts {
		if part == "" {
			return errors.New("hostname contains empty labels")
		}

		if len(part) > 63 {
			return errors.New("hostname label too long (max 63 characters)")
		}

		if strings.HasPrefix(part, "-") || strings.HasSuffix(part, "-") {
			return errors.New("hostname labels cannot start or end with hyphen")
		}

		for _, char := range part {
			isValid := (char >= 'a' && char <= 'z') ||
				(char >= 'A' && char <= 'Z') ||
				(char >= '0' && char <= '9') ||
				char == '-'
			
			if !isValid {
				return fmt.Errorf("hostname contains invalid character: '%c'", char)
			}
		}
	}

	return nil
}

func isValidIPv4(host string) bool {
	parts := strings.Split(host, ".")
	if len(parts) != 4 {
		return false
	}

	for _, part := range parts {
		if len(part) == 0 || len(part) > 3 {
			return false
		}

		num := 0
		for _, char := range part {
			if char < '0' || char > '9' {
				return false
			}
			num = num*10 + int(char-'0')
		}

		if num > 255 {
			return false
		}

		if len(part) > 1 && part[0] == '0' {
			return false
		}
	}

	return true
}