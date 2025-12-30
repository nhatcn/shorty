package url

import (
	"database/sql"
	"time"
	"os"
)

type Repository interface {
	Create(userID int64, originalURL, shortCode, qrURL string, expiresAt time.Time) (int64, error)
	GetByShortCode(shortCode string) (*URL, error)
	GetByID(id int64) (*URL, error)
	FindExistingURL(userID int64, originalURL string) (*URL, error) 
	List(userID int64) ([]*URL, error)
	GetUserStats(userID int64) ([]*URLStats, error)
	DeleteByID(id int64) error
	CountURLsCreatedToday(userID int64) (int, error)
	UpdateShortCodeAndQR(id int64, shortCode, qrURL string) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(userID int64, originalURL, shortCode, qrURL string, expiresAt time.Time) (int64, error) {
	var id int64
	err := r.db.QueryRow(
		"INSERT INTO urls (user_id, original_url, short_code, qr_url, expires_at) VALUES ($1,$2,$3,$4,$5) RETURNING id",
		userID, originalURL, shortCode, qrURL, expiresAt,
	).Scan(&id)
	return id, err
}

func (r *repository) GetByShortCode(shortCode string) (*URL, error) {
	row := r.db.QueryRow(
		"SELECT id, user_id, original_url, short_code, qr_url, created_at, expires_at FROM urls WHERE short_code=$1",
		shortCode,
	)
	u := &URL{}
	if err := row.Scan(&u.ID, &u.UserID, &u.OriginalURL, &u.ShortCode, &u.QRURL, &u.CreatedAt, &u.ExpiresAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func (r *repository) GetByID(id int64) (*URL, error) {
	row := r.db.QueryRow(
		"SELECT id, user_id, original_url, short_code, qr_url, created_at, expires_at FROM urls WHERE id=$1",
		id,
	)
	u := &URL{}
	if err := row.Scan(&u.ID, &u.UserID, &u.OriginalURL, &u.ShortCode, &u.QRURL, &u.CreatedAt, &u.ExpiresAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}


func (r *repository) FindExistingURL(userID int64, originalURL string) (*URL, error) {
	row := r.db.QueryRow(`
		SELECT id, user_id, original_url, short_code, qr_url, created_at, expires_at 
		FROM urls 
		WHERE user_id = $1 AND original_url = $2
		LIMIT 1
	`, userID, originalURL)

	u := &URL{}
	err := row.Scan(&u.ID, &u.UserID, &u.OriginalURL, &u.ShortCode, &u.QRURL, &u.CreatedAt, &u.ExpiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil 
		}
		return nil, err
	}
	return u, nil
}

func (r *repository) List(userID int64) ([]*URL, error) {
	rows, err := r.db.Query(
		"SELECT id, user_id, original_url, short_code, qr_url, created_at, expires_at FROM urls WHERE user_id=$1 ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []*URL
	for rows.Next() {
		u := &URL{}
		if err := rows.Scan(&u.ID, &u.UserID, &u.OriginalURL, &u.ShortCode, &u.QRURL, &u.CreatedAt, &u.ExpiresAt); err != nil {
			return nil, err
		}
		urls = append(urls, u)
	}
	return urls, nil
}


func (r *repository) GetUserStats(userID int64) ([]*URLStats, error) {
	baseURL := os.Getenv("FRONTEND_URL") + "/"
	if baseURL == "" {
		baseURL = "https://shorty-black.vercel.app/"
	}

	query := `
		SELECT 
			u.id,
			u.original_url,
			u.short_code,
			u.qr_url,
			u.created_at,
			u.expires_at,
			COUNT(c.id) as clicks
		FROM urls u
		LEFT JOIN clicks c ON c.url_id = u.id
		WHERE u.user_id = $1
		GROUP BY u.id, u.original_url, u.short_code, u.qr_url, u.created_at, u.expires_at
		ORDER BY u.created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*URLStats
	for rows.Next() {
		var s URLStats
		var shortCode string
		var clicks int64
		if err := rows.Scan(
			&s.ID,
			&s.OriginalURL,
			&shortCode,
			&s.QRURL,
			&s.CreatedAt,
			&s.ExpiresAt,
			&clicks,
		); err != nil {
			return nil, err
		}
		s.Clicks = int(clicks)
		s.ShortURL = baseURL + shortCode
		stats = append(stats, &s)
	}
	return stats, nil
}

func (r *repository) DeleteByID(id int64) error {
	
	_, err := r.db.Exec("DELETE FROM clicks WHERE url_id=$1", id)
	if err != nil {
		return err
	}
	
	_, err = r.db.Exec("DELETE FROM urls WHERE id=$1", id)
	return err
}

func (r *repository) CountURLsCreatedToday(userID int64) (int, error) {
	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM urls WHERE user_id=$1 AND created_at >= CURRENT_DATE",
		userID,
	).Scan(&count)
	return count, err
}
func (r *repository) UpdateShortCodeAndQR(id int64, shortCode, qrURL string) error {
	_, err := r.db.Exec(
		"UPDATE urls SET short_code=$1, qr_url=$2 WHERE id=$3",
		shortCode, qrURL, id,
	)
	return err
}