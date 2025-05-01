package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver
)

type Entry struct {
	PlayerName string `json:"playerName"`
	Turns      int    `json:"turns"`
}

// type Config struct {
// 	Host     string `json:"host"`
// 	Port     string `json:"port"`
// 	Username string `json:"username"`
// 	Password string `json:"password"`
// 	DBName   string `json:"dbname"`
// 	SSLMode  string `json:"sslmode"`
// }

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, Welcome to the server!")
}

func leaderboardHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	err := godotenv.Load()
	consstr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("HOST"), os.Getenv("PORT"), os.Getenv("USER"), os.Getenv("PASSWORD"), os.Getenv("DBNAME"), os.Getenv("SSLMODE"))
    if err != nil {
        log.Fatal("Error loading .env file")
    }
	db, err := sql.Open("postgres", consstr)
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		log.Fatal(err)
		return
	}
	defer db.Close()
	if r.Method == http.MethodPost {
		var entry Entry
		if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			log.Println(err)
			return
		}

		_, err := db.Exec("INSERT INTO results (playerName, turns) VALUES ($1, $2)", entry.PlayerName, entry.Turns)
		if err != nil {
			http.Error(w, "Error inserting data into database", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintln(w, "Entry added successfully")
		return
	}
	if r.Method == http.MethodGet {
		rows, err := db.Query("SELECT playerName, turns FROM results ORDER BY turns ASC LIMIT 5")
		if err != nil {
			http.Error(w, "Error fetching leaderboard data", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		defer rows.Close()
		var leaderboard []Entry
		for rows.Next() {
			var entry Entry
			if err := rows.Scan(&entry.PlayerName, &entry.Turns); err != nil {
				http.Error(w, "Error scanning leaderboard data", http.StatusInternalServerError)
				log.Println(err)
				return
			}
			leaderboard = append(leaderboard, entry)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, "Error iterating over leaderboard data", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(leaderboard); err != nil {
			http.Error(w, "Error encoding leaderboard data", http.StatusInternalServerError)
			log.Println(err)
			return
		}
	}
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/leaderboard", leaderboardHandler)

	log.Println("Starting server on :5000...")
	if err := http.ListenAndServe(":5000", nil); err != nil {
		log.Fatal(err)
	}
	log.Println("Server stopped.")
}
