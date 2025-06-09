# urlshortener

A minimal URL shortening service written in Go, powered by OpenAI for automatic keyword generation and backed by SQLite for storage.

---

## Features

- Accepts long URLs and returns unique keywords
- Generates semantic keywords using OpenAI
- Redirects short keywords to original URLs
- Stores data in a local SQLite database
- Metadata extraction from the target URL
- Simple HTTP interface (no frontend yet)

---

## Requirements

- Go 1.24+
- SQLite (via CGO)
- GCC installed (`sudo apt install build-essential libsqlite3-dev`)
- OpenAI API key

---

## Getting Started

### 1. Clone and prepare

```bash
git clone https://github.com/mattia-m/semantic-urlshortener.git
cd semantic-urlshortener
```

### 2. Create .env file
OPENAI_API_KEY=your-openai-key

### 3. Run the service
```bash
CGO_ENABLED=1 go run .
```
Server will start at: http://localhost:8080

## API

### POST `/shorten`

**Request body:**

```json
{
  "url": "https://example.com"
}
```

**Response:**

```json
{
  "keyword": "example",
  "url": "https://example.com"
}

```

### GET `/:keyword`

Redirects to the original URL.

Example:
```bash
GET http://localhost:8080/example
→ 301 Redirect to https://example.com
```

## Development Notes

- Uses [`github.com/sashabaranov/go-openai`](https://github.com/sashabaranov/go-openai) for ChatGPT integration  
- Loads config from `.env` using [`github.com/joho/godotenv`](https://github.com/joho/godotenv)  
- Requires CGO enabled due to [`github.com/mattn/go-sqlite3`](https://github.com/mattn/go-sqlite3)  

### Project structure
```

├── main.go
├── go.mod
├── pkg/
│ ├── keyword/ # OpenAI keyword generator
│ ├── metadata/ # Metadata extractor
│ └── storage/ # SQLite persistence

```
---

## License

MIT