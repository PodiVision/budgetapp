package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type Transaktion struct {
	ID        int     `json:"id"`
	Typ       string  `json:"typ"`
	Betrag    float64 `json:"betrag"`
	Kategorie string  `json:"kategorie"`
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./budget.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTabelle()

	http.HandleFunc("/einnahmen", corsMiddleware(addEinnahmen))
	http.HandleFunc("/ausgaben", corsMiddleware(addAusgaben))
	http.HandleFunc("/zusammenfassung", corsMiddleware(getZusammenfassung))

	// Routen für das Abrufen und Löschen der Transaktionen
	http.HandleFunc("/transaktionen", corsMiddleware(getTransaktionen))
	http.HandleFunc("/transaktionen/delete", corsMiddleware(deleteTransaktion))

	log.Println("Server gestartet auf Port :8081")
	err = http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func createTabelle() {
	createTabelleSQL := `CREATE TABLE IF NOT EXISTS transaktionen (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        typ TEXT,
        betrag REAL,
        kategorie TEXT
    );`

	anweisung, err := db.Prepare(createTabelleSQL)
	if err != nil {
		log.Fatal(err)
	}
	defer anweisung.Close()

	if _, err := anweisung.Exec(); err != nil {
		log.Fatal(err)
	}
}

func aktiviereCors(writer http.ResponseWriter) {
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE")
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		aktiviereCors(writer)
		if request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusOK)
			return
		}
		next(writer, request)
	}
}

func addEinnahmen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Ungültige Anfragemethode", http.StatusMethodNotAllowed)
		return
	}
	var transaktion Transaktion
	err := json.NewDecoder(r.Body).Decode(&transaktion)
	if err != nil {
		http.Error(w, "Ungültige Eingabe: "+err.Error(), http.StatusBadRequest)
		return
	}
	transaktion.Typ = "Einnahme"
	if err := eingabeTransaktion(transaktion); err != nil {
		http.Error(w, "Transaktion konnte nicht hinzugefügt werden: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func addAusgaben(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "Ungültige Anfragemethode", http.StatusMethodNotAllowed)
		return
	}
	var transaktion Transaktion
	err := json.NewDecoder(request.Body).Decode(&transaktion)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	transaktion.Typ = "Ausgabe"
	if err := eingabeTransaktion(transaktion); err != nil {
		http.Error(writer, "Transaktion konnte nicht hinzugefügt werden: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusCreated)
}

func eingabeTransaktion(transaktion Transaktion) error {
	insertSQL := `INSERT INTO transaktionen (typ, betrag, kategorie) VALUES (?, ?, ?)`
	statement, err := db.Prepare(insertSQL)
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(transaktion.Typ, transaktion.Betrag, transaktion.Kategorie)
	return err
}

// Neue Funktion, um alle Transaktionen abzurufen
func getTransaktionen(writer http.ResponseWriter, request *http.Request) {
	reihen, err := db.Query("SELECT id, typ, betrag, kategorie FROM transaktionen")
	if err != nil {
		http.Error(writer, "Fehler beim Abrufen der Transaktionen: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer reihen.Close()

	var transaktionen []Transaktion
	for reihen.Next() {
		var transaktion Transaktion
		err := reihen.Scan(&transaktion.ID, &transaktion.Typ, &transaktion.Betrag, &transaktion.Kategorie)
		if err != nil {
			http.Error(writer, "Fehler beim Verarbeiten der Daten: "+err.Error(), http.StatusInternalServerError)
			return
		}
		transaktionen = append(transaktionen, transaktion)
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(transaktionen)
}

// Neue Funktion, um eine Transaktion zu löschen
func deleteTransaktion(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodDelete {
		http.Error(writer, "Ungültige Anfragemethode", http.StatusMethodNotAllowed)
		return
	}

	id := request.URL.Query().Get("id")
	if id == "" {
		http.Error(writer, "ID ist erforderlich", http.StatusBadRequest)
		return
	}

	deleteSQL := `DELETE FROM transaktionen WHERE id = ?`
	statement, err := db.Prepare(deleteSQL)
	if err != nil {
		http.Error(writer, "Fehler beim Löschen der Transaktion: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer statement.Close()

	_, err = statement.Exec(id)
	if err != nil {
		http.Error(writer, "Transaktion konnte nicht gelöscht werden: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
}

func getZusammenfassung(writer http.ResponseWriter, request *http.Request) {
	reihen, err := db.Query("SELECT typ, betrag FROM transaktionen")
	if err != nil {
		http.Error(writer, "Fehler beim Abrufen der Zusammenfassung: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer reihen.Close()

	var einkommen, ausgaben float64
	for reihen.Next() {
		var transaktion Transaktion
		err := reihen.Scan(&transaktion.Typ, &transaktion.Betrag)
		if err != nil {
			http.Error(writer, "Fehler beim Verarbeiten der Daten: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if transaktion.Typ == "Einnahme" {
			einkommen += transaktion.Betrag
		} else if transaktion.Typ == "Ausgabe" {
			ausgaben += transaktion.Betrag
		}
	}

	zusammenfassung := map[string]float64{
		"Gesamteinnahmen": einkommen,
		"Gesamtausgaben":  ausgaben,
		"Kontostand":      einkommen - ausgaben,
	}
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(zusammenfassung)
}
