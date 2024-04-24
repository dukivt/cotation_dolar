package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	port            = ":8080"
	cotationURL     = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	timeout         = 200 * time.Millisecond
	timeoutDatabase = 10 * time.Millisecond
)

type Cotation struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func main() {

	db, err := sql.Open("sqlite3", "cotation.db")
	if err != nil {
		log.Fatalf("Error opening database: %v\n", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS cotations (bid TEXT);")
	if err != nil {
		log.Fatalf("Error creating cotations table: %v\n", err)
	}

	http.HandleFunc("/cotacao", handleCotationRequest)
	log.Fatal(http.ListenAndServe(port, nil))
}

func handleCotationRequest(w http.ResponseWriter, _ *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cotation, err := getCotation(ctx)
	if err != nil {
		log.Printf("error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = saveCotation(cotation)
	if err != nil {
		log.Printf("error saving cotation: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(cotation)
	if err != nil {
		log.Printf("error encoding cotation: %v", err)
		return
	}
}

func getCotation(ctx context.Context) (*Cotation, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cotationURL, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	var cotacao Cotation
	err = json.NewDecoder(resp.Body).Decode(&cotacao)
	if err != nil {
		return nil, err
	}

	return &cotacao, nil
}

func saveCotation(c *Cotation) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDatabase)
	defer cancel()

	db, err := sql.Open("sqlite3", "cotation.db")
	if err != nil {
		return err
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	stmt, err := db.Prepare("INSERT INTO cotations (bid) VALUES (?)")
	if err != nil {
		return err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
		}
	}(stmt)

	_, err = stmt.ExecContext(ctx, c.USDBRL.Bid)
	if err != nil {
		return err
	}

	log.Println("Cotação salva com sucesso")
	return nil
}
