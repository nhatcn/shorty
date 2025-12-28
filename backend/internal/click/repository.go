package click

import "database/sql"

type Repository interface {
	Add(urlID int64) error
	Count(urlID int64) (int, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Add(urlID int64) error {
	_, err := r.db.Exec("INSERT INTO clicks (url_id) VALUES ($1)", urlID)
	return err
}

func (r *repository) Count(urlID int64) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM clicks WHERE url_id=$1", urlID).Scan(&count)
	return count, err
}
