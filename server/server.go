package main

import (
	"context"
	"database/sql"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"time"
)

type Quotation struct {
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
}

type QuotationUSDBRL struct {
	USDBRL Quotation `json:"USDBRL"`
}

type HttpError struct {
	Error string `json:"error"`
}

var db *sql.DB

func createDBConnection() *sql.DB {
	dbConn, err := sql.Open("sqlite3", "data.db")
	if err != nil {
		panic(err)
	}
	_, err = dbConn.Exec(`CREATE TABLE IF NOT EXISTS 
    	quotations (
			id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
			code TEXT,
			codein TEXT,
			name TEXT,
			high TEXT,
			low TEXT,
			varBid TEXT,
			pctChange TEXT,
			bid TEXT,
			ask TEXT,
			timestamp TEXT,
			createDate TEXT
		)
	`)

	if err != nil {
		panic(err)
	}

	return dbConn
}

func saveToDatabase(q Quotation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	query := `
		INSERT INTO quotations 
		(code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, createDate) 
		VALUES 
		($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, q.Code, q.Codein, q.Name, q.High, q.Low, q.VarBid, q.PctChange, q.Bid, q.Ask, q.Timestamp, q.CreateDate)

	return err
}

func sendHttpError(w http.ResponseWriter, err string) {
	w.WriteHeader(http.StatusInternalServerError)
	resp, _ := json.Marshal(HttpError{Error: err})
	w.Write(resp)
	log.Println(err)
}

func quotationHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		sendHttpError(w, err.Error())
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		sendHttpError(w, err.Error())
		return
	}
	defer res.Body.Close()

	var quotation QuotationUSDBRL
	err = json.NewDecoder(res.Body).Decode(&quotation)

	if (Quotation{}) == quotation.USDBRL {
		sendHttpError(w, "Empty response from quotation API")
		return
	}

	err = saveToDatabase(quotation.USDBRL)
	if err != nil {
		sendHttpError(w, err.Error())
		return
	}

	resp, err := json.Marshal(quotation.USDBRL)
	if err != nil {
		sendHttpError(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func main() {
	db = createDBConnection()
	defer db.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/cotacao", quotationHandler)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
