// @title        React-Go-App API
// @version      0.1.0
// @description  Auto‑generated Swagger docs
// @host         localhost:8080
// @BasePath     /
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/frallan97/react-go-app-backend/docs"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Message represents one row in messages.
type Message struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// MessageInput represents the input for creating a message.
type MessageInput struct {
	Content string `json:"content"`
}

func main() {
	// Read configuration from environment
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPass := os.Getenv("DB_PASSWORD")
	if dbPass == "" {
		dbPass = "postgres"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "postgres"
	}

	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			dbHost, dbPort, dbUser, dbPass, dbName,
		)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open DB: %v", err)
	}

	// Set up connection pool parameters
	db.SetMaxOpenConns(10)                  // max 10 open connections
	db.SetMaxIdleConns(5)                   // max 5 idle connections
	db.SetConnMaxIdleTime(5 * time.Minute)  // idle timeout
	db.SetConnMaxLifetime(30 * time.Minute) // max lifetime

	// Retry logic: try to ping the DB for up to 30 seconds
	maxWait := 30 * time.Second
	waitInterval := 2 * time.Second
	start := time.Now()
	for {
		err := db.Ping()
		if err == nil {
			break
		}
		if time.Since(start) > maxWait {
			log.Fatalf("unable to ping DB after %v: %v", maxWait, err)
		}
		log.Printf("waiting for DB to be ready: %v", err)
		time.Sleep(waitInterval)
	}
	log.Println("connected to Postgres successfully (using connection pool)")

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler(db))
	mux.HandleFunc("/api/messages", messagesHandler(db))
	// Serve swagger UI at /docs/index.html
	mux.Handle("/docs/", httpSwagger.WrapHandler)

	log.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// healthHandler responds with {"status":"ok"}
// @Summary     Health check
// @Description Returns 200 if DB is reachable
// @Tags        health
// @Produce     json
// @Success     200  {object}  map[string]string
// @Router      /health [get]
func healthHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			http.Error(w, `{"status":"error"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	}
}

// messagesHandler lists or creates messages
// @Summary     List messages
// @Description Get all messages
// @Tags        messages
// @Produce     json
// @Success     200  {array}   Message
// @Router      /api/messages [get]
//
// @Summary     Create message
// @Description Insert a new message
// @Tags        messages
// @Accept      json
// @Produce     json
// @Param       msg  body   MessageInput  true  "message payload"
// @Success     201   {object}  map[string]int
// @Router      /api/messages [post]
func messagesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			rows, err := db.Query(`SELECT id, content, created_at FROM messages ORDER BY id`)
			if err != nil {
				http.Error(w, "db query failed", 500)
				return
			}
			defer rows.Close()

			var msgs []Message
			for rows.Next() {
				var m Message
				if err := rows.Scan(&m.ID, &m.Content, &m.CreatedAt); err != nil {
					http.Error(w, "scan failed", 500)
					return
				}
				msgs = append(msgs, m)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(msgs)

		case http.MethodPost:
			var in MessageInput
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				http.Error(w, "invalid payload", 400)
				return
			}
			var id int
			err := db.QueryRow(
				`INSERT INTO messages(content) VALUES($1) RETURNING id`, in.Content,
			).Scan(&id)
			if err != nil {
				http.Error(w, "insert failed", 500)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]int{"id": id})

		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "method not allowed", 405)
		}
	}
}
