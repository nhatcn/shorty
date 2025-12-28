package url

import "database/sql"
import "url-shortener/internal/click"
type Repository interface {
	Create(userID int64, originalURL, shortCode string) error
	GetByShortCode(shortCode string) (*URL, error)
	List(userID int64) ([]*URL, error)
	ListWithClicks(userID int64, clickRepo click.Repository) ([]*URLStats, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(userID int64, originalURL, shortCode string) error {
	_, err := r.db.Exec("INSERT INTO urls (user_id, original_url, short_code) VALUES ($1,$2,$3)", userID, originalURL, shortCode)
	return err
}

func (r *repository) GetByShortCode(shortCode string) (*URL, error) {
	row := r.db.QueryRow("SELECT id,user_id,original_url,short_code,created_at FROM urls WHERE short_code=$1", shortCode)
	u := &URL{}
	if err := row.Scan(&u.ID, &u.UserID, &u.OriginalURL, &u.ShortCode, &u.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func (r *repository) List(userID int64) ([]*URL, error) {
	rows, err := r.db.Query("SELECT id,user_id,original_url,short_code,created_at FROM urls WHERE user_id=$1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var urls []*URL
	for rows.Next() {
		u := &URL{}
		if err := rows.Scan(&u.ID, &u.UserID, &u.OriginalURL, &u.ShortCode, &u.CreatedAt); err != nil {
			return nil, err
		}
		urls = append(urls, u)
	}
	return urls, nil
}
func (r *repository) ListWithClicks(userID int64, clickRepo click.Repository) ([]*URLStats, error) {
	rows, err := r.db.Query("SELECT id, original_url, short_code FROM urls WHERE user_id=$1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*URLStats
	for rows.Next() {
		var u URL
		if err := rows.Scan(&u.ID, &u.OriginalURL, &u.ShortCode); err != nil {
			return nil, err
		}
		count, _ := clickRepo.Count(u.ID)
		stats = append(stats, &URLStats{
			OriginalURL: u.OriginalURL,
			ShortURL:    "http://localhost:8080/" + u.ShortCode,
			Clicks:      count,
		})
	}
	return stats, nil
}

