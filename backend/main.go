package main

import (
	"encoding/json"
	"net/http"
)

type Transaction struct {
	ID       int     `json:"ID"json:"id"`
	Type     string  `json:"type"`
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
}

var transactions []Transaction

func main() {
	http.HandleFunc("/income", addIncome)
	http.HandleFunc("/expense", addExpense)
	http.HandleFunc("/summary", getSummary)

	http.ListenAndServe(":8080", nil)
}

func addIncome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	var t Transaction
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	t.Type = "income"
	transactions = append(transactions, t)
	w.WriteHeader(http.StatusCreated)
}

func addExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	var t Transaction
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	t.Type = "expense"
	transactions = append(transactions, t)
	w.WriteHeader(http.StatusCreated)
}

func getSummary(w http.ResponseWriter, r *http.Request) {
	income, expenses := 0.0, 0.0
	for _, t := range transactions {
		if t.Type == "income" {
			income += t.Amount
		} else if t.Type == "expense" {
			expenses += t.Amount
		}
	}
	summary := map[string]float64{
		"total_income":   income,
		"total_expenses": expenses,
		"balance":        income - expenses,
	}
	json.NewEncoder(w).Encode(summary)
}
