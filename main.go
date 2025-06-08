package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"urlshortener/pkg/keyword"
	"urlshortener/pkg/metadata"
	"urlshortener/pkg/storage"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Keyword string `json:"keyword"`
	URL     string `json:"url"`
}

type URLService struct {
	storage *storage.URLStorage
	keyword *keyword.Generator
}

func NewURLService() (*URLService, error) {
	// Load .env file
	_ = godotenv.Load()

	// Initialize SQLite database
	db, err := sql.Open("sqlite3", "urls.db")
	if err != nil {
		return nil, err
	}

	// Initialize OpenAI client
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	// Initialize storage
	urlStorage := storage.NewURLStorage(db)
	if err := urlStorage.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %v", err)
	}

	// Initialize keyword generator
	keywordGen := keyword.NewGenerator(openaiKey)

	return &URLService{
		storage: urlStorage,
		keyword: keywordGen,
	}, nil
}

func (s *URLService) handleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate URL
	parsedURL, err := url.Parse(req.URL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Process URL and generate keyword
	keyword, err := s.processURL(r.Context(), req.URL)
	if err != nil {
		log.Printf("Error processing URL: %v", err)
		http.Error(w, "Failed to process URL", http.StatusInternalServerError)
		return
	}

	// Prepare response
	resp := ShortenResponse{
		Keyword: keyword,
		URL:     req.URL,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *URLService) handleRedirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	keyword := r.URL.Path[1:] // Remove leading slash

	// Look up URL
	originalURL, err := s.storage.Get(keyword)
	if err != nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusMovedPermanently)
}

func (s *URLService) processURL(ctx context.Context, inputURL string) (string, error) {
	// Extract metadata
	meta, err := metadata.ExtractFromURL(inputURL)
	if err != nil {
		return "", fmt.Errorf("failed to extract metadata: %v", err)
	}

	// Generate keyword using OpenAI
	keyword, err := s.keyword.GenerateKeyword(ctx, meta.Title, meta.Description, meta.Keywords)
	if err != nil {
		return "", fmt.Errorf("failed to generate keyword: %v", err)
	}

	// Store in database
	if err := s.storage.Store(keyword, inputURL); err != nil {
		return "", fmt.Errorf("failed to store URL: %v", err)
	}

	return keyword, nil
}

func main() {
	service, err := NewURLService()
	if err != nil {
		log.Fatalf("Failed to initialize service: %v", err)
	}

	http.HandleFunc("/shorten", service.handleShorten)
	http.HandleFunc("/", service.handleRedirect)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
