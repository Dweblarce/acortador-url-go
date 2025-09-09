package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	ShortURL string `json:"short_url"`
}

type URLShortener struct {
	urls map[string]string
	mu   sync.RWMutex
}

func NewURLShortener() *URLShortener {
	return &URLShortener{
		urls: make(map[string]string),
	}
}

func generateShortKey() string {
	// Si necesitas un generador de claves más seguro para producción,
	// considera usar "crypto/rand".
	rand.Seed(time.Now().UnixNano())
	chars := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, 6)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func (us *URLShortener) shortenHandler(w http.ResponseWriter, r *http.Request) {
	// Asegura que solo se acepten peticiones POST
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shortKey := generateShortKey()

	us.mu.Lock()
	us.urls[shortKey] = req.URL
	us.mu.Unlock()

	shortURL := fmt.Sprintf("http://localhost:8080/%s", shortKey)

	resp := Response{ShortURL: shortURL}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (us *URLShortener) rootHandler(w http.ResponseWriter, r *http.Request) {
	// Lógica principal del servidor
	path := r.URL.Path

	// Si la ruta es exactamente la raíz, sirve el index.html
	if path == "/" {
		http.ServeFile(w, r, "static/index.html")
		return
	}

	// Si la ruta no es "/", intenta buscar una URL acortada para redirigir
	key := path[1:]
	us.mu.RLock()
	originalURL, ok := us.urls[key]
	us.mu.RUnlock()

	if ok {
		// Si la clave existe, redirige
		http.Redirect(w, r, originalURL, http.StatusFound)
		return
	}

	// Si la ruta no es la raíz ni una clave válida, muestra "404 Not Found"
	http.NotFound(w, r)
}

func main() {
	shortener := NewURLShortener()

	mux := http.NewServeMux()

	// Asigna el manejador principal para la ruta raíz
	mux.HandleFunc("/", shortener.rootHandler)
	
	// Asigna el manejador para la API de acortamiento de URLs
	mux.HandleFunc("/shorten", shortener.shortenHandler)
	
	fmt.Println("Servidor corriendo en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}