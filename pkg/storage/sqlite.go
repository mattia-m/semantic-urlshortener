package storage

import (
	"database/sql"
	"fmt"
	"time"
)

type URLStorage struct {
	db *sql.DB
}

type URL struct {
	Keyword    string
	OriginalURL string
	CreatedAt   time.Time
}

func NewURLStorage(db *sql.DB) *URLStorage {
	return &URLStorage{db: db}
}

func (s *URLStorage) Init() error {
	query := `
		CREATE TABLE IF NOT EXISTS urls (
			keyword TEXT PRIMARY KEY,
			original_url TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := s.db.Exec(query)
	return err
}

func (s *URLStorage) Store(keyword, originalURL string) error {
	// Check if keyword already exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM urls WHERE keyword = ?)", keyword).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check keyword existence: %v", err)
	}
	if exists {
		return fmt.Errorf("keyword already exists: %s", keyword)
	}

	// Insert new URL
	_, err = s.db.Exec(
		"INSERT INTO urls (keyword, original_url) VALUES (?, ?)",
		keyword,
		originalURL,
	)
	if err != nil {
		return fmt.Errorf("failed to store URL: %v", err)
	}

	return nil
}

func (s *URLStorage) Get(keyword string) (string, error) {
	var url string
	err := s.db.QueryRow(
		"SELECT original_url FROM urls WHERE keyword = ?",
		keyword,
	).Scan(&url)

	if err == sql.ErrNoRows {
		return "", fmt.Errorf("URL not found for keyword: %s", keyword)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get URL: %v", err)
	}

	return url, nil
}

func (s *URLStorage) Delete(keyword string) error {
	result, err := s.db.Exec(
		"DELETE FROM urls WHERE keyword = ?",
		keyword,
	)
	if err != nil {
		return fmt.Errorf("failed to delete URL: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("URL not found for keyword: %s", keyword)
	}

	return nil
}
