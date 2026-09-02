package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type Registration struct {
	ID        string    `json:"id"`
	Game      string    `json:"game"`
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
}

const dataDir = "data"
const signupsFile = "data/signups.json"

func init() {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Errore nella creazione della cartella data: %v", err)
	}
}

// Carica tutte le iscrizioni dal file JSON
func loadRegistrations() ([]Registration, error) {
	data, err := ioutil.ReadFile(signupsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []Registration{}, nil
		}
		return nil, err
	}

	var registrations []Registration
	if err := json.Unmarshal(data, &registrations); err != nil {
		return nil, err
	}

	return registrations, nil
}

// Salva le iscrizioni nel file JSON
func saveRegistrations(registrations []Registration) error {
	data, err := json.MarshalIndent(registrations, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(signupsFile, data, 0644)
}

// Handler GET - Carica tutte le iscrizioni
func handleGetSignups(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	registrations, err := loadRegistrations()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(registrations)
}

// Handler POST - Aggiungi una nuova iscrizione
func handlePostSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metodo non consentito", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var reg Registration
	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "JSON non valido"})
		return
	}

	if reg.Name == "" || reg.Game == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Nome e gioco sono obbligatori"})
		return
	}

	reg.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	reg.Timestamp = time.Now()

	registrations, err := loadRegistrations()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	registrations = append(registrations, reg)

	if err := saveRegistrations(registrations); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reg)
}

// Handler DELETE - Elimina una iscrizione
func handleDeleteSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Metodo non consentito", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	id := r.URL.Query().Get("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID mancante"})
		return
	}

	registrations, err := loadRegistrations()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Trova e rimuovi l'iscrizione
	var found bool
	for i, reg := range registrations {
		if reg.ID == id {
			registrations = append(registrations[:i], registrations[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Iscrizione non trovata"})
		return
	}

	if err := saveRegistrations(registrations); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Iscrizione eliminata"})
}

// Handler per CORS preflight
func handleCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(http.StatusOK)
}

func main() {
	// Registra gli handler
	http.HandleFunc("/api/signups", handleGetSignups)
	http.HandleFunc("/api/signup", handlePostSignup)
	http.HandleFunc("/api/signup/", handleDeleteSignup)
	http.HandleFunc("/api/", handleCORS)

	fmt.Println("Server in ascolto su http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Errore nel server: %v", err)
	}
}
