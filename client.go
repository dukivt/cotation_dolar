package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	serverURL = "http://localhost:8080/cotacao"
	timeout   = 300 * time.Millisecond
)

type Cotation struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func main() {

	log.Println("inicializando")
	defer log.Println("requisiçao finalizada")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, serverURL, nil)
	if err != nil {
		log.Printf("erro ao criar requisiçao: %v\n", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("requisiçao nao foi enviada: %v\n", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Printf("servidor nao ta respondendo: %s\n", resp.Status)
	}

	var cotacao Cotation
	err = json.NewDecoder(resp.Body).Decode(&cotacao)
	if err != nil {
		log.Printf("erro no json: %v\n", err)
	}

	if cotacao.USDBRL.Bid == "" {
		log.Printf("bid nao ta retornando\n")
	}

	bid, err := strconv.ParseFloat(cotacao.USDBRL.Bid, 64)
	if err != nil {
		log.Printf("erro ao passar o valor %v\n", err)
	}

	file, err := os.Create("cotacao.txt")
	if err != nil {
		log.Printf("erro ao criar o arquivo: %v\n", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	_, err = fmt.Fprintf(file, "Dólar: %.2f", bid)
	if err != nil {
		log.Printf("erro ao salvar arquivo: %v\n", err)
	}
}
