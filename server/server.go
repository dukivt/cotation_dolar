package main

import (
	"context"
	"database/sql"
	"encoding/json"
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
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

func main() {
	http.HandleFunc("/cotacao", handleCotationRequest)
	log.Fatal(http.ListenAndServe(port, nil))
}

func handleCotationRequest(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cotation, err := GetCotation(ctx)
	if err != nil {
		log.Printf("erro: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	saveCotation(cotation)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cotation)
}

func GetCotation(ctx context.Context) (*Cotation, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cotationURL, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var cotacao Cotation
	err = json.NewDecoder(resp.Body).Decode(&cotacao)
	if err != nil {
		return nil, err
	}

	return &cotacao, nil
}

func saveCotation(c *Cotation) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDatabase)
	defer cancel()

	db, err := sql.Open("sqlite3", "cotation.db")
	if err != nil {
		log.Printf("erro: %v", err)
		return
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, "INSERT INTO cotations (bid) VALUES (?)", c.USDBRL.Bid)
	if err != nil {
		log.Printf("falha ao salvar: %v", err)
		return
	}

	log.Println("cotacao salva com sucesso")
}
