package user

import "database/sql"

type Repository interface {
	Create(username, password string) error
	GetByUsername(username string) (*User, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(username, password string) error {
	_, err := r.db.Exec("INSERT INTO users (username, password) VALUES ($1,$2)", username, password)
	return err
}

func (r *repository) GetByUsername(username string) (*User, error) {
	row := r.db.QueryRow("SELECT id, username, password FROM users WHERE username=$1", username)
	u := &User{}
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}
