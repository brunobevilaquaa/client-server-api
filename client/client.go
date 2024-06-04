package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Quotation struct {
	Bid   string `json:"bid"`
	Error string `json:"error"`
}

func saveToFile(q Quotation) {
	f, err := os.OpenFile("cotacao.txt", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	data := fmt.Sprintf("Dólar: %s\n", q.Bid)

	if _, err = f.WriteString(data); err != nil {
		panic(err)
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	var quotation Quotation
	err = json.NewDecoder(res.Body).Decode(&quotation)
	if err != nil {
		panic(err)
	}

	if quotation.Error != "" {
		panic(quotation.Error)
	}

	saveToFile(quotation)
}
