package ContentRepository

import (
	"database/sql"
)

// Repository struct'ı veritabanı bağlantısını tutar.
type Repository struct {
	db *sql.DB
}

// NewRepository yeni bir Repository instance'ı oluşturur.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}
