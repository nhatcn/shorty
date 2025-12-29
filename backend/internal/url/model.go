package url

import "time"

type URL struct {
	ID          int64
	UserID      int64
	OriginalURL string
	ShortCode   string
	QRURL       string
	CreatedAt   time.Time
	ExpiresAt   time.Time
}
